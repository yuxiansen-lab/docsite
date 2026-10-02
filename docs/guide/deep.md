# 深度示例

这是一篇用于演示**三级菜单**的文档。

## 代码示例

```go
package main

import (
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "hello from go")
	})
	http.ListenAndServe(":8080", nil)
}
```

## 任务列表

- [x] 支持三级菜单
- [x] 选中高亮
- [ ] 快捷键导航（规划中）

## 引用

> 简洁是最终的完美。
> —— 列奥纳多·达·芬奇
