/*
表结构说明：
各类模型表字段信息主要分为：
1. 主键id                        // id_generator 生成的ID
2. 云供应商id                     // 云供应商ID (vendor)
3. 模型特定字段信息Spec            // 需要用户特殊定义的字段 (Spec)
4. 模型差异字段                   // 云资源模型差异字段 (Extension)
5. 外键id                        // 和当前模型有关联关系的模型主键id (Attachment)
6. 关联资源冗余字段                // 和当前模型有关联的子资源等其他资源字段信息 （OtherSpec）
7. 创建信息（CreatedRevision）、创建及修正信息（Revision）
注:
    1. 字段需要按照上述分类进行排序和分类。
    2. 字段说明统一参照 pkg/dal/table 目录下的数据结构定义说明。
    3. varchar字符类型实际存储大小为 Len + 存储长度大小(1/2字节)，但是索引是根据设定的varchar长度进行建立的，
    如需要对字段建立索引，注意存储消耗。varchar类型字段长度从小于255范围，扩展到大于255范围，因为记录varchar
    实际长度的字符需要从 1byte -> 2byte，会进行锁表。所以，表字段跨255范围扩展，需确认影响。
    4. 各类表的name以及namespace字段采用varchar第一范围最大值(255)进行存储，memo字段采用第二范围最小值(256)进行存储，
    非必要禁止跨界。
    5. HCM云资源主键ID后缀统一为'_id', 云资源云上主键ID后缀统一为'_cid'。e.g: vpc_id(hcm系统vpc唯一ID) vpc_cid(云上vpc唯一ID)
    6. 云资源关联关系统一通过关联关系表存储，避免后期对接云的资源关联关系和其他云不一致，导致db数据迁移，但仅限云资源的关联关系，
    其他场景根据实际情况去设置。且关联关系表中仅存储hcm云资源唯一ID即可。关联关系表名为 aTable_bTable_rel，e.g: cvm_vpc_rel。
*/
create database if not exists hcm;

use hcm;

create table if not exists `id_generator`
(
    `resource` varchar(64) not null,
    `max_id`   varchar(64) not null,
    primary key (`resource`)
    ) engine = innodb
    default charset = utf8mb4;

insert into id_generator(`resource`, `max_id`)
values ('account', '0'),
       ('security_group', '0'),
       ('tcloud_security_group_rule', '0'),
       ('aws_security_group_rule', '0'),
       ('azure_security_group_rule', '0'),
       ('huawei_security_group_rule', '0'),
       ('gcp_firewall_rule', '0'),
       ('vpc', '0'),
       ('subnet', '0'),
       ('disk', '0')
ON DUPLICATE KEY UPDATE resource=resource;

create table if not exists `audit`
(
    `id`                      bigint(1) unsigned not null auto_increment,
    `res_id`                  varchar(64)                 default '',
    `cloud_res_id`            varchar(255)                default '',
    `res_name`                varchar(255)                default '',
    `res_type`                varchar(50)        not null,
    `action`                  varchar(20)        not null,
    `bk_biz_id`               bigint(1)          not null default -1,
    `vendor`                  varchar(16)       default '',
    `account_id`              varchar(64)                 default '',
    `operator`                varchar(64)        not null,
    `source`                  varchar(20)        not null,
    `rid`                     varchar(64)        not null,
    `app_code`                varchar(64)                 default '',
    `detail`                  json                        default null,
    `created_at`              timestamp          not null default current_timestamp,
    primary key (`id`)
) engine = innodb
  default charset = utf8mb4;

create table if not exists `account`
(
    `id`             varchar(64) not null,
    `vendor`         varchar(16) not null,
    `name`           varchar(64) not null,
    `managers`       json        not null,
    `department_ids` json     not null,
    `type`           varchar(32) not null,
    `site`           varchar(32) not null,
    `sync_status`    varchar(32) not null,
    `price`          varchar(16)          default '',
    `price_unit`     varchar(8)           default '',
    `memo`           varchar(255)         default '',
    `extension`      json        not null,
    `creator`        varchar(64) not null,
    `reviser`        varchar(64) not null,
    `created_at`     timestamp   not null default current_timestamp,
    `updated_at`     timestamp   not null default current_timestamp on update current_timestamp,
    primary key (`id`),
    unique key `idx_uk_name` (`name`)
    ) engine = innodb
    default charset = utf8mb4;

