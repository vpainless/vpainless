package hosting_test.instances

import data.hosting.instances.allow
import data.hosting.instances.partial

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

test_clients_should_be_able_to_create_instance if {
	request := {
		"principal": {
			"id": "11000000-0000-0000-0000-000000000000",
			"group_id": "00000000-0000-0000-0000-000000000011",
			"role": "client"
		},
		"action": "create",
		"resource": {
			"user_id": "11000000-0000-0000-0000-000000000000"
		}
	}

	allow with input as request
}


test_groupless_clients_should_not_be_able_to_create_instance if {
	request := {
		"principal": {
			"id": "11000000-0000-0000-0000-000000000000",
			"role": "client"
		},
		"action": "create",
		"resource": {
			"user_id": "11000000-0000-0000-0000-000000000000"
		}
	}

	not allow with input as request
}

test_admins_should_not_be_able_to_create_instance if {
	request := {
		"principal": {
			"id": "11000000-0000-0000-0000-000000000000",
			"role": "admin"
		},
		"action": "create",
		"resource": {
			"user_id": "11000000-0000-0000-0000-000000000000"
		}
	}

	not allow with input as request
}

############# Action: get
test_clients_should_be_able_to_get_their_instance if {
	request := {
		"principal": {
			"id": "11000000-0000-0000-0000-000000000000",
			"group_id": "00000000-0000-0000-0000-000000000011",
			"role": "client",
		},
		"action": "get",
		"resource": {
			"id": "18b68320-fd20-4ec5-b84c-5f678cdf46fd"
		}
	}

	allow with input as request

	partial = {
		"condition": "user_id = ?",
		"values": [input.principal.id],
	} with input as request
}

############# Action: list
test_clients_should_be_able_to_list_their_instance if {
	request := {
		"principal": {
			"id": "11000000-0000-0000-0000-000000000000",
			"group_id": "00000000-0000-0000-0000-000000000011",
			"role": "client",
		},
		"action": "list"
	}

	allow with input as request

	partial = {
		"condition": "user_id = ?",
		"values": [input.principal.id],
	} with input as request
}

test_admins_should_be_able_to_list_their_client_instances if {
	request := {
		"principal": {
			"id": "11000000-0000-0000-0000-000000000000",
			"group_id": "00000000-0000-0000-0000-000000000011",
			"role": "admin",
		},
		"action": "list",
	}

	allow with input as request

	partial = {
		"condition": "group_id = ?",
		"values": [input.principal.group_id],
	} with input as request
}


############# Action: delete
test_clients_should_be_able_to_delete_their_instance if {
	request := {
		"principal": {
			"id": "11000000-0000-0000-0000-000000000000",
			"group_id": "00000000-0000-0000-0000-000000000011",
			"role": "client",
		},
		"action": "delete",
		"resource": {
			"id": "00000000-1111-0000-0000-000000000000"
		}
	}

	allow with input as request

	partial = {
		"condition": "user_id = ?",
		"values": [input.principal.id],
	} with input as request
}


test_clients_should_be_able_to_delete_their_clients_instances if {
	request := {
		"principal": {
			"id": "11000000-0000-0000-0000-000000000000",
			"group_id": "00000000-0000-0000-0000-000000000011",
			"role": "admin",
		},
		"action": "delete",
		"resource": {
			"id": "00000000-1111-0000-0000-000000000000"
		}
	}

	allow with input as request

	partial := {
		"condition": "user_id = any (select id from users where group_id = ?)",
		"values": [input.principal.group_id],
	} with input as request
}
