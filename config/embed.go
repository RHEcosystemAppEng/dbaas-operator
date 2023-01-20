package embeddedconfigs

import (
	_ "embed"
)

// EnvImages image data from the embedded yaml file
//
//go:embed default/manager-env-images.yaml
var EnvImages []byte
