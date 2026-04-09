import { h } from 'vue';
import { Tag } from 'bkui-vue';
import { Model, Column } from '@/decorator';
import type { IPermissionTemplateItem } from '@/store/cloud-account-manage/permission-template';

@Model()
export class TableColumnTcloud {
  @Column('string', { name: '模板名称' })
  name: string;

  @Column('string', {
    name: '模板类型',
    render: ({ row }: { row: IPermissionTemplateItem }) => {
      const cloudType = row.extension?.cloud_type;
      if (cloudType === 2) {
        return h(Tag, { radius: '4px' }, '云系统预设');
      }
      if (cloudType === 1) {
        if (row.policy_library_id) {
          return h(Tag, { radius: '4px', theme: 'success' }, '与策略库同步');
        }
        return h(Tag, { radius: '4px', theme: 'warning' }, '云自定义');
      }
      return '--';
    },
  })
  'extension.cloud_type': number;

  @Column('string', { name: '所属二级账号ID' })
  account_id: string;

  @Column('string', { name: '关联三级账号数' })
  associated_sub_account_count: number;

  @Column('string', { name: '模板描述' })
  memo: string;

  @Column('user', { name: '创建人' })
  creator: string;

  @Column('datetime', { name: '创建时间', sort: true })
  created_at: string;

  @Column('user', { name: '更新人' })
  reviser: string;

  @Column('datetime', { name: '更新时间', sort: true })
  updated_at: string;
}
