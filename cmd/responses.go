package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/the20100/typeform-cli/internal/config"
	"github.com/the20100/typeform-cli/internal/output"
	"github.com/the20100/typeform-cli/internal/validate"
)

var responseCmd = &cobra.Command{
	Use:   "response",
	Short: "Manage Typeform form responses",
}

// ---- response list ----

var (
	responseListPageSize    int
	responseListSince       string
	responseListUntil       string
	responseListAfter       string
	responseListBefore      string
	responseListSort        string
	responseListQuery       string
	responseListResponseType string
	responseListIncludedIDs string
)

var responseListCmd = &cobra.Command{
	Use:   "list <form-id>",
	Short: "List responses for a form",
	Long: `Retrieve responses for a Typeform form.

The form must belong to an allowed workspace.

Examples:
  typeform response list abc123
  typeform response list abc123 --page-size 100 --json
  typeform response list abc123 --since 2024-01-01T00:00:00Z
  typeform response list abc123 --response-type completed`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := validate.ResourceID(args[0]); err != nil {
			return fmt.Errorf("invalid form ID: %w", err)
		}
		if err := ensureConfig(); err != nil {
			return err
		}
		if err := checkFormWorkspace(args[0]); err != nil {
			return err
		}

		params := buildParams(
			"page_size", intToStr(responseListPageSize),
			"since", responseListSince,
			"until", responseListUntil,
			"after", responseListAfter,
			"before", responseListBefore,
			"sort", responseListSort,
			"query", responseListQuery,
			"response_type", responseListResponseType,
			"included_response_ids", responseListIncludedIDs,
		)

		resp, err := client.ListResponses(args[0], params)
		if err != nil {
			return err
		}

		if output.IsJSON(cmd) {
			if fieldsFlag != "" {
				return printFilteredJSON(resp, fieldsFlag, output.IsPretty(cmd))
			}
			return output.PrintJSON(resp, output.IsPretty(cmd))
		}

		fmt.Printf("Total: %d responses, %d pages\n\n", resp.TotalItems, resp.PageCount)
		if len(resp.Items) == 0 {
			fmt.Println("No responses found.")
			return nil
		}
		headers := []string{"RESPONSE ID", "SUBMITTED AT", "ANSWERS"}
		rows := make([][]string, len(resp.Items))
		for i, item := range resp.Items {
			rows[i] = []string{
				item.ResponseID,
				output.FormatTime(item.SubmittedAt),
				fmt.Sprintf("%d", len(item.Answers)),
			}
		}
		output.PrintTable(headers, rows)
		return nil
	},
}

// ---- response delete ----

var responseDeleteCmd = &cobra.Command{
	Use:   "delete <form-id> <response-ids...>",
	Short: "Delete responses from a form (blocked in secure mode)",
	Long: `Delete specific responses from a form.

The form must belong to an allowed workspace.
Provide response tokens as arguments (space-separated).

Examples:
  typeform response delete abc123 token1 token2`,
	Args: cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		formID := args[0]
		responseIDs := args[1:]

		if err := validate.ResourceID(formID); err != nil {
			return fmt.Errorf("invalid form ID: %w", err)
		}
		for _, id := range responseIDs {
			if err := validate.ResourceID(id); err != nil {
				return fmt.Errorf("invalid response ID %q: %w", id, err)
			}
		}
		if err := ensureConfig(); err != nil {
			return err
		}
		if err := config.CheckSecureMode(cfg, "delete"); err != nil {
			return err
		}
		if err := checkFormWorkspace(formID); err != nil {
			return err
		}
		if dryRunFlag {
			fmt.Printf("dry-run: would DELETE responses [%s] from form %s\n", strings.Join(responseIDs, ", "), formID)
			return nil
		}
		if err := client.DeleteResponses(formID, responseIDs); err != nil {
			return err
		}
		fmt.Printf("Deleted %d responses from form %s\n", len(responseIDs), formID)
		return nil
	},
}

func init() {
	responseListCmd.Flags().IntVar(&responseListPageSize, "page-size", 25, "Max responses per page (max 1000)")
	responseListCmd.Flags().StringVar(&responseListSince, "since", "", "Filter from date (ISO 8601)")
	responseListCmd.Flags().StringVar(&responseListUntil, "until", "", "Filter until date (ISO 8601)")
	responseListCmd.Flags().StringVar(&responseListAfter, "after", "", "Pagination: after this token")
	responseListCmd.Flags().StringVar(&responseListBefore, "before", "", "Pagination: before this token")
	responseListCmd.Flags().StringVar(&responseListSort, "sort", "", "Sort by field,direction (e.g. submitted_at,desc)")
	responseListCmd.Flags().StringVar(&responseListQuery, "query", "", "Search string matched against all answers")
	responseListCmd.Flags().StringVar(&responseListResponseType, "response-type", "", "Filter: started, partial, completed")
	responseListCmd.Flags().StringVar(&responseListIncludedIDs, "included-ids", "", "Comma-separated response IDs to include")

	responseCmd.AddCommand(responseListCmd, responseDeleteCmd)
	rootCmd.AddCommand(responseCmd)

	RegisterSchema("responses.list", SchemaEntry{
		Command:     "typeform response list <form-id>",
		Description: "List responses for a form (must be in allowed workspace)",
		Args:        []SchemaArg{{Name: "form-id", Required: true, Desc: "Form ID"}},
		Flags: []SchemaFlag{
			{Name: "--page-size", Type: "int", Default: "25", Desc: "Max responses per page (max 1000)"},
			{Name: "--since", Type: "string", Desc: "Filter from date (ISO 8601)"},
			{Name: "--until", Type: "string", Desc: "Filter until date (ISO 8601)"},
			{Name: "--after", Type: "string", Desc: "Pagination cursor: after this token"},
			{Name: "--before", Type: "string", Desc: "Pagination cursor: before this token"},
			{Name: "--sort", Type: "string", Desc: "Sort by field,direction (e.g. submitted_at,desc)"},
			{Name: "--query", Type: "string", Desc: "Search text matched against all answers"},
			{Name: "--response-type", Type: "string", Desc: "Filter: started, partial, completed"},
			{Name: "--included-ids", Type: "string", Desc: "Comma-separated response IDs to include"},
		},
		Examples: []string{
			"typeform response list abc123",
			"typeform response list abc123 --page-size 100 --json",
			"typeform response list abc123 --since 2024-01-01T00:00:00Z",
		},
		Mutating: false,
	})
	RegisterSchema("responses.delete", SchemaEntry{
		Command:     "typeform response delete <form-id> <response-ids...>",
		Description: "Delete responses from a form (blocked in secure mode)",
		Args: []SchemaArg{
			{Name: "form-id", Required: true, Desc: "Form ID"},
			{Name: "response-ids", Required: true, Desc: "Space-separated response tokens"},
		},
		Mutating: true,
	})
}
