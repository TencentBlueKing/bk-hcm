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
    1. 新增 CLB 独占集群表 load_balancer_exclusive_cluster
*/

START TRANSACTION;

-- 1. CLB独占集群表
CREATE TABLE IF NOT EXISTS `load_balancer_exclusive_cluster`
(
    `id`                 varchar(64)  NOT NULL COMMENT 'HCM主键ID',
    `cloud_id`           varchar(64)  NOT NULL COMMENT '云上集群ID，如tgw-38feq8c6',
    `name`               varchar(255) NOT NULL DEFAULT '' COMMENT '集群名称',
    `vendor`             varchar(16)  NOT NULL COMMENT '云厂商',
    `account_id`         varchar(64)  NOT NULL COMMENT '账号ID',
    `bk_biz_id`          bigint       NOT NULL DEFAULT '-1' COMMENT '业务ID，-1表示未分配',
    `region`             varchar(20)  NOT NULL COMMENT '地域',
    `zone`               varchar(64)  NOT NULL DEFAULT '' COMMENT '可用区',
    `cluster_type`       varchar(16)  NOT NULL COMMENT '集群类型：TGW四层/STGW七层/VPCGW内网',
    `cluster_tag`        varchar(128) NOT NULL DEFAULT '' COMMENT '集群标签，空表示未打标签',
    `network`            varchar(16)  NOT NULL DEFAULT '' COMMENT '网络类型：Public/Private/Hybrid',
    `isp`                varchar(16)  NOT NULL DEFAULT '' COMMENT '运营商：BGP/CMCC/CUCC/CTCC/INTERNAL/MIX',
    `egress`             varchar(64)  NOT NULL DEFAULT '' COMMENT '网络出口，如center_egress1',
    `ip_version`         varchar(16)  NOT NULL DEFAULT '' COMMENT 'IP版本',
    `max_conn`           bigint                DEFAULT NULL COMMENT '最大连接数，STGW可能无值',
    `clb_resource_count` bigint       NOT NULL DEFAULT '0' COMMENT '集群内已有CLB实例数',
    `extension`          json         NOT NULL COMMENT '云上扩展字段',
    `memo`               varchar(255)          DEFAULT '' COMMENT '备注',
    `tenant_id`          varchar(64)  NOT NULL DEFAULT 'default' COMMENT '租户ID',
    `creator`            varchar(64)  NOT NULL COMMENT '创建者',
    `reviser`            varchar(64)  NOT NULL COMMENT '更新者',
    `created_at`         timestamp    NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at`         timestamp    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `idx_uk_cloud_id_account_id_tenant_id` (`cloud_id`, `account_id`, `tenant_id`),
    KEY `idx_bk_biz_id` (`bk_biz_id`),
    KEY `idx_cluster_tag` (`cluster_tag`),
    KEY `idx_region_cluster_type` (`region`, `cluster_type`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4
  COLLATE = utf8mb4_bin COMMENT ='CLB独占集群表';

INSERT INTO id_generator(`resource`, `max_id`)
VALUES ('load_balancer_exclusive_cluster', '0');

CREATE OR REPLACE VIEW `hcm_version`(`hcm_ver`, `sql_ver`) AS
SELECT 'v9.9.9' as `hcm_ver`, '9999' as `sql_ver`;

COMMIT;
