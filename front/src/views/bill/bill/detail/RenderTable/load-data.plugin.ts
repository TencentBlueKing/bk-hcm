import useColumns from '@/views/resource/resource-manage/hooks/use-columns';

import { VendorEnum } from '@/common/constant';
import {
  billDetailAwsColumns,
  billDetailAzureColumns,
  billDetailGcpColumns,
  billDetailHuaweiColumns,
  billDetailZenlayerColumns,
} from './columns';
import { injectBizField } from '@/utils';

export const getColumns = (vendor: VendorEnum) => {
  let columns;
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

  injectBizField(columns, 5);

  const settings = generateColumnsSettings(columns);

  return { columns, settings };
};
