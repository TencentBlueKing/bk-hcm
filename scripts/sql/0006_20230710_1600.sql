/*
    SQLVER=0006,HCMVER=v1.1.18

    Notes:
        1. 开始将SQL文件与HCM版本对应，并通过在数据库中创建view来记录当前数据库对应版本。
        2. 将HCM表改为区分大小写排序规则(utf8mb4_bin)。
*/

CREATE
OR REPLACE VIEW `hcm_version`(`hcm_ver`,`sql_ver`) AS
SELECT 'v1.1.18' as `hcm_ver`, '0006' as `sql_ver`;

-- 修改排序规则前需将外键删除，否则会两表字段冲突
alter table disk_cvm_rel drop foreign key disk_cvm_rel_cvm_id;
alter table disk_cvm_rel drop foreign key disk_cvm_rel_disk_id;
alter table security_group_cvm_rel drop foreign key security_group_cvm_rel_security_group_id;
alter table security_group_cvm_rel drop foreign key security_group_cvm_rel_cvm_id;
alter table eip_cvm_rel drop foreign key eip_cvm_rel_eip_id;
alter table eip_cvm_rel drop foreign key eip_cvm_rel_cvm_id;
alter table network_interface_cvm_rel drop foreign key network_interface_cvm_rel_network_id;
alter table network_interface_cvm_rel drop foreign key network_interface_cvm_rel_cvm_id;

-- 修改表排序规则
ALTER TABLE eip_cvm_rel CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_bin;
ALTER TABLE security_group_cvm_rel CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_bin;
ALTER TABLE account_biz_rel CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_bin;
ALTER TABLE disk_cvm_rel CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_bin;
ALTER TABLE network_interface_cvm_rel CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_bin;
ALTER TABLE account CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_bin;
ALTER TABLE account_bill_config CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_bin;
ALTER TABLE application CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_bin;
ALTER TABLE approval_process CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_bin;
ALTER TABLE audit CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_bin;
ALTER TABLE aws_region CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_bin;
ALTER TABLE aws_route CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_bin;
ALTER TABLE aws_security_group_rule CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_bin;
ALTER TABLE azure_region CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_bin;
ALTER TABLE azure_resource_group CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_bin;
ALTER TABLE azure_route CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_bin;
ALTER TABLE azure_security_group_rule CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_bin;
ALTER TABLE cvm CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_bin;
ALTER TABLE disk CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_bin;
ALTER TABLE eip CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_bin;
ALTER TABLE gcp_firewall_rule CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_bin;
ALTER TABLE gcp_region CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_bin;
ALTER TABLE gcp_route CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_bin;
ALTER TABLE huawei_region CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_bin;
ALTER TABLE huawei_route CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_bin;
ALTER TABLE huawei_security_group_rule CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_bin;
ALTER TABLE id_generator CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_bin;
ALTER TABLE image CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_bin;
ALTER TABLE network_interface CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_bin;
ALTER TABLE recycle_record CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_bin;
ALTER TABLE route_table CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_bin;
ALTER TABLE security_group CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_bin;
ALTER TABLE subnet CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_bin;
ALTER TABLE tcloud_region CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_bin;
ALTER TABLE tcloud_route CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_bin;
ALTER TABLE tcloud_security_group_rule CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_bin;
ALTER TABLE vpc CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_bin;
ALTER TABLE zone CONVERT TO CHARACTER SET utf8mb4 COLLATE utf8mb4_bin;

-- 修改完排序规则后，将外键重新添加
alter table disk_cvm_rel
    add constraint disk_cvm_rel_cvm_id foreign key (disk_id) REFERENCES disk (id) ON DELETE CASCADE;
alter table disk_cvm_rel
    add constraint disk_cvm_rel_disk_id foreign key (cvm_id) REFERENCES cvm (id) ON DELETE CASCADE;
alter table security_group_cvm_rel
    add constraint security_group_cvm_rel_security_group_id foreign key (security_group_id) REFERENCES security_group (id) ON DELETE CASCADE;
alter table security_group_cvm_rel
    add constraint security_group_cvm_rel_cvm_id foreign key (cvm_id) REFERENCES cvm (id) ON DELETE CASCADE;
alter table eip_cvm_rel
    add constraint eip_cvm_rel_eip_id foreign key (eip_id) REFERENCES eip (id) ON DELETE CASCADE;
alter table eip_cvm_rel
    add constraint eip_cvm_rel_cvm_id foreign key (cvm_id) REFERENCES cvm (id) ON DELETE CASCADE;
alter table network_interface_cvm_rel
    add constraint network_interface_cvm_rel_network_id foreign key (network_interface_id) REFERENCES network_interface (id) ON DELETE CASCADE;
alter table network_interface_cvm_rel
    add constraint network_interface_cvm_rel_cvm_id foreign key (cvm_id) REFERENCES cvm (id) ON DELETE CASCADE;
