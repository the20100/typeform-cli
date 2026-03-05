package cmd

import (
	"fmt"
	"os"
	"runtime"

	"github.com/spf13/cobra"
	"github.com/the20100/typeform-cli/internal/api"
	"github.com/the20100/typeform-cli/internal/config"
)

var (
	jsonFlag   bool
	prettyFlag bool
	fieldsFlag string
	dryRunFlag bool
	client     *api.Client
	cfg        *config.Config
)

var rootCmd = &cobra.Command{
	Use:   "typeform",
	Short: "Typeform CLI — manage Typeform via the API",
	Long: `typeform is a CLI tool for the Typeform API.

It outputs JSON when piped (for agent use) and human-readable tables in a terminal.

WORKSPACE LOCK: This CLI only operates on workspaces listed in the config file.
SECURE MODE: When enabled, only read and create operations are allowed.

Token resolution order:
  1. TYPEFORM_API_KEY env var
  2. Config file  (~/.config/typeform/config.json  via: typeform auth set-key)

Examples:
  typeform auth set-key <token>
  typeform forms list
  typeform responses list <form-id>
  typeform schema forms.list`,
	SilenceUsage: true,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&jsonFlag, "json", false, "Force JSON output")
	rootCmd.PersistentFlags().BoolVar(&prettyFlag, "pretty", false, "Force pretty-printed JSON output (implies --json)")
	rootCmd.PersistentFlags().StringVar(&fieldsFlag, "fields", "", "Comma-separated list of fields to include in response")
	rootCmd.PersistentFlags().BoolVar(&dryRunFlag, "dry-run", false, "Validate the request locally without hitting the API")

	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if isAuthCommand(cmd) || cmd.Name() == "info" || cmd.Name() == "update" || cmd.Name() == "schema" {
			return nil
		}

		var err error
		cfg, err = config.Load()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		key, err := resolveAPIKey()
		if err != nil {
			return err
		}
		client = api.NewClient(key)
		return nil
	}

	rootCmd.AddCommand(infoCmd)
}

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Show tool info: config path, auth status, and environment",
	Run: func(cmd *cobra.Command, args []string) {
		printInfo()
	},
}

func printInfo() {
	fmt.Printf("typeform — Typeform CLI\n\n")
	exe, _ := os.Executable()
	fmt.Printf("  binary:  %s\n", exe)
	fmt.Printf("  os/arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
	fmt.Println()
	fmt.Println("  config paths by OS:")
	fmt.Printf("    macOS:    ~/Library/Application Support/typeform/config.json\n")
	fmt.Printf("    Linux:    ~/.config/typeform/config.json\n")
	fmt.Printf("    Windows:  %%AppData%%\\typeform\\config.json\n")
	fmt.Printf("  config:   %s\n", config.Path())
	fmt.Println()

	c, _ := config.Load()
	if c != nil {
		fmt.Printf("  workspaces: %v\n", c.Workspaces)
		fmt.Printf("  secure_mode: %v\n", c.SecureMode)
	}
	fmt.Println()
	fmt.Printf("  TYPEFORM_API_KEY = %s\n", maskOrEmpty(os.Getenv("TYPEFORM_API_KEY")))
}

func maskOrEmpty(v string) string {
	if v == "" {
		return "(not set)"
	}
	if len(v) <= 8 {
		return "***"
	}
	return v[:4] + "..." + v[len(v)-4:]
}

func resolveEnv(names ...string) string {
	for _, name := range names {
		if v := os.Getenv(name); v != "" {
			return v
		}
	}
	return ""
}

func resolveAPIKey() (string, error) {
	if k := resolveEnv(
		"TYPEFORM_API_KEY",
		"TYPEFORM_KEY",
		"TYPEFORM_TOKEN",
		"TYPEFORM_API_TOKEN",
		"TYPEFORM_API",
		"TYPEFORM_SECRET",
	); k != "" {
		return k, nil
	}
	if cfg == nil {
		var err error
		cfg, err = config.Load()
		if err != nil {
			return "", fmt.Errorf("failed to load config: %w", err)
		}
	}
	if cfg.APIKey != "" {
		return cfg.APIKey, nil
	}
	return "", fmt.Errorf("not authenticated — run: typeform auth set-key <token>\nor set TYPEFORM_API_KEY env var")
}

func isAuthCommand(cmd *cobra.Command) bool {
	if cmd.Name() == "auth" {
		return true
	}
	p := cmd.Parent()
	for p != nil {
		if p.Name() == "auth" {
			return true
		}
		p = p.Parent()
	}
	return false
}

// ensureConfig loads config if not already loaded.
func ensureConfig() error {
	if cfg != nil {
		return nil
	}
	var err error
	cfg, err = config.Load()
	return err
}
