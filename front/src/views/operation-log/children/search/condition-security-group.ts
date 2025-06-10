/* eslint-disable @typescript-eslint/no-unused-vars */
import { Model, Column } from '@/decorator';
import { OPERATION_LOG_ACTION, OPERATION_LOG_ACTION_NAME } from '@/views/operation-log/constants';
import type { OperationLogAction, OperationLogResourceType } from '@/views/operation-log/typings';
import { SearchCondition } from './condition';

@Model('operation-log/search-condition-security-group')
export class SearchConditionSecurityGroup extends SearchCondition {
  static actionOption = {
    [OPERATION_LOG_ACTION.CREATE]: OPERATION_LOG_ACTION_NAME[OPERATION_LOG_ACTION.CREATE],
    [OPERATION_LOG_ACTION.UPDATE]: OPERATION_LOG_ACTION_NAME[OPERATION_LOG_ACTION.UPDATE],
    [OPERATION_LOG_ACTION.DELETE]: OPERATION_LOG_ACTION_NAME[OPERATION_LOG_ACTION.DELETE],
  };

  @Column('enum', { name: '操作方式', option: SearchConditionSecurityGroup.actionOption, index: 1 })
  action: OperationLogAction;

  @Column('enum', { apiOnly: true })
  res_type: OperationLogResourceType;
}
