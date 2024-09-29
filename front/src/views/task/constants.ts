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
