import { useI18n } from 'vue-i18n';
import { FilterType } from '@/typings';
import { exportBillsRootAccountSummary } from '@/api/bill';

import BillsExportButton from '../../components/bills-export-button';

export const renderOperationComp = (bill_year: number, bill_month: number, filter: FilterType) => {
  const { t } = useI18n();

  return (
    <BillsExportButton
      cb={() => exportBillsRootAccountSummary({ bill_year, bill_month, export_limit: 200000, filter })}
      title={t('账单汇总-一级账号')}
      content={t('导出当月一级账号的账单数据')}
    />
  );
};
