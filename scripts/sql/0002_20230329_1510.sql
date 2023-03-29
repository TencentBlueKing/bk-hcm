alter table azure_security_group_rule drop index `idx_uk_name`;

alter table azure_security_group_rule add unique key `idx_uk_name` (`name`, `cloud_security_group_id`);