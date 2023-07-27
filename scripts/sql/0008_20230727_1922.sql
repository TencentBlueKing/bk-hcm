/*
    SQLVER=0008,HCMVER=v1.1.21

    Notes:
        1. gcp_firewall_rule调整name唯一索引为name,account_id联合唯一索引。
*/

alter table security_group modify column name varchar (255) not null;
alter table gcp_firewall_rule modify column name varchar (65) not null;
