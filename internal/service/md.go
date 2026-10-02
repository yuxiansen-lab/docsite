// md.go — 一个零依赖、基于标准库的 Markdown 渲染器。
//
// 支持：ATX 标题、段落、无序/有序列表、任务列表、引用、
// 围栏代码块（``` 与 ~~~）、表格、水平线、图片、内联链接、
// 行内代码、粗体/斜体。
//
// 设计目标：语义清晰、输出稳定的 HTML（配合 content.css 使用），
// 并为标题提取 TOC 锚点。
package service

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// mdDoc 是渲染中间产物，携带标题元信息。
type mdDoc struct {
	HTML string
	Toc  []tocEntry
}

type tocEntry struct {
	Level int
	Text  string
	ID    string
}

// uniqueID 生成标题锚点；同名标题依次追加 -1、-2… 保证唯一。
func uniqueID(text string, seen map[string]int) string {
	id := slugify(text)
	if n, ok := seen[id]; ok {
		seen[id] = n + 1
		return fmt.Sprintf("%s-%d", id, n+1)
	}
	seen[id] = 0
	return id
}

var (
	reBold     = regexp.MustCompile(`\*\*(.+?)\*\*|__(.+?)__`)
	reItalic   = regexp.MustCompile(`\*(\S(?:.*?\S)?)\*|_(\S(?:.*?\S)?)_`)
	reCode     = regexp.MustCompile("`([^`]+)`")
	reCodeSlot = regexp.MustCompile("\x00(\\d+)\x00")
	reLink     = regexp.MustCompile(`\[([^\]]+)\]\(([^)\s]+)(?:\s+["'](.*?)["'])?\)`)
	reImg      = regexp.MustCompile(`!\[([^\]]*)\]\(([^)\s]+)(?:\s+["'](.*?)["'])?\)`)
	reAutolink = regexp.MustCompile(`&lt;(https?://[^\s&]+)&gt;`)
	reFence    = regexp.MustCompile("^(?:```|~~~)\\s*([\\w-]*)\\s*$")
)

// renderMarkdown 将整篇 Markdown 文本转换为 HTML 字符串并提取 TOC。
func renderMarkdown(src string) mdDoc {
	return renderWith(src, map[string]int{})
}

// renderWith 在给定的锚点去重表下渲染；引用块复用同一张表，
// 保证整篇文档的 heading id 全局唯一。
func renderWith(src string, seen map[string]int) mdDoc {
	lines := splitLines(src)
	var b strings.Builder
	var toc []tocEntry

	i := 0
	for i < len(lines) {
		line := lines[i]

		// 围栏代码块
		if reFence.MatchString(strings.TrimRight(line, " \t")) {
			block, n := readFencedBlock(lines, i)
			b.WriteString(block)
			i += n
			continue
		}

		// 空行
		if strings.TrimSpace(line) == "" {
			i++
			continue
		}

		// 水平线
		if isHr(line) {
			b.WriteString("<hr />\n")
			i++
			continue
		}

		// ATX 标题
		if level, text, ok := parseHeading(line); ok {
			id := uniqueID(text, seen)
			toc = append(toc, tocEntry{Level: level, Text: text, ID: id})
			fmt.Fprintf(&b, "<h%d id=%q>%s</h%d>\n", level, id, inline(text), level)
			i++
			continue
		}

		// 引用
		if strings.HasPrefix(line, ">") {
			block, n, sub := readBlockquote(lines, i, seen)
			b.WriteString(block)
			toc = append(toc, sub...)
			i += n
			continue
		}

		// 表格
		if looksLikeTable(lines, i) {
			block, n := readTable(lines, i)
			b.WriteString(block)
			i += n
			continue
		}

		// 列表
		if isListItem(line) {
			block, n := readList(lines, i)
			b.WriteString(block)
			i += n
			continue
		}

		// 段落：吸收后续非空、非块级行
		para := line
		i++
		for i < len(lines) && strings.TrimSpace(lines[i]) != "" &&
			!isBlockStart(lines[i]) {
			para += "\n" + lines[i]
			i++
		}
		b.WriteString("<p>" + inline(para) + "</p>\n")
	}

	return mdDoc{HTML: b.String(), Toc: toc}
}

