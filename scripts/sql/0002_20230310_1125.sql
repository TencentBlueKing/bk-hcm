alter table huawei_region add service varchar(20) not null;

alter table account
    drop column department_ids;

SET FOREIGN_KEY_CHECKS = 0;
alter table disk_cvm_rel
    add constraint DISK_CVM_ID foreign key (disk_id) REFERENCES disk (id) ON DELETE CASCADE;
alter table disk_cvm_rel
    add constraint CVM_DISK_ID foreign key (cvm_id) REFERENCES cvm (id) ON DELETE CASCADE;

alter table eip_cvm_rel
    add constraint EIP_CVM_ID foreign key (eip_id) REFERENCES eip (id) ON DELETE CASCADE;
alter table eip_cvm_rel
    add constraint CVM_EIP_ID foreign key (cvm_id) REFERENCES cvm (id) ON DELETE CASCADE;
SET FOREIGN_KEY_CHECKS = 1;

alter table azure_security_group_rule
    change column `cloud_source_security_group_ids` `cloud_source_app_security_group_ids` json default null;
alter table azure_security_group_rule
    change column `cloud_destination_security_group_ids` `cloud_destination_app_security_group_ids` json default null;
