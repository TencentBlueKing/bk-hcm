/*
    SQLVER=0010,HCMVER=v1.1.25

    Notes:
        1. 修复aws_security_group_rule表memo字段错误的非空约束。
        2. 修复gcp同步创建主机时可用区超过限制长度问题
*/


start transaction;

alter table aws_security_group_rule modify column memo varchar(255) default '';

alter table cvm modify column zone varchar(64) default '';

CREATE OR REPLACE VIEW `hcm_version`(`hcm_ver`, `sql_ver`) AS SELECT 'v1.1.25' as `hcm_ver`, '0010' as `sql_ver`;

commit;