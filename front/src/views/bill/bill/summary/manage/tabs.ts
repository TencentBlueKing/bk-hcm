import i18n from '@/language/i18n';
import PrimaryAccount from '@/views/bill/bill/summary/primary';
import SubAccount from '@/views/bill/bill/summary/sub';

const { t } = i18n.global;

export const baseTabs = [
  { name: 'primary', label: t('一级账号'), Component: PrimaryAccount },
  { name: 'sub', label: t('二级账号'), Component: SubAccount },
];
