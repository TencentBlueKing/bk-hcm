/*
    SQLVER=9999,HCMVER=v9.9.9

    Notes:
    1. 修改`tcloud_lb_url_rule`表，增加`bk_biz_id`和`account_id`字段
*/

START TRANSACTION;

ALTER TABLE `tcloud_lb_url_rule`
    ADD COLUMN `bk_biz_id` bigint NOT NULL DEFAULT 0 COMMENT '业务ID' AFTER `cloud_lb_id`,
    ADD COLUMN `account_id` varchar(64) NOT NULL DEFAULT '' COMMENT '账号ID' AFTER `bk_biz_id`,
    ADD INDEX `idx_bk_biz_id` (`bk_biz_id`),
    ADD INDEX `idx_account_id` (`account_id`);

-- 刷新历史数据的bk_biz_id和account_id
UPDATE `tcloud_lb_url_rule` t
INNER JOIN `load_balancer` lb ON t.lb_id = lb.id
SET 
    t.bk_biz_id = lb.bk_biz_id,
    t.account_id = lb.account_id
WHERE 
    t.bk_biz_id = 0 OR t.account_id = '';

CREATE OR REPLACE VIEW `hcm_version`(`hcm_ver`, `sql_ver`) AS
SELECT 'v9.9.9' as `hcm_ver`, '9999' as `sql_ver`;

COMMIT;