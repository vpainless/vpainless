begin;

PRAGMA foreign_keys = ON;
attach database 'data/access.db' as access;
attach database 'data/hosting.db' as hosting;

create table if not exists access.groups (
	id uuid not null primary key default (gen_uuid_v4()),
	name text not null
);

create table if not exists access.users (
	id uuid not null primary key default (gen_uuid_v4()),
	group_id uuid,
	username text unique not null,
	password text not null,
	role text not null default 'client',
	foreign key (group_id) references groups(id)
);

create table if not exists hosting.groups (
	id uuid not null primary key default (gen_uuid_v4()),
	name text not null,
	provider_name text not null,
	provider_url text not null,
	provider_apikey text not null,
	default_xray_template uuid not null,
	default_ssh_key uuid not null,
	default_startup_script uuid not null,

	foreign key (default_xray_template) references xray_templates(id),
	foreign key (default_ssh_key) references ssh_keys(id),
	foreign key (default_startup_script) references startup_scripts(id)
);

create table if not exists hosting.users (
	id uuid not null primary key default (gen_uuid_v4()),
	group_id uuid not null,
	role text not null default 'client',
	foreign key (group_id) references groups(id)
);

create table if not exists hosting.xray_templates (
	id uuid not null primary key default (gen_uuid_v4()),
	group_id uuid not null,
	base text not null
);

create table if not exists hosting.ssh_keys (
	id uuid not null primary key default (gen_uuid_v4()),
	group_id uuid not null,
	remote_id uuid,
	name text not null,
	private_key blob not null,
	public_key blob not null
);

create table if not exists hosting.startup_scripts (
	id uuid not null primary key default (gen_uuid_v4()),
	group_id uuid not null,
	remote_id uuid,
	content text not null
);

create table if not exists hosting.instances (
	id uuid not null primary key default (gen_uuid_v4()),
	user_id uuid not null,
	remote_id uuid,
	ip text,
	status text not null,
	connection_str text,
	private_key blob not null,
	created_at text not null,
	updated_at text not null,
	deleted_at text,
	foreign key (user_id) references users(id)
);

create unique index hosting.idx_unique_user_id_not_deleted on instances (user_id) where deleted_at is null;

commit;
