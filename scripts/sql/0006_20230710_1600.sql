/*
    SQLVER=0006,HCMVER=v1.1.18
    开始将SQL文件与HCM版本对应，并通过在数据库中创建view来记录当前数据库对应版本
*/

CREATE OR REPLACE VIEW `hcm_version`(`hcm_ver`,`sql_ver`) AS SELECT 'v1.1.18' as `hcm_ver`, '0006' as `sql_ver`;