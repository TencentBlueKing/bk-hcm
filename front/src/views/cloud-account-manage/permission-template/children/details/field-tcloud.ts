import { h } from 'vue';
import { Tag } from 'bkui-vue';
import { Model, Column } from '@/decorator';
import type { IPermissionTemplateItem } from '@/store/cloud-account-manage/permission-template';
import { getTypeData } from '@/views/cloud-account-manage/permission-template/utils';

@Model()
export class DetailsFieldTcloud {
  @Column('string', {
    name: '模板名称',
    group: '基本信息',
  })
  name: string;

  @Column('string', {
    name: '模板类型',
    group: '基本信息',
    meta: {
      display: {
        render: (value: IPermissionTemplateItem) => {
          const { label, theme } = getTypeData(value);
          return h(Tag, { radius: '4px', theme }, label);
        },
      },
    },
  })
  'extension.cloud_type': number;

  @Column('string', {
    name: '所属二级账号',
    group: '基本信息',
  })
  account_id: string;

  @Column('string', {
    name: '关联三级账号数',
    group: '基本信息',
  })
  associated_sub_account_count: number;

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

  @Column('string', { name: '模板描述', group: '基本信息' })
  memo: string;

  @Column('json', { name: '', group: '权限模板' })
  policy_document: string;
}
