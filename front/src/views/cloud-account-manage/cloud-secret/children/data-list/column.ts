/* eslint-disable @typescript-eslint/no-unused-vars */
import { Model, Column } from '@/decorator';
import { SECRET_STATUS_OPTIONS, SUB_ACCOUNT_TYPE_OPTIONS } from '../search/condition';

@Model('cloud-secret/table-column')
export class TableColumn {
  @Column('string', {
    name: '密钥ID',
    index: 0,
    width: 140,
  })
  cloud_secret_id: string;

  @Column('enum', {
    name: '密钥状态',
    option: SECRET_STATUS_OPTIONS,
    index: 1,
    width: 100,
  })
  status: string;

  @Column('string', {
    name: '所属三级账号ID',
    index: 2,
    width: 140,
  })
  cloud_sub_account_id: string;

  @Column('enum', {
    name: '三级账号类型',
    option: SUB_ACCOUNT_TYPE_OPTIONS,
    index: 3,
    width: 120,
  })
  console_login: number;

  @Column('string', {
    name: '三级账号负责人',
    index: 4,
    width: 120,
  })
  sub_account_manager: string;

  @Column('string', {
    name: '所属二级账号ID',
    index: 5,
    width: 140,
  })
  cloud_main_account_id: string;

  @Column('string', {
    name: '二级账号负责人',
    index: 6,
    width: 120,
  })
  account_manager: string;

  @Column('datetime', {
    name: '创建时间',
    sort: true,
    index: 7,
    width: 170,
  })
  cloud_created_at: string;

  @Column('datetime', {
    name: '最近访问时间',
    sort: true,
    index: 8,
    width: 170,
  })
  last_used_time: string;

  @Column('datetime', {
    name: '密钥禁用时间',
    sort: true,
    index: 9,
    width: 170,
  })
  disabled_time: string;
}
