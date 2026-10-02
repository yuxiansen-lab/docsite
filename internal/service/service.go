// Package service implements the domain logic: config loading, Markdown
// rendering with table-of-contents extraction, document lookup and search.
package service

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	docsite "docsite"
	"docsite/internal/model"
)

var uidRe = regexp.MustCompile(`[^\p{L}\p{N}-]+`)

// slugify 生成标题锚点 id；前端 markdown.js 使用同一规则。
func slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = uidRe.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if s == "" {
		s = "section"
	}
	return s
}

// 由根包嵌入的资产树派生出各子文件系统。
var (
	configFS fs.FS
	webFS    fs.FS
	docsFS   fs.FS
	assetsFS fs.FS
)

func init() {
	configFS = sub(docsite.FS, "config")
	webFS = sub(docsite.FS, "web")
	docsFS = sub(docsite.FS, "docs")
	assetsFS = sub(docsite.FS, "docs/assets")
}

func sub(fsys fs.FS, dir string) fs.FS {
	s, err := fs.Sub(fsys, dir)
	if err != nil {
		return fsys
	}
	return s
}

// Service bundles all read-only domain data behind a stable API.
type Service struct {
	site     model.SiteConfig
	menu     *model.MenuResponse
	docsRoot fs.FS

	mu    sync.RWMutex
	cache map[string]*cachedDoc
}

type cachedDoc struct {
	resp      model.DocResponse
	updatedAt time.Time
	raw       []byte
}

// New loads site config, the menu tree and primes the document cache.
func New() (*Service, error) {
	var site model.SiteConfig
	siteBytes, err := fs.ReadFile(configFS, "site.json")
	if err != nil {
		return nil, fmt.Errorf("read site config: %w", err)
	}
	if err := json.Unmarshal(siteBytes, &site); err != nil {
		return nil, fmt.Errorf("parse site config: %w", err)
	}

	menuBytes, err := fs.ReadFile(configFS, "menu.json")
	if err != nil {
		return nil, fmt.Errorf("read menu config: %w", err)
	}
	var menu model.MenuResponse
	if err := json.Unmarshal(menuBytes, &menu); err != nil {
		return nil, fmt.Errorf("parse menu config: %w", err)
	}
	sortMenus(menu.Items)

	docsRoot := docsFS

	s := &Service{
		site:     site,
		menu:     &menu,
		docsRoot: docsRoot,
		cache:    make(map[string]*cachedDoc),
	}

	// Prime the cache so startup failures are fatal, not per-request.
	var paths []string
	_ = fs.WalkDir(docsRoot, ".", func(p string, d fs.DirEntry, err error) error {
		if err == nil && !d.IsDir() && strings.HasSuffix(p, ".md") {
			paths = append(paths, p)
		}
		return nil
	})
	for _, p := range paths {
		if _, err := s.getDoc(p); err != nil {
			return nil, fmt.Errorf("render %s: %w", p, err)
		}
	}
	return s, nil
}

// Site returns the site configuration.
func (s *Service) Site() model.SiteConfig { return s.site }

// Menu returns the menu tree.
func (s *Service) Menu() *model.MenuResponse { return s.menu }

// Web returns the embedded static frontend filesystem.
func (s *Service) Web() fs.FS { return webFS }

// Assets returns the embedded docs/assets filesystem.
func (s *Service) Assets() fs.FS { return assetsFS }

// normalizePath rejects path traversal and resolves to a canonical
// slash-separated path inside the docs root.
func normalizePath(p string) (string, bool) {
	p = strings.TrimPrefix(p, "/")
	p = strings.ReplaceAll(p, "\\", "/")
	for _, seg := range strings.Split(p, "/") {
		if seg == ".." || seg == "." || seg == "" {
			return "", false
		}
	}
	if strings.Contains(p, "\x00") {
		return "", false
	}
	if !strings.HasSuffix(p, ".md") {
		p += ".md"
	}
	return p, true
}

func (s *Service) getDoc(normalized string) (*cachedDoc, error) {
	key := normalized
	s.mu.RLock()
	if c, ok := s.cache[key]; ok {
		s.mu.RUnlock()
		return c, nil
	}
	s.mu.RUnlock()

	raw, err := fs.ReadFile(s.docsRoot, key)
	if err != nil {
		return nil, err
	}

	doc := renderMarkdown(string(raw))

	// 标题：取第一个 H1，否则用文件名。
	title := ""
	for _, t := range doc.Toc {
		if t.Level == 1 {
			title = t.Text
			break
		}
	}
	if title == "" {
		title = baseName(key)
	}

	var toc []model.TocItem
	for _, t := range doc.Toc {
		id := t.ID
		if id == "" {
			id = slugify(t.Text)
		}
		toc = append(toc, model.TocItem{
			Level: t.Level,
			Text:  t.Text,
			ID:    id,
		})
	}

	s.mu.Lock()
	c := &cachedDoc{
		resp: model.DocResponse{
			Path:      key,
			Title:     title,
			HTML:      doc.HTML,
			TOC:       toc,
			UpdatedAt: time.Now().UTC().Format(time.RFC3339),
		},
		updatedAt: time.Now(),
		raw:       raw,
	}
	s.cache[key] = c
	s.mu.Unlock()
	return c, nil
}