// ---------- 块级辅助 ----------

func splitLines(s string) []string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	return strings.Split(s, "\n")
}

func isBlockStart(line string) bool {
	if strings.TrimSpace(line) == "" {
		return true
	}
	if _, _, ok := parseHeading(line); ok {
		return true
	}
	if strings.HasPrefix(line, ">") || isListItem(line) || isHr(line) {
		return true
	}
	if reFence.MatchString(line) || looksLikeFenceStart(line) {
		return true
	}
	return false
}

func parseHeading(line string) (level int, text string, ok bool) {
	if !strings.HasPrefix(line, "#") {
		return 0, "", false
	}
	n := 0
	for n < len(line) && line[n] == '#' {
		n++
	}
	if n > 6 || n == len(line) || (n < len(line) && line[n] != ' ') {
		return 0, "", false
	}
	text = strings.TrimSpace(line[n:])
	if end := strings.LastIndex(text, " #"); end >= 0 && isAllHashes(text[end+1:]) {
		text = strings.TrimSpace(text[:end])
	}
	return n, text, true
}

func isAllHashes(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r != '#' {
			return false
		}
	}
	return true
}

func isHr(line string) bool {
	t := strings.TrimLeft(line, " \t")
	if len(t) < 3 {
		return false
	}
	c := t[0]
	if c != '-' && c != '*' && c != '_' {
		return false
	}
	s := strings.Trim(t, " \t-*_")
	return s == "" && len(t) >= 3
}

func isListItem(line string) bool {
	t := strings.TrimLeft(line, " \t")
	if strings.HasPrefix(t, "- ") || strings.HasPrefix(t, "* ") || strings.HasPrefix(t, "+ ") {
		return true
	}
	if len(t) >= 2 && (t[0] >= '0' && t[0] <= '9') {
		for k := 1; k < len(t); k++ {
			if t[k] >= '0' && t[k] <= '9' {
				continue
			}
			return t[k] == '.' || t[k] == ')'
		}
	}
	return false
}

// readFencedBlock 读取围栏代码块，返回包裹好的 <pre> 与消费行数。
func readFencedBlock(lines []string, i int) (string, int) {
	start := lines[i]
	fence := "```"
	if strings.HasPrefix(start, "~~~") {
		fence = "~~~"
	}
	lang := ""
	if m := reFence.FindStringSubmatch(strings.TrimRight(start, " \t")); m != nil {
		lang = m[1]
	}
	var sb []string
	j := i + 1
	for j < len(lines) && !strings.HasPrefix(lines[j], fence) {
		sb = append(sb, lines[j])
		j++
	}
	// j 指向结束围栏或 EOF
	content := strings.Join(sb, "\n")
	langLabel := "text"
	if lang != "" {
		langLabel = lang
	}
	html := "<div class=\"codeblock\">" +
		"<div class=\"codeblock__head\"><span>" + htmlEscape(langLabel) + "</span>" +
		"<button type=\"button\" class=\"codeblock__copy\" data-copy>" +
		"<svg viewBox=\"0 0 24 24\" width=\"13\" height=\"13\" fill=\"none\" stroke=\"currentColor\" stroke-width=\"2\" stroke-linecap=\"round\" stroke-linejoin=\"round\" aria-hidden=\"true\"><rect x=\"9\" y=\"9\" width=\"13\" height=\"13\" rx=\"2\"/><path d=\"M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1\"/></svg>" +
		"<span>复制</span></button></div>" +
		"<pre><code class=\"language-" + htmlEscape(langLabel) + "\">" +
		htmlEscape(content) + "</code></pre></div>\n"
	return html, j - i + 1
}

func looksLikeFenceStart(line string) bool {
	return strings.HasPrefix(line, "```") || strings.HasPrefix(line, "~~~")
}

