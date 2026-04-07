// Package commands provides the Cobra command structure for the Linear CLI.
package commands

import (
	"crypto/tls"
	"fmt"
	"log"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/chainguard-dev/clog"
	"github.com/spf13/cobra"

	"github.com/chainguard-sandbox/go-linear/v2/cmd/linear/commands/attachment"
	"github.com/chainguard-sandbox/go-linear/v2/cmd/linear/commands/comment"
	"github.com/chainguard-sandbox/go-linear/v2/cmd/linear/commands/cycle"
	"github.com/chainguard-sandbox/go-linear/v2/cmd/linear/commands/document"
	"github.com/chainguard-sandbox/go-linear/v2/cmd/linear/commands/favorite"
	"github.com/chainguard-sandbox/go-linear/v2/cmd/linear/commands/initiative"
	"github.com/chainguard-sandbox/go-linear/v2/cmd/linear/commands/issue"
	"github.com/chainguard-sandbox/go-linear/v2/cmd/linear/commands/label"
	"github.com/chainguard-sandbox/go-linear/v2/cmd/linear/commands/notification"
	"github.com/chainguard-sandbox/go-linear/v2/cmd/linear/commands/organization"
	"github.com/chainguard-sandbox/go-linear/v2/cmd/linear/commands/project"
	"github.com/chainguard-sandbox/go-linear/v2/cmd/linear/commands/reaction"
	"github.com/chainguard-sandbox/go-linear/v2/cmd/linear/commands/roadmap"
	"github.com/chainguard-sandbox/go-linear/v2/cmd/linear/commands/state"
	"github.com/chainguard-sandbox/go-linear/v2/cmd/linear/commands/status"
	"github.com/chainguard-sandbox/go-linear/v2/cmd/linear/commands/team"
	"github.com/chainguard-sandbox/go-linear/v2/cmd/linear/commands/template"
	"github.com/chainguard-sandbox/go-linear/v2/cmd/linear/commands/user"
	"github.com/chainguard-sandbox/go-linear/v2/cmd/linear/commands/viewer"
	"github.com/chainguard-sandbox/go-linear/v2/cmd/linear/commands/webhook"
	"github.com/chainguard-sandbox/go-linear/v2/pkg/linear"
)

// clientConfig holds parsed configuration from environment variables.
// Read once at initialization to avoid repeated os.Getenv() calls.
type clientConfig struct {
	baseURL                string
	timeout                time.Duration
	retryAttempts          int
	retryInitial           time.Duration
	retryMax               time.Duration
	circuitBreakerFailures int
	circuitBreakerTimeout  time.Duration
	tlsMinVersion          uint16

	// Observability
	logger         *clog.Logger // nil = disabled
	metricsEnabled bool
}

