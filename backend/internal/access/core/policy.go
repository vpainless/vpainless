package core

import (
	_ "embed"
)

//go:embed policy/users.rego
var usersModule string

//go:embed policy/groups.rego
var groupsModule string

func policies() map[string]string {
	return map[string]string{
		"access/users.rego":  usersModule,
		"access/groups.rego": groupsModule,
	}
}
