package api

// TypeformError is returned when the API responds with an error.
type TypeformError struct {
	StatusCode  int    `json:"status_code"`
	Code        string `json:"code"`
	Description string `json:"description"`
}

func (e *TypeformError) Error() string {
	if e.Description != "" {
		return e.Description
	}
	return e.Code
}

// --- Workspace ---

type Workspace struct {
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	Default   bool            `json:"default"`
	Shared    bool            `json:"shared"`
	Forms     *WorkspaceRef   `json:"forms,omitempty"`
	Self      *WorkspaceRef   `json:"self,omitempty"`
	Members   []WorkspaceMember `json:"members,omitempty"`
	AccountID string          `json:"account_id,omitempty"`
}

type WorkspaceRef struct {
	Count int    `json:"count,omitempty"`
	Href  string `json:"href,omitempty"`
}

type WorkspaceMember struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

type WorkspacePatch struct {
	Op    string `json:"op"`
	Path  string `json:"path"`
	Value any    `json:"value,omitempty"`
}

// --- Form ---

type Form struct {
	ID          string        `json:"id"`
	Title       string        `json:"title"`
	Type        string        `json:"type,omitempty"`
	LastUpdated string        `json:"last_updated_at,omitempty"`
	CreatedAt   string        `json:"created_at,omitempty"`
	Settings    *FormSettings `json:"settings,omitempty"`
	Theme       *ThemeRef     `json:"theme,omitempty"`
	Workspace   *ThemeRef     `json:"workspace,omitempty"`
	Fields      []Field       `json:"fields,omitempty"`
	Links       *FormLinks    `json:"_links,omitempty"`
	Self        *FormSelf     `json:"self,omitempty"`
}

type FormSettings struct {
	Language           string `json:"language,omitempty"`
	IsPublic           bool   `json:"is_public,omitempty"`
	ProgressBar        string `json:"progress_bar,omitempty"`
	ShowProgressBar    bool   `json:"show_progress_bar,omitempty"`
	IsTrial            bool   `json:"is_trial,omitempty"`
	ShowTypeformBranding bool `json:"show_typeform_branding,omitempty"`
}

type ThemeRef struct {
	Href string `json:"href,omitempty"`
}

type FormLinks struct {
	Display string `json:"display,omitempty"`
}

type FormSelf struct {
	Href string `json:"href,omitempty"`
}

type Field struct {
	ID          string      `json:"id,omitempty"`
	Ref         string      `json:"ref,omitempty"`
	Title       string      `json:"title"`
	Type        string      `json:"type"`
	Properties  any         `json:"properties,omitempty"`
	Validations any         `json:"validations,omitempty"`
}

type FormCreateRequest struct {
	Title         string      `json:"title"`
	Type          string      `json:"type,omitempty"`
	WorkspaceHref string      `json:"-"` // used to set workspace.href
	Workspace     *ThemeRef   `json:"workspace,omitempty"`
	Settings      *FormSettings `json:"settings,omitempty"`
	Fields        []Field     `json:"fields,omitempty"`
}

// --- Response ---

type ResponseList struct {
	TotalItems int              `json:"total_items"`
	PageCount  int              `json:"page_count"`
	Items      []FormResponse   `json:"items"`
}

type FormResponse struct {
	ResponseID  string            `json:"response_id"`
	LandingID   string            `json:"landing_id"`
	Token       string            `json:"token"`
	LandedAt    string            `json:"landed_at"`
	SubmittedAt string            `json:"submitted_at"`
	Metadata    *ResponseMetadata `json:"metadata,omitempty"`
	Answers     []Answer          `json:"answers,omitempty"`
	Calculated  *Calculated       `json:"calculated,omitempty"`
	Hidden      map[string]string `json:"hidden,omitempty"`
	Variables   []Variable        `json:"variables,omitempty"`
}

type ResponseMetadata struct {
	UserAgent string `json:"user_agent,omitempty"`
	Platform  string `json:"platform,omitempty"`
	Referer   string `json:"referer,omitempty"`
	NetworkID string `json:"network_id,omitempty"`
	Browser   string `json:"browser,omitempty"`
}

type Answer struct {
	Field   AnswerField `json:"field"`
	Type    string      `json:"type"`
	Text    string      `json:"text,omitempty"`
	Email   string      `json:"email,omitempty"`
	Number  float64     `json:"number,omitempty"`
	Boolean bool        `json:"boolean,omitempty"`
	Date    string      `json:"date,omitempty"`
	URL     string      `json:"url,omitempty"`
	Choice  *Choice     `json:"choice,omitempty"`
	Choices *Choices    `json:"choices,omitempty"`
}

type AnswerField struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	Ref  string `json:"ref,omitempty"`
}

type Choice struct {
	ID    string `json:"id,omitempty"`
	Label string `json:"label,omitempty"`
	Ref   string `json:"ref,omitempty"`
}

type Choices struct {
	IDs    []string `json:"ids,omitempty"`
	Labels []string `json:"labels,omitempty"`
	Refs   []string `json:"refs,omitempty"`
}

type Calculated struct {
	Score int `json:"score,omitempty"`
}

type Variable struct {
	Key    string  `json:"key"`
	Type   string  `json:"type"`
	Number float64 `json:"number,omitempty"`
	Text   string  `json:"text,omitempty"`
}

// --- Theme ---

type Theme struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Visibility  string          `json:"visibility,omitempty"`
	Font        string          `json:"font,omitempty"`
	HasTransparentButton bool   `json:"has_transparent_button,omitempty"`
	Colors      *ThemeColors    `json:"colors,omitempty"`
	Background  *ThemeBackground `json:"background,omitempty"`
}

type ThemeColors struct {
	Question   string `json:"question,omitempty"`
	Answer     string `json:"answer,omitempty"`
	Button     string `json:"button,omitempty"`
	Background string `json:"background,omitempty"`
}

type ThemeBackground struct {
	Href       string `json:"href,omitempty"`
	Layout     string `json:"layout,omitempty"`
	Brightness int    `json:"brightness,omitempty"`
}

// --- Image ---

type Image struct {
	ID       string `json:"id"`
	Src      string `json:"src"`
	FileName string `json:"file_name"`
	Width    int    `json:"width,omitempty"`
	Height   int    `json:"height,omitempty"`
	MediaType string `json:"media_type,omitempty"`
	HasAlpha bool   `json:"has_alpha,omitempty"`
	AvgColor string `json:"avg_color,omitempty"`
}

// --- Webhook ---

type Webhook struct {
	ID        string `json:"id"`
	FormID    string `json:"form_id"`
	Tag       string `json:"tag"`
	URL       string `json:"url"`
	Enabled   bool   `json:"enabled"`
	Secret    string `json:"secret,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`
}

type WebhookCreateRequest struct {
	URL     string `json:"url"`
	Enabled bool   `json:"enabled"`
	Secret  string `json:"secret,omitempty"`
}
