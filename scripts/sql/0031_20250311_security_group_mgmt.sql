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
    SQLVER=0031,HCMVER=v1.8.0

    Notes:
    1. 增加资源使用范围表
    2. 安全组增加负责人、管理业务、管理类型字段
*/

START TRANSACTION;

--  1. 增加资源使用范围表
CREATE TABLE `res_usage_biz_rel`
(
    `rel_id`         bigint unsigned NOT NULL AUTO_INCREMENT,
    `res_type`       varchar(64)     NOT NULL COMMENT '资源类型',
    `res_id`         varchar(64)     NOT NULL COMMENT '资源ID',
    `usage_biz_id`   bigint          NOT NULL COMMENT '使用业务ID',
    `res_vendor`     varchar(64)     NOT NULL DEFAULT '' COMMENT '云资源厂商',
    `res_cloud_id`   varchar(255)    NOT NULL DEFAULT '' COMMENT '云资源ID',
    `rel_creator`    varchar(64)     not null comment '创建者',
    `rel_created_at` timestamp       not null default current_timestamp comment '创建时间',
    PRIMARY KEY (`rel_id`),
    UNIQUE KEY `idx_uk_res_type_usage_biz_id_res_id` (`res_type`, `usage_biz_id`, `res_id`),
    KEY idx_res_type_res_id_usage_biz_id (res_type, res_id, usage_biz_id)
);


ALTER TABLE security_group
    ADD COLUMN mgmt_type varchar(64) NOT NULL DEFAULT '' COMMENT '管理类型' AFTER account_id;
ALTER TABLE security_group
    ADD COLUMN mgmt_biz_id bigint NOT NULL DEFAULT -1 COMMENT '管理业务ID' AFTER mgmt_type;
ALTER TABLE security_group
    ADD COLUMN manager varchar(64) NOT NULL DEFAULT '' COMMENT '负责人' AFTER mgmt_biz_id;
ALTER TABLE security_group
    ADD COLUMN bak_manager varchar(64) NOT NULL DEFAULT '' COMMENT '备份负责人' AFTER manager;

ALTER TABLE security_group_common_rel
    ADD INDEX idx_security_group_id (security_group_id);
ALTER TABLE security_group_common_rel
    CHANGE COLUMN vendor res_vendor varchar(16);

alter table tcloud_security_group_rule
    ADD INDEX
        idx_security_group_id_cloud_target_security_group_id_region (security_group_id, cloud_target_security_group_id, region);
alter table huawei_security_group_rule
    ADD INDEX
        idx_security_group_id_cloud_remote_group_id_region (security_group_id, cloud_remote_group_id, region);
alter table aws_security_group_rule
    ADD INDEX
        idx_security_group_id_cloud_target_security_group_id_region (security_group_id, cloud_target_security_group_id, region);


CREATE OR REPLACE VIEW `hcm_version`(`hcm_ver`, `sql_ver`) AS
SELECT 'v1.8.0' as `hcm_ver`, '0031' as `sql_ver`;

COMMIT;
