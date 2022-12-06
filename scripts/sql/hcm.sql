/*
表结构说明：
各类模型表字段信息主要分为：
1. 主键id                        // 自增主键ID
2. 模型特定字段信息Spec            // 需要用户特殊定义的字段 (Spec)
3. 模型差异字段                   // 云资源模型差异字段 (Extension)
4. 外键id                        // 和当前模型有关联关系的模型主键id (Attachment)
5. 关联资源冗余字段                 // 和当前模型由关联的子资源等其他资源字段信息 （OtherSpec）
6. 创建信息（CreatedRevision）、创建及修正信息（Revision）

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

create table if not exists `audit`
(
    `id`         bigint(1) unsigned not null auto_increment,

    # Spec
    `res_type`   varchar(50)        not null,
    `res_id`     bigint(1) unsigned not null,
    `action`     varchar(20)        not null,
    `rid`        varchar(64)        not null,
    `app_code`   varchar(64)        default '',
    `detail`     json               default null,
    `bk_biz_id`  bigint(1) unsigned default 0,
    `account_id` bigint(1) unsigned default 0,
    `tenant_id`  varchar(64)        default '',

    # Revision
    `operator`   varchar(64)        not null,
    `created_at` timestamp          not null,

    primary key (`id`)
) engine = innodb
  default charset = utf8mb4;

create table if not exists `account`
(
    `id`              bigint(1) unsigned  not null auto_increment,
    `name`            varchar(64)         not null,
    `vendor`          varchar(16)         not null,
    `managers`        json                not null,
    `department_id`   int(11)             not null,
    `type`            varchar(32)         not null,
    `sync_status`     varchar(32)         not null,
    `price`           varchar(16)         default '',
    `price_unit`      varchar(8)          default '',
    `extension`       json                not null,
    `creator`         varchar(64)         not null,
    `reviser`         varchar(64)         not null,
    `created_at`      timestamp           not null default now(),
    `updated_at`      timestamp           not null default now(),
    `memo`            varchar(255)        default '',
    primary key (`id`),
    unique key `idx_uk_name` (`name`)
) engine=innodb default charset=utf8mb4;

create table if not exists `account_biz_rel` (
    `id`              bigint(1) unsigned  not null auto_increment,
    `bk_biz_id`       int(11)             not null,
    `account_id`      bigint(1) unsigned  not null,
    `creator`         varchar(64)         not null,
    `reviser`         varchar(64)         not null,
    `created_at`      timestamp           not null default now(),
    `updated_at`      timestamp           not null default now(),
    primary key (`id`),
    unique key `idx_uk_bk_biz_id_account_id` (`bk_biz_id`,`account_id`)
) engine=innodb default charset=utf8mb4;
