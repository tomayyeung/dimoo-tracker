// Package assets embeds frontend templates and static files for the Go server.
package assets

import "embed"

// Files contains the embedded HTML templates and static assets served by the app.
//
//go:embed templates/*.html templates/partials/*.html static/*.css static/*.js
var Files embed.FS
