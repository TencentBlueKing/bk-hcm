/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2022 THL A29 Limited,
 * a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on
 * an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the
 * specific language governing permissions and limitations under the License.
 *
 * We undertake not to change the open source license (MIT license) applicable
 *
 * to the current version of the project delivered to anyone in the future.
 */

/*
    SQLVER=0018,HCMVER=v1.5.0

    Notes:
    1. 支持腾讯云负载均衡
*/

START TRANSACTION;


--  1. 负载均衡表
create table `load_balancer`
(
    `id`                     varchar(64)  not null,
    `cloud_id`               varchar(255) not null,
    `name`                   varchar(255) not null,
    `vendor`                 varchar(16)  not null,

    `bk_biz_id`              bigint       not null default -1,
    `account_id`             varchar(64)  not null,

    `region`                 varchar(20)  not null,
    `zones`                  json         not null,
    `backup_zones`           json                  default null,
    `lb_type`                varchar(64)  not null,
    `ip_version`             varchar(64)  not null default '',

    `vpc_id`                 varchar(255) not null,
    `cloud_vpc_id`           varchar(255) not null,
    `cloud_subnet_id`        varchar(255) not null,
    `subnet_id`              varchar(255) not null,

    `private_ipv4_addresses` json                  default null,
    `private_ipv6_addresses` json                  default null,
    `public_ipv4_addresses`  json                  default null,
    `public_ipv6_addresses`  json                  default null,

    `domain`                 varchar(255) not null,
    `status`                 varchar(64)  not null,
    `memo`                   varchar(255)          default '',
    `cloud_created_time`     varchar(64)           default '',
    `cloud_status_time`      varchar(64)           default '',
    `cloud_expired_time`     varchar(64)           default '',
    `extension`              json         not null,

    `creator`                varchar(64)  not null,
    `reviser`                varchar(64)  not null,
    `created_at`             timestamp    not null default current_timestamp,
    `updated_at`             timestamp    not null default current_timestamp on update current_timestamp,
    primary key (`id`),
    unique key `idx_uk_cloud_id_vendor` (`cloud_id`, `vendor`)
) engine = innodb
  default charset = utf8mb4
  collate = utf8mb4_bin comment ='负载均衡表';


-- 2. 通用安全组资源关联表
create table `security_group_common_rel`
(
    `id`                bigint unsigned not null auto_increment,
    `vendor`            varchar(16)     not null,
    `res_id`            varchar(64)     not null,
    `res_type`          varchar(64)     not null,
    `security_group_id` varchar(64)     not null,
    `priority`          int             not null,

    `creator`           varchar(64)     not null,
    `reviser`           varchar(64)     not null,
    `created_at`        timestamp       not null default current_timestamp,
    `updated_at`        timestamp       not null default current_timestamp on update current_timestamp,
    primary key (`id`),
    unique key `idx_uk_vendor_res_type_res_id_sg_id`
        (`vendor`, `res_type`, `res_id`, `security_group_id`)
) engine = innodb
  default charset = utf8mb4
  collate = utf8mb4_bin comment ='通用安全组资源关联表';


-- 3. 负载均衡监听器
create table `load_balancer_listener`
(
    `id`             varchar(64)  not null,
    `cloud_id`       varchar(255) not null,
    `name`           varchar(255) not null,
    `vendor`         varchar(16)  not null,

    `account_id`     varchar(64)  not null,
    `bk_biz_id`      bigint(1)    not null default -1,

    `lb_id`          varchar(255) not null,
    `cloud_lb_id`    varchar(255) not null,
    `protocol`       varchar(64)  not null,
    `port`           bigint       not null,
    `default_domain` varchar(255)          default null,
    `zones`          json,
    `sni_switch`     bigint                default 0,
    `extension`      json                  default null,
    `memo`           varchar(255)          default '',

    `creator`        varchar(64)  not null,
    `reviser`        varchar(64)  not null,
    `created_at`     timestamp    not null default current_timestamp,
    `updated_at`     timestamp    not null default current_timestamp on update current_timestamp,
    primary key (`id`),
    unique key `idx_uk_cloud_id_vendor` (`cloud_id`, `vendor`),
    key `idx_lb_id`(`lb_id`)

) engine = innodb
  default charset = utf8mb4
  collate = utf8mb4_bin comment ='负载均衡监听器';

