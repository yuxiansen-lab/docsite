# API 参考

所有 API 均为无状态 GET 请求，返回 `application/json`。

## 服务端接口

以下接口由 Go 单二进制实时生成。

### `GET /api/config`

返回站点配置。

```json
{ "title": "DocSite", "themeColor": "#6366f1" }
```

### `GET /api/menu`

返回菜单树。

### `GET /api/doc?path=xxx.md`

返回渲染后的文档。

| 字段 | 说明 |
| ---- | ---- |
| `path` | 规范化路径 |
| `title` | 文档标题（首个 H1） |
| `html` | 渲染后的 HTML |
| `toc` | 目录 `[{level, text, id}]` |
| `prev` / `next` | 上一篇/下一篇 |
| `updatedAt` | 渲染时间 |

### `GET /api/search?q=关键词`

全文搜索，返回前 20 条命中。

## 静态 JSON 接口

下面这组路径在**服务端与静态导出中完全一致**，前端实际使用的就是它们。
这样同一份前端既能跑在 Go 单二进制上，也能跑在 GitHub Pages 等纯静态托管上。

| 路径 | 说明 |
| ---- | ---- |
| `/api/config.json` | 站点配置 |
| `/api/menu.json` | 菜单树 |
| `/api/docs.json` | 全部文档（HTML + TOC + 上下篇） |
| `/api/search-index.json` | 客户端搜索索引 |

静态模式下由 `go run ./cmd/build -out dist` 预生成，其余能力（Markdown 渲染、
TOC 抽取、上下篇计算）全部在构建期完成；搜索移到浏览器端执行。

## 静态资源

- `/docs/assets/*`：文档内图片与附件
- 其余路径由前端 SPA 接管

## 部署方式

| 方式 | 地址 | 说明 |
| ---- | ---- | ---- |
| GitHub Pages | <https://yuxiansen-lab.github.io/docsite/> | 纯静态，推送 `main` 后自动部署 |
| 单二进制 | 自定 `-addr` | `go build -o docsite ./cmd/server` |
