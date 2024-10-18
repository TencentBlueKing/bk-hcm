import { type Columns } from 'bkui-vue/lib/table/props';
import i18n from '@/language/i18n';
import { costColumns } from '../../constants/columns';

const { t } = i18n.global;

export const billsProductSummaryColumns: Columns = [
  { label: t('一级账号名称'), field: 'root_account_name' },
  ...costColumns,
];
