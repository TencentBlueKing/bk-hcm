/*
    SQLVER=0010,HCMVER=v1.1.24,dev

    Notes:
        1. 修复错误的非空约束
        2. 独立存储aws每个账号的地域信息
*/
start transaction;

alter table aws_security_group_rule modify column memo varchar(255) default '';

delete from aws_region;
alter table aws_region add column account_id varchar(64) not null;
alter table aws_region drop index `idx_uk_region_id_status`;
alter table aws_region add unique key `idx_uk_account_id_region_id` (`account_id`, `region_id`);

CREATE OR REPLACE VIEW `hcm_version`(`hcm_ver`, `sql_ver`) AS SELECT 'v1.1.24' as `hcm_ver`, '0010' as `sql_ver`;
commit;