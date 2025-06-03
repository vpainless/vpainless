package access.groups

import rego.v1

# Default deny
default allow := false

default partial := {
	"condition": "",
	"values": [],
}

# verbs := ["get", "list", "create", "update", "delete"]
# TODO: move create user policy here when we add method and path varialbes to the request.

################ GET

################ CREATE
# users should be able to register in the system by themselves
allow if {
	input.action == "create"
	input.principal.role == "client"
	not input.principal.group_id
}
