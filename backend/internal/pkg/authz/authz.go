package authz

import "github.com/gofrs/uuid/v5"

type (
	// Role specifies the set of privileges a user in our system can have.
	Role    string
	ctxtype string
)

const (
	// Client is the least privileged user in the system.
	Client Role = "client"
	// Admin is a user that have higher privileges in the group he belongs.
	Admin Role    = "admin"
	authz ctxtype = "authz"
)

// Principal encapsulates the authorization info of the logged in user,
// including it's ID, group ID and role.
type Principal struct {
	ID      uuid.UUID
	GroupID uuid.UUID
	Role    Role
}

// Verb is the action performed on resources.
// e.g. "delete" "user" where delete is the verb and "user" is the resource.
type Verb string

const (
	Get    Verb = "get"
	List   Verb = "list"
	Create Verb = "create"
	Update Verb = "update"
	Delete Verb = "delete"
)

// Resourcer is an interface used to identify a resource.
// Resources belong to a group. These groups are used
// to group related rego expressions into a module.
type Resourcer interface {
	Resource() (group string, value any)
}

// Resource is a helper structure that implements Resourcer.
type Resource struct {
	Group string
	Value any
}

func (r Resource) Resource() (string, any) {
	return r.Group, r.Value
}

// ResourceFunc is a helper function that implements Resourcer
type ResourceFunc func() (string, any)

func (f ResourceFunc) Resource() (string, any) {
	return f()
}

// ResourceID is a helper function that creates resources from uuids.
func ResourceID(resourceGroup string, id uuid.UUID) Resource {
	return Resource{
		Group: resourceGroup,
		Value: map[string]string{
			"id": id.String(),
		},
	}
}
