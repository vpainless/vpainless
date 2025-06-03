package hosting.instances

import rego.v1

# Default deny
default allow := false

default partial := {
	"condition": "",
	"values": [],
}

################ Create
# Allow users to create instances
allow if {
	input.action = "create"
	input.principal.id = input.resource.user_id
	input.principal.group_id
	input.principal.role = "client"
}

################ Get
# Users should be able to see their instances
allow if {
	input.action = "get"
	input.principal.id
	input.principal.group_id
	input.principal.role in ["client", "admin"]
	input.resource.id
}

partial := clause if {
	input.action = "get"
	input.principal.id
	input.principal.group_id
	input.principal.role in ["client", "admin"]
	input.resource.id
	clause := {
		"condition": "user_id = ?",
		"values": [input.principal.id],
	}
}


################ List
# Clients should be able to list their instance
allow if {
	input.action = "list"
	input.principal.id
	input.principal.group_id
	input.principal.role = "client"
}

partial := clause if {
	input.action = "list"
	input.principal.id
	input.principal.group_id
	input.principal.role = "client"
	clause := {
		"condition": "user_id = ?",
		"values": [input.principal.id],
	}
}

# Admins should be able to list their clients instances.
allow if {
	input.action = "list"
	input.principal.id
	input.principal.group_id
	input.principal.role = "admin"
}

partial := clause if {
	input.action = "list"
	input.principal.id
	input.principal.group_id
	input.principal.role = "admin"
	clause := {
		"condition": "group_id = ?",
		"values": [input.principal.group_id],
	}
}


################ Delete
# Clients should be able to delete their instance
allow if {
	input.action = "delete"
	input.principal.id
	input.principal.group_id
	input.principal.role = "client"
	input.resource.id
}

partial := clause if {
	input.action = "delete"
	input.principal.id
	input.principal.group_id
	input.principal.role = "client"
	input.resource.id
	clause := {
		"condition": "user_id = ?",
		"values": [input.principal.id],
	}
}


# Admins should be able to delete instances that belong to their clients
allow if {
	input.action = "delete"
	input.principal.id
	input.principal.group_id
	input.principal.role = "admin"
	input.resource.id
}

partial := clause if {
	input.action = "delete"
	input.principal.id
	input.principal.group_id
	input.principal.role = "admin"
	input.resource.id
	clause := {
		"condition": "user_id in (select id from users where group_id = ?)",
		"values": [input.principal.group_id],
	}
}
