import { PropType, Ref, defineComponent, inject, onMounted, ref } from 'vue';
import './index.scss';

import { Button } from 'bkui-vue';
import BillsExportButton from '../../components/bills-export-button';
import ImportBillDetailDialog from '../ImportBillDetailDialog';

import { useI18n } from 'vue-i18n';
import useColumns from '@/views/resource/resource-manage/hooks/use-columns';
import { useTable } from '@/hooks/useTable/useTable';
import { VendorEnum } from '@/common/constant';
import { exportBillsItems, reqBillsItemList, reqBillsRootAccountSummaryList } from '@/api/bill';
import { QueryRuleOPEnum, RulesItem } from '@/typings';

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

    // 是否可以导入
    const disableImport = ref(true);

    onMounted(() => {
      const checkBillsState = async () => {
        const res = await reqBillsRootAccountSummaryList({
          bill_year: bill_year.value,
          bill_month: bill_month.value,
          filter: {
            op: QueryRuleOPEnum.AND,
            rules: [
              { field: 'vendor', op: QueryRuleOPEnum.EQ, value: VendorEnum.ZENLAYER },
              { field: 'state', op: QueryRuleOPEnum.NEQ, value: 'accounting' },
            ],
          },
          page: { count: true, start: 0, limit: 0 },
        });
        // 所有zenlayer账号都处在accounting 核算中的状态，才能进行导入
        disableImport.value = res.data.count > 0;
      };

      props.vendor === VendorEnum.ZENLAYER && checkBillsState();
    });

    expose({ reloadTable });

    return () => (
      <div class='bill-detail-render-table-container'>
        <CommonTable>
          {{
            operation: () => (
              <>
                {props.vendor === VendorEnum.ZENLAYER && (
                  <Button
                    onClick={() => importBillDetailDialogRef.value.triggerShow(true)}
                    disabled={disableImport.value}
                    v-bk-tooltips={{
                      content: t('所有zenlayer账号都处在accounting 核算中的状态，才能进行导入'),
                      disabled: !disableImport.value,
                    }}>
                    {t('导入')}
                  </Button>
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
                  title={t('账单汇总-一级账号')}
                  content={t('导出当月一级账号的账单数据')}
                />
              </>
            ),
          }}
        </CommonTable>
        <ImportBillDetailDialog ref={importBillDetailDialogRef} vendor={props.vendor} onUpdateTable={getListData} />
      </div>
    );
  },
});