// loadClientConfig reads and parses environment variables once.
// Called during root command initialization.
func loadClientConfig() *clientConfig {
	cfg := &clientConfig{
		// Defaults tuned for Linear's API characteristics:
		// - Rate limits: 250,000 points/hour for API keys, 1,500 requests/hour per user
		// - Max query complexity: 10,000 points per query
		// - Uses leaky bucket algorithm for rate limiting
		timeout:                30 * time.Second,
		retryAttempts:          3,
		retryInitial:           1 * time.Second,
		retryMax:               30 * time.Second,
		circuitBreakerFailures: 5,
		circuitBreakerTimeout:  60 * time.Second,
		tlsMinVersion:          tls.VersionTLS12,
	}

	// Custom API endpoint (optional)
	if baseURL := os.Getenv("LINEAR_BASE_URL"); baseURL != "" {
		cfg.baseURL = baseURL
	}

	// Request timeout
	if timeoutStr := os.Getenv("LINEAR_TIMEOUT"); timeoutStr != "" {
		if timeout, err := time.ParseDuration(timeoutStr); err == nil {
			cfg.timeout = timeout
		} else {
			log.Printf("Warning: Invalid LINEAR_TIMEOUT, using default (30s)")
		}
	}

	// Retry configuration
	if attemptsStr := os.Getenv("LINEAR_RETRY_ATTEMPTS"); attemptsStr != "" {
		if attempts, err := strconv.Atoi(attemptsStr); err == nil && attempts >= 0 {
			cfg.retryAttempts = attempts
		}
	}
	if initialStr := os.Getenv("LINEAR_RETRY_INITIAL"); initialStr != "" {
		if initial, err := time.ParseDuration(initialStr); err == nil {
			cfg.retryInitial = initial
		}
	}
	if maxStr := os.Getenv("LINEAR_RETRY_MAX"); maxStr != "" {
		if maxDuration, err := time.ParseDuration(maxStr); err == nil {
			cfg.retryMax = maxDuration
		}
	}

	// Circuit breaker configuration
	if cbFailuresStr := os.Getenv("LINEAR_CIRCUIT_BREAKER_FAILURES"); cbFailuresStr != "" {
		if failures, err := strconv.Atoi(cbFailuresStr); err == nil && failures > 0 {
			cfg.circuitBreakerFailures = failures
		}
	}
	if cbTimeoutStr := os.Getenv("LINEAR_CIRCUIT_BREAKER_TIMEOUT"); cbTimeoutStr != "" {
		if timeout, err := time.ParseDuration(cbTimeoutStr); err == nil {
			cfg.circuitBreakerTimeout = timeout
		}
	}

	// TLS configuration
	if tlsMinStr := os.Getenv("LINEAR_TLS_MIN_VERSION"); tlsMinStr != "" {
		switch tlsMinStr {
		case "1.2":
			cfg.tlsMinVersion = tls.VersionTLS12
		case "1.3":
			cfg.tlsMinVersion = tls.VersionTLS13
		default:
			log.Printf("Warning: Invalid LINEAR_TLS_MIN_VERSION, using TLS 1.2")
		}
	}

	// Logging configuration
	if logLevel := os.Getenv("LINEAR_LOG_LEVEL"); logLevel != "" {
		var level slog.Level
		switch strings.ToLower(logLevel) {
		case "debug":
			level = slog.LevelDebug
		case "info":
			level = slog.LevelInfo
		case "warn", "warning":
			level = slog.LevelWarn
		case "error":
			level = slog.LevelError
		default:
			log.Printf("Warning: Invalid LINEAR_LOG_LEVEL, logging disabled")
			return cfg
		}

		// Create JSON handler writing to stderr (not stdout)
		handler := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
			Level: level,
		})
		cfg.logger = clog.New(handler)
	}

	// Metrics configuration
	if metricsStr := os.Getenv("LINEAR_METRICS_ENABLED"); metricsStr != "" {
		if enabled, err := strconv.ParseBool(metricsStr); err == nil {
			cfg.metricsEnabled = enabled
		} else {
			log.Printf("Warning: Invalid LINEAR_METRICS_ENABLED, must be true/false")
		}
	}

	return cfg
}

// buildClientOptions constructs Linear client options from parsed config.
// Config is loaded once at initialization, this function just converts to options.
func buildClientOptions(cfg *clientConfig) []linear.Option {
	opts := []linear.Option{}

	// Base URL (optional)
	if cfg.baseURL != "" {
		opts = append(opts, linear.WithBaseURL(cfg.baseURL))
	}

	// Always apply timeout
	opts = append(opts, linear.WithTimeout(cfg.timeout))

	// Apply retry if enabled
	if cfg.retryAttempts > 0 {
		opts = append(opts, linear.WithRetry(
			cfg.retryAttempts,
			cfg.retryInitial,
			cfg.retryMax,
		))
	}

	// Always apply circuit breaker and TLS config
	opts = append(opts,
		linear.WithCircuitBreaker(&linear.CircuitBreaker{
			MaxFailures:  cfg.circuitBreakerFailures,
			ResetTimeout: cfg.circuitBreakerTimeout,
		}),
		linear.WithTLSConfig(&tls.Config{
			MinVersion: cfg.tlsMinVersion, // #nosec G402 - configurable, defaults to TLS 1.2
		}),
	)

	// Apply logger if configured
	if cfg.logger != nil {
		opts = append(opts, linear.WithLogger(cfg.logger))
	}

	// Apply metrics if enabled
	if cfg.metricsEnabled {
		// Call EnableMetrics() once to register global collectors
		linear.EnableMetrics()
		opts = append(opts, linear.WithMetrics())
	}

	return opts
}

