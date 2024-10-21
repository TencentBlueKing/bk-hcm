import { useI18n } from 'vue-i18n';
import useColumns from '@/views/resource/resource-manage/hooks/use-columns';
import { useBusinessMapStore } from '@/store/useBusinessMap';

import { VendorEnum } from '@/common/constant';
import {
  billDetailAwsColumns,
  billDetailAzureColumns,
  billDetailGcpColumns,
  billDetailHuaweiColumns,
  billDetailZenlayerColumns,
} from './columns';

export const getColumns = (vendor: VendorEnum) => {
  let columns;
  const { t } = useI18n();
  const businessMapStore = useBusinessMapStore();
  const { generateColumnsSettings } = useColumns(null);

  switch (vendor) {
    case VendorEnum.AWS:
      columns = billDetailAwsColumns.slice();
      break;
    case VendorEnum.AZURE:
      columns = billDetailAzureColumns.slice();
      break;
    case VendorEnum.GCP:
      columns = billDetailGcpColumns.slice();
      break;
    case VendorEnum.HUAWEI:
      columns = billDetailHuaweiColumns.slice();
      break;
    case VendorEnum.ZENLAYER:
      columns = billDetailZenlayerColumns.slice();
      break;
  }

  columns.splice(5, 0, {
    label: t('业务名称'),
    field: 'bk_biz_id',
    render: ({ data }: any) => businessMapStore.businessMap.get(data.bk_biz_id) || '未分配',
  });

  const settings = generateColumnsSettings(columns);

  return { columns, settings };
};
