import { type Columns } from 'bkui-vue/lib/table/props';
import i18n from '@/language/i18n';
import { VendorEnum, VendorMap } from '@/common/constant';
import { costColumns } from '../../constants/columns';

const { t } = i18n.global;

export const billsMainAccountSummaryColumns: Columns = [
  { label: t('二级账号ID'), field: 'main_account_cloud_id' },
  { label: t('二级账号名称'), field: 'main_account_name' },
  { label: t('云厂商'), field: 'vendor', render: ({ cell }: { cell: VendorEnum }) => VendorMap[cell] },
  { label: t('一级账号名称'), field: 'root_account_name' },
  ...costColumns,
];
