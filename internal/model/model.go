// Package model defines the data types shared between handlers and services.
package model

// SiteConfig is the root of config/site.json.
type SiteConfig struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	ThemeColor  string `json:"themeColor"`
	DefaultDoc  string `json:"defaultDoc"`
	LogoText    string `json:"logoText"`
}

// MenuItem is one node of the menu tree (config/menu.json).
//
//   - Action "doc" loads the Markdown document at Path.
//   - Action "link" opens URL in a new tab.
//   - Action "callback" invokes the frontend function named Callback.
//   - Action "toggle" (or no action with Children) just expands/collapses.
type MenuItem struct {
	ID       string      `json:"id"`
	Title    string      `json:"title"`
	Icon     string      `json:"icon,omitempty"`
	Path     string      `json:"path,omitempty"`
	URL      string      `json:"url,omitempty"`
	Callback string      `json:"callback,omitempty"`
	Action   string      `json:"action,omitempty"`
	Order    int         `json:"order,omitempty"`
	Expanded bool        `json:"expanded,omitempty"`
	Hidden   bool        `json:"hidden,omitempty"`
	Children []*MenuItem `json:"children,omitempty"`
}

// MenuResponse wraps the menu tree returned by /api/menu.
type MenuResponse struct {
	Items []*MenuItem `json:"items"`
}

// TocItem is one entry of a document's table of contents.
type TocItem struct {
	Level int    `json:"level"`
	Text  string `json:"text"`
	ID    string `json:"id"`
}

// DocResponse is the payload of /api/doc.
type DocResponse struct {
	Path      string    `json:"path"`
	Title     string    `json:"title"`
	HTML      string    `json:"html"`
	TOC       []TocItem `json:"toc"`
	Prev      *NavItem  `json:"prev"`
	Next      *NavItem  `json:"next"`
	UpdatedAt string    `json:"updatedAt"`
}

// NavItem describes a previous/next document link.
type NavItem struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Path  string `json:"path"`
}

// SearchHit is one search result.
type SearchHit struct {
	Path    string `json:"path"`
	Title   string `json:"title"`
	Snippet string `json:"snippet"`
}
