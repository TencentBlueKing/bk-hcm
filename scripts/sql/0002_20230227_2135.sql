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
create
database if not exists hcm;

use
hcm;

insert into id_generator(`resource`, `max_id`)
values ('network_interface', '0');

-- ----------------------------
-- Table structure for network_interface
-- ----------------------------
CREATE TABLE `network_interface`
(
    `id`              varchar(64)  NOT NULL COMMENT '主键',
    `account_id`      varchar(64)  NOT NULL COMMENT '账号ID',
    `vendor`          varchar(32)  NOT NULL DEFAULT '' COMMENT '云厂商标识',
    `name`            varchar(64)  NOT NULL COMMENT '网络接口名称',
    `region`          varchar(255) NOT NULL DEFAULT '' COMMENT '区域/地域',
    `zone`            varchar(255) NOT NULL DEFAULT '' COMMENT '可用区',
    `cloud_id`        varchar(255)          DEFAULT '' COMMENT '网卡端口所属网络ID',
    `vpc_id`          varchar(255) NOT NULL DEFAULT '' COMMENT 'VPC的ID',
    `cloud_vpc_id`    varchar(255) NOT NULL DEFAULT '' COMMENT '云VPC的ID',
    `subnet_id`       varchar(64)  NOT NULL DEFAULT '' COMMENT '子网的ID',
    `cloud_subnet_id` varchar(255)          DEFAULT '' COMMENT '云子网的ID',
    `private_ip`      varchar(64)           DEFAULT '' COMMENT '内网IP',
    `public_ip`       varchar(64)           DEFAULT '' COMMENT '公网IP',
    `bk_biz_id`       bigint                DEFAULT '-1',
    `instance_id`     varchar(255)          DEFAULT '' COMMENT '关联的实例id',
    `extension`       json                  DEFAULT NULL COMMENT '扩展字段',
    `creator`         varchar(64)           DEFAULT '' COMMENT '创建人',
    `reviser`         varchar(64)           DEFAULT '' COMMENT '修改人',
    `created_at`      timestamp    NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at`      timestamp    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4 COMMENT = '网络接口表';
