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
