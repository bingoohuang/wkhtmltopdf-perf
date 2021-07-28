package wkp

import (
	"embed"
	"io/fs"
)

//go:embed assets
var AssetsFS embed.FS

var Assets, _ = fs.Sub(AssetsFS, "assets")
