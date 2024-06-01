/*
    SQLVER=0017,HCMVER=v1.5.0

    Notes:
    1. 添加证书托管表
*/

START TRANSACTION;

create table if not exists `ssl_cert`
(
    `id`                  varchar(64)  not null,
    `cloud_id`            varchar(255) not null,
    `name`                varchar(255) not null,
    `vendor`              varchar(16)  not null,
    `bk_biz_id`           bigint       not null default '-1',
    `account_id`          varchar(64)  not null,
    `domain`              json         not null,
    `cert_type`           varchar(16)  not null,
    `cert_status`         varchar(64)  not null,
    `encrypt_algorithm`   varchar(64) not null default '',
    `cloud_created_time`  timestamp    not null default current_timestamp,
    `cloud_expired_time`  timestamp    not null default current_timestamp,
    `memo`                varchar(255)          default '',
    `creator`             varchar(64)  not null,
    `reviser`             varchar(64)  not null,
    `created_at`          timestamp    not null default current_timestamp,
    `updated_at`          timestamp    not null default current_timestamp on update current_timestamp,
    primary key (`id`),
    unique key `idx_uk_bk_biz_id_cloud_id` (`bk_biz_id`, `cloud_id`)
) engine = innodb
  default charset = utf8mb4
  collate utf8mb4_bin comment ='证书托管表';

insert into id_generator(`resource`, `max_id`)
values ('ssl_cert', '0');

CREATE OR REPLACE VIEW `hcm_version`(`hcm_ver`, `sql_ver`) AS
SELECT 'v1.5.0' as `hcm_ver`, '0017' as `sql_ver`;

COMMIT