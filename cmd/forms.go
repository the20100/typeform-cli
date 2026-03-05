package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/the20100/typeform-cli/internal/api"
	"github.com/the20100/typeform-cli/internal/config"
	"github.com/the20100/typeform-cli/internal/output"
	"github.com/the20100/typeform-cli/internal/validate"
)

var formCmd = &cobra.Command{
	Use:   "form",
	Short: "Manage Typeform forms",
}

// ---- form list ----

var (
	formListSearch      string
	formListPage        int
	formListPageSize    int
	formListWorkspaceID string
)

var formListCmd = &cobra.Command{
	Use:   "list",
	Short: "List forms (filtered to allowed workspaces)",
	Long: `List Typeform forms.

Forms are filtered to only show those belonging to allowed workspaces.
Use --workspace-id to filter to a specific workspace.

Examples:
  typeform form list
  typeform form list --workspace-id abc123
  typeform form list --search "survey" --json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := ensureConfig(); err != nil {
			return err
		}

		// If workspace specified, validate it's allowed
		if formListWorkspaceID != "" {
			if err := config.IsWorkspaceAllowed(cfg, formListWorkspaceID); err != nil {
				return err
			}
		}

		params := buildParams(
			"search", formListSearch,
			"page", intToStr(formListPage),
			"page_size", intToStr(formListPageSize),
			"workspace_id", formListWorkspaceID,
		)

		items, err := client.ListForms(params)
		if err != nil {
			return err
		}

		// If no workspace filter, filter to forms in allowed workspaces
		if formListWorkspaceID == "" {
			items = filterFormsByWorkspace(items)
		}

		if output.IsJSON(cmd) {
			if fieldsFlag != "" {
				return printFilteredJSON(items, fieldsFlag, output.IsPretty(cmd))
			}
			return output.PrintJSON(items, output.IsPretty(cmd))
		}
		printFormsTable(items)
		return nil
	},
}

// ---- form get ----

var formGetCmd = &cobra.Command{
	Use:   "get <form-id>",
	Short: "Get details of a specific form",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := validate.ResourceID(args[0]); err != nil {
			return fmt.Errorf("invalid form ID: %w", err)
		}
		if err := ensureConfig(); err != nil {
			return err
		}
		// Verify form belongs to allowed workspace
		if err := checkFormWorkspace(args[0]); err != nil {
			return err
		}
		form, err := client.GetForm(args[0])
		if err != nil {
			return err
		}
		if output.IsJSON(cmd) {
			if fieldsFlag != "" {
				return printFilteredJSON(form, fieldsFlag, output.IsPretty(cmd))
			}
			return output.PrintJSON(form, output.IsPretty(cmd))
		}
		lang := ""
		if form.Settings != nil {
			lang = form.Settings.Language
		}
		output.PrintKeyValue([][]string{
			{"ID", form.ID},
			{"Title", form.Title},
			{"Type", form.Type},
			{"Language", lang},
			{"Last Updated", output.FormatTime(form.LastUpdated)},
			{"Created At", output.FormatTime(form.CreatedAt)},
			{"Fields", fmt.Sprintf("%d", len(form.Fields))},
		})
		return nil
	},
}

// ---- form create ----

var (
	formCreateWorkspace string
	formCreateType      string
	formCreateParams    string
)

var formCreateCmd = &cobra.Command{
	Use:   "create <title>",
	Short: "Create a new form",
	Long: `Create a new Typeform form.

Supports two input modes:
  Human-first: typeform form create "My Survey" --workspace-id abc123
  Agent-first: typeform form create "My Survey" --workspace-id abc123 --params '{"fields":[...]}'

Examples:
  typeform form create "Customer Feedback" --workspace-id abc123
  typeform form create "Quiz" --workspace-id abc123 --type quiz`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := validate.SafeString(args[0], 512); err != nil {
			return fmt.Errorf("invalid form title: %w", err)
		}
		if err := ensureConfig(); err != nil {
			return err
		}
		if err := config.CheckSecureMode(cfg, "create"); err != nil {
			return err
		}
		if formCreateWorkspace == "" {
			return fmt.Errorf("--workspace-id is required (must be in allowed list)")
		}
		if err := config.IsWorkspaceAllowed(cfg, formCreateWorkspace); err != nil {
			return err
		}

		req := &api.FormCreateRequest{
			Title:         args[0],
			Type:          formCreateType,
			WorkspaceHref: fmt.Sprintf("%s/workspaces/%s", "https://api.typeform.com", formCreateWorkspace),
		}

		// Merge --params if provided
		if formCreateParams != "" {
			if err := validate.JSONPayload(formCreateParams); err != nil {
				return fmt.Errorf("invalid --params: %w", err)
			}
			if err := json.Unmarshal([]byte(formCreateParams), req); err != nil {
				return fmt.Errorf("parsing --params: %w", err)
			}
			// Re-set title and workspace (user args take precedence)
			req.Title = args[0]
			req.WorkspaceHref = fmt.Sprintf("%s/workspaces/%s", "https://api.typeform.com", formCreateWorkspace)
		}

		if dryRunFlag {
			data, _ := json.MarshalIndent(req, "", "  ")
			fmt.Printf("dry-run: would POST /forms with:\n%s\n", string(data))
			return nil
		}

		form, err := client.CreateForm(req)
		if err != nil {
			return err
		}
		if output.IsJSON(cmd) {
			return output.PrintJSON(form, output.IsPretty(cmd))
		}
		fmt.Printf("Created form %q (ID: %s)\n", form.Title, form.ID)
		return nil
	},
}

// ---- form update ----

var (
	formUpdateParams string
)

var formUpdateCmd = &cobra.Command{
	Use:   "update <form-id>",
	Short: "Update a form (full replace via PUT)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := validate.ResourceID(args[0]); err != nil {
			return fmt.Errorf("invalid form ID: %w", err)
		}
		if err := ensureConfig(); err != nil {
			return err
		}
		if err := config.CheckSecureMode(cfg, "update"); err != nil {
			return err
		}
		if err := checkFormWorkspace(args[0]); err != nil {
			return err
		}
		if formUpdateParams == "" {
			return fmt.Errorf("--params is required (JSON body for PUT /forms/{id})")
		}
		if err := validate.JSONPayload(formUpdateParams); err != nil {
			return fmt.Errorf("invalid --params: %w", err)
		}
		var payload any
		if err := json.Unmarshal([]byte(formUpdateParams), &payload); err != nil {
			return fmt.Errorf("parsing --params: %w", err)
		}
		if dryRunFlag {
			fmt.Printf("dry-run: would PUT /forms/%s\n", args[0])
			return nil
		}
		form, err := client.UpdateForm(args[0], payload)
		if err != nil {
			return err
		}
		if output.IsJSON(cmd) {
			return output.PrintJSON(form, output.IsPretty(cmd))
		}
		fmt.Printf("Updated form %q (ID: %s)\n", form.Title, form.ID)
		return nil
	},
}

// ---- form patch ----

var (
	formPatchParams string
)

var formPatchCmd = &cobra.Command{
	Use:   "patch <form-id>",
	Short: "Partially update a form (PATCH)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := validate.ResourceID(args[0]); err != nil {
			return fmt.Errorf("invalid form ID: %w", err)
		}
		if err := ensureConfig(); err != nil {
			return err
		}
		if err := config.CheckSecureMode(cfg, "update"); err != nil {
			return err
		}
		if err := checkFormWorkspace(args[0]); err != nil {
			return err
		}
		if formPatchParams == "" {
			return fmt.Errorf("--params is required (JSON body for PATCH /forms/{id})")
		}
		if err := validate.JSONPayload(formPatchParams); err != nil {
			return fmt.Errorf("invalid --params: %w", err)
		}
		var payload any
		if err := json.Unmarshal([]byte(formPatchParams), &payload); err != nil {
			return fmt.Errorf("parsing --params: %w", err)
		}
		if dryRunFlag {
			fmt.Printf("dry-run: would PATCH /forms/%s\n", args[0])
			return nil
		}
		form, err := client.PatchForm(args[0], payload)
		if err != nil {
			return err
		}
		if output.IsJSON(cmd) {
			return output.PrintJSON(form, output.IsPretty(cmd))
		}
		fmt.Printf("Patched form %q (ID: %s)\n", form.Title, form.ID)
		return nil
	},
}

// ---- form delete ----

var formDeleteCmd = &cobra.Command{
	Use:   "delete <form-id>",
	Short: "Delete a form (blocked in secure mode)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := validate.ResourceID(args[0]); err != nil {
			return fmt.Errorf("invalid form ID: %w", err)
		}
		if err := ensureConfig(); err != nil {
			return err
		}
		if err := config.CheckSecureMode(cfg, "delete"); err != nil {
			return err
		}
		if err := checkFormWorkspace(args[0]); err != nil {
			return err
		}
		if dryRunFlag {
			fmt.Printf("dry-run: would DELETE /forms/%s\n", args[0])
			return nil
		}
		if err := client.DeleteForm(args[0]); err != nil {
			return err
		}
		fmt.Printf("Deleted form %s\n", args[0])
		return nil
	},
}

// checkFormWorkspace verifies a form belongs to an allowed workspace by fetching it.
func checkFormWorkspace(formID string) error {
	form, err := client.GetForm(formID)
	if err != nil {
		return err
	}
	if form.Workspace == nil || form.Workspace.Href == "" {
		return fmt.Errorf("form %s has no workspace — cannot verify workspace lock", formID)
	}
	// Extract workspace ID from href like https://api.typeform.com/workspaces/abc123
	parts := strings.Split(form.Workspace.Href, "/")
	if len(parts) > 0 {
		wsID := parts[len(parts)-1]
		return config.IsWorkspaceAllowed(cfg, wsID)
	}
	return fmt.Errorf("cannot extract workspace ID from form href: %s", form.Workspace.Href)
}

// filterFormsByWorkspace filters forms to only those in allowed workspaces.
func filterFormsByWorkspace(forms []api.Form) []api.Form {
	if cfg == nil || len(cfg.Workspaces) == 0 {
		return nil
	}
	allowed := make(map[string]bool)
	for _, w := range cfg.Workspaces {
		allowed[w] = true
	}
	var filtered []api.Form
	for _, f := range forms {
		if f.Workspace != nil && f.Workspace.Href != "" {
			parts := strings.Split(f.Workspace.Href, "/")
			if len(parts) > 0 {
				wsID := parts[len(parts)-1]
				if allowed[wsID] {
					filtered = append(filtered, f)
				}
			}
		}
	}
	return filtered
}

func intToStr(n int) string {
	if n <= 0 {
		return ""
	}
	return fmt.Sprintf("%d", n)
}

func init() {
	formListCmd.Flags().StringVar(&formListSearch, "search", "", "Search forms by title")
	formListCmd.Flags().IntVar(&formListPage, "page", 0, "Page number")
	formListCmd.Flags().IntVar(&formListPageSize, "page-size", 0, "Results per page (max 200)")
	formListCmd.Flags().StringVar(&formListWorkspaceID, "workspace-id", "", "Filter by workspace ID (must be in allowed list)")

	formCreateCmd.Flags().StringVar(&formCreateWorkspace, "workspace-id", "", "Workspace ID (required, must be in allowed list)")
	formCreateCmd.Flags().StringVar(&formCreateType, "type", "", "Form type: quiz, classification, score, branching")
	formCreateCmd.Flags().StringVar(&formCreateParams, "params", "", "Raw JSON payload for advanced form creation")

	formUpdateCmd.Flags().StringVar(&formUpdateParams, "params", "", "Raw JSON body for PUT /forms/{id}")
	formPatchCmd.Flags().StringVar(&formPatchParams, "params", "", "Raw JSON body for PATCH /forms/{id}")

	formCmd.AddCommand(formListCmd, formGetCmd, formCreateCmd, formUpdateCmd, formPatchCmd, formDeleteCmd)
	rootCmd.AddCommand(formCmd)

	RegisterSchema("forms.list", SchemaEntry{
		Command:     "typeform form list",
		Description: "List forms (filtered to allowed workspaces)",
		Flags: []SchemaFlag{
			{Name: "--search", Type: "string", Desc: "Search forms by title"},
			{Name: "--workspace-id", Type: "string", Desc: "Filter by workspace ID (must be in allowed list)"},
			{Name: "--page", Type: "int", Desc: "Page number"},
			{Name: "--page-size", Type: "int", Desc: "Results per page (max 200)"},
			{Name: "--fields", Type: "string", Desc: "Comma-separated fields to include in JSON output"},
		},
		Examples: []string{
			"typeform form list",
			"typeform form list --workspace-id abc123 --json",
			"typeform form list --search survey",
		},
		Mutating: false,
	})
	RegisterSchema("forms.get", SchemaEntry{
		Command:     "typeform form get <form-id>",
		Description: "Get details of a specific form (must be in allowed workspace)",
		Args:        []SchemaArg{{Name: "form-id", Required: true, Desc: "Form ID"}},
		Mutating:    false,
	})
	RegisterSchema("forms.create", SchemaEntry{
		Command:     "typeform form create <title>",
		Description: "Create a new form in an allowed workspace",
		Args:        []SchemaArg{{Name: "title", Required: true, Desc: "Form title"}},
		Flags: []SchemaFlag{
			{Name: "--workspace-id", Type: "string", Required: true, Desc: "Workspace ID (must be in allowed list)"},
			{Name: "--type", Type: "string", Desc: "Form type: quiz, classification, score, branching"},
			{Name: "--params", Type: "string", Desc: "Raw JSON payload for advanced form creation"},
		},
		Mutating: true,
	})
	RegisterSchema("forms.update", SchemaEntry{
		Command:     "typeform form update <form-id> --params '{...}'",
		Description: "Full update a form via PUT (blocked in secure mode)",
		Args:        []SchemaArg{{Name: "form-id", Required: true, Desc: "Form ID"}},
		Flags:       []SchemaFlag{{Name: "--params", Type: "string", Required: true, Desc: "JSON body"}},
		Mutating:    true,
	})
	RegisterSchema("forms.patch", SchemaEntry{
		Command:     "typeform form patch <form-id> --params '{...}'",
		Description: "Partially update a form via PATCH (blocked in secure mode)",
		Args:        []SchemaArg{{Name: "form-id", Required: true, Desc: "Form ID"}},
		Flags:       []SchemaFlag{{Name: "--params", Type: "string", Required: true, Desc: "JSON body"}},
		Mutating:    true,
	})
	RegisterSchema("forms.delete", SchemaEntry{
		Command:     "typeform form delete <form-id>",
		Description: "Delete a form (blocked in secure mode)",
		Args:        []SchemaArg{{Name: "form-id", Required: true, Desc: "Form ID"}},
		Mutating:    true,
	})
}