// NewRootCommand creates the root command for the Linear CLI.
func NewRootCommand() *cobra.Command {
	var apiKey string
	var verbose bool

	// Load configuration from environment variables ONCE at initialization
	cfg := loadClientConfig()

	rootCmd := &cobra.Command{
		Use:     "go-linear",
		Version: linear.Version,
		Short:   "Linear MCP server for AI agents",
		Long: `Linear MCP server provides AI agents with command-line access to Linear.

Optimized for both human users and AI agents via MCP (Model Context Protocol).
Supports parameter-rich commands for complex queries without multi-step workflows.

Examples:
  # List my urgent issues
  go-linear issue list --assignee=me --priority=1

  # Find completed issues from yesterday
  go-linear issue list --team=Engineering --completed-after=yesterday --completed-before=today

  # Get user's completed work
  go-linear user completed --user=alice@company.com --completed-after=7d

Environment Variables:
  LINEAR_API_KEY                     Linear API key (required)
  LINEAR_BASE_URL                    Custom API endpoint (default: https://api.linear.app/graphql)
  LINEAR_TIMEOUT                     Request timeout (default: 30s)
  LINEAR_RETRY_ATTEMPTS              Number of retry attempts (default: 3)
  LINEAR_RETRY_INITIAL               Initial retry backoff (default: 1s)
  LINEAR_RETRY_MAX                   Maximum retry backoff (default: 30s)
  LINEAR_CIRCUIT_BREAKER_FAILURES    Failures before circuit opens (default: 5)
  LINEAR_CIRCUIT_BREAKER_TIMEOUT     Circuit breaker reset timeout (default: 60s)
  LINEAR_TLS_MIN_VERSION             Minimum TLS version: 1.2 or 1.3 (default: 1.2)
  LINEAR_LOG_LEVEL                   Logging level: debug|info|warn|error (default: disabled)
  LINEAR_METRICS_ENABLED             Enable Prometheus metrics: true|false (default: false)`,
		SilenceUsage: true, // Don't print usage on errors (reduces context bloat for AI agents)
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Initialize Linear client for all subcommands
			if apiKey == "" {
				apiKey = os.Getenv("LINEAR_API_KEY")
			}
			if apiKey == "" {
				return fmt.Errorf("LINEAR_API_KEY environment variable or --api-key flag required")
			}

			// Validation only - actual client creation happens in subcommands
			return nil
		},
	}

	// Global flags
	rootCmd.PersistentFlags().StringVar(&apiKey, "api-key", "", "Linear API key (or set LINEAR_API_KEY env var)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose debug logging")

	// Get API key for subcommands - uses pre-loaded config
	getClient := func() (*linear.Client, error) {
		key := apiKey
		if key == "" {
			key = os.Getenv("LINEAR_API_KEY")
		}
		if key == "" {
			return nil, fmt.Errorf("LINEAR_API_KEY environment variable or --api-key flag required")
		}

		// Override logger if --verbose flag is set
		cfgToUse := cfg
		if verbose && cfg.logger == nil {
			// Create temporary config with debug logger
			handler := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
				Level: slog.LevelDebug,
			})
			// Clone config to avoid mutation
			cfgCopy := *cfg
			cfgCopy.logger = clog.New(handler)
			cfgToUse = &cfgCopy
		}

		// Build client options from config
		opts := buildClientOptions(cfgToUse)

		return linear.NewClient(key, opts...)
	}

	// Add subcommands (ordered alphabetically)
	rootCmd.AddCommand(attachment.NewAttachmentCommand(getClient))
	rootCmd.AddCommand(comment.NewCommentCommand(getClient))
	rootCmd.AddCommand(cycle.NewCycleCommand(getClient))
	rootCmd.AddCommand(document.NewDocumentCommand(getClient))
	rootCmd.AddCommand(favorite.NewFavoriteCommand(getClient))
	rootCmd.AddCommand(initiative.NewInitiativeCommand(getClient))
	rootCmd.AddCommand(issue.NewIssueCommand(getClient))
	rootCmd.AddCommand(label.NewLabelCommand(getClient))
	rootCmd.AddCommand(notification.NewNotificationCommand(getClient))
	rootCmd.AddCommand(organization.NewOrganizationCommand(getClient))
	rootCmd.AddCommand(project.NewProjectCommand(getClient))
	rootCmd.AddCommand(reaction.NewReactionCommand(getClient))
	rootCmd.AddCommand(roadmap.NewRoadmapCommand(getClient))
	rootCmd.AddCommand(state.NewStateCommand(getClient))
	rootCmd.AddCommand(status.NewStatusCommand())
	rootCmd.AddCommand(team.NewTeamCommand(getClient))
	rootCmd.AddCommand(template.NewTemplateCommand(getClient))
	rootCmd.AddCommand(user.NewUserCommand(getClient))
	rootCmd.AddCommand(viewer.NewViewerCommand(getClient))
	rootCmd.AddCommand(webhook.NewWebhookCommand(getClient))

	return rootCmd
}
