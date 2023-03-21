alter table huawei_region
    add service varchar(20) not null;

alter table account
    drop column department_ids;

SET FOREIGN_KEY_CHECKS = 0;
alter table disk_cvm_rel
    add constraint disk_cvm_rel_cvm_id foreign key (disk_id) REFERENCES disk (id) ON DELETE CASCADE;
alter table disk_cvm_rel
    add constraint disk_cvm_rel_disk_id foreign key (cvm_id) REFERENCES cvm (id) ON DELETE CASCADE;

alter table eip_cvm_rel
    add constraint eip_cvm_rel_eip_id foreign key (eip_id) REFERENCES eip (id) ON DELETE CASCADE;
alter table eip_cvm_rel
    add constraint eip_cvm_rel_cvm_id foreign key (cvm_id) REFERENCES cvm (id) ON DELETE CASCADE;

alter table security_group_cvm_rel
    add constraint security_group_cvm_rel_security_group_id foreign key (security_group_id) REFERENCES security_group (id) ON DELETE CASCADE;
alter table security_group_cvm_rel
    add constraint security_group_cvm_rel_cvm_id foreign key (cvm_id) REFERENCES cvm (id) ON DELETE CASCADE;

SET FOREIGN_KEY_CHECKS = 1;

alter table azure_security_group_rule
    change column `cloud_source_security_group_ids` `cloud_source_app_security_group_ids` json default null;
alter table azure_security_group_rule
    change column `cloud_destination_security_group_ids` `cloud_destination_app_security_group_ids` json default null;

alter table gcp_firewall_rule add vpc_self_link varchar(255) default '';

alter table network_interface
    add `vpc_self_link` varchar(255) default '' after `cloud_vpc_id`;

# recycle record related table structure
create table if not exists `recycle_record`
(
    `id`           bigint(1) unsigned not null auto_increment,
    `task_id`      varchar(64)        not null,
    `vendor`       varchar(32)        not null,
    `res_type`     varchar(64)        not null,
    `res_id`       varchar(64)        not null,
    `cloud_res_id` varchar(255)       not null,
    `res_name`     varchar(255)                default '',
    `bk_biz_id`    bigint(1)          not null,
    `account_id`   varchar(64)        not null,
    `region`       varchar(255)       not null,
    `detail`       json               not null,
    `status`       varchar(32)        not null,
    `creator`      varchar(64)        not null,
    `reviser`      varchar(64)        not null,
    `created_at`   timestamp          not null default current_timestamp,
    `updated_at`   timestamp          not null default current_timestamp on update current_timestamp,
    primary key (`id`),
    unique key `idx_res_type_res_id` (`res_type`, `res_id`),
    unique key `idx_res_type_vendor_cloud_res_id` (`res_type`, `vendor`, `cloud_res_id`)
) engine = innodb
  default charset = utf8mb4;

insert into id_generator(`resource`, `max_id`)
values ('recycle_record', '0');

alter table cvm
    add recycle_status varchar(32) default '';

alter table disk
    add recycle_status varchar(32) default '';