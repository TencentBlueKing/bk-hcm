insert into id_generator(`resource`, `max_id`)
values ('account_bill_config', '0');

CREATE TABLE `account_bill_config`
(
    `id`                  varchar(64) not null,
    `vendor`              varchar(16) not null default '',
    `account_id`          varchar(64) not null,
    `cloud_database_name` varchar(64)          default '',
    `cloud_table_name`    varchar(64)          default '',
    `status`              tinyint unsigned default '0',
    `err_msg`             json                 default NULL,
    `extension`           json                 default NULL,
    `creator`             varchar(64)          default '',
    `reviser`             varchar(64)          default '',
    `created_at`          timestamp   not null default current_timestamp,
    `updated_at`          timestamp   not null default current_timestamp on update current_timestamp,
    primary key (`id`),
    unique key `idx_uk_vendor_account_id` (`vendor`, `account_id`)
) engine = innodb
  default charset = utf8mb4;