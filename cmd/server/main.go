// Package main boots the docsite server.
//
// docsite is a minimal, configuration-driven Markdown documentation system:
// a single Go binary that serves a static web frontend, a JSON menu tree,
// site configuration and rendered Markdown documents from an embedded
// filesystem.
package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"docsite/internal/handler"
)

func main() {
	addr := flag.String("addr", ":8080", "listen address, e.g. :8080 or 127.0.0.1:3000")
	flag.Parse()

	srv, err := handler.New()
	if err != nil {
		log.Fatalf("docsite: %v", err)
	}

	log.Printf("docsite listening on http://%s", *addr)
	if err := http.ListenAndServe(*addr, srv); err != nil && err != http.ErrServerClosed {
		log.Printf("server stopped: %v", err)
		os.Exit(1)
	}
}
