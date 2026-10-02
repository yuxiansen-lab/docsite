package service

import (
	"strings"
	"testing"
)

func TestRenderHeadingsAndTOC(t *testing.T) {
	src := "# 标题\n\n## 小节\n\ntext\n"
	doc := renderMarkdown(src)

	if len(doc.Toc) != 2 {
		t.Fatalf("want 2 toc entries, got %d", len(doc.Toc))
	}
	if doc.Toc[0].Level != 1 || doc.Toc[0].Text != "标题" {
		t.Errorf("unexpected first toc entry: %+v", doc.Toc[0])
	}
	if !strings.Contains(doc.HTML, `<h1 id="标题">标题</h1>`) {
		t.Errorf("h1 not rendered with id: %s", doc.HTML)
	}
	if !strings.Contains(doc.HTML, `<h2 id="小节">小节</h2>`) {
		t.Errorf("h2 not rendered with id: %s", doc.HTML)
	}
}

func TestDuplicateHeadingsGetUniqueIDs(t *testing.T) {
	src := "## 重复\n\n## 重复\n\n## 重复\n"
	doc := renderMarkdown(src)

	ids := []string{doc.Toc[0].ID, doc.Toc[1].ID, doc.Toc[2].ID}
	want := []string{"重复", "重复-1", "重复-2"}
	for i := range want {
		if ids[i] != want[i] {
			t.Errorf("heading %d: want id %q, got %q", i, want[i], ids[i])
		}
	}
	// HTML 中的 id 必须与 TOC 一致
	for _, id := range want {
		if !strings.Contains(doc.HTML, `id="`+id+`"`) {
			t.Errorf("HTML missing id %q: %s", id, doc.HTML)
		}
	}
}

func TestFencedCodeBlock(t *testing.T) {
	src := "```go\nfunc main() {}\n```\n"
	doc := renderMarkdown(src)

	if !strings.Contains(doc.HTML, `class="language-go"`) {
		t.Errorf("missing language class: %s", doc.HTML)
	}
	if !strings.Contains(doc.HTML, "codeblock__copy") {
		t.Errorf("missing copy button: %s", doc.HTML)
	}
	// 代码内容必须被 HTML 转义，不能原样注入
	if !strings.Contains(doc.HTML, "func main() {}") {
		t.Errorf("code body lost: %s", doc.HTML)
	}
}

func TestCodeBodyIsEscaped(t *testing.T) {
	src := "```html\n<script>alert(1)</script>\n```\n"
	doc := renderMarkdown(src)

	if strings.Contains(doc.HTML, "<script>") {
		t.Errorf("raw <script> leaked into output: %s", doc.HTML)
	}
	if !strings.Contains(doc.HTML, "&lt;script&gt;") {
		t.Errorf("code body was not escaped: %s", doc.HTML)
	}
}

func TestInlineCodeIsNotFormatted(t *testing.T) {
	src := "使用 `a*b*c` 与 **粗体**。\n"
	doc := renderMarkdown(src)

	if !strings.Contains(doc.HTML, "<code>a*b*c</code>") {
		t.Errorf("inline code mangled: %s", doc.HTML)
	}
	if !strings.Contains(doc.HTML, "<strong>粗体</strong>") {
		t.Errorf("bold not rendered: %s", doc.HTML)
	}
}

func TestTable(t *testing.T) {
	src := "| A | B |\n| --- | ---: |\n| 1 | 2 |\n"
	doc := renderMarkdown(src)

	for _, want := range []string{"<table>", "<th>A</th>", "<td>1</td>", `align="right"`} {
		if !strings.Contains(doc.HTML, want) {
			t.Errorf("table missing %q: %s", want, doc.HTML)
		}
	}
}

func TestTaskList(t *testing.T) {
	src := "- [x] done\n- [ ] todo\n"
	doc := renderMarkdown(src)

	if !strings.Contains(doc.HTML, `class="task-list-item"`) {
		t.Errorf("missing task-list-item class: %s", doc.HTML)
	}
	if !strings.Contains(doc.HTML, "checked") {
		t.Errorf("checked state lost: %s", doc.HTML)
	}
}

func TestParagraphEscapesHTML(t *testing.T) {
	doc := renderMarkdown("普通文本 <img src=x onerror=alert(1)>\n")
	if strings.Contains(doc.HTML, "<img src=x") {
		t.Errorf("raw HTML injected from paragraph: %s", doc.HTML)
	}
}

func TestNormalizePath(t *testing.T) {
	cases := []struct {
		in      string
		want    string
		wantOK  bool
		comment string
	}{
		{"guide/intro.md", "guide/intro.md", true, "normal"},
		{"guide/intro", "guide/intro.md", true, "extension added"},
		{"../config/site.json", "", false, "parent traversal"},
		{"guide/../../etc/passwd", "", false, "nested traversal"},
		{"a//b.md", "", false, "empty segment"},
		{"..\\windows\\win.ini", "", false, "backslash traversal"},
		// 前导 / 会被剥离，结果仍被限制在 docs/ 内（读取时自然 404）
		{"/etc/passwd", "etc/passwd.md", true, "absolute path stays inside docs"},
	}
	for _, c := range cases {
		got, ok := normalizePath(c.in)
		if ok != c.wantOK {
			t.Errorf("%s: normalizePath(%q) ok=%v, want %v", c.comment, c.in, ok, c.wantOK)
			continue
		}
		if ok && got != c.want {
			t.Errorf("%s: normalizePath(%q)=%q, want %q", c.comment, c.in, got, c.want)
		}
	}
}

