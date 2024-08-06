/*
    SQLVER=9999,HCMVER=v9.9.9

    Notes:
    1. 修改`application`表，增加`bk_biz_ids`字段
*/

START TRANSACTION;

-- 增加`bk_biz_ids`字段
alter table application
    add bk_biz_ids json after status;

CREATE OR REPLACE VIEW `hcm_version`(`hcm_ver`, `sql_ver`) AS
SELECT 'v9.9.9' as `hcm_ver`, '9999' as `sql_ver`;

COMMIT