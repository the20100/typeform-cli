package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const apiBase = "https://api.typeform.com"

type Client struct {
	token      string
	httpClient *http.Client
}

func NewClient(token string) *Client {
	return &Client{
		token:      token,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *Client) doRequest(req *http.Request) ([]byte, error) {
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode >= 400 {
		var apiErr TypeformError
		if json.Unmarshal(body, &apiErr) == nil && apiErr.Description != "" {
			apiErr.StatusCode = resp.StatusCode
			return nil, &apiErr
		}
		return nil, &TypeformError{
			StatusCode:  resp.StatusCode,
			Description: fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(body)),
		}
	}
	return body, nil
}

func (c *Client) buildURL(path string, params url.Values) string {
	u, _ := url.Parse(apiBase + path)
	if params != nil && len(params) > 0 {
		u.RawQuery = params.Encode()
	}
	return u.String()
}

func (c *Client) Get(path string, params url.Values) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, c.buildURL(path, params), nil)
	if err != nil {
		return nil, err
	}
	return c.doRequest(req)
}

func (c *Client) Post(path string, params url.Values, payload any) ([]byte, error) {
	var body io.Reader
	if payload != nil {
		data, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("encoding request: %w", err)
		}
		body = bytes.NewReader(data)
	}
	req, err := http.NewRequest(http.MethodPost, c.buildURL(path, params), body)
	if err != nil {
		return nil, err
	}
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return c.doRequest(req)
}

func (c *Client) Put(path string, params url.Values, payload any) ([]byte, error) {
	var body io.Reader
	if payload != nil {
		data, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("encoding request: %w", err)
		}
		body = bytes.NewReader(data)
	}
	req, err := http.NewRequest(http.MethodPut, c.buildURL(path, params), body)
	if err != nil {
		return nil, err
	}
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return c.doRequest(req)
}

func (c *Client) Patch(path string, params url.Values, payload any) ([]byte, error) {
	var body io.Reader
	if payload != nil {
		data, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("encoding request: %w", err)
		}
		body = bytes.NewReader(data)
	}
	req, err := http.NewRequest(http.MethodPatch, c.buildURL(path, params), body)
	if err != nil {
		return nil, err
	}
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return c.doRequest(req)
}

func (c *Client) Delete(path string, params url.Values) ([]byte, error) {
	req, err := http.NewRequest(http.MethodDelete, c.buildURL(path, params), nil)
	if err != nil {
		return nil, err
	}
	return c.doRequest(req)
}

// ---- Workspace methods ----

func (c *Client) ListWorkspaces(params url.Values) ([]Workspace, error) {
	body, err := c.Get("/workspaces", params)
	if err != nil {
		return nil, err
	}
	var resp struct {
		Items []Workspace `json:"items"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}
	return resp.Items, nil
}

func (c *Client) GetWorkspace(id string) (*Workspace, error) {
	body, err := c.Get("/workspaces/"+id, nil)
	if err != nil {
		return nil, err
	}
	var ws Workspace
	return &ws, json.Unmarshal(body, &ws)
}

func (c *Client) CreateWorkspace(name string) (*Workspace, error) {
	body, err := c.Post("/workspaces", nil, map[string]string{"name": name})
	if err != nil {
		return nil, err
	}
	var ws Workspace
	return &ws, json.Unmarshal(body, &ws)
}

func (c *Client) UpdateWorkspace(id string, patches []WorkspacePatch) (*Workspace, error) {
	body, err := c.Patch("/workspaces/"+id, nil, patches)
	if err != nil {
		return nil, err
	}
	var ws Workspace
	return &ws, json.Unmarshal(body, &ws)
}

func (c *Client) DeleteWorkspace(id string) error {
	_, err := c.Delete("/workspaces/"+id, nil)
	return err
}

// ---- Form methods ----

func (c *Client) ListForms(params url.Values) ([]Form, error) {
	body, err := c.Get("/forms", params)
	if err != nil {
		return nil, err
	}
	var resp struct {
		TotalItems int    `json:"total_items"`
		PageCount  int    `json:"page_count"`
		Items      []Form `json:"items"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}
	return resp.Items, nil
}

func (c *Client) GetForm(id string) (*Form, error) {
	body, err := c.Get("/forms/"+id, nil)
	if err != nil {
		return nil, err
	}
	var f Form
	return &f, json.Unmarshal(body, &f)
}

