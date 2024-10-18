import { Ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { exportBillsBizSummary } from '@/api/bill';

import BillsExportButton from '../../components/bills-export-button';

export const renderOperationComp = (bill_year: number, bill_month: number, searchRef: Ref<any>) => {
  const { t } = useI18n();

  return (
    <BillsExportButton
      cb={() =>
        exportBillsBizSummary({
          bill_year,
          bill_month,
          export_limit: 200000,
          bk_biz_ids: searchRef.value.rules.find((rule: any) => rule.field === 'bk_biz_id')?.value || [],
        })
      }
      title={t('账单汇总-业务')}
      content={t('导出当月业务的账单数据')}
    />
  );
};
