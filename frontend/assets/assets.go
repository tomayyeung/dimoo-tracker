package assets

import "embed"

//go:embed templates/*.html templates/partials/*.html static/*.css
var Files embed.FS
