/* eslint-disable @typescript-eslint/no-unused-vars */
import { Model, Column } from '@/decorator';
import { QueryRuleOPEnum } from '@/typings';
import { TaskClbType, TaskStatus, TaskSource } from '@/views/task/typings';
import { TASK_CLB_TYPE_NAME, TASK_SOURCE_NAME, TASK_STATUS_NAME } from '@/views/task/constants';
import { ResourceTypeEnum } from '@/common/resource-constant';

@Model('task/search.view')
export class SearchView {
  @Column('string')
  resource: ResourceTypeEnum;

  @Column('account', { name: '云账号', index: 0 })
  account_ids: string;

  @Column('enum', { name: '任务状态', option: TASK_STATUS_NAME, index: 2 })
  state: TaskStatus;

  @Column('enum', { name: '任务来源', option: TASK_SOURCE_NAME, index: 3 })
  source: TaskSource;

  @Column('datetime', { name: '操作时间', index: 4 })
  created_at: string;

  @Column('user', { name: '操作人', index: 5 })
  creator: string;
}

@Model('task/search-clb.view')
export class SearchClbView extends SearchView {
  @Column('enum', { name: '任务类型', index: 1, option: TASK_CLB_TYPE_NAME, op: QueryRuleOPEnum.JSON_OVERLAPS })
  operations: TaskClbType;
}
