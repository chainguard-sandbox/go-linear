// Package graphql provides a raw GraphQL passthrough command — a supported
// escape hatch for running arbitrary GraphQL documents through go-linear's
// authenticated, resilient client.
package graphql

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/parser"

	"github.com/chainguard-sandbox/go-linear/v2/internal/cli"
	"github.com/chainguard-sandbox/go-linear/v2/pkg/linear"
)

const longDescription = `Run an arbitrary GraphQL document through the Linear client.

This is a supported escape hatch: the document is sent through the same
authenticated, resilient client as every other command (auth, retries, rate
limiting, metrics) and the raw "data" object is printed as JSON.

Non-guarantees (this is a RAW passthrough):
  - No name→ID resolution (pass real IDs, not names).
  - No field defaults and no --fields pruning; you get exactly what you select.
  - No response-shape stability; output tracks Linear's API verbatim.
  - No query-complexity reduction.

Mutations are rejected unless --allow-mutation is given. The check runs before
any network call by parsing the document and inspecting the selected operation.

Authentication comes from the existing credential provider (LINEAR_API_KEY or
--api-key); never pass credentials to this command.

Variables (--var) are typed to avoid string-concatenation into the query:
  name=value           string (default)
  name:int=42          integer
  name:float=1.5       float
  name:bool=true       boolean
  name:json=[1,2,3]    raw JSON (object, array, etc.)

Examples:
  # Nested field absent from any baked selection set:
  go-linear graphql --query 'query { viewer { organization { id name } } }'

  # Query with typed variables:
  go-linear graphql \
    --query 'query($id:String!){ issue(id:$id){ title project { lead { email } } } }' \
    --var id=abc-123

  # Read the document from a file or stdin:
  go-linear graphql --query @query.graphql
  echo 'query { viewer { id } }' | go-linear graphql --query -

  # Variables from a JSON file, overridden by an inline --var:
  go-linear graphql --query @q.graphql --vars-file @vars.json --var first:int=50

  # A mutation (must opt in):
  go-linear graphql --allow-mutation \
    --query 'mutation { issueUpdate(id:"abc", input:{priority:1}){ success } }'`

// NewGraphQLCommand creates the graphql command.
func NewGraphQLCommand(clientFactory cli.ClientFactory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "graphql",
		Short: "Run a raw GraphQL query or mutation (escape hatch)",
		Long:  longDescription,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return run(cmd, clientFactory)
		},
	}

	cmd.Flags().String("query", "", "GraphQL document; inline, @file to read a file, or - for stdin")
	cmd.Flags().StringArray("var", nil, "variable as name[:type]=value (type: string|int|float|bool|json); repeatable")
	cmd.Flags().String("vars-file", "", "read variables from a JSON file (@file.json)")
	cmd.Flags().String("operation-name", "", "operation to run when the document defines several")
	cmd.Flags().Bool("allow-mutation", false, "permit the document to run a mutation")
	cmd.Flags().Bool("compact", false, "print compact (single-line) JSON instead of indented")
	cmd.Flags().Int64("max-response-bytes", 0, "cap the response size in bytes (0 = default 10MB)")

	return cmd
}

func run(cmd *cobra.Command, clientFactory cli.ClientFactory) error {
	queryFlag, _ := cmd.Flags().GetString("query")
	varFlags, _ := cmd.Flags().GetStringArray("var")
	varsFile, _ := cmd.Flags().GetString("vars-file")
	operationName, _ := cmd.Flags().GetString("operation-name")
	allowMutation, _ := cmd.Flags().GetBool("allow-mutation")
	compact, _ := cmd.Flags().GetBool("compact")
	maxResponseBytes, _ := cmd.Flags().GetInt64("max-response-bytes")

	query, err := loadQuery(queryFlag, cmd.InOrStdin())
	if err != nil {
		return err
	}
	if strings.TrimSpace(query) == "" {
		return errors.New("--query is required")
	}

	vars, err := buildVars(varsFile, varFlags)
	if err != nil {
		return err
	}

	// Gate mutations before any network call.
	if err := checkOperation(query, operationName, allowMutation); err != nil {
		return err
	}

	client, err := clientFactory()
	if err != nil {
		return err
	}
	defer client.Close()

	var opts []linear.ExecuteOption
	if operationName != "" {
		opts = append(opts, linear.WithOperationName(operationName))
	}
	if maxResponseBytes > 0 {
		opts = append(opts, linear.WithMaxResponseBytes(maxResponseBytes))
	}

	var data json.RawMessage
	// The returned error is printed once by cobra (the root command silences
	// usage, not errors); *GraphQLResponseError renders its messages and paths.
	if err := client.Execute(cmd.Context(), query, vars, &data, opts...); err != nil {
		return err
	}

	return printJSON(cmd.OutOrStdout(), data, compact)
}

