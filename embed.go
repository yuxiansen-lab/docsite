// Package docsite embeds the runtime assets (config, docs, web) so the
// binary is fully self-contained.
//
// The embed directives must live at the module root: go:embed patterns are
// resolved relative to the directory of the source file, so a directive
// inside internal/service could not reach config/ or docs/.
package docsite

import "embed"

// FS holds the three asset trees shipped with the binary:
//
//	config/  site.json, menu.json
//	docs/    Markdown documents and docs/assets static files
//	web/     the static frontend (HTML/CSS/JS)
//
//go:embed config docs web
var FS embed.FS
