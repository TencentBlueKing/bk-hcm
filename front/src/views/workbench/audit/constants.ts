export enum AuditActionEnum {
  CREATE = 'create',
  MOUNT = 'associate',
  UPDATE = 'update',
  DELETE = 'delete',
  UNMOUNT = 'disassociate',
  APPLY = 'apply',
  ASSIGN = 'assign',
  REBOOT = 'reboot',
  START = 'start',
  STOP = 'stop',
  RESET_PWD = 'reset_pwd',
  RECYCLE = 'recycle',
}

export enum AuditActionNameEnum {
  CREATE = '创建',
  MOUNT = '挂载',
  UPDATE = '更新',
  DELETE = '删除',
  UNMOUNT = '卸载',
  APPLY = '申请',
  ASSIGN = '分配',
  EDIT = '编辑',
  REBOOT = '重启',
  START = '开机',
  STOP = '关机',
  RESET_PWD = '重置密码',
  RECYCLE = '回收',
}

export enum AuditSourceEnum {
  API_CALL = 'api_call',
  BACKGROUND_SYNC = 'background_sync'
}

export const AUDIT_SOURCE_MAP = {
  [AuditSourceEnum.API_CALL]: 'API调用',
  [AuditSourceEnum.BACKGROUND_SYNC]: '后台同步',
};

export const AUDIT_ACTION_MAP = {
  [AuditActionEnum.CREATE]: AuditActionNameEnum.CREATE,
  [AuditActionEnum.MOUNT]: AuditActionNameEnum.MOUNT,
  [AuditActionEnum.UNMOUNT]: AuditActionNameEnum.UNMOUNT,
  [AuditActionEnum.UPDATE]: AuditActionNameEnum.UPDATE,
  [AuditActionEnum.DELETE]: AuditActionNameEnum.DELETE,
  [AuditActionEnum.APPLY]: AuditActionNameEnum.APPLY,
  [AuditActionEnum.ASSIGN]: AuditActionNameEnum.ASSIGN,
  [AuditActionEnum.REBOOT]: AuditActionNameEnum.REBOOT,
  [AuditActionEnum.START]: AuditActionNameEnum.START,
  [AuditActionEnum.STOP]: AuditActionNameEnum.STOP,
  [AuditActionEnum.RESET_PWD]: AuditActionNameEnum.RESET_PWD,
};
