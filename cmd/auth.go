package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/the20100/typeform-cli/internal/config"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage Typeform authentication",
}

var authSetKeyCmd = &cobra.Command{
	Use:   "set-key <personal-access-token>",
	Short: "Save a Typeform personal access token to the config file",
	Long: `Save a Typeform personal access token to the local config file.

Get your token from: https://admin.typeform.com/account#/section/tokens

The token is stored at:
  macOS:   ~/Library/Application Support/typeform/config.json
  Linux:   ~/.config/typeform/config.json
  Windows: %AppData%\typeform\config.json

You can also set the TYPEFORM_API_KEY env var instead of using this command.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]
		if len(key) < 8 {
			return fmt.Errorf("API token looks too short")
		}
		// Load existing config to preserve workspaces and secure_mode
		existing, _ := config.Load()
		if existing == nil {
			existing = &config.Config{}
		}
		existing.APIKey = key
		if err := config.Save(existing); err != nil {
			return fmt.Errorf("saving config: %w", err)
		}
		fmt.Printf("API token saved to %s\n", config.Path())
		fmt.Printf("Token: %s\n", maskOrEmpty(key))
		return nil
	},
}

var authStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current authentication status",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := config.Load()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}
		fmt.Printf("Config: %s\n\n", config.Path())
		if envKey := os.Getenv("TYPEFORM_API_KEY"); envKey != "" {
			fmt.Println("Token source: TYPEFORM_API_KEY env var (takes priority over config)")
			fmt.Printf("Token:        %s\n", maskOrEmpty(envKey))
		} else if c.APIKey != "" {
			fmt.Println("Token source: config file")
			fmt.Printf("Token:        %s\n", maskOrEmpty(c.APIKey))
		} else {
			fmt.Println("Status: not authenticated")
			fmt.Printf("\nRun: typeform auth set-key <your-token>\n")
			fmt.Printf("Or:  export TYPEFORM_API_KEY=<your-token>\n")
		}
		fmt.Printf("\nWorkspaces: %v\n", c.Workspaces)
		fmt.Printf("Secure mode: %v\n", c.SecureMode)
		return nil
	},
}

var authLogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Remove the saved API token from the config file",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := config.Clear(); err != nil {
			return fmt.Errorf("removing config: %w", err)
		}
		fmt.Println("API token removed from config.")
		return nil
	},
}

func init() {
	authCmd.AddCommand(authSetKeyCmd, authStatusCmd, authLogoutCmd)
	rootCmd.AddCommand(authCmd)
}
