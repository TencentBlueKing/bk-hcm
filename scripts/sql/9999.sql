/*
    SQLVER=9999,HCMVER=v9.9.9

    Notes:
        1. 添加子账号表。
        2. 修复aws_security_group_rule表memo字段错误的非空约束
        3. aws_region表增加account_id字段，独立存储aws每个账号的地域信息
        4. 添加账号同步详情表
*/
start transaction;

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
  COLLATE utf8mb4_bin;

insert into id_generator(`resource`, `max_id`)
values ('sub_account', '0');

alter table aws_security_group_rule
    modify column memo varchar(255) default '';

delete
from aws_region;
alter table aws_region
    add column account_id varchar(64) not null;

alter table aws_region
    drop index `idx_uk_region_id_status`;

alter table aws_region
    add unique key `idx_uk_account_id_region_id` (`account_id`, `region_id`);

CREATE OR REPLACE VIEW `hcm_version`(`hcm_ver`, `sql_ver`) AS
SELECT 'v9.9.9' as `hcm_ver`, '9999' as `sql_ver`;

create table if not exists `account_sync_detail`
(
    `id`         varchar(64)  not null,
    `vendor`     varchar(16)  not null,
    `account_id` varchar(64)  not null,
    `res_name`   varchar(64)  not null,
    `res_status` varchar(64)  not null,
    `res_end_time` varchar(64)  default '',
    `res_failed_reason` json  default null,
    `creator`    varchar(64)  not null,
    `reviser`    varchar(64)  not null,
    `created_at` timestamp    not null default current_timestamp,
    `updated_at` timestamp    not null default current_timestamp on update current_timestamp,
    primary key (`id`),
    unique key `idx_uk_vendor_account_id_res_name` (`vendor`, `account_id`, `res_name`)
) engine = innodb
  default charset = utf8mb4
  collate utf8mb4_bin;

insert into id_generator(`resource`, `max_id`) values ('account_sync_detail', '0');

commit;
