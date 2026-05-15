/* eslint-disable @typescript-eslint/no-unused-vars */
import { Model, Column } from '@/decorator';
import { QueryRuleOPEnum } from '@/typings';
import {
  OPERATION_LOG_ACTION_NAME,
  OPERATION_LOG_RESOURCE_TYPE_NAME,
  OPERATION_LOG_RES_TYPES,
} from '@/views/operation-log/constants';
import type { OperationLogAction, OperationLogResourceType } from '@/views/operation-log/typings';
import { SearchCondition } from './condition';

@Model('operation-log/search-condition-all')
export class SearchConditionAll extends SearchCondition {
  @Column('enum', {
    name: '资源类型',
    option: OPERATION_LOG_RESOURCE_TYPE_NAME,
    index: 1,
    meta: {
      search: {
        filterRules(value: OperationLogResourceType) {
          const val = OPERATION_LOG_RES_TYPES[value as keyof typeof OPERATION_LOG_RES_TYPES] || [value];
          return {
            field: 'res_type',
            op: QueryRuleOPEnum.IN,
            value: val,
          };
        },
      },
    },
  })
  declare res_type: OperationLogResourceType;

  @Column('enum', { name: '操作方式', option: OPERATION_LOG_ACTION_NAME, index: 2 })
  action: OperationLogAction;
}
