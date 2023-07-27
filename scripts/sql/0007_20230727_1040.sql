/*
    SQLVER=0007,HCMVER=v1.1.20

    Notes:
        1. gcp_firewall_rule调整name唯一索引为name,account_id联合唯一索引。
*/

alter table gcp_firewall_rule drop index idx_uk_name;
alter table gcp_firewall_rule add unique key `idx_uk_account_id_name` (`account_id`, `name`);
