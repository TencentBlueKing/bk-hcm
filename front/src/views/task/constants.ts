import { TaskClbType, TaskStatus, TaskSource, TaskDetailStatus } from './typings';

export const TASK_CLB_TYPE_NAME = {
  [TaskClbType.CREATE_L4_LISTENER]: '创建监听器-TCP/UDP',
  [TaskClbType.CREATE_L7_LISTENER]: '创建监听器-HTTP/HTTPS',
  [TaskClbType.CREATE_L7_RULE]: '创建URL规则-HTTP/HTTPS',
  [TaskClbType.DELETE_LISTENER]: '删除监听器-TCP/UDP/HTTP/HTTPS',
  [TaskClbType.BINDING_L4_RS]: '绑定RS-TCP/UDP',
  [TaskClbType.BINDING_L7_RS]: '绑定RS-HTTP/HTTPS',
  [TaskClbType.UNBIND_LAYER4_RS]: '解绑RS-TCP/UDP',
  [TaskClbType.MODIFY_LAYER4_RS_WEIGHT]: '权重调整-TCP/UDP',
  [TaskClbType.MODIFY_LAYER7_RS_WEIGHT]: '权重调整-HTTP/HTTPS',
};

export const TASK_TYPE_NAME = {
  ...TASK_CLB_TYPE_NAME,
};

export const TASK_STATUS_NAME = {
  [TaskStatus.RUNNING]: '执行中',
  [TaskStatus.FAILED]: '失败',
  [TaskStatus.SUCCESS]: '成功',
  [TaskStatus.DELIVER_PARTIAL]: '部分成功',
  [TaskStatus.CANCELED]: '已取消',
};

export const TASK_SOURCE_NAME = {
  [TaskSource.SOPS]: '标准运维',
  [TaskSource.EXCEL]: 'Excel导入',
  [TaskSource.API]: '页面操作',
};

export const TASK_DETAIL_STATUS_NAME = {
  [TaskDetailStatus.INIT]: '待执行',
  [TaskDetailStatus.RUNNING]: '运行',
  [TaskDetailStatus.FAILED]: '失败',
  [TaskDetailStatus.SUCCESS]: '成功',
  [TaskDetailStatus.CANCEL]: '取消',
};
