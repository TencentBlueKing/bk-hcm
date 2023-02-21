export enum AuditActionEnum {
  CREATE = 'create',
  UPDATE = 'update',
  DELETE = 'delete',
  RESTART = 'restart',
}

export enum AuditActionNameEnum {
  CREATE = '创建',
  UPDATE = '更新',
  DELETE = '删除',
  RESTART = '重启',
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
  [AuditActionEnum.UPDATE]: AuditActionNameEnum.UPDATE,
  [AuditActionEnum.DELETE]: AuditActionNameEnum.DELETE,
  [AuditActionEnum.RESTART]: AuditActionNameEnum.RESTART,
};
