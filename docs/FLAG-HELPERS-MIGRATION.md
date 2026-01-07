# Flag Helpers Migration Guide

This guide explains how to migrate commands to use the flag helper structs in `internal/cli/flags.go`.

## Why Migrate?

**Benefits:**
- ✅ **Reduce duplication**: ~4 lines per command → 1 line
- ✅ **Type safety**: Direct struct access vs string parsing
- ✅ **Consistent validation**: Built-in `Validate()` methods
- ✅ **Better testability**: Pass structs to `run*()` functions
- ✅ **Cleaner code**: No more `cmd.Flags().GetString()` calls

## Available Helpers

### 1. OutputFlags
Handles `--output` and `--fields` flags.

**Usage:**
```go
flags := &cli.OutputFlags{}
flags.Bind(cmd, "defaults (...) | none | defaults,extra | ...")

// In RunE:
if err := flags.Validate(); err != nil {
    return err
}
switch flags.Output {
case "json":
    // use flags.Fields
case "table":
    // ...
}
```

### 2. PaginationFlags
Handles `--limit` and `--after` flags.

**Usage:**
```go
flags := &cli.PaginationFlags{}
flags.Bind(cmd, 50) // default limit

// In RunE:
first := flags.LimitPtr()  // *int64
after := flags.AfterPtr()  // *string (nil if empty)
```

### 3. ConfirmationFlags
Handles `--yes` flag for destructive operations.

**Usage:**
```go
flags := &cli.ConfirmationFlags{}
flags.Bind(cmd)

// In RunE:
if !flags.Yes {
    // prompt for confirmation
}
```

## Migration Patterns

### Pattern 1: Simple Get Command (OutputFlags only)

**Before:**
```go
func NewGetCommand(clientFactory cli.ClientFactory) *cobra.Command {
    cmd := &cobra.Command{
        RunE: func(cmd *cobra.Command, args []string) error {
            // ...
            return runGet(cmd, client, args[0])
        },
    }

    cmd.Flags().StringP("output", "o", "table", "Output format: json|table")
    cmd.Flags().String("fields", "", "defaults (...)")
    return cmd
}

func runGet(cmd *cobra.Command, client *linear.Client, id string) error {
    output, _ := cmd.Flags().GetString("output")
    fieldsSpec, _ := cmd.Flags().GetString("fields")

    switch output {
    case "json":
        // use fieldsSpec
    case "table":
        // ...
    default:
        return fmt.Errorf("unsupported output format: %s", output)
    }
}
```

**After:**
```go
func NewGetCommand(clientFactory cli.ClientFactory) *cobra.Command {
    flags := &cli.OutputFlags{}

    cmd := &cobra.Command{
        RunE: func(cmd *cobra.Command, args []string) error {
            // ...
            return runGet(cmd, client, args[0], flags)
        },
    }

    flags.Bind(cmd, "defaults (...)")
    return cmd
}

func runGet(cmd *cobra.Command, client *linear.Client, id string, flags *cli.OutputFlags) error {
    if err := flags.Validate(); err != nil {
        return err
    }

    switch flags.Output {
    case "json":
        // use flags.Fields
    case "table":
        // ...
    default:
        return fmt.Errorf("unsupported output format: %s", flags.Output)
    }
}
```

**Example:** `cmd/linear/commands/issue/get.go`

### Pattern 2: List Command (OutputFlags + PaginationFlags)

**Before:**
```go
func NewListCommand(clientFactory cli.ClientFactory) *cobra.Command {
    cmd := &cobra.Command{
        RunE: func(cmd *cobra.Command, args []string) error {
            // ...
            return runList(cmd, client)
        },
    }

    cmd.Flags().IntP("limit", "l", 50, "Number of items to return")
    cmd.Flags().StringP("output", "o", "table", "Output format: json|table")
    cmd.Flags().String("fields", "", "defaults (...)")
    return cmd
}

func runList(cmd *cobra.Command, client *linear.Client) error {
    limit, _ := cmd.Flags().GetInt("limit")
    first := int64(limit)

    output, _ := cmd.Flags().GetString("output")
    fieldsSpec, _ := cmd.Flags().GetString("fields")

    items, err := client.Items(ctx, &first, nil)
    // ...
}
```

**After:**
```go
func NewListCommand(clientFactory cli.ClientFactory) *cobra.Command {
    outputFlags := &cli.OutputFlags{}
    paginationFlags := &cli.PaginationFlags{}

    cmd := &cobra.Command{
        RunE: func(cmd *cobra.Command, args []string) error {
            // ...
            return runList(cmd, client, outputFlags, paginationFlags)
        },
    }

    paginationFlags.Bind(cmd, 50)
    outputFlags.Bind(cmd, "defaults (...)")
    return cmd
}

func runList(cmd *cobra.Command, client *linear.Client, outputFlags *cli.OutputFlags, paginationFlags *cli.PaginationFlags) error {
    if err := outputFlags.Validate(); err != nil {
        return err
    }

    first := paginationFlags.LimitPtr()
    after := paginationFlags.AfterPtr()

    items, err := client.Items(ctx, first, after)
    // ...
}
```