-- 4. 负载均衡七层规则
create table `tcloud_lb_url_rule`
(
    `id`                    varchar(64)  not null,
    `cloud_id`              varchar(255) not null,
    `name`                  varchar(255) not null,
    `rule_type`             varchar(64)  not null,

    `lb_id`                 varchar(255) not null,
    `cloud_lb_id`           varchar(255) not null,
    `lbl_id`                varchar(255) not null,
    `cloud_lbl_id`          varchar(255) not null,
    `target_group_id`       varchar(255)          default '',
    `cloud_target_group_id` varchar(255)          default '',

    `domain`                varchar(255)          default '',
    `url`                   varchar(255)          default '',
    `scheduler`             varchar(64)  not null,
    `session_type`          varchar(64)           default '',
    `session_expire`        bigint                default 0,
    `health_check`          json                  default null,
    `certificate`           json                  default null,
    `memo`                  varchar(255)          default '',


    `creator`               varchar(64)  not null,
    `reviser`               varchar(64)  not null,
    `created_at`            timestamp    not null default current_timestamp,
    `updated_at`            timestamp    not null default current_timestamp on update current_timestamp,
    primary key (`id`),
    unique key `idx_uk_cloud_id_lbl_id` (`cloud_id`, `lbl_id`)
) engine = innodb
  default charset = utf8mb4
  collate = utf8mb4_bin comment ='负载均衡七层规则';


-- 5. 负载均衡目标
create table `load_balancer_target`
(
    `id`                    varchar(64)  not null,
    `account_id`            varchar(64)  not null,

    `inst_type`             varchar(255) not null,
    `inst_id`               varchar(255) not null,
    `cloud_inst_id`         varchar(255) not null,
    `inst_name`             varchar(255) not null,

    `target_group_id`       varchar(255)          default '',
    `cloud_target_group_id` varchar(255)          default '',

    `port`                  bigint       not null,
    `weight`                bigint       not null,
    `private_ip_address`    json         not null,
    `public_ip_address`     json         not null,
    `cloud_vpc_ids`         json                  default null,
    `zone`                  varchar(255) not null default '',
    `memo`                  varchar(255)          default '',

    `creator`               varchar(64)  not null,
    `reviser`               varchar(64)  not null,
    `created_at`            timestamp    not null default current_timestamp,
    `updated_at`            timestamp    not null default current_timestamp on update current_timestamp,

    primary key (`id`),
    unique key `idx_uk_cloud_target_group_id_cloud_inst_id_port` (`cloud_target_group_id`, `cloud_inst_id`, `port`)
) engine = innodb
  default charset = utf8mb4
  collate = utf8mb4_bin comment ='负载均衡目标';


-- 6. 负载均衡目标组
create table `load_balancer_target_group`
(
    `id`                varchar(64)  not null,
    `cloud_id`          varchar(255) not null,
    `name`              varchar(255) not null,
    `vendor`            varchar(16)  not null,
    `target_group_type` varchar(16)  not null,

    `account_id`        varchar(64)  not null,
    `bk_biz_id`         bigint(1)    not null default -1,
    `vpc_id`            varchar(255) not null,
    `cloud_vpc_id`      varchar(255) not null,

    `region`            varchar(20)  not null,
    `protocol`          varchar(64)  not null,
    `port`              bigint       not null,
    `weight`            bigint       not null default -1,
    `health_check`      json                  default null,

    `memo`              varchar(255)          default '',
    `extension`         json         not null default ('{}'),

    `creator`           varchar(64)  not null,
    `reviser`           varchar(64)  not null,
    `created_at`        timestamp    not null default current_timestamp,
    `updated_at`        timestamp    not null default current_timestamp on update current_timestamp,
    primary key (`id`),
    unique key `idx_uk_cloud_id_vendor` (`cloud_id`, `vendor`)
) engine = innodb
  default charset = utf8mb4
  collate = utf8mb4_bin comment ='负载均衡目标组';

