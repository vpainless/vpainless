begin;

drop table if exists access.users;
drop table if exists access.groups;

drop table if exists hosting.instances;
drop table if exists hosting.users;

drop table if exists hosting.xray_templates;
drop table if exists hosting.ssh_keys;
drop table if exists hosting.startup_scripts;
drop table if exists hosting.groups;

commit;

detach database access;
detach database hosting;
