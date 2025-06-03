package access_test.users

import data.access.users.allow
import data.access.users.partial

nil_partial := {
	"condition": "",
	"values": [],
}

test_default_allow if {
	allow == false
}

test_default_partial if {
	partial == nil_partial
}

# TODO: write a test to ensure admins always should have a group id

############# Action: get

input_request_get(uid, pid, role) := request if {
	request := {
		"principal": {
			"group_id": "00000000-3a1c-4768-a723-cad22e955848",
			"id": pid,
			"role": role,
		},
		"action": "get",
		"resource": {"id": uid},
	}
}

test_client_view_himself if {
	request := input_request_get(
		"22222222-3e3c-49e7-852f-1b516782681d",
		"22222222-3e3c-49e7-852f-1b516782681d",
		"client",
	)
	allow with input as request
	partial == nil_partial with input as request
}

test_client_cannot_view_others if {
	request := input_request_get(
		"11111111-3e3c-49e7-852f-1b516782681d",
		"22222222-3e3c-49e7-852f-1b516782681d",
		"client",
	)
	not allow with input as request
	partial == nil_partial with input as request
}

test_admin_partially_view_others if {
	request := input_request_get(
		"11111111-3e3c-49e7-852f-1b516782681d",
		"22222222-3e3c-49e7-852f-1b516782681d",
		"admin",
	)
	allow with input as request
	partial == {
		"condition": "group_id = ?",
		"values": [input.principal.group_id],
	} with input as request
}

############# Action: list
test_clients_should_be_able_to_list_themselves if {
	request := {
		"principal": {
			"id": "11000000-0000-0000-0000-000000000000",
			"group_id": "00000000-0000-0000-0000-111111111111",
			"role": "client",
		},
		"action": "list"
	}

	allow with input as request

	partial == {
		"condition": "id = ?",
		"values": [input.principal.id],
	} with input as request
}

test_admins_should_be_able_to_list_their_users if {
	request := {
		"principal": {
			"id": "11000000-0000-0000-0000-000000000000",
			"group_id": "00000000-0000-0000-0000-111111111111",
			"role": "admin",
		},
		"action": "list"
	}

	allow with input as request
	partial == {
		"condition": "group_id = ?",
		"values": [input.principal.group_id]
	} with input as request
}

############# Action: create
test_users_should_be_able_to_register if {
	request := {"action": "create"}

	allow with input as request
}

test_registered_clients_should_not_be_able_to_create_users if {
	request := {
		"principal": {
			"id": "11000000-0000-0000-0000-000000000000",
			"role": "client",
		},
		"action": "create",
	}

	not allow with input as request
}

test_admins_should_be_able_to_create_users_in_their_groups if {
	request := {
		"principal": {
			"id": "11000000-0000-0000-0000-000000000000",
			"group_id": "00000000-0000-0000-0000-000000000011",
			"role": "admin",
		},
		"action": "create",
		"resource": {"group_id": "00000000-0000-0000-0000-000000000011"},
	}

	allow with input as request
}

test_admins_should_not_be_able_to_create_users_in_others_groups if {
	request := {
		"principal": {
			"id": "11000000-0000-0000-0000-000000000000",
			"group_id": "00000000-0000-0000-0000-000000000011",
			"role": "admin",
		},
		"action": "create",
		"resource": {"group_id": "00000000-0000-0000-0000-000000000022"},
	}

	not allow with input as request
}

############# Action: update

test_clients_update_their_user_or_password if {
	request := {
		"principal": {
			"id": "11000000-0000-0000-0000-000000000000",
			"role": "client",
		},
		"action": "update",
		"resource": {
			"id": "11000000-0000-0000-0000-000000000000",
			"old_role": "client",
			"new_role": "client",
		},
	}

	allow with input as request
}

test_clients_cannot_upgrade_their_role if {
	request := {
		"principal": {
			"id": "11000000-0000-0000-0000-000000000000",
			"group_id": "00000000-0000-0000-0000-000000000011",
			"role": "client",
		},
		"action": "update",
		"resource": {
			"id": "11000000-0000-0000-0000-000000000000",
			"old_group_id": "00000000-0000-0000-0000-000000000011",
			"new_group_id": "00000000-0000-0000-0000-000000000011",
			"old_role": "client",
			"new_role": "admin",
		},
	}

	not allow with input as request
}

test_clients_cannot_change_their_group if {
	request := {
		"principal": {
			"id": "11000000-0000-0000-0000-000000000000",
			"group_id": "00000000-0000-0000-0000-000000000011",
			"role": "client",
		},
		"action": "update",
		"resource": {
			"id": "11000000-0000-0000-0000-000000000000",
			"old_group_id": "00000000-0000-0000-0000-000000000011",
			"new_group_id": "00000000-0000-0000-0000-000000000022",
			"old_role": "client",
			"new_role": "client",
		},
	}

	not allow with input as request
}

