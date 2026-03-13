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
    1. 新增 res_plan_demand_gpu_order 表
    2. 新增 res_plan_demand_gpu_suborder 表
    3. 新增 res_plan_demand_gpu_template 表
*/

START TRANSACTION;

-- ----------------------------
-- res_plan_demand_gpu_order: GPU需求提报-主单表
-- ----------------------------
CREATE TABLE `res_plan_demand_gpu_order` (
    `id`              VARCHAR(64)  NOT NULL                                        COMMENT '需求主单ID',
    `bk_biz_id`       BIGINT       NOT NULL                                        COMMENT '业务ID',
    `op_product_id`   BIGINT       NOT NULL                                        COMMENT '运营产品ID',
    `op_product_name` VARCHAR(64)  NOT NULL                                        COMMENT '运营产品名称',
    `template_id`     VARCHAR(32)  NOT NULL                                        COMMENT '模版ID',
    `status`          VARCHAR(32)  NOT NULL                                        COMMENT '状态(INIT:待评审 PENDING:评审中 DONE:已评审 REJECT:部分已驳回 REJECT_ALL:全部已驳回 TERMINATE:已终止)',
    `remark`          VARCHAR(255) DEFAULT ''                                      COMMENT '备注',
    `creator`         VARCHAR(64)  NOT NULL                                        COMMENT '创建人',
    `reviser`         VARCHAR(64)  DEFAULT ''                                      COMMENT '修改人',
    `created_at`      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP              COMMENT '创建时间',
    `updated_at`      DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',

    PRIMARY KEY (`id`),
    KEY `idx_bk_biz_id_op_product_id_status` (`bk_biz_id`, `op_product_id`, `status`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COMMENT = 'GPU需求提报-主单表';

-- ----------------------------
-- res_plan_demand_gpu_suborder: GPU需求提报-子单表
-- ----------------------------
CREATE TABLE `res_plan_demand_gpu_suborder` (
    `id`           VARCHAR(64)  NOT NULL                                        COMMENT '需求子单ID',
    `order_id`     VARCHAR(64)  NOT NULL                                        COMMENT '需求主单ID',
    `bk_biz_id`    BIGINT       NOT NULL                                        COMMENT '业务ID',
    `demand_type`  VARCHAR(64)  NOT NULL                                        COMMENT '需求分类',
    `demand_year`  BIGINT       DEFAULT 0                                       COMMENT '需求年',
    `demand_month` BIGINT       DEFAULT 0                                       COMMENT '需求月',
    `gpu_num`      BIGINT       DEFAULT 0                                       COMMENT 'GPU预算卡数',
    `qpm_max`      BIGINT       DEFAULT 0                                       COMMENT '峰值QPM',
    `status`       VARCHAR(32)  NOT NULL                                        COMMENT '状态(INIT:待评审 PENDING:评审中 DONE:已评审 REJECT:已驳回 TERMINATE:已终止)',
    `comment`      JSON         DEFAULT NULL                                    COMMENT '评审意见',
    `extension`    JSON         NOT NULL,
    `remark`       VARCHAR(255) DEFAULT ''                                      COMMENT '备注',
    `creator`      VARCHAR(64)  NOT NULL                                        COMMENT '创建人',
    `reviser`      VARCHAR(64)  DEFAULT ''                                      COMMENT '修改人',
    `created_at`   DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP              COMMENT '创建时间',
    `updated_at`   DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',

    PRIMARY KEY (`id`),
    KEY `idx_order_id_bk_biz_id_status` (`order_id`, `bk_biz_id`, `status`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COMMENT = 'GPU需求提报-子单表';

-- ----------------------------
-- res_plan_demand_gpu_template: GPU需求提报-模版配置表
-- ----------------------------
CREATE TABLE `res_plan_demand_gpu_template` (
    `id`         VARCHAR(64)  NOT NULL                                        COMMENT '模版ID',
    `tpl_schema` JSON         NOT NULL                                        COMMENT '模版内容(一个Excel对应一条记录)',
    `remark`     VARCHAR(255) DEFAULT ''                                      COMMENT '备注',
    `creator`    VARCHAR(64)  NOT NULL                                        COMMENT '创建人',
    `reviser`    VARCHAR(64)  DEFAULT ''                                      COMMENT '修改人',
    `created_at` DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP              COMMENT '创建时间',
    `updated_at` DATETIME     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',

    PRIMARY KEY (`id`)
) ENGINE = InnoDB DEFAULT CHARSET = utf8mb4 COMMENT = 'GPU需求提报-模版配置表';

-- ----------------------------
-- 初始化ID生成器
-- ----------------------------
INSERT INTO `id_generator` (`resource`, `max_id`)
VALUES ('res_plan_demand_gpu_order', '0'),
       ('res_plan_demand_gpu_suborder', '0'),
       ('res_plan_demand_gpu_template', '0');

CREATE OR REPLACE VIEW `hcm_version` (`hcm_ver`, `sql_ver`) AS
SELECT 'v9.9.9' AS `hcm_ver`, '9999' AS `sql_ver`;

COMMIT;
