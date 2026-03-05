package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/the20100/typeform-cli/internal/config"
	"github.com/the20100/typeform-cli/internal/output"
	"github.com/the20100/typeform-cli/internal/validate"
)

var themeCmd = &cobra.Command{
	Use:   "theme",
	Short: "Manage Typeform themes",
}

// ---- theme list ----

var (
	themeListPage     int
	themeListPageSize int
)

var themeListCmd = &cobra.Command{
	Use:   "list",
	Short: "List themes",
	RunE: func(cmd *cobra.Command, args []string) error {
		params := buildParams(
			"page", intToStr(themeListPage),
			"page_size", intToStr(themeListPageSize),
		)
		items, err := client.ListThemes(params)
		if err != nil {
			return err
		}
		if output.IsJSON(cmd) {
			if fieldsFlag != "" {
				return printFilteredJSON(items, fieldsFlag, output.IsPretty(cmd))
			}
			return output.PrintJSON(items, output.IsPretty(cmd))
		}
		printThemesTable(items)
		return nil
	},
}

// ---- theme get ----

var themeGetCmd = &cobra.Command{
	Use:   "get <theme-id>",
	Short: "Get details of a specific theme",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := validate.ResourceID(args[0]); err != nil {
			return fmt.Errorf("invalid theme ID: %w", err)
		}
		theme, err := client.GetTheme(args[0])
		if err != nil {
			return err
		}
		if output.IsJSON(cmd) {
			return output.PrintJSON(theme, output.IsPretty(cmd))
		}
		output.PrintKeyValue([][]string{
			{"ID", theme.ID},
			{"Name", theme.Name},
			{"Visibility", theme.Visibility},
			{"Font", theme.Font},
		})
		return nil
	},
}

// ---- theme create ----

var themeCreateParams string

var themeCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new theme",
	Long: `Create a new Typeform theme.

Requires --params with the full JSON body.

Examples:
  typeform theme create --params '{"name":"My Theme","colors":{"question":"#000000","answer":"#333333","button":"#FF6B6B","background":"#FFFFFF"}}'`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := ensureConfig(); err != nil {
			return err
		}
		if err := config.CheckSecureMode(cfg, "create"); err != nil {
			return err
		}
		if themeCreateParams == "" {
			return fmt.Errorf("--params is required (JSON body for POST /themes)")
		}
		if err := validate.JSONPayload(themeCreateParams); err != nil {
			return fmt.Errorf("invalid --params: %w", err)
		}
		var payload any
		if err := json.Unmarshal([]byte(themeCreateParams), &payload); err != nil {
			return fmt.Errorf("parsing --params: %w", err)
		}
		if dryRunFlag {
			fmt.Println("dry-run: would POST /themes")
			return nil
		}
		theme, err := client.CreateTheme(payload)
		if err != nil {
			return err
		}
		if output.IsJSON(cmd) {
			return output.PrintJSON(theme, output.IsPretty(cmd))
		}
		fmt.Printf("Created theme %q (ID: %s)\n", theme.Name, theme.ID)
		return nil
	},
}

// ---- theme update ----

var themeUpdateParams string

var themeUpdateCmd = &cobra.Command{
	Use:   "update <theme-id>",
	Short: "Update a theme (blocked in secure mode)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := validate.ResourceID(args[0]); err != nil {
			return fmt.Errorf("invalid theme ID: %w", err)
		}
		if err := ensureConfig(); err != nil {
			return err
		}
		if err := config.CheckSecureMode(cfg, "update"); err != nil {
			return err
		}
		if themeUpdateParams == "" {
			return fmt.Errorf("--params is required (JSON body for PUT /themes/{id})")
		}
		if err := validate.JSONPayload(themeUpdateParams); err != nil {
			return fmt.Errorf("invalid --params: %w", err)
		}
		var payload any
		if err := json.Unmarshal([]byte(themeUpdateParams), &payload); err != nil {
			return fmt.Errorf("parsing --params: %w", err)
		}
		if dryRunFlag {
			fmt.Printf("dry-run: would PUT /themes/%s\n", args[0])
			return nil
		}
		theme, err := client.UpdateTheme(args[0], payload)
		if err != nil {
			return err
		}
		if output.IsJSON(cmd) {
			return output.PrintJSON(theme, output.IsPretty(cmd))
		}
		fmt.Printf("Updated theme %q (ID: %s)\n", theme.Name, theme.ID)
		return nil
	},
}

// ---- theme delete ----

var themeDeleteCmd = &cobra.Command{
	Use:   "delete <theme-id>",
	Short: "Delete a theme (blocked in secure mode)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := validate.ResourceID(args[0]); err != nil {
			return fmt.Errorf("invalid theme ID: %w", err)
		}
		if err := ensureConfig(); err != nil {
			return err
		}
		if err := config.CheckSecureMode(cfg, "delete"); err != nil {
			return err
		}
		if dryRunFlag {
			fmt.Printf("dry-run: would DELETE /themes/%s\n", args[0])
			return nil
		}
		if err := client.DeleteTheme(args[0]); err != nil {
			return err
		}
		fmt.Printf("Deleted theme %s\n", args[0])
		return nil
	},
}

func init() {
	themeListCmd.Flags().IntVar(&themeListPage, "page", 0, "Page number")
	themeListCmd.Flags().IntVar(&themeListPageSize, "page-size", 0, "Results per page")

	themeCreateCmd.Flags().StringVar(&themeCreateParams, "params", "", "Raw JSON body for POST /themes")
	themeUpdateCmd.Flags().StringVar(&themeUpdateParams, "params", "", "Raw JSON body for PUT /themes/{id}")

	themeCmd.AddCommand(themeListCmd, themeGetCmd, themeCreateCmd, themeUpdateCmd, themeDeleteCmd)
	rootCmd.AddCommand(themeCmd)

	RegisterSchema("themes.list", SchemaEntry{
		Command:     "typeform theme list",
		Description: "List themes",
		Flags: []SchemaFlag{
			{Name: "--page", Type: "int", Desc: "Page number"},
			{Name: "--page-size", Type: "int", Desc: "Results per page"},
		},
		Mutating: false,
	})
	RegisterSchema("themes.get", SchemaEntry{
		Command:     "typeform theme get <theme-id>",
		Description: "Get details of a specific theme",
		Args:        []SchemaArg{{Name: "theme-id", Required: true, Desc: "Theme ID"}},
		Mutating:    false,
	})
	RegisterSchema("themes.create", SchemaEntry{
		Command:     "typeform theme create --params '{...}'",
		Description: "Create a new theme",
		Flags:       []SchemaFlag{{Name: "--params", Type: "string", Required: true, Desc: "JSON body for theme creation"}},
		Mutating:    true,
	})
	RegisterSchema("themes.update", SchemaEntry{
		Command:     "typeform theme update <theme-id> --params '{...}'",
		Description: "Update a theme (blocked in secure mode)",
		Args:        []SchemaArg{{Name: "theme-id", Required: true, Desc: "Theme ID"}},
		Flags:       []SchemaFlag{{Name: "--params", Type: "string", Required: true, Desc: "JSON body"}},
		Mutating:    true,
	})
	RegisterSchema("themes.delete", SchemaEntry{
		Command:     "typeform theme delete <theme-id>",
		Description: "Delete a theme (blocked in secure mode)",
		Args:        []SchemaArg{{Name: "theme-id", Required: true, Desc: "Theme ID"}},
		Mutating:    true,
	})
}
