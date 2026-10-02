# 欢迎使用 DocSite

> 极简主义、现代化、丝滑流畅的文档展示系统。

## 特性

- **单二进制部署** — 静态资源、文档与配置全部嵌入，`go build` 即得可运行程序
- **配置驱动** — 菜单、站点标题、主题色均由 JSON 配置生成
- **多级子母菜单** — 支持任意层级，带展开动画与选中高亮
- **Markdown 渲染** — 代码高亮、一键复制、TOC 目录、上下篇导航
- **亮/暗主题** — 默认跟随系统，可手动切换并持久化
- **全文搜索** — 跨所有文档的即时搜索

## 快速开始

```bash
# 克隆并构建
git clone https://example.com/docsite.git
cd docsite
go build ./cmd/server

# 运行
./server -addr :8080
```

打开 <http://localhost:8080> 即可。

## 项目结构

| 目录 | 说明 |
| ---- | ---- |
| `cmd/server` | 入口程序 |
| `internal/` | 业务逻辑（handler / service / model） |
| `web/` | 前端静态资源（嵌入二进制） |
| `config/` | `menu.json` 与 `site.json` |
| `docs/` | Markdown 文档与静态资源 |
