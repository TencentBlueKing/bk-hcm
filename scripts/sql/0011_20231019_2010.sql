/*
    SQLVER=0011,HCMVER=v1.1.27

    Notes:
        1. 修复tcloud_route表中不同路由表下路由策略ID重复的问题
 */
start transaction;

alter table tcloud_route drop key idx_uk_cloud_id;
alter table tcloud_route
    add constraint idx_uk_cloud_route_table_id_cloud_id unique ( cloud_route_table_id,cloud_id);

CREATE OR REPLACE VIEW `hcm_version`(`hcm_ver`, `sql_ver`) AS
SELECT 'v1.1.27' as `hcm_ver`, '0011' as `sql_ver`;

commit;
