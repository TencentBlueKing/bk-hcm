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
    SQLVER=9999,HCMVER=v9.9.9

    Notes:
    1. 增加资源使用范围表
    2. 安全组增加负责人、管理业务、管理类型字段
*/

START TRANSACTION;

--  1. 增加资源使用范围表
CREATE TABLE `res_usage_biz_rel`
(
    `rel_id`       bigint unsigned NOT NULL AUTO_INCREMENT,
    `res_type`     varchar(64)     NOT NULL COMMENT '资源类型',
    `res_id`       varchar(64)     NOT NULL COMMENT '资源ID',
    `usage_biz_id` bigint          NOT NULL COMMENT '使用业务ID',
    `creator`      varchar(64)     not null comment '创建者',
    `reviser`      varchar(64)     not null comment '更新者',
    `created_at`   timestamp       not null default current_timestamp comment '创建时间',
    `updated_at`   timestamp       not null default current_timestamp on update current_timestamp comment '更新时间',
    PRIMARY KEY (`rel_id`),
    UNIQUE KEY `idx_uk_res_type_usage_biz_id_res_id` (`res_type`, `usage_biz_id`, `res_id`),
    KEY idx_res_type_res_id_usage_biz_id (res_type, res_id, usage_biz_id)
);

alter table security_group
    add column mgmt_type varchar(64) NOT NULL DEFAULT '' COMMENT '管理类型' AFTER account_id;
alter table security_group
    add column mgmt_biz_id bigint NOT NULL DEFAULT -1 COMMENT '管理业务ID' AFTER mgmt_type;
alter table security_group
    add column manager varchar(64) NOT NULL DEFAULT '' COMMENT '负责人' AFTER mgmt_biz_id;
alter table security_group
    add column bak_manager varchar(64) NOT NULL DEFAULT '' COMMENT '备份负责人' AFTER manager;



CREATE OR REPLACE VIEW `hcm_version`(`hcm_ver`, `sql_ver`) AS
SELECT 'v9.9.9' as `hcm_ver`, '9999' as `sql_ver`;

COMMIT;
