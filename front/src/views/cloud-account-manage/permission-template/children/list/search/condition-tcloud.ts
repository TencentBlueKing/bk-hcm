import { Model, Column } from '@/decorator';
import { toArray } from '@/common/util';
import { PERMISSION_TEMPLATE_TYPE_MAP } from '@/views/cloud-account-manage/permission-template/constants';

@Model()
export class SearchConditionTcloud {
  @Column('string', {
    name: '权限模板ID',
    format: (value: string | string[]) => toArray(value).map((val) => String(val)),
  })
  cloud_ids: string[];

  @Column('string', {
    name: '权限模板名称',
    format: (value: string | string[]) => toArray(value).map((val) => String(val)),
  })
  names: string[];

  @Column('string', {
    name: '所属三级账号ID',
    converter: (value: string | string[]) => ({
      extension: {
        cloud_sub_account_ids: toArray(value).map((val) => String(val)),
      },
    }),
  })
  'extension.cloud_sub_account_ids': string[];

  @Column('string', {
    name: '所属二级账号ID',
    converter: (value: string | string[]) => ({
      extension: {
        cloud_main_account_ids: toArray(value).map((val) => String(val)),
      },
    }),
  })
  'extension.cloud_main_account_ids': string[];

  @Column('enum', {
    name: '模板类型',
    option: PERMISSION_TEMPLATE_TYPE_MAP,
    props: {
      multiple: false,
    },
  })
  permission_template_type: string;

  @Column('user', { name: '创建人' })
  creator: string;

  @Column('user', { name: '更新人' })
  reviser: string;
}
