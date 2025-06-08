/* eslint-disable @typescript-eslint/no-unused-vars */
import { Model, Column } from '@/decorator';
import { OPERATION_LOG_ACTION_NAME, OPERATION_LOG_SOURCE_NAME } from '@/views/operation-log/constants';
import type { OperationLogAction, OperationLogSource } from '@/views/operation-log/typings';

@Model('operation-log/table-column')
export class TableColumn {
  @Column('datetime', { name: '操作时间', sort: true, index: 0 })
  created_at: string;

  @Column('string', { name: '资源类型', index: 0 })
  res_type: string;

  @Column('string', { name: '资源名称', index: 0 })
  res_name: string;

  @Column('enum', { name: '操作方式', option: OPERATION_LOG_ACTION_NAME, index: 0 })
  action: OperationLogAction;

  @Column('enum', { name: '操作来源', option: OPERATION_LOG_SOURCE_NAME, index: 0 })
  source: OperationLogSource;

  @Column('business', { name: '所属业务', index: 0 })
  bk_biz_id: number;

  @Column('string', { name: '云账号', defaultHidden: true, index: 0 })
  account_id: string;

  @Column('user', { name: '操作人', index: 0 })
  operator: string;
}
