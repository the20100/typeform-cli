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

var webhookCmd = &cobra.Command{
	Use:   "webhook",
	Short: "Manage Typeform webhooks",
}

// ---- webhook list ----

var webhookListCmd = &cobra.Command{
	Use:   "list <form-id>",
	Short: "List webhooks for a form",
	Args:  cobra.ExactArgs(1),
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
		items, err := client.ListWebhooks(args[0])
		if err != nil {
			return err
		}
		if output.IsJSON(cmd) {
			return output.PrintJSON(items, output.IsPretty(cmd))
		}
		printWebhooksTable(items)
		return nil
	},
}

// ---- webhook get ----

var webhookGetCmd = &cobra.Command{
	Use:   "get <form-id> <tag>",
	Short: "Get details of a specific webhook",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := validate.ResourceID(args[0]); err != nil {
			return fmt.Errorf("invalid form ID: %w", err)
		}
		if err := validate.ResourceID(args[1]); err != nil {
			return fmt.Errorf("invalid webhook tag: %w", err)
		}
		if err := ensureConfig(); err != nil {
			return err
		}
		if err := checkFormWorkspace(args[0]); err != nil {
			return err
		}
		wh, err := client.GetWebhook(args[0], args[1])
		if err != nil {
			return err
		}
		if output.IsJSON(cmd) {
			return output.PrintJSON(wh, output.IsPretty(cmd))
		}
		output.PrintKeyValue([][]string{
			{"Tag", wh.Tag},
			{"URL", wh.URL},
			{"Enabled", output.FormatBool(wh.Enabled)},
			{"Created", output.FormatTime(wh.CreatedAt)},
			{"Updated", output.FormatTime(wh.UpdatedAt)},
		})
		return nil
	},
}

// ---- webhook create ----

var (
	webhookCreateURL     string
	webhookCreateEnabled bool
	webhookCreateSecret  string
)

var webhookCreateCmd = &cobra.Command{
	Use:   "create <form-id> <tag>",
	Short: "Create or update a webhook",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		formID := args[0]
		tag := args[1]
		if err := validate.ResourceID(formID); err != nil {
			return fmt.Errorf("invalid form ID: %w", err)
		}
		if err := validate.ResourceID(tag); err != nil {
			return fmt.Errorf("invalid webhook tag: %w", err)
		}
		if err := ensureConfig(); err != nil {
			return err
		}
		if err := config.CheckSecureMode(cfg, "create"); err != nil {
			return err
		}
		if err := checkFormWorkspace(formID); err != nil {
			return err
		}
		if webhookCreateURL == "" {
			return fmt.Errorf("--url is required")
		}

		req := &api.WebhookCreateRequest{
			URL:     webhookCreateURL,
			Enabled: webhookCreateEnabled,
			Secret:  webhookCreateSecret,
		}

		if dryRunFlag {
			data, _ := json.MarshalIndent(req, "", "  ")
			fmt.Printf("dry-run: would PUT /forms/%s/webhooks/%s with:\n%s\n", formID, tag, string(data))
			return nil
		}

		wh, err := client.CreateOrUpdateWebhook(formID, tag, req)
		if err != nil {
			return err
		}
		if output.IsJSON(cmd) {
			return output.PrintJSON(wh, output.IsPretty(cmd))
		}
		fmt.Printf("Created/updated webhook %q for form %s\n", wh.Tag, formID)
		return nil
	},
}

// ---- webhook delete ----

var webhookDeleteCmd = &cobra.Command{
	Use:   "delete <form-id> <tag>",
	Short: "Delete a webhook (blocked in secure mode)",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := validate.ResourceID(args[0]); err != nil {
			return fmt.Errorf("invalid form ID: %w", err)
		}
		if err := validate.ResourceID(args[1]); err != nil {
			return fmt.Errorf("invalid webhook tag: %w", err)
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
			fmt.Printf("dry-run: would DELETE /forms/%s/webhooks/%s\n", args[0], args[1])
			return nil
		}
		if err := client.DeleteWebhook(args[0], args[1]); err != nil {
			return err
		}
		fmt.Printf("Deleted webhook %q from form %s\n", args[1], args[0])
		return nil
	},
}

func init() {
	webhookCreateCmd.Flags().StringVar(&webhookCreateURL, "url", "", "Webhook destination URL (required)")
	webhookCreateCmd.Flags().BoolVar(&webhookCreateEnabled, "enabled", true, "Enable the webhook")
	webhookCreateCmd.Flags().StringVar(&webhookCreateSecret, "secret", "", "Webhook secret for signature verification")

	webhookCmd.AddCommand(webhookListCmd, webhookGetCmd, webhookCreateCmd, webhookDeleteCmd)
	rootCmd.AddCommand(webhookCmd)

	RegisterSchema("webhooks.list", SchemaEntry{
		Command:     "typeform webhook list <form-id>",
		Description: "List webhooks for a form (must be in allowed workspace)",
		Args:        []SchemaArg{{Name: "form-id", Required: true, Desc: "Form ID"}},
		Mutating:    false,
	})
	RegisterSchema("webhooks.get", SchemaEntry{
		Command:     "typeform webhook get <form-id> <tag>",
		Description: "Get details of a specific webhook",
		Args: []SchemaArg{
			{Name: "form-id", Required: true, Desc: "Form ID"},
			{Name: "tag", Required: true, Desc: "Webhook tag"},
		},
		Mutating: false,
	})
	RegisterSchema("webhooks.create", SchemaEntry{
		Command:     "typeform webhook create <form-id> <tag> --url <url>",
		Description: "Create or update a webhook",
		Args: []SchemaArg{
			{Name: "form-id", Required: true, Desc: "Form ID"},
			{Name: "tag", Required: true, Desc: "Webhook tag"},
		},
		Flags: []SchemaFlag{
			{Name: "--url", Type: "string", Required: true, Desc: "Webhook destination URL"},
			{Name: "--enabled", Type: "bool", Default: "true", Desc: "Enable the webhook"},
			{Name: "--secret", Type: "string", Desc: "Webhook secret for signature verification"},
		},
		Mutating: true,
	})
	RegisterSchema("webhooks.delete", SchemaEntry{
		Command:     "typeform webhook delete <form-id> <tag>",
		Description: "Delete a webhook (blocked in secure mode)",
		Args: []SchemaArg{
			{Name: "form-id", Required: true, Desc: "Form ID"},
			{Name: "tag", Required: true, Desc: "Webhook tag"},
		},
		Mutating: true,
	})
}