// loadQuery resolves the --query flag: "@path" reads a file, "-" reads stdin,
// anything else is treated as the literal document.
func loadQuery(flag string, stdin io.Reader) (string, error) {
	switch {
	case flag == "-":
		b, err := io.ReadAll(stdin)
		if err != nil {
			return "", fmt.Errorf("read query from stdin: %w", err)
		}
		return string(b), nil
	case strings.HasPrefix(flag, "@"):
		path := strings.TrimPrefix(flag, "@")
		if path == "-" {
			b, err := io.ReadAll(stdin)
			if err != nil {
				return "", fmt.Errorf("read query from stdin: %w", err)
			}
			return string(b), nil
		}
		b, err := os.ReadFile(path) // #nosec G304 - user-provided query file is expected
		if err != nil {
			return "", fmt.Errorf("read query file: %w", err)
		}
		return string(b), nil
	default:
		return flag, nil
	}
}

// buildVars merges variables from a JSON file (base) with typed --var flags,
// where --var overrides file entries on conflict.
func buildVars(varsFile string, varFlags []string) (map[string]any, error) {
	vars := map[string]any{}

	if varsFile != "" {
		path := strings.TrimPrefix(varsFile, "@")
		b, err := os.ReadFile(path) // #nosec G304 - user-provided vars file is expected
		if err != nil {
			return nil, fmt.Errorf("read vars file: %w", err)
		}
		if err := json.Unmarshal(b, &vars); err != nil {
			return nil, fmt.Errorf("parse vars file %q as a JSON object: %w", path, err)
		}
		// JSON "null" unmarshals into a nil map without error; a later vars[k]=v
		// would then panic. Reject any non-object (null, array, scalar).
		if vars == nil {
			return nil, fmt.Errorf("vars file %q must be a JSON object, got null", path)
		}
	}

	for _, raw := range varFlags {
		name, value, err := parseVar(raw)
		if err != nil {
			return nil, err
		}
		vars[name] = value
	}

	if len(vars) == 0 {
		return nil, nil
	}
	return vars, nil
}

// parseVar parses a single --var argument of the form name[:type]=value.
func parseVar(raw string) (string, any, error) {
	key, value, ok := strings.Cut(raw, "=")
	if !ok {
		return "", nil, fmt.Errorf("invalid --var %q: expected name[:type]=value", raw)
	}

	name := key
	typ := "string"
	if n, t, hasType := strings.Cut(key, ":"); hasType {
		name = n
		typ = t
	}
	if name == "" {
		return "", nil, fmt.Errorf("invalid --var %q: empty variable name", raw)
	}

	switch typ {
	case "string", "":
		return name, value, nil
	case "int":
		n, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return "", nil, fmt.Errorf("invalid --var %q: %q is not an int", raw, value)
		}
		return name, n, nil
	case "float":
		f, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return "", nil, fmt.Errorf("invalid --var %q: %q is not a float", raw, value)
		}
		return name, f, nil
	case "bool":
		b, err := strconv.ParseBool(value)
		if err != nil {
			return "", nil, fmt.Errorf("invalid --var %q: %q is not a bool", raw, value)
		}
		return name, b, nil
	case "json":
		var v any
		if err := json.Unmarshal([]byte(value), &v); err != nil {
			return "", nil, fmt.Errorf("invalid --var %q: value is not valid JSON: %w", raw, err)
		}
		return name, v, nil
	default:
		return "", nil, fmt.Errorf("invalid --var %q: unknown type %q (want string|int|float|bool|json)", raw, typ)
	}
}

// checkOperation parses the document and rejects mutations (unless allowed) and
// subscriptions, selecting the operation by name when the document has several.
func checkOperation(query, operationName string, allowMutation bool) error {
	doc, err := parser.ParseQuery(&ast.Source{Name: "graphql", Input: query})
	if err != nil {
		return fmt.Errorf("parse query: %w", err)
	}

	if len(doc.Operations) == 0 {
		return errors.New("no operation found in document")
	}

	var op *ast.OperationDefinition
	switch {
	case operationName != "":
		op = doc.Operations.ForName(operationName)
		if op == nil {
			return fmt.Errorf("operation %q not found in document", operationName)
		}
	case len(doc.Operations) == 1:
		op = doc.Operations[0]
	default:
		return errors.New("document defines multiple operations; select one with --operation-name")
	}

	switch op.Operation {
	case ast.Mutation:
		if !allowMutation {
			return errors.New("refusing to run a mutation without --allow-mutation")
		}
	case ast.Subscription:
		return errors.New("subscriptions are not supported")
	case ast.Query:
	default:
		// Backstop: a future gqlparser operation type is not implicitly allowed.
		return fmt.Errorf("unsupported operation type %q", op.Operation)
	}

	return nil
}

func printJSON(w io.Writer, data json.RawMessage, compact bool) error {
	if len(data) == 0 {
		_, err := fmt.Fprintln(w, "null")
		return err
	}
	var buf bytes.Buffer
	if compact {
		if err := json.Compact(&buf, data); err != nil {
			return fmt.Errorf("format JSON: %w", err)
		}
	} else if err := json.Indent(&buf, data, "", "  "); err != nil {
		return fmt.Errorf("format JSON: %w", err)
	}
	buf.WriteByte('\n')
	_, err := w.Write(buf.Bytes())
	return err
}
