import i18n from '@/language/i18n';
import { SecurityGroupManageType } from '@/store/security-group';

const { t } = i18n.global;

export const MGMT_TYPE_MAP: Record<string, string> = {
  [SecurityGroupManageType.BIZ]: t('业务管理'),
  [SecurityGroupManageType.PLATFORM]: t('平台管理'),
  [SecurityGroupManageType.UNKNOWN]: t('未确认'),
};
