import i18n from '@/language/i18n';
import { costColumns } from '../../constants/columns';
import { IColumns } from '@/views/resource/resource-manage/hooks/use-columns';

const { t } = i18n.global;

export const billsProductSummaryColumns: IColumns = [
  { label: t('一级账号名称'), field: 'root_account_name' },
  ...costColumns,
];
