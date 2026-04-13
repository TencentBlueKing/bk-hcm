import { Model, Column } from '@/decorator';
import { toArray } from '@/common/util';

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
    format: (value: string | string[]) => toArray(value).map((val) => String(val)),
  })
  cloud_sub_account_ids: string[];

  @Column('string', {
    name: '所属二级账号ID',
    format: (value: string | string[]) => toArray(value).map((val) => String(val)),
  })
  cloud_account_ids: string[];

  @Column('user', { name: '创建人' })
  creator: string;

  @Column('user', { name: '更新人' })
  reviser: string;
}