**Example:** `cmd/linear/commands/team/list.go`

### Pattern 3: Delete Command (ConfirmationFlags)

**Before:**
```go
func NewDeleteCommand(clientFactory cli.ClientFactory) *cobra.Command {
    cmd := &cobra.Command{
        RunE: func(cmd *cobra.Command, args []string) error {
            // ...
            yes, _ := cmd.Flags().GetBool("yes")
            if !yes {
                // confirmation prompt
            }

            output, _ := cmd.Flags().GetString("output")
            if output == "json" {
                // ...
            }
        },
    }

    cmd.Flags().BoolP("yes", "y", false, "Skip confirmation")
    cmd.Flags().StringP("output", "o", "table", "Output format: json|table")
    return cmd
}
```

**After:**
```go
func NewDeleteCommand(clientFactory cli.ClientFactory) *cobra.Command {
    confirmFlags := &cli.ConfirmationFlags{}
    var output string

    cmd := &cobra.Command{
        RunE: func(cmd *cobra.Command, args []string) error {
            // ...
            if !confirmFlags.Yes {
                // confirmation prompt
            }

            if output == "json" {
                // ...
            }
        },
    }

    confirmFlags.Bind(cmd)
    cmd.Flags().StringVarP(&output, "output", "o", "table", "Output format: json|table")
    return cmd
}
```

**Example:** `cmd/linear/commands/label/delete.go`

**Note:** For simple output handling (no fields), use a local `var output string` instead of OutputFlags.

## Migration Checklist

For each command to migrate:

1. **Identify which helpers apply:**
   - Has `--output` + `--fields`? → OutputFlags
   - Has `--limit` + `--after`? → PaginationFlags
   - Has `--yes`? → ConfirmationFlags

2. **Update `NewXCommand`:**
   - Create flag struct(s) before `cobra.Command`
   - Call `.Bind()` after command definition
   - Remove old `cmd.Flags().*` calls
   - Pass flag structs to `runX()` function

3. **Update `runX` function:**
   - Add flag struct parameters
   - Add validation: `if err := flags.Validate(); err != nil`
   - Replace `cmd.Flags().GetX()` with `flags.Field`
   - Use helper methods: `.LimitPtr()`, `.AfterPtr()`

4. **Test:**
   - `go build ./cmd/linear/commands/...`
   - Run the command manually
   - Check tests still pass

## Commands to Migrate

### High Priority (use OutputFlags + PaginationFlags)

List commands:
- [ ] `issue/list.go`
- [ ] `user/list.go`
- [ ] `cycle/list.go`
- [ ] `project/list.go`
- [ ] `comment/list.go`
- [ ] `attachment/list.go`
- [ ] `initiative/list.go`
- [ ] `document/list.go`
- [ ] `state/list.go`
- [ ] `roadmap/list.go`
- [ ] `template/list.go`

### Medium Priority (use OutputFlags)

Get commands:
- [ ] `team/get.go`
- [ ] `user/get.go`
- [ ] `cycle/get.go`
- [ ] `project/get.go`
- [ ] `comment/get.go`
- [ ] `attachment/get.go`
- [ ] `initiative/get.go`
- [ ] `document/get.go`
- [ ] `state/get.go`
- [ ] `roadmap/get.go`
- [ ] `template/get.go`
- [ ] `label/get.go`

Create/Update commands:
- [ ] `issue/create.go`
- [ ] `issue/update.go`
- [ ] `issue/search.go`
- [ ] `team/create.go`
- [ ] `cycle/create.go`
- [ ] `project/create.go`
- [ ] `comment/create.go`
- [ ] `label/create.go`
- [ ] All other create/update commands

### Low Priority (use ConfirmationFlags)

Delete commands:
- [ ] `issue/delete.go`
- [ ] `team/delete.go`
- [ ] `cycle/archive.go`
- [ ] `project/delete.go`
- [ ] `comment/delete.go`
- [ ] `attachment/delete.go`
- [ ] All other delete commands

## Testing

After migration, ensure:

1. **Build succeeds:**
   ```bash
   go build ./cmd/linear/...
   ```

2. **Tests pass:**
   ```bash
   go test ./cmd/linear/commands/...
   ```

3. **Manual testing:**
   ```bash
   ./bin/go-linear <entity> <command> --output=json
   ./bin/go-linear <entity> <command> --output=table
   ```

4. **MCP tools generation:**
   ```bash
   ./bin/go-linear mcp tools > mcp-tools.json
   jq 'length' mcp-tools.json  # Should still be 74
   ```

## Notes

- **Don't break existing tests**: Update test mocks to use the new signatures
- **Gradual migration**: You can migrate commands incrementally
- **OutputFlags validation**: Always call `flags.Validate()` early in `runX()`
- **Simple output**: If a command only has `--output` (no `--fields`), use `var output string` instead of OutputFlags

## Examples Completed

✅ `cmd/linear/commands/issue/get.go` - OutputFlags
✅ `cmd/linear/commands/team/list.go` - OutputFlags + PaginationFlags
✅ `cmd/linear/commands/label/delete.go` - ConfirmationFlags

Use these as reference when migrating similar commands.