// readBlockquote 读取连续引用行；复用外层的 seen 表以保证 id 唯一，
// 并把引用块内的标题一并汇入 TOC。
func readBlockquote(lines []string, i int, seen map[string]int) (string, int, []tocEntry) {
	var sb []string
	j := i
	for j < len(lines) && strings.HasPrefix(strings.TrimLeft(lines[j], " \t"), ">") {
		body := strings.TrimLeft(lines[j], " \t")
		body = strings.TrimPrefix(body, ">")
		body = strings.TrimPrefix(body, " ")
		sb = append(sb, body)
		j++
	}
	inner := renderWith(strings.Join(sb, "\n"), seen)
	return "<blockquote>\n" + inner.HTML + "</blockquote>\n", j - i, inner.Toc
}

// looksLikeTable 判断 i 行是否为表头且 i+1 行为分隔行。
func looksLikeTable(lines []string, i int) bool {
	if i+1 >= len(lines) {
		return false
	}
	if !strings.Contains(lines[i], "|") {
		return false
	}
	return isTableDivider(lines[i+1])
}

func isTableDivider(line string) bool {
	t := strings.TrimSpace(line)
	if !strings.Contains(t, "|") && !strings.Contains(t, "-") {
		return false
	}
	cells := splitRow(t)
	for _, c := range cells {
		c = strings.TrimSpace(c)
		if !regexp.MustCompile(`^:?-+:?$`).MatchString(c) {
			return false
		}
	}
	return len(cells) >= 2
}

func splitRow(line string) []string {
	// 去掉首尾的 |，然后按 | 分割（不处理转义，保持简单）
	t := strings.TrimSpace(line)
	t = strings.TrimPrefix(t, "|")
	t = strings.TrimSuffix(t, "|")
	return strings.Split(t, "|")
}

func readTable(lines []string, i int) (string, int) {
	header := splitRow(lines[i])
	aligns := splitRow(lines[i+1])
	for k := range aligns {
		aligns[k] = normalizeAlign(aligns[k])
	}
	j := i + 2
	var rows [][]string
	for j < len(lines) && strings.Contains(lines[j], "|") {
		rows = append(rows, splitRow(lines[j]))
		j++
	}

	var b strings.Builder
	b.WriteString("<div class=\"table-wrap\"><table>\n<thead>\n<tr>\n")
	for k, h := range header {
		b.WriteString("<th" + alignAttr(k, aligns) + ">" + inline(strings.TrimSpace(h)) + "</th>\n")
	}
	b.WriteString("</tr>\n</thead>\n<tbody>\n")
	for _, r := range rows {
		b.WriteString("<tr>\n")
		for k := 0; k < len(header); k++ {
			cell := ""
			if k < len(r) {
				cell = strings.TrimSpace(r[k])
			}
			b.WriteString("<td" + alignAttr(k, aligns) + ">" + inline(cell) + "</td>\n")
		}
		b.WriteString("</tr>\n")
	}
	b.WriteString("</tbody>\n</table></div>\n")
	return b.String(), j - i
}

func alignAttr(idx int, aligns []string) string {
	if idx < len(aligns) && aligns[idx] != "" {
		return fmt.Sprintf(` align=%q`, aligns[idx])
	}
	return ""
}

func normalizeAlign(s string) string {
	s = strings.TrimSpace(s)
	left := strings.HasPrefix(s, ":")
	right := strings.HasSuffix(s, ":")
	switch {
	case left && right:
		return "center"
	case right:
		return "right"
	case left:
		return "left"
	}
	return ""
}

// readList 读取连续列表项（支持任务列表）。
func readList(lines []string, i int) (string, int) {
	ordered := looksOrdered(lines[i])
	var items []string
	j := i
	for j < len(lines) && isListItem(lines[j]) {
		items = append(items, stripBullet(lines[j]))
		j++
	}

	var b strings.Builder
	tag := "ul"
	if ordered {
		tag = "ol"
	}
	b.WriteString("<" + tag + ">\n")
	for _, it := range items {
		// 任务列表项
		if task, ok := parseTask(it); ok {
			checked := " checked"
			if !task.checked {
				checked = ""
			}
			b.WriteString("<li class=\"task-list-item\"><input type=\"checkbox\" disabled" + checked + ">" +
				"<span>" + inline(task.text) + "</span></li>\n")
		} else {
			b.WriteString("<li>" + inline(it) + "</li>\n")
		}
	}
	b.WriteString("</" + tag + ">\n")
	return b.String(), j - i
}

