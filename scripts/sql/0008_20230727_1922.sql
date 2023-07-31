/*
    SQLVER=0008,HCMVER=v1.1.21

    Notes:
        1. 调整security_group/gcp_firewall_rule name字段长度。
*/

alter table security_group modify column name varchar (255) not null;
alter table gcp_firewall_rule modify column name varchar (65) not null;
