import i18n from '@/language/i18n';
import { formatBillCost } from '@/utils';
import { IColumns } from '@/views/resource/resource-manage/hooks/use-columns';

const { t } = i18n.global;

export const costColumns: IColumns = [
  {
    label: t('已确认账单人民币（元）'),
    field: 'current_month_rmb_cost_synced',
    render: ({ cell }: any) => formatBillCost(cell),
    sort: true,
  },
  {
    label: t('已确认账单美金（美元）'),
    field: 'current_month_cost_synced',
    render: ({ cell }: any) => formatBillCost(cell),
    sort: true,
  },
  {
    label: t('当前账单人民币（元）'),
    field: 'current_month_rmb_cost',
    render: ({ cell }: any) => formatBillCost(cell),
    sort: true,
  },
  {
    label: t('当前账单美金（美元）'),
    field: 'current_month_cost',
    render: ({ cell }: any) => formatBillCost(cell),
    sort: true,
  },
  {
    label: t('调账人民币（元）'),
    field: 'adjustment_rmb_cost',
    render: ({ cell }: any) => formatBillCost(cell),
    sort: true,
  },
  {
    label: t('调账美金（美元）'),
    field: 'adjustment_cost',
    render: ({ cell }: any) => formatBillCost(cell),
    sort: true,
  },
];
