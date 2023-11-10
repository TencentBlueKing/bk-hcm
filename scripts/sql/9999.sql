/*
    SQLVER=9999,HCMVER=v9.9.9

    Notes:
        1. 添加子账号表。
        2. 修复aws_security_group_rule表memo字段错误的非空约束
        3. aws_region表增加account_id字段，独立存储aws每个账号的地域信息
        4. 添加业务收藏表
        5. 添加账号同步详情表
        6. 资源下账号粒度管理回收保留时间
        7. 子账号表新增账号类型字段
        8. 新增任务流表
        9. 新增任务表
        10. EIP、硬盘、网络接口增加回收状态recycle_status字段
        11. 镜像添加操作系统类型字段
*/
start transaction;

CREATE OR REPLACE VIEW `hcm_version`(`hcm_ver`, `sql_ver`) AS
SELECT 'v9.9.9' as `hcm_ver`, '9999' as `sql_ver`;

-- 1. 添加子账号表
create table if not exists `sub_account`
(
    `id`         varchar(64)  not null,
    `cloud_id`   varchar(255) not null,
    `name`       varchar(255) not null,
    `vendor`     varchar(16)  not null,
    `site`       varchar(32)  not null,
    `account_id` varchar(64)  not null,
    `extension`  json         not null,
    `managers`   json,
    `bk_biz_ids` json,
    `memo`       varchar(255)          default '',
    `creator`    varchar(64)  not null,
    `reviser`    varchar(64)  not null,
    `created_at` timestamp    not null default current_timestamp,
    `updated_at` timestamp    not null default current_timestamp on update current_timestamp,
    primary key (`id`),
    unique key `idx_uk_vendor_cloud_id` (`vendor`, `cloud_id`)
) engine = innodb
  default charset = utf8mb4
  collate utf8mb4_bin;

insert into id_generator(`resource`, `max_id`)
values ('sub_account', '0');

-- 2. 修复aws_security_group_rule表memo字段错误的非空约束
alter table aws_security_group_rule
    modify column memo varchar(255) default '';

-- 3. aws_region表增加account_id字段，独立存储aws每个账号的地域信息
delete
from aws_region;
alter table aws_region
    add column account_id varchar(64) not null;

alter table aws_region
    drop index `idx_uk_region_id_status`;

alter table aws_region
    add unique key `idx_uk_account_id_region_id` (`account_id`, `region_id`);

-- 4. 添加业务收藏表
create table if not exists `user_collection`
(
    `id`         varchar(64) not null,
    `user`       varchar(64) not null,
    `res_type`   varchar(50) not null,
    `res_id`     varchar(64) not null,
    `creator`    varchar(64) not null,
    `created_at` timestamp   not null default current_timestamp,
    primary key (`id`),
    unique key `idx_uk_user_res_type_res_id` (`user`, `res_type`, `res_id`)
) engine = innodb
  default charset = utf8mb4
  collate utf8mb4_bin;

insert into id_generator(`resource`, `max_id`)
values ('user_collection', '0');

-- 5. 添加账号同步详情表
create table if not exists `account_sync_detail`
(
    `id`                varchar(64) not null,
    `vendor`            varchar(16) not null,
    `account_id`        varchar(64) not null,
    `res_name`          varchar(64) not null,
    `res_status`        varchar(64) not null,
    `res_end_time`      varchar(64)          default '',
    `res_failed_reason` json                 default null,
    `creator`           varchar(64) not null,
    `reviser`           varchar(64) not null,
    `created_at`        timestamp   not null default current_timestamp,
    `updated_at`        timestamp   not null default current_timestamp on update current_timestamp,
    primary key (`id`),
    unique key `idx_uk_vendor_account_id_res_name` (`vendor`, `account_id`, `res_name`)
) engine = innodb
  default charset = utf8mb4
  collate utf8mb4_bin;

insert into id_generator(`resource`, `max_id`)
values ('account_sync_detail', '0');

-- 6. 资源下账号粒度管理回收保留时间
alter table account
    add column recycle_reserve_time bigint default -1;
alter table recycle_record
    add column recycled_at timestamp not null default current_timestamp;

-- 7. 子账号表新增账号类型字段
alter table sub_account
    add column account_type varchar(64) default '';

-- 8. 新增任务流表
create table if not exists `async_flow`
(
    `id`         varchar(64) not null,
    `name`       varchar(64) not null,
    `state`      varchar(16) not null,
    `reason`     json                 default null,
    `share_data` json                 default null,
    `memo`       varchar(64) not null,
    `worker`     varchar(64) not null,
    `creator`    varchar(64) not null,
    `reviser`    varchar(64) not null,
    `created_at` timestamp   not null default current_timestamp,
    `updated_at` timestamp   not null default current_timestamp on update current_timestamp,
    primary key (`id`)
) engine = innodb
  default charset = utf8mb4
  collate utf8mb4_bin;

insert into id_generator(`resource`, `max_id`)
values ('async_flow', '0');

-- 9. 新增任务表
create table if not exists `async_flow_task`
(
    `id`          varchar(64) not null,
    `flow_id`     varchar(64) not null,
    `flow_name`   varchar(64) not null,
    `action_id`   varchar(64) not null,
    `action_name` varchar(64) not null,
    `params`      json                 default null,
    `retry`       json        not null,
    `depend_on`   varchar(64)          default '',
    `state`       varchar(16) not null,
    `reason`      json                 default null,
    `result`      json                 default null,
    `creator`     varchar(64) not null,
    `reviser`     varchar(64) not null,
    `created_at`  timestamp   not null default current_timestamp,
    `updated_at`  timestamp   not null default current_timestamp on update current_timestamp,
    primary key (`id`)
) engine = innodb
  default charset = utf8mb4
  collate utf8mb4_bin;

insert into id_generator(`resource`, `max_id`)
values ('async_flow_task', '0');

-- 10. EIP、硬盘、网络接口增加回收状态recycle_status字段
alter table eip
    add column `recycle_status` varchar(32) default '';
alter table network_interface
    add column `recycle_status` varchar(32) default '';

-- 11. 镜像添加操作系统类型字段
alter table image
    add column `os_type` varchar(32) default '';

-- 12. 回收记录增加回收类型字段，用于标记关联资源回收
alter table recycle_record
    add column `recycle_type` varchar(64) default '';

commit;
