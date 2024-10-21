import i18n from '@/language/i18n';
import { timeFormatter } from '@/common/util';
import { formatBillCost } from '@/utils';
import { BILL_ADJUSTMENT_STATE__MAP, BILL_ADJUSTMENT_TYPE__MAP, CURRENCY_MAP } from '@/constants';
import { IColumns } from '@/views/resource/resource-manage/hooks/use-columns';

const { t } = i18n.global;

export const billAdjustColumns: IColumns = [
  { type: 'selection', label: '', width: 30, minWidth: 30 },
  { label: t('更新时间'), field: 'updated_at', render: ({ cell }: any) => timeFormatter(cell) },
  { label: t('调账ID'), field: 'id' },
  { label: t('二级账号名称'), field: 'main_account_cloud_id' },
  {
    label: t('调账类型'),
    field: 'type',
    render: ({ cell }: any) => (
      <bk-tag theme={cell === 'increase' ? 'success' : 'danger'}>{BILL_ADJUSTMENT_TYPE__MAP[cell]}</bk-tag>
    ),
  },
  { label: t('操作人'), field: 'operator' },
  { label: t('金额'), field: 'cost', render: ({ cell }: any) => formatBillCost(cell) },
  { label: t('币种'), field: 'currency', render: ({ cell }: any) => CURRENCY_MAP[cell] || '--' },
  {
    label: t('调账状态'),
    field: 'state',
    width: 100,
    render: ({ cell }: any) => (
      <bk-tag theme={cell === 'confirmed' ? 'success' : undefined}>{BILL_ADJUSTMENT_STATE__MAP[cell]}</bk-tag>
    ),
  },
];