test_clients_can_leave_their_group if {
	request := {
		"principal": {
			"id": "11000000-0000-0000-0000-000000000000",
			"group_id": "00000000-0000-0000-0000-000000000011",
			"role": "client",
		},
		"action": "update",
		"resource": {
			"id": "11000000-0000-0000-0000-000000000000",
			"old_group_id": "00000000-0000-0000-0000-000000000011",
			"old_role": "client",
			"new_role": "client",
		},
	}

	allow with input as request
}

test_clients_cannot_edit_other_clients if {
	request := {
		"principal": {
			"id": "11000000-0000-0000-0000-000000000000",
			"role": "client",
		},
		"action": "update",
		"resource": {
			"id": "22000000-0000-0000-0000-000000000000",
			"old_role": "client",
			"new_role": "client",
		},
	}

	not allow with input as request
}

test_admins_can_edit_their_clients if {
	request := {
		"principal": {
			"id": "11000000-0000-0000-0000-000000000000",
			"group_id": "00000000-0000-0000-0000-000000000011",
			"role": "admin",
		},
		"action": "update",
		"resource": {
			"id": "22000000-0000-0000-0000-000000000000",
			"old_group_id": "00000000-0000-0000-0000-000000000011",
			"new_group_id": "00000000-0000-0000-0000-000000000011",
			"old_role": "client",
			"new_role": "client",
		},
	}

	allow with input as request
}

# TODO: admins should be able to bring groupless clients in
# Not decided about this one yet. Instead they should be able to create
# new users, so that should be on create action.

# admins should be able to edit other admings in the group,
# including demoting
test_admins_editing_other_admins_in_their_group if {
	request := {
		"principal": {
			"id": "11000000-0000-0000-0000-000000000000",
			"group_id": "00000000-0000-0000-0000-000000000011",
			"role": "admin",
		},
		"action": "update",
		"resource": {
			"id": "22000000-0000-0000-0000-000000000000",
			"old_group_id": "00000000-0000-0000-0000-000000000011",
			"new_group_id": "00000000-0000-0000-0000-000000000011",
			"old_role": "admin",
			"new_role": "client",
		},
	}

	allow with input as request
}

# admins should be able to kick users out of their group
# this includes changing their user and password too!
test_admins_kicking_users_out if {
	request := {
		"principal": {
			"id": "11000000-0000-0000-0000-000000000000",
			"group_id": "00000000-0000-0000-0000-000000000011",
			"role": "admin",
		},
		"action": "update",
		"resource": {
			"id": "22000000-0000-0000-0000-000000000000",
			"old_group_id": "00000000-0000-0000-0000-000000000011",
			"old_role": "admin",
			"new_role": "client",
		},
	}

	allow with input as request
}

# admins should be able to kick users out of their group
# however, when doing that, users should be demoted to client
test_admins_kicking_users_out_invalid_role if {
	request := {
		"principal": {
			"id": "11000000-0000-0000-0000-000000000000",
			"group_id": "00000000-0000-0000-0000-000000000011",
			"role": "admin",
		},
		"action": "update",
		"resource": {
			"id": "22000000-0000-0000-0000-000000000000",
			"old_group_id": "00000000-0000-0000-0000-000000000011",
			"new_role": "admin",
		},
	}

	not allow with input as request
}

# admins should be able to promote their clients
test_admins_promoting_clients if {
	request := {
		"principal": {
			"id": "11000000-0000-0000-0000-000000000000",
			"group_id": "00000000-0000-0000-0000-000000000011",
			"role": "admin",
		},
		"action": "update",
		"resource": {
			"id": "22000000-0000-0000-0000-000000000000",
			"old_group_id": "00000000-0000-0000-0000-000000000011",
			"new_group_id": "00000000-0000-0000-0000-000000000011",
			"old_role": "client",
			"new_role": "admin",
		},
	}

	allow with input as request
}

# admins should not be able to edit users of other groups
test_admins_should_not_be_able_to_edit_clients_of_other_groups if {
	request := {
		"principal": {
			"id": "11000000-0000-0000-0000-000000000000",
			"group_id": "00000000-0000-0000-0000-000000000011",
			"role": "admin",
		},
		"action": "update",
		"resource": {
			"id": "22000000-0000-0000-0000-000000000000",
			"old_group_id": "00000000-0000-0000-0000-000000000022",
		},
	}

	not allow with input as request
}
