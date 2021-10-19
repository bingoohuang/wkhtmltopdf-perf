package wkp

import (
	"embed"
	"io/fs"
)

// AssetsFS is the assets.
//go:embed assets
var AssetsFS embed.FS

// Assets ...
var Assets, _ = fs.Sub(AssetsFS, "assets")
