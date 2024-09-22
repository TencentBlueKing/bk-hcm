import { VendorMap } from '@/common/constant';
import { ResourceProperty } from '@/common/resource-constant';
import { TaskType, TaskStatus, TaskSource } from './typings';

export const TASKT_TYPE_NAME = {
  [TaskType.CREATE_L4_LISTENER]: '创建监听器-TCP/UDP',
  [TaskType.CREATE_L7_LISTENER]: '创建监听器-HTTP/HTTPS',
  [TaskType.CREATE_URL_FILTER]: '创建URL规则-HTTP/HTTPS',
};

export const TASKT_STATUS_NAME = {
  [TaskStatus.RUNNING]: '执行中',
  [TaskStatus.FAILED]: '失败',
  [TaskStatus.SUCCESS]: '成功',
  [TaskStatus.DELIVER_PARTIAL]: '部分成功',
  [TaskStatus.CANCELED]: '已取消',
};

export const TASKT_SOURCE_NAME = {
  [TaskSource.SOPS]: '标准运维',
  [TaskSource.EXCEL]: '页面操作',
};

export const COMMON_PROPERTIES: ResourceProperty[] = [
  {
    id: 'account_id',
    name: '云账号',
    type: 'account',
  },
  {
    id: 'vendor',
    name: '云厂商',
    type: 'enum',
    option: VendorMap,
  },
];

export const TASK_BASE_PROPERIES: ResourceProperty[] = [
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
];