func looksOrdered(line string) bool {
	t := strings.TrimLeft(line, " \t")
	if len(t) < 2 || t[0] < '0' || t[0] > '9' {
		return false
	}
	for k := 1; k < len(t); k++ {
		if t[k] >= '0' && t[k] <= '9' {
			continue
		}
		return t[k] == '.' || t[k] == ')'
	}
	return false
}

func stripBullet(line string) string {
	t := strings.TrimLeft(line, " \t")
	if m := regexp.MustCompile(`^[-*+] `).FindStringIndex(t); m != nil {
		return t[m[1]:]
	}
	if m := regexp.MustCompile(`^\d+[.)] `).FindStringIndex(t); m != nil {
		return t[m[1]:]
	}
	return t
}

type taskItem struct {
	checked bool
	text    string
}

func parseTask(s string) (taskItem, bool) {
	if !strings.HasPrefix(s, "[ ] ") && !strings.HasPrefix(s, "[x] ") && !strings.HasPrefix(s, "[X] ") {
		return taskItem{}, false
	}
	checked := s[1] == 'x' || s[1] == 'X'
	return taskItem{checked: checked, text: s[4:]}, true
}

// ---------- 行内渲染 ----------

func inline(s string) string {
	// 1) 先把行内代码抽成占位符，避免其中的 * _ [ ] 被后续规则误处理
	var codes []string
	s = reCode.ReplaceAllStringFunc(s, func(m string) string {
		codes = append(codes, reCode.FindStringSubmatch(m)[1])
		return fmt.Sprintf("\x00%d\x00", len(codes)-1)
	})

	// 2) 转义 HTML
	s = htmlEscape(s)

	// 3) 图片（必须在链接之前，否则会被链接规则抢先匹配）
	s = reImg.ReplaceAllStringFunc(s, func(m string) string {
		parts := reImg.FindStringSubmatch(m)
		alt, src, title := parts[1], parts[2], parts[3]
		t := ""
		if title != "" {
			t = fmt.Sprintf(` title=%q`, title)
		}
		return fmt.Sprintf(`<img src=%q alt=%q%s />`, src, alt, t)
	})

	// 4) 链接
	s = reLink.ReplaceAllStringFunc(s, func(m string) string {
		parts := reLink.FindStringSubmatch(m)
		label, href, title := parts[1], parts[2], parts[3]
		t := ""
		if title != "" {
			t = fmt.Sprintf(` title=%q`, title)
		}
		rel := ""
		if strings.HasPrefix(href, "http") {
			rel = ` rel="noopener"`
		}
		return fmt.Sprintf(`<a href=%q%s%s>%s</a>`, href, t, rel, label)
	})

	// 5) 自动链接（此时 < > 已转义为实体）
	s = reAutolink.ReplaceAllStringFunc(s, func(m string) string {
		url := strings.TrimSuffix(strings.TrimPrefix(m, "&lt;"), "&gt;")
		return fmt.Sprintf(`<a href=%q rel="noopener">%s</a>`, url, url)
	})

	// 6) 粗体，再斜体
	s = reBold.ReplaceAllStringFunc(s, func(m string) string {
		parts := reBold.FindStringSubmatch(m)
		inner := parts[1]
		if inner == "" {
			inner = parts[2]
		}
		return "<strong>" + inner + "</strong>"
	})

	s = reItalic.ReplaceAllStringFunc(s, func(m string) string {
		parts := reItalic.FindStringSubmatch(m)
		inner := parts[1]
		if inner == "" {
			inner = parts[2]
		}
		return "<em>" + inner + "</em>"
	})

	// 7) 最后还原行内代码：放在强调之后，代码里的 * _ 才不会被当成标记
	s = reCodeSlot.ReplaceAllStringFunc(s, func(m string) string {
		idx, err := strconv.Atoi(m[1 : len(m)-1])
		if err != nil || idx < 0 || idx >= len(codes) {
			return m
		}
		return "<code>" + htmlEscape(codes[idx]) + "</code>"
	})

	return s
}

func htmlEscape(s string) string {
	r := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
		`"`, "&quot;",
	)
	return r.Replace(s)
}
