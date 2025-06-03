package access_test.groups

import data.access.groups.allow
import data.access.groups.partial

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

############# Action: create

test_admins_should_not_be_able_to_create_groups if {
	request := {
		"principal": {
			"id": "11000000-0000-0000-0000-000000000000",
			"role": "admin",
		},
		"action": "create",
	}

	not allow with input as request
}

test_clients_that_have_a_group_should_not_be_able_to_create_a_group if {
	request := {
		"principal": {
			"id": "11000000-0000-0000-0000-000000000000",
			"group_id": "00000000-0000-0000-0000-000000000011",
			"role": "client",
		},
		"action": "create",
	}

	not allow with input as request
}

test_groupless_clients_should_be_able_to_create_a_group if {
	request := {
		"principal": {
			"id": "11000000-0000-0000-0000-000000000000",
			"role": "client",
		},
		"action": "create",
	}

	allow with input as request
}
