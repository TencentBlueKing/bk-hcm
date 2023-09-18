/*
    SQLVER=9999,HCMVER=v9.9.9

    Notes:
        1. 新增任务流表
        2. 新增任务表
*/

CREATE OR REPLACE VIEW `hcm_version`(`hcm_ver`, `sql_ver`) AS SELECT 'v9.9.9' as `hcm_ver`, '9999' as `sql_ver`;

-- 1. 新增任务流表
create table if not exists `async_flow`
(   
    `id`                 varchar(64) not null,  
    `name`               varchar(64) not null,  
    `state`              varchar(16) not null,
    `reason`             json default null,
    `share_data`         json default null,
    `memo`               varchar(64) not null,
    `creator`            varchar(64) not null,
    `reviser`            varchar(64) not null,
    `created_at`         timestamp not null default current_timestamp,
    `updated_at`         timestamp not null default current_timestamp on update current_timestamp,
    primary key (`id`)
) engine = innodb
  default charset = utf8mb4
  collate utf8mb4_bin;

insert into id_generator(`resource`, `max_id`) values ('async_flow', '0');

-- 2. 新增任务表
create table if not exists `async_flow_task` 
(
    `id`                 varchar(64) not null,
    `flow_id`            varchar(64) not null,
    `flow_name`          varchar(64) not null, 
    `action_name`        varchar(64) not null, 
    `params`             json default null,
    `retry_count`        bigint not null,
    `timeout_secs`       bigint not null,
    `depend_on`          varchar(64) default '',
    `state`              varchar(16) not null,
    `memo`               varchar(64) not null,
    `reason`             json default null,
    `share_data`         json default null,
    `creator`            varchar(64) not null,
    `reviser`            varchar(64) not null,
    `created_at`         timestamp not null default current_timestamp,
    `updated_at`         timestamp not null default current_timestamp on update current_timestamp,
    primary key (`id`)
) engine = innodb
  default charset = utf8mb4
  collate utf8mb4_bin;

insert into id_generator(`resource`, `max_id`) values ('async_flow_task', '0');