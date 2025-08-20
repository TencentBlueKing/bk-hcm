
/*
    SQLVER=0034,HCMVER=v1.5.0

    Notes:
    1. -- 为tcloud_lb_url_rule表添加bk_biz_id字段
*/

START TRANSACTION;

alter table `tcloud_lb_url_rule`
add column `bk_biz_id` bigint not null default 0 COMMENT '业务ID' after `cloud_lb_id`,
add index `idx_bk_biz_id` (`bk_biz_id`);

alter table "tcloud_lb_url_rule"
add column "account_id" varchar(64) not null default '' COMMENT '账号ID' after `bk_biz_id`;
add index "idx_account_id" ("account_id");
COMMIT;
