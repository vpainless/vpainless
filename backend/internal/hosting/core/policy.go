package core

import (
	_ "embed"
)

//go:embed policy/instances.rego
var instancesModule string

func policies() map[string]string {
	return map[string]string{
		"access/instances.rego": instancesModule,
	}
}
