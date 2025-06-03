package access.users

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
# Allow uses to view tmeselves
allow if {
	input.action = "get"
	input.principal.id == input.resource.id
}

# Allow admins view users of their own group
# admins cannot view already existing clients
# if they are not in their group, but
# they can create new clients and onboard them
allow if {
	input.action = "get"
	input.principal.role == "admin"
	input.principal.id != input.resource.id
}

partial := clause if {
	input.action = "get"
	input.principal.role == "admin"
	input.principal.id != input.resource.id
	clause := {
		"condition": "group_id = ?",
		"values": [input.principal.group_id],
	}
}

################ LIST
# Allow clients to view themselves
# TODO: write test
allow if {
	input.action = "list"
	input.principal.role == "client"
}

partial := clause if {
	input.action = "list"
	input.principal.role == "client"
	clause := {
		"condition": "id = ?",
		"values": [input.principal.id],
	}
}

# Allow admins to list users in ther group
# TODO: write test
allow if {
	input.action = "list"
	input.principal.group_id
	input.principal.role == "admin"
}

partial := clause if {
	input.action = "list"
	input.principal.group_id
	input.principal.role == "admin"
	clause := {
		"condition": "group_id = ?",
		"values": [input.principal.group_id],
	}
}

################ CREATE
# users should be able to register in the system by themselves
allow if {
	input.action = "create"
	not input.resource
	not input.principal
}

# admins should be able to create users in ther own group
allow if {
	input.action = "create"
	input.principal.role = "admin"
	input.principal.group_id = input.resource.group_id
}

################ UPDATE
# TODO: review all these policies.

# input format
# {
#   "principal": {
#     "group_id": "00000000-3a1c-4768-a723-cad22e955848",
#     "id": "11000000-3e3c-49e7-852f-1b516782681d",
#     "role": "admin"
#   },
#   "action": "update",
#   "resource": {
#     "id": "11000000-3e3c-49e7-852f-1b516782681d",
#     "old_group_id": "506357a8-2288-4bc3-b798-abae4ebf9d5e",
#     "old_role": "client",
#     "new_group_id": "506357a8-2288-4bc3-b798-abae4ebf9d5e",
#     "new_role": "client"
#   }
# }
# admins editing themselves without changing their group
allow if {
	input.action == "update"
	input.principal.role == "admin"
	input.resource.new_role in ["admin", "client"]
	input.principal.id == input.resource.id
	input.principal.group_id == input.resource.new_group_id
}

# clients editing themselves, without changing their role or group
allow if {
	input.action == "update"
	input.principal.role == "client"
	input.resource.new_role in ["client"]
	input.principal.id == input.resource.id
	object.get(input, ["resource", "new_group_id"], "null") in ["null", object.get(input, ["principal", "group_id"], "null")]
}

# admins are allowed editing other users in their group
allow if {
	input.action == "update"
	input.principal.role == "admin"
	input.resource.new_role in ["client", "admin"]
	input.resource.old_group_id = input.principal.group_id
	input.resource.new_group_id = input.principal.group_id
}

# admins are allowed to kick users out of their group
allow if {
	input.action == "update"
	input.principal.role == "admin"
	input.resource.new_role in ["client"]
	input.resource.old_group_id = input.principal.group_id
	not input.resource.new_group_id
}
