/*
 * TencentBlueKing is pleased to support the open source community by making
 * 蓝鲸智云 - 混合云管理平台 (BlueKing - Hybrid Cloud Management System) available.
 * Copyright (C) 2024 THL A29 Limited,
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
    SQLVER=9999,HCMVER=v9.9.9

    Notes:
    1. 修改`load_balancer_target`表: 增加ip, 以 `cloud_target_group_id`,`ip`,`port` 为唯一键
    2. 修改`target_group_listener_rule_rel`表，增加`vendor`字段
    3. 修改`ssl_cert`表，`cloud_created_time`,`cloud_expired_time`字段类型为`varchar(32)`
*/

START TRANSACTION;

/* first use default value from private_ip_address*/
alter table load_balancer_target
    add ip varchar(255) not null default (private_ip_address ->> '$[0]') not null after inst_type;
/* second remove default value definition */
alter table load_balancer_target
    modify ip varchar(255) not null default '';


alter table load_balancer_target
    drop key idx_uk_cloud_target_group_id_cloud_inst_id_port;

alter table load_balancer_target
    add constraint idx_uk_cloud_target_group_id_ip_port_cloud_inst_id
        unique (cloud_target_group_id, ip, port, cloud_inst_id);

alter table target_group_listener_rule_rel
    add vendor varchar(16) default 'tcloud' not null after id;
alter table target_group_listener_rule_rel
    modify vendor varchar(16) not null default '';

alter table ssl_cert
    modify cloud_created_time varchar(32) default '' not null;

alter table ssl_cert
    modify cloud_expired_time varchar(32) default '' not null;


CREATE OR REPLACE VIEW `hcm_version`(`hcm_ver`, `sql_ver`) AS
SELECT 'v9.9.9' as `hcm_ver`, '9999' as `sql_ver`;

COMMIT