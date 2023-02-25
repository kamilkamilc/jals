package views

import (
	_ "embed"
)

//go:embed templates/index.gohtml
var IndexTemplate []byte
