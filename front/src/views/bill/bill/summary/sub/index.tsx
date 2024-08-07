import { Ref, defineComponent, inject, onMounted, ref, watch } from 'vue';
import Search from '../../components/search';
import Button from '../../components/button';
import Amount from '../../components/amount';
import { useTable } from '@/hooks/useTable/useTable';
import useColumns from '@/views/resource/resource-manage/hooks/use-columns';
import { reqBillsMainAccountSummaryList, reqBillsMainAccountSummarySum } from '@/api/bill';
import { QueryRuleOPEnum, RulesItem } from '@/typings';
import { useRoute } from 'vue-router';
import { BILL_BIZS_KEY, BILL_MAIN_ACCOUNTS_KEY } from '@/constants';

export default defineComponent({
  name: 'SubAccountTabPanel',
  setup() {
    const route = useRoute();
    const bill_year = inject<Ref<number>>('bill_year');
    const bill_month = inject<Ref<number>>('bill_month');

    const searchRef = ref();
    const amountRef = ref();

    const { columns } = useColumns('billsMainAccountSummary');
    const { CommonTable, getListData, clearFilter, filter } = useTable({
      searchOptions: { disabled: true },
      tableOptions: {
        columns,
      },
      requestOption: {
        sortOption: {
          sort: 'current_month_rmb_cost',
          order: 'DESC',
        },
        apiMethod: reqBillsMainAccountSummaryList,
        extension: () => ({
          bill_year: bill_year.value,
          bill_month: bill_month.value,
        }),
        immediate: false,
      },
    });

    const reloadTable = (rules: RulesItem[]) => {
      clearFilter();
      getListData(rules);
    };

    watch([bill_year, bill_month], () => {
      searchRef.value.handleSearch();
    });

    watch(filter, () => {
      amountRef.value.refreshAmountInfo();
    });

    onMounted(() => {
      // 只有业务、二级账号有保存的需求
      const rules = [];
      if (route.query[BILL_MAIN_ACCOUNTS_KEY]) {
        rules.push({
          field: 'main_account_id',
          op: QueryRuleOPEnum.IN,
          value: JSON.parse(atob(route.query[BILL_MAIN_ACCOUNTS_KEY] as string)),
        });
      }
      if (route.query[BILL_BIZS_KEY]) {
        rules.push({
          field: 'bk_biz_id',
          op: QueryRuleOPEnum.IN,
          value: JSON.parse(atob(route.query[BILL_BIZS_KEY] as string)),
        });
      }
      reloadTable(rules);
    });

    return () => (
      <>
        <Search
          ref={searchRef}
          searchKeys={['vendor', 'root_account_id', 'product_id', 'main_account_id']}
          onSearch={reloadTable}
          autoSelectMainAccount
        />
        <div class='p24' style={{ height: 'calc(100% - 238px)' }}>
          <CommonTable>
            {{
              operation: () => <Button noSyncBtn />,
              operationBarEnd: () => (
                <Amount
                  ref={amountRef}
                  api={reqBillsMainAccountSummarySum}
                  payload={() => ({ bill_year: bill_year.value, bill_month: bill_month.value, filter })}
                />
              ),
            }}
          </CommonTable>
        </div>
      </>
    );
  },
});
