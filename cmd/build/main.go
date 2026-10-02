// Command build 把 DocSite 导出为纯静态站点，用于 GitHub Pages 等静态托管。
//
// 产出的 dist/ 目录结构：
//
//	dist/
//	  index.html                 前端入口（资源引用均为相对路径）
//	  css/  js/  assets/         前端静态资源
//	  api/config.json            站点配置
//	  api/menu.json              菜单树
//	  api/docs.json              全部文档（含正文 HTML、TOC、上/下篇）
//	  api/search-index.json      客户端搜索索引
//	  docs/assets/               文档内图片等静态资源
//	  .nojekyll                  阻止 GitHub Pages 的 Jekyll 处理
//
// 服务端 /api/*.json 路由与这些文件内容一致，因此同一份前端在
// “Go 单二进制” 与 “纯静态托管” 两种模式下行为完全相同。
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"

	"docsite/internal/model"
	"docsite/internal/service"
)

func main() {
	out := flag.String("out", "dist", "输出目录")
	flag.Parse()

	if err := run(*out); err != nil {
		log.Fatalf("build: %v", err)
	}
}

func run(out string) error {
	svc, err := service.New()
	if err != nil {
		return err
	}

	// 从干净目录开始，避免残留旧文件
	if err := os.RemoveAll(out); err != nil {
		return fmt.Errorf("清理输出目录: %w", err)
	}
	if err := os.MkdirAll(out, 0o755); err != nil {
		return fmt.Errorf("创建输出目录: %w", err)
	}

	// 1) 前端资源
	if err := copyFS(svc.Web(), out); err != nil {
		return fmt.Errorf("导出前端: %w", err)
	}

	// 2) 文档静态资源 -> docs/assets
	if err := copyFS(svc.Assets(), filepath.Join(out, "docs", "assets")); err != nil {
		return fmt.Errorf("导出文档资源: %w", err)
	}

	// 3) JSON 接口
	apiDir := filepath.Join(out, "api")
	if err := os.MkdirAll(apiDir, 0o755); err != nil {
		return err
	}

	docs := svc.AllDocs()
	writes := []struct {
		name string
		v    any
	}{
		{"config.json", svc.Site()},
		{"menu.json", svc.Menu()},
		{"docs.json", model.DocsBundle{Docs: docs}},
		{"search-index.json", model.SearchIndex{Items: svc.SearchIndex()}},
	}
	for _, w := range writes {
		if err := writeJSON(filepath.Join(apiDir, w.name), w.v); err != nil {
			return err
		}
	}

	// 4) .nojekyll：否则 Pages 会忽略以下划线开头的文件
	if err := os.WriteFile(filepath.Join(out, ".nojekyll"), nil, 0o644); err != nil {
		return err
	}

	fmt.Printf("已导出 %d 篇文档 -> %s\n", len(docs), out)
	return nil
}

// copyFS 把 fs.FS 的内容递归复制到磁盘目录。
func copyFS(src fs.FS, dst string) error {
	return fs.WalkDir(src, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		target := filepath.Join(dst, filepath.FromSlash(p))

		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}

		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}

		in, err := src.Open(p)
		if err != nil {
			return err
		}
		defer in.Close()

		outFile, err := os.Create(target)
		if err != nil {
			return err
		}
		defer outFile.Close()

		if _, err := io.Copy(outFile, in); err != nil {
			return err
		}
		return nil
	})
}

func writeJSON(path string, v any) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}
