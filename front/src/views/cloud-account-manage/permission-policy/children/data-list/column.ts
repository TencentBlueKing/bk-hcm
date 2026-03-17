/* eslint-disable @typescript-eslint/no-unused-vars */
import { Model, Column } from '@/decorator';
// import { SITE_TYPE_OPTIONS, SYNC_STATUS_OPTIONS } from '../search/condition';
import { ITcloudExtension } from '../../typings';
import { FLAG_OPTIONS } from '../../constants';

@Model('cloud-account-manage/table-column')
export class TableColumn {
  @Column('string', {
    name: '策略库名称',
    index: 0,
    width: 150,
  })
  name: string;

  @Column('string', {
    name: '策略库描述',
    index: 1,
    width: 180,
  })
  desc: string;

  @Column('number', {
    name: '关联二级账号数',
    index: 2,
    width: 110,
    render: ({ row }: { row: { extension: ITcloudExtension } }) => {
      return FLAG_OPTIONS[row?.extension?.login_flag] || '--';
    },
  })
  'extension.login_flag': number;

  @Column('user', {
    name: '创建人',
    sort: true,
    index: 3,
    width: 170,
  })
  creaor: string;

  @Column('datetime', {
    name: '创建时间',
    sort: true,
    index: 4,
    width: 170,
  })
  creat_at: string;

  @Column('user', {
    name: '更新人',
    sort: true,
    index: 5,
    width: 170,
  })
  updator: string;

  @Column('datetime', {
    name: '更新时间',
    sort: true,
    index: 6,
    width: 170,
  })
  updated_at: string;
}
