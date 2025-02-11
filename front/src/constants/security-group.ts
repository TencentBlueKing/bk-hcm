import i18n from '@/language/i18n';

const { t } = i18n.global;

export const MGMT_TYPE_MAP: Record<string, string> = {
  biz: t('业务管理'),
  platform: t('平台管理'),
  '': t('未确认'),
};
