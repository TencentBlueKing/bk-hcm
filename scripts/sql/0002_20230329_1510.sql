alter table azure_security_group_rule drop index `idx_uk_name`;

alter table azure_security_group_rule add unique key `idx_uk_name` (`name`, `cloud_security_group_id`);

alter table `eip` drop column `instance_id`;
alter table `eip` drop column `instance_type`;
alter table `account` drop column  `sync_status`;
alter table `recycle_record` drop index `idx_res_type_vendor_cloud_res_id`;
alter table `recycle_record` drop index `idx_res_type_res_id`;
alter table `recycle_record` modify column `id` varchar(64)  not null;
insert into id_generator(`resource`, `max_id`) values ('recycle_record_task_id', '0');
