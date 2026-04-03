import type { ISecretActionConfig, SecretActionType } from './typings';

export const SECRET_STATUS_MAP: Record<string, { class: string; text: string; iconName: string }> = {
  enabled: { class: 'status-enabled', text: '已启用', iconName: 'normal' },
  disabled: { class: 'status-disabled', text: '已禁用', iconName: 'unknown' },
};

export const CONSOLE_LOGIN_MAP: Record<number, string> = {
  0: '编程账号',
  1: '控制台账号',
};

export const SECRET_ACTION_CONFIG: Record<SecretActionType, ISecretActionConfig> = {
  disable: {
    type: 'disable',
    title: '确认禁用密钥？',
    alertType: 'warning',
    alertMessage: '禁用此密钥后，腾讯云将拒绝此密钥的所有请求。\n禁用密钥预计15分钟内生效。',
    confirmText: '禁用',
    confirmTheme: 'danger',
  },
  enable: {
    type: 'enable',
    title: '确认启用密钥？',
    alertType: 'warning',
    alertMessage: '开启密钥时，请确保密钥处于安全状态',
    confirmText: '启用',
    confirmTheme: 'primary',
  },
  delete: {
    type: 'delete',
    title: '确认删除密钥？',
    alertType: 'error',
    alertMessage: '删除此密钥后，腾讯云将永久拒绝此密钥的所有请求。\n此操作不可恢复，请谨慎操作。',
    confirmText: '删除',
    confirmTheme: 'danger',
  },
};
