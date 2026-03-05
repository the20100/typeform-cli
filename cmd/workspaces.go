package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/the20100/typeform-cli/internal/api"
	"github.com/the20100/typeform-cli/internal/config"
	"github.com/the20100/typeform-cli/internal/output"
	"github.com/the20100/typeform-cli/internal/validate"
)

var workspaceCmd = &cobra.Command{
	Use:   "workspace",
	Short: "Manage Typeform workspaces",
}

// ---- workspace list ----

var workspaceListCmd = &cobra.Command{
	Use:   "list",
	Short: "List workspaces (filtered to allowed workspaces only)",
	Long: `List Typeform workspaces.

Only workspaces listed in the config are shown (workspace lock).

Examples:
  typeform workspace list
  typeform workspace list --json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := ensureConfig(); err != nil {
			return err
		}
		items, err := client.ListWorkspaces(nil)
		if err != nil {
			return err
		}
		// Filter to allowed workspaces only
		filtered := filterAllowedWorkspaces(items)
		if output.IsJSON(cmd) {
			if fieldsFlag != "" {
				return printFilteredJSON(filtered, fieldsFlag, output.IsPretty(cmd))
			}
			return output.PrintJSON(filtered, output.IsPretty(cmd))
		}
		printWorkspacesTable(filtered)
		return nil
	},
}

// ---- workspace get ----

var workspaceGetCmd = &cobra.Command{
	Use:   "get <workspace-id>",
	Short: "Get details of a specific workspace",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := validate.ResourceID(args[0]); err != nil {
			return fmt.Errorf("invalid workspace ID: %w", err)
		}
		if err := ensureConfig(); err != nil {
			return err
		}
		if err := config.IsWorkspaceAllowed(cfg, args[0]); err != nil {
			return err
		}
		ws, err := client.GetWorkspace(args[0])
		if err != nil {
			return err
		}
		if output.IsJSON(cmd) {
			return output.PrintJSON(ws, output.IsPretty(cmd))
		}
		output.PrintKeyValue([][]string{
			{"ID", ws.ID},
			{"Name", ws.Name},
			{"Default", output.FormatBool(ws.Default)},
			{"Shared", output.FormatBool(ws.Shared)},
			{"Account ID", ws.AccountID},
		})
		return nil
	},
}

// ---- workspace create ----

var workspaceCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new workspace",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := validate.SafeString(args[0], 256); err != nil {
			return fmt.Errorf("invalid workspace name: %w", err)
		}
		if err := ensureConfig(); err != nil {
			return err
		}
		if err := config.CheckSecureMode(cfg, "create"); err != nil {
			return err
		}
		if dryRunFlag {
			fmt.Printf("dry-run: would create workspace %q\n", args[0])
			return nil
		}
		ws, err := client.CreateWorkspace(args[0])
		if err != nil {
			return err
		}
		if output.IsJSON(cmd) {
			return output.PrintJSON(ws, output.IsPretty(cmd))
		}
		fmt.Printf("Created workspace %q (ID: %s)\n", ws.Name, ws.ID)
		fmt.Println("Note: add this workspace ID to your config to use it with this CLI.")
		return nil
	},
}

// ---- workspace update ----

var workspaceUpdateName string

var workspaceUpdateCmd = &cobra.Command{
	Use:   "update <workspace-id>",
	Short: "Update a workspace (rename)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := validate.ResourceID(args[0]); err != nil {
			return fmt.Errorf("invalid workspace ID: %w", err)
		}
		if err := ensureConfig(); err != nil {
			return err
		}
		if err := config.IsWorkspaceAllowed(cfg, args[0]); err != nil {
			return err
		}
		if err := config.CheckSecureMode(cfg, "update"); err != nil {
			return err
		}
		if workspaceUpdateName == "" {
			return fmt.Errorf("--name is required")
		}
		if err := validate.SafeString(workspaceUpdateName, 256); err != nil {
			return fmt.Errorf("invalid workspace name: %w", err)
		}
		patches := []api.WorkspacePatch{
			{Op: "replace", Path: "/name", Value: workspaceUpdateName},
		}
		if dryRunFlag {
			data, _ := json.MarshalIndent(patches, "", "  ")
			fmt.Printf("dry-run: would PATCH /workspaces/%s with:\n%s\n", args[0], string(data))
			return nil
		}
		ws, err := client.UpdateWorkspace(args[0], patches)
		if err != nil {
			return err
		}
		if output.IsJSON(cmd) {
			return output.PrintJSON(ws, output.IsPretty(cmd))
		}
		fmt.Printf("Updated workspace %q (ID: %s)\n", ws.Name, ws.ID)
		return nil
	},
}

// ---- workspace delete ----

var workspaceDeleteCmd = &cobra.Command{
	Use:   "delete <workspace-id>",
	Short: "Delete a workspace",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := validate.ResourceID(args[0]); err != nil {
			return fmt.Errorf("invalid workspace ID: %w", err)
		}
		if err := ensureConfig(); err != nil {
			return err
		}
		if err := config.IsWorkspaceAllowed(cfg, args[0]); err != nil {
			return err
		}
		if err := config.CheckSecureMode(cfg, "delete"); err != nil {
			return err
		}
		if dryRunFlag {
			fmt.Printf("dry-run: would DELETE /workspaces/%s\n", args[0])
			return nil
		}
		if err := client.DeleteWorkspace(args[0]); err != nil {
			return err
		}
		fmt.Printf("Deleted workspace %s\n", args[0])
		return nil
	},
}

func filterAllowedWorkspaces(items []api.Workspace) []api.Workspace {
	if cfg == nil || len(cfg.Workspaces) == 0 {
		return nil // no workspaces configured = show nothing
	}
	allowed := make(map[string]bool)
	for _, w := range cfg.Workspaces {
		allowed[w] = true
	}
	var filtered []api.Workspace
	for _, item := range items {
		if allowed[item.ID] {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

func init() {
	workspaceUpdateCmd.Flags().StringVar(&workspaceUpdateName, "name", "", "New workspace name")

	workspaceCmd.AddCommand(workspaceListCmd, workspaceGetCmd, workspaceCreateCmd, workspaceUpdateCmd, workspaceDeleteCmd)
	rootCmd.AddCommand(workspaceCmd)

	RegisterSchema("workspaces.list", SchemaEntry{
		Command:     "typeform workspace list",
		Description: "List workspaces (filtered to allowed workspaces in config)",
		Flags: []SchemaFlag{
			{Name: "--fields", Type: "string", Desc: "Comma-separated fields to include"},
		},
		Examples: []string{"typeform workspace list", "typeform workspace list --json"},
		Mutating: false,
	})
	RegisterSchema("workspaces.get", SchemaEntry{
		Command:     "typeform workspace get <workspace-id>",
		Description: "Get details of a specific workspace",
		Args:        []SchemaArg{{Name: "workspace-id", Required: true, Desc: "Workspace ID (must be in allowed list)"}},
		Mutating:    false,
	})
	RegisterSchema("workspaces.create", SchemaEntry{
		Command:     "typeform workspace create <name>",
		Description: "Create a new workspace",
		Args:        []SchemaArg{{Name: "name", Required: true, Desc: "Workspace name"}},
		Mutating:    true,
	})
	RegisterSchema("workspaces.update", SchemaEntry{
		Command:     "typeform workspace update <workspace-id> --name <new-name>",
		Description: "Rename a workspace (blocked in secure mode)",
		Args:        []SchemaArg{{Name: "workspace-id", Required: true, Desc: "Workspace ID (must be in allowed list)"}},
		Flags:       []SchemaFlag{{Name: "--name", Type: "string", Required: true, Desc: "New workspace name"}},
		Mutating:    true,
	})
	RegisterSchema("workspaces.delete", SchemaEntry{
		Command:     "typeform workspace delete <workspace-id>",
		Description: "Delete a workspace (blocked in secure mode)",
		Args:        []SchemaArg{{Name: "workspace-id", Required: true, Desc: "Workspace ID (must be in allowed list)"}},
		Mutating:    true,
	})
}
