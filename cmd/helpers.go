package cmd

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/the20100/typeform-cli/internal/api"
	"github.com/the20100/typeform-cli/internal/output"
)

// buildParams creates a url.Values from alternating key/value pairs,
// skipping pairs where the value is empty.
func buildParams(pairs ...string) url.Values {
	p := url.Values{}
	for i := 0; i+1 < len(pairs); i += 2 {
		if pairs[i+1] != "" {
			p.Set(pairs[i], pairs[i+1])
		}
	}
	return p
}

// printFilteredJSON filters JSON output to only include specified fields.
func printFilteredJSON(v any, fields string, pretty bool) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	fieldList := strings.Split(fields, ",")
	for i := range fieldList {
		fieldList[i] = strings.TrimSpace(fieldList[i])
	}

	// Try as array first
	var arr []map[string]any
	if json.Unmarshal(data, &arr) == nil {
		filtered := make([]map[string]any, len(arr))
		for i, item := range arr {
			filtered[i] = filterFields(item, fieldList)
		}
		return output.PrintJSON(filtered, pretty)
	}

	// Try as single object
	var obj map[string]any
	if json.Unmarshal(data, &obj) == nil {
		return output.PrintJSON(filterFields(obj, fieldList), pretty)
	}

	// Fallback: print as-is
	return output.PrintJSON(v, pretty)
}

func filterFields(obj map[string]any, fields []string) map[string]any {
	result := make(map[string]any)
	for _, f := range fields {
		if v, ok := obj[f]; ok {
			result[f] = v
		}
	}
	return result
}

func printFormsTable(items []api.Form) {
	if len(items) == 0 {
		fmt.Println("No forms found.")
		return
	}
	headers := []string{"ID", "TITLE", "LAST UPDATED"}
	rows := make([][]string, len(items))
	for i, item := range items {
		rows[i] = []string{
			item.ID,
			output.Truncate(item.Title, 44),
			output.FormatTime(item.LastUpdated),
		}
	}
	output.PrintTable(headers, rows)
}

func printWorkspacesTable(items []api.Workspace) {
	if len(items) == 0 {
		fmt.Println("No workspaces found.")
		return
	}
	headers := []string{"ID", "NAME", "DEFAULT", "SHARED"}
	rows := make([][]string, len(items))
	for i, item := range items {
		rows[i] = []string{
			item.ID,
			output.Truncate(item.Name, 44),
			output.FormatBool(item.Default),
			output.FormatBool(item.Shared),
		}
	}
	output.PrintTable(headers, rows)
}

func printThemesTable(items []api.Theme) {
	if len(items) == 0 {
		fmt.Println("No themes found.")
		return
	}
	headers := []string{"ID", "NAME", "VISIBILITY"}
	rows := make([][]string, len(items))
	for i, item := range items {
		rows[i] = []string{
			item.ID,
			output.Truncate(item.Name, 44),
			item.Visibility,
		}
	}
	output.PrintTable(headers, rows)
}

func printImagesTable(items []api.Image) {
	if len(items) == 0 {
		fmt.Println("No images found.")
		return
	}
	headers := []string{"ID", "FILENAME", "MEDIA TYPE"}
	rows := make([][]string, len(items))
	for i, item := range items {
		rows[i] = []string{
			item.ID,
			output.Truncate(item.FileName, 44),
			item.MediaType,
		}
	}
	output.PrintTable(headers, rows)
}

func printWebhooksTable(items []api.Webhook) {
	if len(items) == 0 {
		fmt.Println("No webhooks found.")
		return
	}
	headers := []string{"TAG", "URL", "ENABLED"}
	rows := make([][]string, len(items))
	for i, item := range items {
		rows[i] = []string{
			item.Tag,
			output.Truncate(item.URL, 60),
			output.FormatBool(item.Enabled),
		}
	}
	output.PrintTable(headers, rows)
}
