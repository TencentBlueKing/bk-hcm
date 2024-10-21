import i18n from '@/language/i18n';
import { timeFormatter } from '@/common/util';
import { formatBillCost } from '@/utils';
import { VendorEnum, VendorMap } from '@/common/constant';
import { BILL_TYPE__MAP_HW, CURRENCY_MAP } from '@/constants';
import { IColumns } from '@/views/resource/resource-manage/hooks/use-columns';

const { t } = i18n.global;

export const billDetailBaseColumns: IColumns = [
  {
    label: t('核算日期'),
    field: 'bill_date',
    render: ({ data: { bill_year, bill_month, bill_day } }: any) =>
      timeFormatter(new Date(bill_year, bill_month - 1, bill_day), 'YYYYMMDD'),
  },
  { label: t('ID'), field: 'id', isDefaultShow: true },
  { label: t('一级账号ID'), field: 'root_account_id', isDefaultShow: true },
  { label: t('二级账号ID'), field: 'main_account_id', isDefaultShow: true },
  {
    label: t('云厂商'),
    field: 'vendor',
    isDefaultShow: true,
    render: ({ cell }: { cell: VendorEnum }) => VendorMap[cell],
  },
  { label: t('币种'), field: 'currency', isDefaultShow: true, render: ({ cell }: any) => CURRENCY_MAP[cell] },
  {
    label: t('本期应付金额'),
    field: 'cost',
    isDefaultShow: true,
    render: ({ cell }: any) => formatBillCost(cell),
    sort: true,
  },
  { label: t('资源类型编码'), field: 'hc_product_code', isDefaultShow: true },
  { label: t('产品名称'), field: 'hc_product_name', isDefaultShow: true },
  { label: t('预留实例使用量'), field: 'res_amount', isDefaultShow: true },
  { label: t('预留实例使用单位'), field: 'res_amount_unit', isDefaultShow: true },
];

export const billDetailAwsColumns: IColumns = [...billDetailBaseColumns];

export const billDetailAzureColumns: IColumns = [...billDetailBaseColumns];

export const billDetailGcpColumns: IColumns = [...billDetailBaseColumns];

export const billDetailHuaweiColumns: IColumns = [
  ...billDetailBaseColumns,
  { label: t('使用量类型'), field: 'extension.usage_type' },
  { label: t('使用量'), field: 'extension.usage' },
  { label: t('使用量度量单位'), field: 'extension.unit' },
  { label: t('云服务类型编码'), field: 'extension.cloud_service_type' },
  { label: t('云服务类型名称'), field: 'extension.cloud_service_type_name' },
  { label: t('云服务区编码'), field: 'extension.region' },
  { label: t('云服务区名称'), field: 'extension.region_name' },
  { label: t('资源类型编码'), field: 'extension.resource_type' },
  { label: t('资源类型名称'), field: 'extension.resource_type_name' },
  { label: t('计费模式'), field: 'extension.charge_mode' },
  { label: t('账单类型'), field: 'extension.bill_type', render: ({ cell }: any) => BILL_TYPE__MAP_HW[cell] },
];

export const billDetailZenlayerColumns: IColumns = [...billDetailBaseColumns];
