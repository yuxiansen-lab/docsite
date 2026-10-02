# 配置

DocSite 由两个 JSON 文件驱动。

## site.json

```json
{
  "title": "DocSite",
  "description": "极简现代文档展示系统",
  "themeColor": "#6366f1",
  "defaultDoc": "index.md",
  "logoText": "DS"
}
```

- `themeColor`：CSS 变量 `--accent` 的来源
- `defaultDoc`：首次访问时加载的文档

## menu.json

每个菜单项支持以下字段：

| 字段 | 类型 | 说明 |
| ---- | ---- | ---- |
| `id` | string | 唯一标识 |
| `title` | string | 显示名称 |
| `icon` | string | 内置图标名 |
| `action` | string | `doc` / `link` / `callback` / `toggle` |
| `path` | string | 文档路径（action=doc） |
| `url` | string | 外部链接（action=link） |
| `order` | int | 排序权重，越小越靠前 |
| `expanded` | bool | 默认展开 |
| `hidden` | bool | 隐藏但不删除 |
| `children` | array | 子菜单 |

## 修改后生效

配置文件通过 `embed.FS` 编译进二进制，修改后需重新构建。