// TestNormalizePathCannotEscape 是不变量测试：任何输入都不得解析出绝对
// 路径，也不得出现 ".." 路径段，否则就能读出 docs/ 之外的文件。
// 注意：像 "..%2Fsecret.md" 这样的整体文件名不含独立 ".." 段，只是
// 一个普通文件名，本身无法穿越目录（HTTP 层已先做百分号解码）。
func TestNormalizePathCannotEscape(t *testing.T) {
	hostile := []string{
		"../secret.md", "../../etc/passwd", "..%2Fsecret.md",
		"a/../../b.md", "....//secret.md", "\\..\\secret.md",
		"C:/Windows/win.ini", "/etc/passwd", "//server/share/x.md",
		"guide/./../../x.md", "a/..", "..",
	}
	for _, in := range hostile {
		got, ok := normalizePath(in)
		if !ok {
			continue // 直接拒绝，安全
		}
		if strings.HasPrefix(got, "/") || strings.Contains(got, "\\") {
			t.Errorf("normalizePath(%q)=%q is not a relative slash path", in, got)
			continue
		}
		for _, seg := range strings.Split(got, "/") {
			if seg == ".." {
				t.Errorf("normalizePath(%q)=%q contains a parent segment", in, got)
				break
			}
		}
	}
}

func TestSlugify(t *testing.T) {
	cases := map[string]string{
		"介绍":            "介绍",
		"从源码运行（开发）":     "从源码运行-开发",
		"Hello, World!": "hello-world",
		"***":           "section",
	}
	for in, want := range cases {
		if got := slugify(in); got != want {
			t.Errorf("slugify(%q)=%q, want %q", in, got, want)
		}
	}
}

func TestServiceLoadsEmbeddedConfigAndDocs(t *testing.T) {
	s, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	if s.Site().Title == "" {
		t.Error("site title is empty")
	}
	if len(s.Menu().Items) == 0 {
		t.Fatal("menu is empty")
	}

	// 菜单里声明的文档都应能渲染
	for _, item := range s.docOrder() {
		if _, err := s.Doc(item.Path); err != nil {
			t.Errorf("Doc(%q) failed: %v", item.Path, err)
		}
	}
}

func TestDocNotFound(t *testing.T) {
	s, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	if _, err := s.Doc("does/not/exist.md"); err == nil {
		t.Error("expected error for missing document")
	}
}

// TestSiteRelative 保证文档内绝对资源路径被改写为相对路径，
// 否则部署在 GitHub Pages 子路径下图片会 404。
func TestSiteRelative(t *testing.T) {
	cases := map[string]string{
		"/docs/assets/a.svg":        "docs/assets/a.svg",
		"/docs/assets/sub/b.png":    "docs/assets/sub/b.png",
		"docs/assets/a.svg":         "docs/assets/a.svg",
		"https://example.com/a.png": "https://example.com/a.png",
		"/other/x.png":              "/other/x.png",
	}
	for in, want := range cases {
		if got := siteRelative(in); got != want {
			t.Errorf("siteRelative(%q)=%q, want %q", in, got, want)
		}
	}
}

func TestImageSrcIsRelative(t *testing.T) {
	doc := renderMarkdown("![图](/docs/assets/architecture.svg)\n")
	if !strings.Contains(doc.HTML, `src="docs/assets/architecture.svg"`) {
		t.Errorf("image src not made relative: %s", doc.HTML)
	}
	if strings.Contains(doc.HTML, `src="/docs/assets/`) {
		t.Errorf("image src still absolute: %s", doc.HTML)
	}
}

// TestAllDocsAndSearchIndex 覆盖静态导出所依赖的两个聚合接口。
func TestAllDocsAndSearchIndex(t *testing.T) {
	s, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	docs := s.AllDocs()
	if len(docs) == 0 {
		t.Fatal("AllDocs returned nothing")
	}
	// 菜单顺序：第一篇应为首页，且首篇没有上一篇
	if docs[0].Path != "index.md" {
		t.Errorf("first doc = %q, want index.md", docs[0].Path)
	}
	if docs[0].Prev != nil {
		t.Error("first doc should not have a prev")
	}
	// 相邻文档应互为上/下篇
	if len(docs) > 1 {
		if docs[0].Next == nil || docs[0].Next.Path != docs[1].Path {
			t.Errorf("docs[0].Next = %+v, want %q", docs[0].Next, docs[1].Path)
		}
		if docs[1].Prev == nil || docs[1].Prev.Path != docs[0].Path {
			t.Errorf("docs[1].Prev = %+v, want %q", docs[1].Prev, docs[0].Path)
		}
	}

	idx := s.SearchIndex()
	if len(idx) == 0 {
		t.Fatal("SearchIndex returned nothing")
	}
	for _, it := range idx {
		if it.Path == "" || it.Title == "" || it.Text == "" {
			t.Errorf("incomplete search entry: %+v", it)
		}
	}
}
