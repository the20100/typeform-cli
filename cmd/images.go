package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/the20100/typeform-cli/internal/config"
	"github.com/the20100/typeform-cli/internal/output"
	"github.com/the20100/typeform-cli/internal/validate"
)

var imageCmd = &cobra.Command{
	Use:   "image",
	Short: "Manage Typeform images",
}

// ---- image list ----

var imageListCmd = &cobra.Command{
	Use:   "list",
	Short: "List images",
	RunE: func(cmd *cobra.Command, args []string) error {
		items, err := client.ListImages()
		if err != nil {
			return err
		}
		if output.IsJSON(cmd) {
			if fieldsFlag != "" {
				return printFilteredJSON(items, fieldsFlag, output.IsPretty(cmd))
			}
			return output.PrintJSON(items, output.IsPretty(cmd))
		}
		printImagesTable(items)
		return nil
	},
}

// ---- image get ----

var imageGetCmd = &cobra.Command{
	Use:   "get <image-id>",
	Short: "Get details of a specific image",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := validate.ResourceID(args[0]); err != nil {
			return fmt.Errorf("invalid image ID: %w", err)
		}
		img, err := client.GetImage(args[0])
		if err != nil {
			return err
		}
		if output.IsJSON(cmd) {
			return output.PrintJSON(img, output.IsPretty(cmd))
		}
		output.PrintKeyValue([][]string{
			{"ID", img.ID},
			{"Filename", img.FileName},
			{"Media Type", img.MediaType},
			{"Size", fmt.Sprintf("%dx%d", img.Width, img.Height)},
			{"Src", img.Src},
		})
		return nil
	},
}

// ---- image delete ----

var imageDeleteCmd = &cobra.Command{
	Use:   "delete <image-id>",
	Short: "Delete an image (blocked in secure mode)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := validate.ResourceID(args[0]); err != nil {
			return fmt.Errorf("invalid image ID: %w", err)
		}
		if err := ensureConfig(); err != nil {
			return err
		}
		if err := config.CheckSecureMode(cfg, "delete"); err != nil {
			return err
		}
		if dryRunFlag {
			fmt.Printf("dry-run: would DELETE /images/%s\n", args[0])
			return nil
		}
		if err := client.DeleteImage(args[0]); err != nil {
			return err
		}
		fmt.Printf("Deleted image %s\n", args[0])
		return nil
	},
}

func init() {
	imageCmd.AddCommand(imageListCmd, imageGetCmd, imageDeleteCmd)
	rootCmd.AddCommand(imageCmd)

	RegisterSchema("images.list", SchemaEntry{
		Command:     "typeform image list",
		Description: "List images",
		Mutating:    false,
	})
	RegisterSchema("images.get", SchemaEntry{
		Command:     "typeform image get <image-id>",
		Description: "Get details of a specific image",
		Args:        []SchemaArg{{Name: "image-id", Required: true, Desc: "Image ID"}},
		Mutating:    false,
	})
	RegisterSchema("images.delete", SchemaEntry{
		Command:     "typeform image delete <image-id>",
		Description: "Delete an image (blocked in secure mode)",
		Args:        []SchemaArg{{Name: "image-id", Required: true, Desc: "Image ID"}},
		Mutating:    true,
	})
}