func (c *Client) CreateForm(req *FormCreateRequest) (*Form, error) {
	if req.WorkspaceHref != "" {
		req.Workspace = &ThemeRef{Href: req.WorkspaceHref}
	}
	body, err := c.Post("/forms", nil, req)
	if err != nil {
		return nil, err
	}
	var f Form
	return &f, json.Unmarshal(body, &f)
}

func (c *Client) UpdateForm(id string, payload any) (*Form, error) {
	body, err := c.Put("/forms/"+id, nil, payload)
	if err != nil {
		return nil, err
	}
	var f Form
	return &f, json.Unmarshal(body, &f)
}

func (c *Client) PatchForm(id string, payload any) (*Form, error) {
	body, err := c.Patch("/forms/"+id, nil, payload)
	if err != nil {
		return nil, err
	}
	var f Form
	return &f, json.Unmarshal(body, &f)
}

func (c *Client) DeleteForm(id string) error {
	_, err := c.Delete("/forms/"+id, nil)
	return err
}

// ---- Response methods ----

func (c *Client) ListResponses(formID string, params url.Values) (*ResponseList, error) {
	body, err := c.Get("/forms/"+formID+"/responses", params)
	if err != nil {
		return nil, err
	}
	var resp ResponseList
	return &resp, json.Unmarshal(body, &resp)
}

func (c *Client) DeleteResponses(formID string, responseIDs []string) error {
	params := url.Values{}
	for _, id := range responseIDs {
		params.Add("included_tokens", id)
	}
	_, err := c.Delete("/forms/"+formID+"/responses", params)
	return err
}

// ---- Theme methods ----

func (c *Client) ListThemes(params url.Values) ([]Theme, error) {
	body, err := c.Get("/themes", params)
	if err != nil {
		return nil, err
	}
	var resp struct {
		TotalItems int     `json:"total_items"`
		PageCount  int     `json:"page_count"`
		Items      []Theme `json:"items"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}
	return resp.Items, nil
}

func (c *Client) GetTheme(id string) (*Theme, error) {
	body, err := c.Get("/themes/"+id, nil)
	if err != nil {
		return nil, err
	}
	var t Theme
	return &t, json.Unmarshal(body, &t)
}

func (c *Client) CreateTheme(payload any) (*Theme, error) {
	body, err := c.Post("/themes", nil, payload)
	if err != nil {
		return nil, err
	}
	var t Theme
	return &t, json.Unmarshal(body, &t)
}

func (c *Client) UpdateTheme(id string, payload any) (*Theme, error) {
	body, err := c.Put("/themes/"+id, nil, payload)
	if err != nil {
		return nil, err
	}
	var t Theme
	return &t, json.Unmarshal(body, &t)
}

func (c *Client) DeleteTheme(id string) error {
	_, err := c.Delete("/themes/"+id, nil)
	return err
}

// ---- Image methods ----

func (c *Client) ListImages() ([]Image, error) {
	body, err := c.Get("/images", nil)
	if err != nil {
		return nil, err
	}
	var images []Image
	return images, json.Unmarshal(body, &images)
}

func (c *Client) GetImage(id string) (*Image, error) {
	body, err := c.Get("/images/"+id, nil)
	if err != nil {
		return nil, err
	}
	var img Image
	return &img, json.Unmarshal(body, &img)
}

func (c *Client) DeleteImage(id string) error {
	_, err := c.Delete("/images/"+id, nil)
	return err
}

// ---- Webhook methods ----

func (c *Client) ListWebhooks(formID string) ([]Webhook, error) {
	body, err := c.Get("/forms/"+formID+"/webhooks", nil)
	if err != nil {
		return nil, err
	}
	var resp struct {
		Items []Webhook `json:"items"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}
	return resp.Items, nil
}

func (c *Client) GetWebhook(formID, tag string) (*Webhook, error) {
	body, err := c.Get("/forms/"+formID+"/webhooks/"+tag, nil)
	if err != nil {
		return nil, err
	}
	var wh Webhook
	return &wh, json.Unmarshal(body, &wh)
}

func (c *Client) CreateOrUpdateWebhook(formID, tag string, req *WebhookCreateRequest) (*Webhook, error) {
	body, err := c.Put("/forms/"+formID+"/webhooks/"+tag, nil, req)
	if err != nil {
		return nil, err
	}
	var wh Webhook
	return &wh, json.Unmarshal(body, &wh)
}

func (c *Client) DeleteWebhook(formID, tag string) error {
	_, err := c.Delete("/forms/"+formID+"/webhooks/"+tag, nil)
	return err
}
