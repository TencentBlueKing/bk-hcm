import { Model, Column } from '@/decorator';
import { QueryRuleOPEnum } from '@/typings';
import {
  OPERATION_LOG_RESOURCE_TYPE_NAME,
  OPERATION_LOG_RES_TYPES,
  OPERATION_LOG_SOURCE_NAME,
} from '@/views/operation-log/constants';
import type { OperationLogResourceType, OperationLogSource } from '@/views/operation-log/typings';

@Model('operation-log/search-condition')
export class SearchCondition {
  @Column('datetime', { name: '操作时间', index: 0 })
  created_at: string;

  @Column('string', {
    name: '资源名称',
    meta: {
      search: {
        filterRules(value: string | string[]) {
          if (Array.isArray(value) && value.length > 1) {
            return {
              op: QueryRuleOPEnum.OR,
              rules: value.map((val) => ({ field: 'res_name', op: QueryRuleOPEnum.CS, value: val })),
            };
          }
          if (Array.isArray(value) && value.length === 1) {
            return { field: 'res_name', op: QueryRuleOPEnum.CS, value: value[0] };
          }
          return { field: 'res_name', op: QueryRuleOPEnum.CS, value };
        },
      },
    },
    index: 3,
  })
  res_name: string;

  @Column('enum', { name: '操作来源', option: OPERATION_LOG_SOURCE_NAME, index: 3 })
  source: OperationLogSource;

  @Column('string', { name: '云账号', index: 3 })
  account_id: string;

  @Column('user', { name: '操作人', index: 3 })
  operator: string;

  @Column('business', { name: '所属业务', index: 3 })
  bk_biz_id: number;

  @Column('enum', {
    apiOnly: true,
    name: '资源类型',
    option: OPERATION_LOG_RESOURCE_TYPE_NAME,
    index: 1,
    meta: {
      search: {
        filterRules(value: OperationLogResourceType | OperationLogResourceType[]) {
          // use-search-qs.get 会按 IN 操作符把标量包成数组，这里需要兼容两种入参，避免 [["cvm"]] 二次包裹
          const toResTypes = (val: OperationLogResourceType) =>
            OPERATION_LOG_RES_TYPES[val as keyof typeof OPERATION_LOG_RES_TYPES] || [val];
          return {
            field: 'res_type',
            op: QueryRuleOPEnum.IN,
            value: Array.isArray(value) ? value.flatMap(toResTypes) : toResTypes(value),
          };
        },
      },
    },
  })
  res_type: OperationLogResourceType;
}
