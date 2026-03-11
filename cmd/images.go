package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/the20100/typeform-cli/internal/api"
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

// ---- image upload ----

var (
	imageUploadFileName string
)

var imageUploadCmd = &cobra.Command{
	Use:   "upload <url-or-file>",
	Short: "Upload an image from a URL or local file",
	Long: `Upload an image to Typeform from a public URL or local file path.
The uploaded image gets a Typeform-hosted URL that can be used in form fields,
welcome screens, and thank-you screens via the attachment.href property.

Supported formats: JPEG, PNG, GIF, BMP. WebP is NOT supported by Typeform.

Examples:
  typeform image upload https://example.com/photo.png
  typeform image upload ./logo.png
  typeform image upload https://example.com/photo.png --file-name "my-logo.png"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		source := args[0]
		if dryRunFlag {
			fmt.Printf("dry-run: would POST /images (source: %s)\n", source)
			return nil
		}

		var (
			img *api.Image
			err error
		)

		if strings.HasPrefix(source, "http://") || strings.HasPrefix(source, "https://") {
			img, err = client.UploadImageFromURL(source, imageUploadFileName)
		} else {
			img, err = client.UploadImageFromFile(source)
		}
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

func init() {
	imageUploadCmd.Flags().StringVar(&imageUploadFileName, "file-name", "", "Override the file name stored in Typeform")

	imageCmd.AddCommand(imageListCmd, imageGetCmd, imageUploadCmd, imageDeleteCmd)
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
	RegisterSchema("images.upload", SchemaEntry{
		Command:     "typeform image upload <url-or-file>",
		Description: "Upload an image from a URL or local file. Returns the Typeform-hosted URL to use in form attachments.",
		Args:        []SchemaArg{{Name: "url-or-file", Required: true, Desc: "Public image URL or local file path"}},
		Flags:       []SchemaFlag{{Name: "--file-name", Type: "string", Desc: "Override the stored file name"}},
		Examples: []string{
			"typeform image upload https://example.com/logo.png",
			"typeform image upload ./photo.jpg --file-name brand-photo.jpg",
			"typeform image upload https://cdn.example.com/hero.png --pretty",
		},
		Mutating: true,
	})
	RegisterSchema("images.delete", SchemaEntry{
		Command:     "typeform image delete <image-id>",
		Description: "Delete an image (blocked in secure mode)",
		Args:        []SchemaArg{{Name: "image-id", Required: true, Desc: "Image ID"}},
		Mutating:    true,
	})
}
