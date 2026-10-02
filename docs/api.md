# API 参考

所有 API 均为无状态 GET 请求，返回 `application/json`。

## `GET /api/config`

返回站点配置。

```json
{ "title": "DocSite", "themeColor": "#6366f1" }
```

## `GET /api/menu`

返回菜单树。

## `GET /api/doc?path=xxx.md`

返回渲染后的文档。

| 字段 | 说明 |
| ---- | ---- |
| `path` | 规范化路径 |
| `title` | 文档标题（首个 H1） |
| `html` | 渲染后的 HTML |
| `toc` | 目录 `[{level, text, id}]` |
| `prev` / `next` | 上一篇/下一篇 |
| `updatedAt` | 渲染时间 |

## `GET /api/search?q=关键词`

全文搜索，返回前 20 条命中。

## 静态资源

- `/docs/assets/*`：文档内图片与附件
- 其余路径由前端 SPA 接管
