import { QueryRuleOPEnum } from '@/typings/common';
import { Loading, SearchSelect, Table } from 'bkui-vue';
import type { Column } from 'bkui-vue/lib/table/props';
import { ISearchItem } from 'bkui-vue/lib/search-select/utils';
import { defineComponent, reactive, ref, watch } from 'vue';
import './index.scss';
import Empty from '@/components/empty';
import { useAccountStore, useResourceStore } from '@/store';
import { useWhereAmI } from '../useWhereAmI';

export interface IProp {
  columns: Array<Column>;
  searchData: Array<ISearchItem>;
  type: string; // 资源类型
  tableData?: Array<Record<string, any>>; // 临时看看效果
  noSearch?: boolean; // 是否不需要搜索
  tableExtraOptions?: object; // 额外的表格属性及事件
}

export const useTable = (props: IProp) => {
  const { isBusinessPage } = useWhereAmI();
  const resourceStore = useResourceStore();
  const accountStore = useAccountStore();
  const searchVal = ref('');
  const dataList = ref([]);
  const isLoading = ref(false);
  const pagination = reactive({
    start: 0,
    limit: 10,
    count: 100,
  });
  const filter = reactive({
    op: QueryRuleOPEnum.AND,
    rules: [],
  });
  const handlePageLimitChange = (v: number) => {
    pagination.limit = v;
  };
  const handlePageValueCHange = (v: number) => {
    pagination.start = (v - 1) * pagination.limit;
  };
  const getListData = async (customRules: Array<{
    op: QueryRuleOPEnum,
    field: string,
    value: string | number,
  }> = []) => {
    if (props.tableData) {
      dataList.value = props.tableData;
      return;
    }
    isLoading.value = true;
    const [detailsRes, countRes] = await Promise.all([false, true].map(isCount => resourceStore.list({
      page: {
        limit: isCount ? 0 : pagination.limit,
        start: isCount ? 0 : pagination.start,
        count: isCount,
      },
      filter: {
        op: filter.op,
        rules: [...filter.rules, ...customRules],
      },
    }, props.type)));
    dataList.value = detailsRes?.data?.details;
    pagination.count = countRes?.data?.count;
    isLoading.value = false;
  };
  const CommonTable = defineComponent({
    setup(_props, { slots }) {
      return () => (
        <>
          <section class='operation-wrap'>
            <div class='operate-btn-groups'>{slots.operation?.()}</div>
            {!props.noSearch && (
              <SearchSelect class='w500 common-search-selector' v-model={searchVal.value} data={props.searchData} />
            )}
          </section>
          <Loading loading={isLoading.value} class='loading-table-container'>
            <Table
              class='table-container'
              data={dataList.value}
              columns={props.columns}
              pagination={pagination}
              remotePagination
              showOverflowTooltip
              {...(props.tableExtraOptions || {})}
              onPageLimitChange={handlePageLimitChange}
              onPageValueChange={handlePageValueCHange}
              onColumnSort={() => {}}
              onColumnFilter={() => {}}>
                {{
                  empty: () => {
                    if (isLoading.value) return null;
                    return <Empty />;
                  },
                }}
              </Table>
          </Loading>
        </>
      );
    },
  });
  watch(
    () => pagination,
    () => {
      getListData();
    },
    {
      deep: true,
    },
  );
  watch(
    [
      () => searchVal.value,
      () => accountStore.bizs,
    ],
    ([searchVal, bizs]) => {
      if (isBusinessPage && !bizs) return;
      filter.rules = Array.isArray(searchVal) ? searchVal.map((val: any) => ({
        field: val?.id,
        op: val?.id === 'domain' ? QueryRuleOPEnum.JSON_CONTAINS : QueryRuleOPEnum.EQ,
        value: val?.values?.[0]?.id,
      })) : [];
      getListData();
    },
    {
      immediate: true,
    },
  );

  return {
    CommonTable,
    getListData,
  };
};
