/*
    SQLVER=0010,HCMVER=v1.1.25

    Notes:
        1. 修复aws_security_group_rule表memo字段错误的非空约束。
*/


start transaction;
alter table aws_security_group_rule modify column memo varchar(255) default '';
CREATE OR REPLACE VIEW `hcm_version`(`hcm_ver`, `sql_ver`) AS SELECT 'v1.1.25' as `hcm_ver`, '0010' as `sql_ver`;
commit;