-- 7. 目标组监听器关系表
create table `target_group_listener_rule_rel`
(
    `id`                     varchar(64) not null,
    `listener_rule_id`       varchar(64) not null,
    `listener_rule_type`     varchar(64) not null,
    `cloud_listener_rule_id` varchar(64) not null,
    `target_group_id`        varchar(64) not null,
    `cloud_target_group_id`  varchar(64) not null,
    `lb_id`                  varchar(64) not null,
    `cloud_lb_id`            varchar(64) not null,
    `lbl_id`                 varchar(64) not null,
    `cloud_lbl_id`           varchar(64) not null,
    `binding_status`         varchar(64) not null,
    `detail`                 json                 default null,

    `creator`                varchar(64) not null,
    `reviser`                varchar(64) not null,
    `created_at`             timestamp   not null default current_timestamp,
    `updated_at`             timestamp   not null default current_timestamp on update current_timestamp,
    primary key (`id`),
    unique key `idx_uk_target_group_id_listener_rule_id_listener_rule_type` (`target_group_id`, `listener_rule_id`, `listener_rule_type`),
    key `idx_lbid_binding_status_rule_type`(`lb_id`, `binding_status`, `listener_rule_type`)
) engine = innodb
  default charset = utf8mb4
  collate = utf8mb4_bin comment ='目标组监听器关系表';

-- 8. 资源与异步任务关系表
create table `resource_flow_rel`
(
    `id`         varchar(64) not null,
    `res_id`     varchar(64) not null,
    `res_type`   varchar(64) not null,
    `flow_id`    varchar(64) not null,
    `task_type`  varchar(64) not null,
    `status`     varchar(64) not null,

    `creator`    varchar(64) not null,
    `reviser`    varchar(64) not null,
    `created_at` timestamp   not null default current_timestamp,
    `updated_at` timestamp   not null default current_timestamp on update current_timestamp,
    primary key (`id`),
    unique key `idx_uk_res_id_flow_id` (`res_id`, `res_type`, `flow_id`)
) engine = innodb
  default charset = utf8mb4
  collate = utf8mb4_bin comment ='资源与异步任务的关系表';


-- 9. 资源与异步任务锁定的表
create table `resource_flow_lock`
(
    `res_type`   varchar(64) not null,
    `res_id`     varchar(64) not null,
    `owner`      varchar(64) not null,

    `creator`    varchar(64) not null,
    `reviser`    varchar(64) not null,
    `created_at` timestamp   not null default current_timestamp,
    `updated_at` timestamp   not null default current_timestamp on update current_timestamp,
    primary key (`res_type`, `res_id`)
) engine = innodb
  default charset = utf8mb4
  collate = utf8mb4_bin comment ='资源与异步任务锁定的表';

insert into id_generator(`resource`, `max_id`)
values ('load_balancer', '0'),
       ('security_group_common_rel', '0'),
       ('load_balancer_listener', '0'),
       ('tcloud_lb_url_rule', '0'),
       ('load_balancer_target', '0'),
       ('load_balancer_target_group', '0'),
       ('target_group_listener_rule_rel', '0'),
       ('resource_flow_rel', '0');


CREATE OR REPLACE VIEW `hcm_version`(`hcm_ver`, `sql_ver`) AS
SELECT 'v1.5.0' as `hcm_ver`, '0018' as `sql_ver`;

COMMIT