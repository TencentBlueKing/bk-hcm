/* eslint-disable @typescript-eslint/no-unused-vars */
import { Model, Column } from '@/decorator';
import { QueryRuleOPEnum } from '@/typings';
import { OPERATION_LOG_ACTION, OPERATION_LOG_ACTION_NAME } from '@/views/operation-log/constants';
import type { OperationLogAction, OperationLogResourceType } from '@/views/operation-log/typings';
import { SearchCondition } from './condition';

@Model('operation-log/search-condition-clb')
export class SearchConditionClb extends SearchCondition {
  static actionOption = {
    [OPERATION_LOG_ACTION.CREATE]: OPERATION_LOG_ACTION_NAME[OPERATION_LOG_ACTION.CREATE],
    [OPERATION_LOG_ACTION.UPDATE]: OPERATION_LOG_ACTION_NAME[OPERATION_LOG_ACTION.UPDATE],
    [OPERATION_LOG_ACTION.DELETE]: OPERATION_LOG_ACTION_NAME[OPERATION_LOG_ACTION.DELETE],
    [OPERATION_LOG_ACTION.ASSIGN]: OPERATION_LOG_ACTION_NAME[OPERATION_LOG_ACTION.ASSIGN],
  };

  @Column('enum', { name: '操作方式', option: SearchConditionClb.actionOption, index: 1 })
  action: OperationLogAction;

  @Column('enum', {
    name: '任务类型',
    option: { '': '异步任务' },
    meta: { search: { enableEmpty: true, op: QueryRuleOPEnum.JSON_NEQ } },
    index: 3,
  })
  'detail.data.res_flow.flow_id': string;

  @Column('enum', {
    apiOnly: true,
    meta: {
      search: {
        filterRules() {
          return {
            field: 'res_type',
            op: QueryRuleOPEnum.IN,
            value: ['load_balancer', 'url_rule', 'listener', 'url_rule_domain', 'target_group'],
          };
        },
      },
    },
  })
  res_type: OperationLogResourceType;
}
