/*
    SQLVER=9999,HCMVER=v9.9.9

    Notes:
        1. 添加自研云配额表
        2. 添加业务配额表
*/

-- TODO: 补充sql版本管理信息
-- 自研云配额表
create table if not exists `tcloud_private_quota`
(
    `id`                     varchar(64) not null,
    `account_id`             varchar(64) not null,
    `cloud_id`               bigint      not null,
    `platform`               varchar(64) not null,
    `city`                   varchar(64) not null,
    `zone`                   varchar(64) not null,
    `zone_name`              varchar(64) not null,
    `total_quota`            bigint      not null,
    `user_quota`             bigint      not null,
    `resource_type`          varchar(64) not null,
    `cvm_amount`             bigint      not null,
    `available_quota`        bigint      not null,
    `urgent_user_quota`      bigint      not null,
    `urgent_cvm_amount`      bigint      not null,
    `urgent_available_quota` bigint      not null,
    `uin`                    varchar(64) not null,
    `instance_specs`         varchar(64) not null,
    `instance_types`         varchar(64) not null,
    `creator`                varchar(64) not null,
    `reviser`                varchar(64) not null,
    `created_at`             timestamp   not null default current_timestamp,
    `updated_at`             timestamp   not null default current_timestamp on update current_timestamp,
    primary key (`id`)
) engine = innodb
  default charset = utf8mb4
  collate utf8mb4_bin;

-- 业务额度
create table if not exists `biz_quota`
(
    `id`             varchar(64) not null,
    `cloud_quota_id` varchar(64) not null,
    `account_id`     varchar(64) not null,
    `bk_biz_id`      bigint      not null,
    `res_type`       varchar(64) not null,
    `vendor`         varchar(64) not null,
    `region`         varchar(64)          default '',
    `zone`           varchar(64)          default '',
    `levels`         json        not null,
    `dimension`      json        not null,
    `memo`           varchar(255)         default '',
    `creator`        varchar(64) not null,
    `reviser`        varchar(64) not null,
    `created_at`     timestamp   not null default current_timestamp,
    `updated_at`     timestamp   not null default current_timestamp on update current_timestamp,
    primary key (`id`)
) engine = innodb
  default charset = utf8mb4
  collate utf8mb4_bin;
