package embeddedconfigs

import (
	_ "embed"
)

//go:embed default/manager-env-images.yaml
var EnvImages []byte
