# 数据库schema
```sql
create table main.wg_server
(
    id          INTEGER not null
        constraint wg_server_pk
            primary key autoincrement,
    name        TEXT    not null,
    address     TEXT    not null,
    listen_port integer not null,
    private_key TEXT    not null,
    public_key  TEXT    not null,
    mtu         integer not null,
    dns         TEXT,
    post_up     TEXT,
    post_down   TEXT,
    endpoint    TEXT,
    comments    TEXT
);

create table main.wg_clients
(
    id                   INTEGER           not null
        constraint wg_clients_pk
            primary key autoincrement,
    server_id            integer           not null,
    name                 TEXT              not null,
    address              TEXT              not null,
    listen_port          integer,
    private_key          TEXT              not null,
    public_key           TEXT              not null,
    allow_ips            TEXT              not null,
    mtu                  integer           not null,
    dns                  TEXT,
    description          TEXT,
    comments             TEXT,
    disabled             integer default 0 not null,
    persistent_keepalive integer           not null
);

create table main.sys_users
(
    id     integer not null
        constraint sys_users_pk
            primary key autoincrement,
    name   TEXT    not null,
    passwd TEXT    not null,
    roles   TEXT    not null
);

```