create table if not exists `account_biz_rel`
(
    `id`         bigint(1) unsigned not null auto_increment,
    `bk_biz_id`  bigint(1)          not null,
    `account_id` varchar(64)        not null,
    `creator`    varchar(64)        not null,
    `created_at` timestamp          not null default current_timestamp,
    primary key (`id`),
    unique key `idx_uk_bk_biz_id_account_id` (`bk_biz_id`, `account_id`)
    ) engine = innodb
    default charset = utf8mb4;

create table if not exists `security_group`
(
    `id`                      varchar(64)  not null,
    `vendor`                  varchar(16)  not null,
    `cloud_id`                varchar(255) not null,
    `bk_biz_id`               bigint(1)    not null default -1,
    `region`                  varchar(20)  not null,
    `name`                    varchar(60)  not null,
    `account_id`              varchar(64)  not null,
    `memo`                    varchar(255)          default '',
    `association_template_id` varchar(64)           default 0,
    `extension`               json         not null,
    `creator`                 varchar(64)  not null,
    `reviser`                 varchar(64)  not null,
    `created_at`              timestamp    not null default current_timestamp,
    `updated_at`              timestamp    not null default current_timestamp on update current_timestamp,
    primary key (`id`),
    unique key `idx_uk_cloud_id_vendor` (`cloud_id`, `vendor`)
) engine = innodb
  default charset = utf8mb4;

# vpc_security_group_rel is only used of aws.
create table if not exists `vpc_security_group_rel`
(
    `id`                bigint(1) unsigned not null auto_increment,
    `vpc_id`            varchar(64)        not null,
    `security_group_id` varchar(64)        not null,
    `creator`           varchar(64)        not null,
    `created_at`        timestamp          not null default current_timestamp,
    primary key (`id`),
    unique key `idx_uk_vpc_id_security_group_id` (`vpc_id`, `security_group_id`)
) engine = innodb
  default charset = utf8mb4;

# security_group_network_interface_rel is only used of azure.
create table if not exists `security_group_network_interface_rel`
(
    `id`                   bigint(1) unsigned not null auto_increment,
    `security_group_id`    varchar(64)        not null,
    `network_interface_id` varchar(64)        not null,
    `creator`              varchar(64)        not null,
    `created_at`           timestamp          not null default current_timestamp,
    primary key (`id`),
    unique key `idx_uk_security_group_id_network_interface_id` (`security_group_id`, `network_interface_id`)
) engine = innodb
  default charset = utf8mb4;

# security_group_subnet_rel is only used of azure.
create table if not exists `security_group_subnet_rel`
(
    `id`                bigint(1) unsigned not null auto_increment,
    `security_group_id` varchar(64)        not null,
    `subnet_id`         varchar(64)        not null,
    `creator`           varchar(64)        not null,
    `created_at`        timestamp          not null default current_timestamp,
    primary key (`id`),
    unique key `idx_uk_security_group_id_subnet_id` (`security_group_id`, `subnet_id`)
) engine = innodb
  default charset = utf8mb4;

create table if not exists `tcloud_security_group_rule`
(
    `id`                             varchar(64)  not null,
    `cloud_policy_index`             bigint(1)    not null,
    `type`                           varchar(20)  not null,
    `cloud_security_group_id`        varchar(255) not null,
    `security_group_id`              varchar(64)  not null,
    `account_id`                     varchar(64)  not null,
    `region`                         varchar(20)  not null,
    `version`                        varchar(255) not null,
    `action`                         varchar(10)  not null,
    `protocol`                       varchar(10)           default null,
    `port`                           varchar(255)          default null,
    `cloud_service_id`               varchar(255)          default null,
    `cloud_service_group_id`         varchar(255)          default null,
    `ipv4_cidr`                      varchar(255)          default null,
    `ipv6_cidr`                      varchar(255)          default null,
    `cloud_target_security_group_id` varchar(255)          default null,
    `cloud_address_id`               varchar(255)          default null,
    `cloud_address_group_id`         varchar(255)          default null,
    `memo`                           varchar(60)           default null,
    `creator`                        varchar(64)  not null,
    `reviser`                        varchar(64)  not null,
    `created_at`                     timestamp    not null default current_timestamp,
    `updated_at`                     timestamp    not null default current_timestamp on update current_timestamp,
    primary key (`id`),
    unique key `idx_uk_cloud_security_group_id_cloud_policy_index_type` (`cloud_security_group_id`, `cloud_policy_index`, `type`)
    ) engine = innodb
    default charset = utf8mb4;

