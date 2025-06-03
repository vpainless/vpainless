package authz

import (
	"context"
	"testing"

	"github.com/gofrs/uuid/v5"
	"github.com/stretchr/testify/require"
)

var content string = `
package access

import rego.v1

# Default deny
default allow := false

# Allow if the user is requesting their own data
allow if {
	input.principal.role == "user"
	input.resource.type == "users"
	input.action in ["view"]
	input.principal.id == input.resource.value
}

# Allow if the user is an admin and the requested user is in the same group
allow if {
	input.principal.role == "admin"
	input.resource.type == "users"
	input.action in ["view"]
}

partial := clause if {
	input.principal.role == "admin"
	input.resource.type == "users"
	input.action in ["view"]
	clause := {
		"condition": "g.id = ?",
		"values": [input.principal.group_id],
	}
}
`

func TestPolicyMaking(t *testing.T) {
	ctx := context.Background()

	p := Principal{
		ID:      uuid.FromStringOrNil("aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"),
		GroupID: uuid.FromStringOrNil("bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"),
		Role:    Admin,
	}
	resouceID := uuid.FromStringOrNil("ffffffff-ffff-ffff-ffff-ffffffffffff")
	e := NewValidator("access", WithRegoModule("", content))
	expect := Policy{
		Allow: true,
		Partial: Clause{
			Condition: "g.id = ?",
			Values:    []any{"bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"},
		},
	}

	policy, err := e.Can(ctx, p, "view", ResourceID("users", resouceID))
	require.NoError(t, err, "should evaluate correctly")
	require.Equal(t, expect, policy, "policy should match")
}
