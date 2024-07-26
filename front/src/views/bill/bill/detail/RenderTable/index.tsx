import { PropType, Ref, defineComponent, inject, ref, watch } from 'vue';
import './index.scss';

import { Button } from 'bkui-vue';
import BillsExportButton from '../../components/bills-export-button';
import ImportBillDetailDialog from '../ImportBillDetailDialog';

import { useI18n } from 'vue-i18n';
import useColumns from '@/views/resource/resource-manage/hooks/use-columns';
import { useTable } from '@/hooks/useTable/useTable';
import { VendorEnum, VendorMap } from '@/common/constant';
import { exportBillsItems, reqBillsItemList } from '@/api/bill';
import { RulesItem } from '@/typings';

export default defineComponent({
  name: 'BillDetailRenderTable',
  props: { vendor: String as PropType<VendorEnum> },
  setup(props, { expose }) {
    const { t } = useI18n();
    const bill_year = inject<Ref<number>>('bill_year');
    const bill_month = inject<Ref<number>>('bill_month');

    const importBillDetailDialogRef = ref();

    const getColumnName = () => {
      switch (props.vendor) {
        case VendorEnum.AWS:
          return 'billDetailAws';
        case VendorEnum.AZURE:
          return 'billDetailAzure';
        case VendorEnum.GCP:
          return 'billDetailGcp';
        case VendorEnum.HUAWEI:
          return 'billDetailHuawei';
        case VendorEnum.ZENLAYER:
          return 'billDetailZenlayer';
      }
    };

    const { columns, settings } = useColumns(getColumnName());

    const { CommonTable, getListData, clearFilter, filter } = useTable({
      tableOptions: {
        columns,
        extra: {
          settings: settings.value,
        },
      },
      searchOptions: {
        disabled: true,
      },
      requestOption: {
        apiMethod: reqBillsItemList,
        extension: () => ({ vendor: props.vendor, bill_year: bill_year.value, bill_month: bill_month.value }),
        immediate: false,
      },
    });

    const reloadTable = (rules: RulesItem[]) => {
      clearFilter();
      getListData(rules);
    };

    watch([bill_year, bill_month], () => {
      getListData();
    });

    expose({ reloadTable });

    return () => (
      <div class='bill-detail-render-table-container'>
        <CommonTable>
          {{
            operation: () => (
              <>
                {props.vendor === VendorEnum.ZENLAYER && (
                  <Button onClick={() => importBillDetailDialogRef.value.triggerShow(true)}>{t('导入')}</Button>
                )}
                <BillsExportButton
                  cb={() =>
                    exportBillsItems(props.vendor, {
                      bill_year: bill_year.value,
                      bill_month: bill_month.value,
                      export_limit: 200000,
                      filter,
                    })
                  }
                  fileName={t(`账单明细-${VendorMap[props.vendor]}`)}
                />
              </>
            ),
          }}
        </CommonTable>
        <ImportBillDetailDialog ref={importBillDetailDialogRef} vendor={props.vendor} />
      </div>
    );
  },
});
