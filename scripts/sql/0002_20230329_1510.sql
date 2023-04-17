alter table azure_security_group_rule drop index `idx_uk_name`;

alter table azure_security_group_rule add unique key `idx_uk_name` (`name`, `cloud_security_group_id`);

alter table `eip` drop column `instance_id`;
alter table `eip` drop column `instance_type`;
alter table `account` drop column  `sync_status`;
