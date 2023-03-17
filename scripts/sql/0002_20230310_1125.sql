alter table huawei_region add service varchar(20) not null;

alter table account drop column department_ids;

alter table disk_cvm_rel add constraint DISK_ID foreign key(disk_id) REFERENCES disk(id) ON DELETE CASCADE;
alter table disk_cvm_rel add constraint CVM_ID foreign key(cvm_id) REFERENCES cvm(id) ON DELETE CASCADE;

alter table eip_cvm_rel add constraint EIP_ID foreign key(eip_id) REFERENCES eip(id) ON DELETE CASCADE;
alter table eip_cvm_rel add constraint CVM_ID foreign key(cvm_id) REFERENCES cvm(id) ON DELETE CASCADE;