func baseName(p string) string {
	i := strings.LastIndexByte(p, '/')
	if i < 0 {
		return p
	}
	return p[i+1:]
}

// Doc renders (and caches) the requested document.
func (s *Service) Doc(p string) (model.DocResponse, error) {
	norm, ok := normalizePath(p)
	if !ok {
		return model.DocResponse{}, fs.ErrNotExist
	}
	c, err := s.getDoc(norm)
	if err != nil {
		return model.DocResponse{}, err
	}
	resp := c.resp

	flat := s.docOrder()
	idx := -1
	for i, d := range flat {
		if d.Path == norm {
			idx = i
			break
		}
	}
	if idx > 0 {
		resp.Prev = &model.NavItem{ID: flat[idx-1].ID, Title: flat[idx-1].Title, Path: flat[idx-1].Path}
	}
	if idx >= 0 && idx < len(flat)-1 {
		resp.Next = &model.NavItem{ID: flat[idx+1].ID, Title: flat[idx+1].Title, Path: flat[idx+1].Path}
	}
	return resp, nil
}

// docOrder walks the menu tree (DFS, skipping hidden items and non-doc
// actions) and returns every document in reading order.
func (s *Service) docOrder() []model.NavItem {
	var out []model.NavItem
	var walk func(items []*model.MenuItem)
	walk = func(items []*model.MenuItem) {
		for _, it := range items {
			if it.Hidden {
				continue
			}
			if it.Action == "doc" && it.Path != "" {
				out = append(out, model.NavItem{ID: it.ID, Title: it.Title, Path: it.Path})
			}
			if len(it.Children) > 0 {
				walk(it.Children)
			}
		}
	}
	walk(s.menu.Items)
	return out
}

// Search returns up to 20 documents whose title or body contains the query.
func (s *Service) Search(q string) []model.SearchHit {
	q = strings.ToLower(strings.TrimSpace(q))
	if len(q) < 2 {
		return nil
	}
	s.mu.RLock()
	defer s.mu.RUnlock()

	var hits []model.SearchHit
	for path, c := range s.cache {
		lower := strings.ToLower(string(c.raw))
		idx := strings.Index(lower, q)
		if idx < 0 {
			continue
		}
		start := idx
		if start > 60 {
			start -= 60
		}
		end := idx + len(q) + 80
		if end > len(c.raw) {
			end = len(c.raw)
		}
		snippet := strings.TrimSpace(strings.ReplaceAll(string(c.raw)[start:end], "\n", " "))
		hits = append(hits, model.SearchHit{
			Path:    path,
			Title:   c.resp.Title,
			Snippet: snippet,
		})
	}
	sort.Slice(hits, func(i, j int) bool { return hits[i].Path < hits[j].Path })
	if len(hits) > 20 {
		hits = hits[:20]
	}
	return hits
}

// AllDocs 按菜单阅读顺序返回全部文档（含 prev/next）。
// 服务端的 /api/docs.json 与静态导出共用此方法。
func (s *Service) AllDocs() []model.DocResponse {
	out := make([]model.DocResponse, 0, 16)
	for _, item := range s.docOrder() {
		d, err := s.Doc(item.Path)
		if err != nil {
			continue
		}
		out = append(out, d)
	}
	return out
}

// SearchIndex 返回客户端搜索索引，覆盖磁盘上全部 .md（与 Search 范围一致），
// 使静态托管时无需后端也能全文搜索。
func (s *Service) SearchIndex() []model.SearchDoc {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]model.SearchDoc, 0, len(s.cache))
	for path, c := range s.cache {
		out = append(out, model.SearchDoc{
			Path:  path,
			Title: c.resp.Title,
			Text:  string(c.raw),
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Path < out[j].Path })
	return out
}

func sortMenus(items []*model.MenuItem) {
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].Order != items[j].Order {
			return items[i].Order < items[j].Order
		}
		return items[i].Title < items[j].Title
	})
	for _, it := range items {
		if len(it.Children) > 0 {
			sortMenus(it.Children)
		}
	}
}
