//go:build generate

package tools

import (
	// document generation
	_ "github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs"
)

// Format configuration code for use in documentation.
//go:generate tofu fmt -recursive ../examples/
// Generate documentation.
//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs generate --provider-dir ..
