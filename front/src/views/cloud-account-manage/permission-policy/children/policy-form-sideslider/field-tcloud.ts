import { Model, Column } from '@/decorator';

@Model()
export class InfoFieldTcloud {
  @Column('string', {
    name: '权限策略库名称',
    group: '基本信息',
  })
  name: string;

  @Column('string', {
    name: '关联二级账号数',
    group: '基本信息',
  })
  associated_account_count: number;

  @Column('user', {
    name: '创建人',
    group: '基本信息',
  })
  creator: string;

  @Column('datetime', {
    name: '创建时间',
    group: '基本信息',
  })
  created_at: string;

  @Column('user', {
    name: '更新人',
    group: '基本信息',
  })
  reviser: string;

  @Column('datetime', {
    name: '更新时间',
    group: '基本信息',
  })
  updated_at: string;

  @Column('json', { name: '', group: '权限模板' })
  policy_document: string;
}
