/*
   SQLVER=0010,HCMVER=v1.1.23

   Notes:
       1. 添加子账号表。
*/

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
