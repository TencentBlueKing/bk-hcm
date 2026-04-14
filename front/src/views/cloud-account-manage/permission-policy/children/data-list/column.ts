/* eslint-disable @typescript-eslint/no-unused-vars */
import { Model, Column } from '@/decorator';

@Model('permission-policy/table-column')
export class TableColumn {
  @Column('string', {
    name: '策略库名称',
    index: 0,
    width: 180,
  })
  name: string;

  @Column('number', {
    name: '关联二级账号数',
    index: 1,
    width: 140,
  })
  associated_account_count: number;

  @Column('string', {
    name: '策略库描述',
    index: 2,
    minWidth: 300,
  })
  memo: string;

  @Column('string', {
    name: '创建人',
    index: 3,
    width: 120,
  })
  creator: string;

  @Column('datetime', {
    name: '创建时间',
    sort: true,
    index: 4,
    width: 180,
  })
  created_at: string;

  @Column('string', {
    name: '更新人',
    index: 5,
    width: 120,
  })
  reviser: string;

  @Column('datetime', {
    name: '更新时间',
    sort: true,
    index: 6,
    width: 180,
  })
  updated_at: string;
}
