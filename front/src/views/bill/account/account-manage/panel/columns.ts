import { type Columns } from 'bkui-vue/lib/table/props';
import i18n from '@/language/i18n';
import { BILL_SITE_TYPES_MAP, BILL_VENDORS_MAP } from '../constants';

const { t } = i18n.global;

// 一级账号
export const firstAccountColumns: Columns = [
  { label: t('一级帐号ID'), field: 'cloud_id' },
  { label: t('云厂商'), field: 'vendor', render: ({ cell }: any) => BILL_VENDORS_MAP[cell] || '--' },
  { label: t('帐号邮箱'), field: 'email' },
  { label: t('主负责人'), field: 'managers', render: ({ cell }: any) => cell.join(',') },
  { label: t('备注'), field: 'memo' },
];

// 二级账号
export const secondaryAccountColumns = [
  { label: t('二级账号ID'), field: 'cloud_id' },
  { label: t('所属一级帐号'), field: 'parent_account_name' },
  { label: t('云厂商'), field: 'vendor', render: ({ cell }: any) => BILL_VENDORS_MAP[cell] || '--' },
  { label: t('站点类型'), field: 'site', render: ({ cell }: any) => BILL_SITE_TYPES_MAP[cell] },
  { label: t('帐号邮箱'), field: 'email' },
  { label: t('主负责人'), field: 'managers', render: ({ cell }: any) => cell.join(',') },
  { label: t('备注'), field: 'memo' },
];