create table if not exists `aws_security_group_rule`
(
    `id`                             varchar(64)  not null,
    `cloud_id`                       varchar(255) not null,
    `cloud_security_group_id`        varchar(255) not null,
    `cloud_group_owner_id`           varchar(255) not null,
    `account_id`                     varchar(64)  not null,
    `region`                         varchar(20)  not null,
    `security_group_id`              varchar(64)  not null,
    `type`                           varchar(20)  not null,
    `ipv4_cidr`                      varchar(255)          default null,
    `ipv6_cidr`                      varchar(255)          default null,
    `memo`                           varchar(60)           default null,
    `from_port`                      int(1) unsigned       default 0,
    `to_port`                        int(1) unsigned       default 0,
    `protocol`                       varchar(10)           default null,
    `cloud_prefix_list_id`           varchar(255)          default null,
    `cloud_target_security_group_id` varchar(255)          default null,
    `creator`                        varchar(64)  not null,
    `reviser`                        varchar(64)  not null,
    `created_at`                     timestamp    not null default current_timestamp,
    `updated_at`                     timestamp    not null default current_timestamp on update current_timestamp,
    primary key (`id`),
    unique key `idx_uk_cloud_id` (`cloud_id`)
    ) engine = innodb
    default charset = utf8mb4;

create table if not exists `huawei_security_group_rule`
(
    `id`                            varchar(64)  not null,
    `cloud_id`                      varchar(255) not null,
    `type`                          varchar(20)  not null,
    `cloud_security_group_id`       varchar(255) not null,
    `security_group_id`             varchar(64)  not null,
    `account_id`                    varchar(64)  not null,
    `region`                        varchar(20)  not null,
    `action`                        varchar(10)  not null,
    `cloud_project_id`              varchar(255)          default '',
    `memo`                          varchar(255)          default '',
    `protocol`                      varchar(10)           default '',
    `ethertype`                     varchar(10)           default '',
    `cloud_remote_group_id`         varchar(255)          default '',
    `remote_ip_prefix`              varchar(255)          default '',
    `cloud_remote_address_group_id` varchar(255)          default '',
    `port`                          varchar(255)          default '',
    `priority`                      int(1) unsigned       default 0,
    `creator`                       varchar(64)  not null,
    `reviser`                       varchar(64)  not null,
    `created_at`                    timestamp    not null default current_timestamp,
    `updated_at`                    timestamp    not null default current_timestamp on update current_timestamp,
    primary key (`id`),
    unique key `idx_uk_cloud_id` (`cloud_id`)
    ) engine = innodb
    default charset = utf8mb4;

create table if not exists `azure_security_group_rule`
(
    `id`                                   varchar(64)  not null,
    `cloud_id`                             varchar(255) not null,
    `cloud_security_group_id`              varchar(255) not null,
    `account_id`                           varchar(64)  not null,
    `security_group_id`                    varchar(64)  not null,
    `type`                                 varchar(20)  not null,
    `region`                               varchar(20)  not null,
    `provisioning_state`                   varchar(20)  not null,
    `etag`                                 varchar(255)          default '',
    `name`                                 varchar(255)          default '',
    `memo`                                 varchar(140)          default '',
    `destination_address_prefix`           varchar(255)          default '',
    `destination_address_prefixes`         json                  default null,
    `cloud_destination_security_group_ids` json                  default null,
    `destination_port_range`               varchar(255)          default '',
    `destination_port_ranges`              json                  default null,
    `protocol`                             varchar(10)           default '',
    `source_address_prefix`                varchar(255)          default '',
    `source_address_prefixes`              json                  default null,
    `cloud_source_security_group_ids`      json                  default null,
    `source_port_range`                    varchar(255)          default '',
    `source_port_ranges`                   json                  default null,
    `priority`                             bigint(1)             default 0,
    `access`                               varchar(20)           default '',
    `creator`                              varchar(64)  not null,
    `reviser`                              varchar(64)  not null,
    `created_at`                           timestamp    not null default current_timestamp,
    `updated_at`                           timestamp    not null default current_timestamp on update current_timestamp,
    primary key (`id`),
    unique key `idx_uk_cloud_id` (`cloud_id`),
    unique key `idx_uk_name` (`name`)
    ) engine = innodb
    default charset = utf8mb4;

create table if not exists `security_group_tag`
(
    `id`                varchar(64)  not null,
    `security_group_id` varchar(64)  not null,
    `key`               varchar(255) not null,
    `value`             varchar(255)          default '',
    `creator`           varchar(64)  not null,
    `reviser`           varchar(64)  not null,
    `created_at`        timestamp    not null default current_timestamp,
    `updated_at`        timestamp    not null default current_timestamp on update current_timestamp,
    primary key (`id`)
    ) engine = innodb
    default charset = utf8mb4;

