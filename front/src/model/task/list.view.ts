/* eslint-disable @typescript-eslint/no-unused-vars */
import { Model, Column } from '@/decorator';
import { VendorEnum, VendorMap } from '@/common/constant';
import { type TaskType, TaskStatus, TaskSource } from '@/views/task/typings';
import { TASK_TYPE_NAME, TASK_SOURCE_NAME, TASK_STATUS_NAME } from '@/views/task/constants';

@Model('task/list.view')
export class ListView {
  @Column('account', { name: '云账号', sort: true })
  account_ids: string[];

  @Column('enum', { name: '云厂商', option: VendorMap, sort: true })
  vendors: VendorEnum[];

  @Column('datetime', { name: '操作时间', sort: true })
  created_at: string;

  @Column('enum', { name: '任务类型', option: TASK_TYPE_NAME, sort: true })
  operations: TaskType;

  @Column('enum', { name: '任务来源', option: TASK_SOURCE_NAME, sort: true })
  source: TaskSource;

  @Column('user', { name: '操作人' })
  creator: string;

  @Column('enum', {
    name: '任务状态',
    option: TASK_STATUS_NAME,
    meta: {
      display: {
        appearance: 'status',
      },
    },
  })
  state: TaskStatus;

  @Column('number', { name: '总数' })
  count_total: number;

  @Column('number', { name: '成功数' })
  count_success: number;

  @Column('number', { name: '失败数' })
  count_failed: number;
}
