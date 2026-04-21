/* eslint-disable @typescript-eslint/no-unused-vars */
import { Model, Column } from '@/decorator';
import { SITE_TYPE_OPTIONS, SYNC_STATUS_OPTIONS } from '../search/condition';
import { ITcloudExtension } from '../../typings';
import { FLAG_OPTIONS } from '../../constants';
import { ISecondaryAccountItem } from '@/store';
import { h } from 'vue';

@Model('cloud-account-manage/table-column')
export class TableColumn {
  @Column('string', {
    name: '二级账号名称',
    index: 0,
    width: 150,
  })
  name: string;

  @Column('string', {
    name: '二级账号ID',
    index: 1,
    width: 120,
    render: ({ row }: { row: { extension: ITcloudExtension } }) => {
      return row?.extension?.cloud_main_account_id || '--';
    },
  })
  'extension.cloud_main_account_id': string;

  @Column('string', {
    name: '账号邮箱',
    index: 2,
    width: 180,
  })
  email: string;

  @Column('enum', {
    name: '站点类型',
    option: SITE_TYPE_OPTIONS,
    index: 3,
    width: 100,
  })
  site: string;

  @Column('array', {
    name: '负责人',
    index: 4,
    width: 120,
  })
  managers: string[];

  @Column('array', {
    name: '安全负责人',
    index: 5,
    width: 120,
  })
  security_managers: string[];

  @Column('business', {
    name: '使用业务',
    index: 6,
    width: 180,
    meta: {
      display: { appearance: 'tag' },
    },
  })
  usage_biz_ids: number[];

  @Column('string', {
    name: '三级账号数',
    index: 7,
    width: 130,
    sort: true,
    meta: {
      search: {
        props: {
          searchField: 'extension.cloud_main_account_id', // 路由查询关联字段-目标文件
          valueField: 'extension.cloud_main_account_id', // 路由查询值-本文件
        },
      },
    },
  })
  sub_account_count: number;

  @Column('string', {
    name: '密钥数',
    index: 8,
    width: 80,
    sort: true,
    meta: {
      search: {
        props: {
          searchField: 'cloud_main_account_id', // 路由查询关联字段-目标文件
          valueField: 'extension.cloud_main_account_id', // 路由查询值-本文件
        },
      },
    },
  })
  account_secret_count: number;

  @Column('enum', {
    name: '资源纳管',
    option: SYNC_STATUS_OPTIONS,
    index: 9,
    width: 110,
    render: ({ row }: { row: ISecondaryAccountItem }) => {
      const statusMap: Record<string, { class: string; text: string }> = {
        sync_success: { class: 'status-tag status-tag-success', text: '同步成功' },
        sync_failed: { class: 'status-tag status-tag-failed', text: '同步失败' },
        not_sync: { class: 'status-tag status-tag-not-sync', text: '未同步' },
        syncing: { class: 'status-tag status-tag-syncing', text: '同步中' },
        managed: { class: 'status-tag status-tag-managed', text: '已纳管' },
        unmanaged: { class: 'status-tag status-tag-unmanaged', text: '未纳管' },
      };
      const status = statusMap[row.sync_status] || { class: '', text: row.sync_status || '--' };
      return h('span', { class: status.class }, status.text);
    },
  })
  sync_status: string;

  @Column('string', {
    name: '登录保护',
    index: 10,
    width: 110,
    render: ({ row }: { row: { extension: ITcloudExtension } }) => {
      return FLAG_OPTIONS[row?.extension?.login_flag] || '--';
    },
  })
  'extension.login_flag': string;

  @Column('string', {
    name: '操作保护',
    index: 11,
    width: 110,
    render: ({ row }: { row: { extension: ITcloudExtension } }) => {
      return FLAG_OPTIONS[row.extension?.action_flag] || '--';
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
    index: 13,
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
