package embeddedconfigs

import (
	_ "embed"
)

//go:embed default/manager-env-images.yaml
// EnvImages image data from the embedded yaml file
var EnvImages []byte
