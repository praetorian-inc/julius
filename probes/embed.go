package probes

import "embed"

//go:embed *.yaml
var EmbeddedProbes embed.FS
