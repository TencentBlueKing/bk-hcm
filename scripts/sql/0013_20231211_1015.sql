/*
    SQLVER=0013,HCMVER=v1.3.0

    Notes:
    1. 添加云选型方案表
*/

START TRANSACTION;

create table if not exists `cloud_selection_scheme`
(
    `id`                      varchar(64)  not null,
    `bk_biz_id`               bigint       not null,
    `name`                    varchar(255) not null,
    `biz_type`                varchar(64)  not null,
    `vendors`                 json         not null,
    `deployment_architecture` json         not null,
    `cover_ping`              double       not null,
    `composite_score`         double       not null,
    `net_score`               double       not null,
    `cost_score`              double       not null,
    `cover_rate`              double       not null,
    `user_distribution`       json         not null,
    `result_idc_ids`          json         not null,
    `creator`                 varchar(64)  not null,
    `reviser`                 varchar(64)  not null,
    `created_at`              timestamp    not null default current_timestamp,
    `updated_at`              timestamp    not null default current_timestamp on update current_timestamp,
    primary key (`id`),
    unique key `idx_uk_bk_biz_id_name` (`bk_biz_id`, `name`)
) engine = innodb
  default charset = utf8mb4
  collate utf8mb4_bin comment ='云选型方案表';

insert into id_generator(`resource`, `max_id`)
values ('cloud_selection_scheme', '0');

create table if not exists `cloud_selection_biz_type`
(
    `id`                      varchar(64) not null,
    `biz_type`                varchar(64) not null,
    `cover_ping`              double      not null,
    `deployment_architecture` json        not null,
    `creator`                 varchar(64) not null,
    `reviser`                 varchar(64) not null,
    `created_at`              timestamp   not null default current_timestamp,
    `updated_at`              timestamp   not null default current_timestamp on update current_timestamp,
    primary key (`id`),
    unique key `idx_uk_biz_type` (`biz_type`)
) engine = innodb
  default charset = utf8mb4
  collate utf8mb4_bin comment ='云选型业务类型表';

insert into id_generator(`resource`, `max_id`)
values ('cloud_selection_biz_type', '0');

create table if not exists `cloud_selection_idc`
(
    `id`         varchar(64)  not null,
    `bk_biz_id`  bigint       not null,
    `name`       varchar(255) not null,
    `vendor`     varchar(16)  not null,
    `country`    varchar(255) not null,
    `region`     varchar(255) not null,
    `creator`    varchar(64)  not null,
    `reviser`    varchar(64)  not null,
    `created_at` timestamp    not null default current_timestamp,
    `updated_at` timestamp    not null default current_timestamp on update current_timestamp,
    primary key (id),
    unique key `idx_uk_bk_biz_id_name` (`bk_biz_id`, `name`)
) engine = InnoDB
  default charset = utf8mb4
  collate utf8mb4_bin comment ='云选型IDC信息表';

insert into id_generator(`resource`, `max_id`)
values ('cloud_selection_idc', '0');

CREATE OR REPLACE VIEW `hcm_version`(`hcm_ver`, `sql_ver`) AS
SELECT 'v1.3.0' as `hcm_ver`, '0013' as `sql_ver`;

COMMIT