create table if not exists `gcp_firewall_rule`
(
    `id`                      varchar(64)  not null,
    `cloud_id`                varchar(255) not null,
    `name`                    varchar(62)           default '',
    `priority`                bigint(1)             default 0,
    `memo`                    varchar(2048)         default '',
    `cloud_vpc_id`            varchar(255)          default '',
    `vpc_id`                  varchar(64)           default '',
    `account_id`              varchar(64)           default '',
    `source_ranges`           json                  default null,
    `destination_ranges`      json                  default null,
    `source_tags`             json                  default null,
    `target_tags`             json                  default null,
    `source_service_accounts` json                  default null,
    `target_service_accounts` json                  default null,
    `denied`                  json                  default null,
    `allowed`                 json                  default null,
    `type`                    varchar(20)           default '',
    `log_enable`              boolean               default false,
    `disabled`                boolean               default false,
    `self_link`               varchar(255)          default '',
    `bk_biz_id`               bigint(1)    not null default -1,
    `creator`                 varchar(64)  not null,
    `reviser`                 varchar(64)  not null,
    `created_at`              timestamp    not null default current_timestamp,
    `updated_at`              timestamp    not null default current_timestamp on update current_timestamp,
    primary key (`id`),
    unique key `idx_uk_cloud_id` (`cloud_id`),
    unique key `idx_uk_name` (`name`)
) engine = innodb
  default charset = utf8mb4;

create table if not exists `vpc`
(
    `id`          varchar(64)  not null,
    `vendor`      varchar(32)  not null,
    `account_id`  varchar(64)  not null,
    `cloud_id`    varchar(255) not null,
    `name`        varchar(128) not null,
    `category`    varchar(32)  not null,
    `memo`        varchar(255)          default '',
    `bk_cloud_id` bigint(1)             default -1,
    `bk_biz_id`   bigint(1)    not null default -1,
    # Extension
    `extension`   json         not null,
    # Revision
    `creator`     varchar(64)  not null,
    `reviser`     varchar(64)  not null,
    `created_at`  timestamp    not null default current_timestamp,
    `updated_at`  timestamp    not null default current_timestamp on update current_timestamp,
    primary key (`id`),
    unique key `idx_uk_cloud_id_vendor` (`cloud_id`, `vendor`)
) engine = innodb
  default charset = utf8mb4;

create table if not exists `subnet`
(
    `id`           varchar(64)  not null,
    `vendor`       varchar(32)  not null,
    `account_id`   varchar(64)  not null,
    `cloud_vpc_id` varchar(255) not null,
    `cloud_id`     varchar(255) not null,
    `name`         varchar(128) not null,
    `ipv4_cidr`    json         not null,
    `ipv6_cidr`    json         not null,
    `memo`         varchar(255)          default '',
    `vpc_id`       varchar(64)  not null,
    `bk_biz_id`    bigint(1)    not null default -1,
    # Extension
    `extension`    json         not null,
    # Revision
    `creator`      varchar(64)  not null,
    `reviser`      varchar(64)  not null,
    `created_at`   timestamp    not null default current_timestamp,
    `updated_at`   timestamp    not null default current_timestamp on update current_timestamp,
    primary key (`id`),
    unique key `idx_uk_cloud_id_vendor` (`cloud_id`, `vendor`)
) engine = innodb
  default charset = utf8mb4;

create table if not exists `disk`
(
    `id`          varchar(64)        not null,
    `vendor`      varchar(16)        not null,
    `name`        varchar(128)       not null,
    `account_id`  varchar(64)        not null,
    `cloud_id`    varchar(128)       not null,
    `bk_biz_id`   bigint(1)          not null default -1,
    `region`      varchar(128)       not null,
    `zone`        varchar(128)       not null,
    `disk_size`   bigint(1) unsigned not null,
    `disk_type`   varchar(128)       not null,
    `disk_status` varchar(128)       not null,
    `memo`        varchar(255)                default '',
    `extension`   json               not null,
    `creator`     varchar(64)        not null,
    `reviser`     varchar(64)        not null,
    `created_at`  timestamp          not null default current_timestamp,
    `updated_at`  timestamp          not null default current_timestamp on update current_timestamp,
    primary key (`id`)
) engine = innodb
  default charset = utf8mb4;
