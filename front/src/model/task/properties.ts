import { ModelProperty } from '@/model/typings';
import { TASKT_TYPE_NAME, TASKT_SOURCE_NAME, TASKT_STATUS_NAME } from '@/views/task/constants';

export default [
  {
    id: 'created_at',
    name: '操作时间',
    type: 'datetime',
    index: 1,
  },
  {
    id: 'operations',
    name: '任务类型',
    type: 'enum',
    index: 1,
    option: TASKT_TYPE_NAME,
  },
  {
    id: 'source',
    name: '任务来源',
    type: 'enum',
    index: 1,
    option: TASKT_SOURCE_NAME,
  },
  {
    id: 'creator',
    name: '操作人',
    type: 'user',
    index: 1,
  },
  {
    id: 'state',
    name: '任务状态',
    type: 'enum',
    index: 1,
    option: TASKT_STATUS_NAME,
  },
] as ModelProperty[];
