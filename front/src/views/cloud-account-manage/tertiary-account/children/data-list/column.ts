/* eslint-disable @typescript-eslint/no-unused-vars */
import { Model, Column } from '@/decorator';
import { ACCOUNT_TYPE_OPTIONS } from '../search/condition';
import { ITcloudExtension } from '../../typings';
import { FLAG_OPTIONS } from '../../constants';

@Model('tertiary-account/table-column')
export class TableColumn {
  @Column('string', {
    name: '三级账号名称',
    index: 0,
    width: 150,
  })
  name: string;

  @Column('string', {
    name: '三级账号ID',
    index: 1,
    width: 130,
  })
  cloud_id: string;

  @Column('string', {
    name: '所属二级账号ID',
    index: 2,
    width: 150,
    render: ({ row }: { row: { extension: ITcloudExtension } }) => {
      return row?.extension?.cloud_main_account_id || '--';
    },
  })
  'extension.cloud_main_account_id': string;

  @Column('string', {
    name: '账号类型',
    index: 3,
    width: 100,
    render: ({ row }: { row: { extension: ITcloudExtension } }) => {
      const consoleLogin = row?.extension?.console_login;
      return ACCOUNT_TYPE_OPTIONS[consoleLogin] || '--';
    },
  })
  'extension.console_login': string;

  @Column('array', {
    name: '负责人',
    index: 4,
    width: 120,
  })
  managers: string[];

  @Column('business', {
    name: '所属业务',
    index: 5,
    width: 180,
    meta: {
      display: { appearance: 'tag' },
    },
  })
  bk_biz_ids: number[];

  @Column('string', {
    name: '账号邮箱',
    index: 6,
    width: 180,
  })
  email: string;

  @Column('string', {
    name: '手机号',
    index: 7,
    width: 120,
  })
  phone_num: string;

  @Column('string', {
    name: '权限模板数',
    index: 8,
    width: 80,
    sort: true,
    meta: {
      search: {
        props: {
          searchField: 'extension.cloud_sub_account_ids', // 路由查询关联字段-目标文件
          valueField: 'cloud_id', // 路由查询值-本文件
        },
      },
    },
  })
  permission_template_count: number;

  @Column('string', {
    name: '密钥数',
    index: 8,
    width: 80,
    sort: true,
    meta: {
      search: {
        props: {
          searchField: 'cloud_sub_account_id',
          valueField: 'cloud_id',
        },
      },
    },
  })
  sub_account_secret_count: number;

  @Column('string', {
    name: '登录保护',
    index: 9,
    width: 110,
    render: ({ row }: { row: { extension: ITcloudExtension } }) => {
      return FLAG_OPTIONS[row?.extension?.login_flag] || '--';
    },
  })
  'extension.login_flag': string;

  @Column('string', {
    name: '操作保护',
    index: 10,
    width: 110,
    render: ({ row }: { row: { extension: ITcloudExtension } }) => {
      return FLAG_OPTIONS[row?.extension?.action_flag] || '--';
    },
  })
  'extension.action_flag': string;

  @Column('string', {
    name: 'MFA设备绑定',
    index: 11,
    width: 110,
    render: ({ row }: { row: { extension: ITcloudExtension } }) => {
      return row?.extension?.login_flag === 'stoken' || row?.extension?.action_flag === 'stoken' ? '已绑定' : '未绑定';
    },
  })
  mfa_device: string;

  @Column('string', {
    name: '备注',
    index: 12,
    width: 150,
  })
  memo: string;

  @Column('datetime', {
    name: '云上创建时间',
    sort: true,
    index: 14,
    width: 170,
  })
  cloud_created_at: string;

  @Column('datetime', {
    name: '录入时间',
    sort: true,
    index: 15,
    width: 170,
  })
  created_at: string;

  @Column('datetime', {
    name: '更新时间',
    sort: true,
    index: 16,
    width: 170,
  })
  updated_at: string;
}
