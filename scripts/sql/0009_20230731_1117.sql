/*
    SQLVER=0009,HCMVER=v1.1.22

    Notes:
        1. 调整aws_security_group_rule memo字段长度。
*/

alter table aws_security_group_rule modify column memo varchar (255) not null;
