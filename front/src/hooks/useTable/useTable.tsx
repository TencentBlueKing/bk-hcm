import { QueryRuleOPEnum, RulesItem } from '@/typings/common';
import { FilterType } from '@/typings';
import { Loading, SearchSelect, Table } from 'bkui-vue';
import type { Column } from 'bkui-vue/lib/table/props';
import { ISearchItem } from 'bkui-vue/lib/search-select/utils';
import { defineComponent, reactive, ref, watch } from 'vue';
import './index.scss';
import Empty from '@/components/empty';
import { useAccountStore, useResourceStore } from '@/store';
import { useBusinessMapStore } from '@/store/useBusinessMap';
import { useWhereAmI } from '../useWhereAmI';
export interface IProp {
  // search-select 相关字段
  searchOptions: {
    searchData: Array<ISearchItem>; // search-select 可选项
    disabled?: boolean, // 是否禁用 search-select
    extra?: Object, // 其他 search-select 属性/自定义事件, 比如 placeholder, onSearch...
  },
  // table 相关字段
  tableOptions: {
    columns: Array<Column>; // 表格字段
    reviewData?: Array<Record<string, any>>; // 用于预览效果的数据
    extra?: Object, // 其他 table 属性/自定义事件, 比如 settings, onSelectionChange...
  },
  // 请求相关字段
  requestOption: {
    type: string, // 资源类型
    sort?: string, // 需要排序的字段, 与 order 配合使用
    order?: 'ASC' | 'DESC', // 排序方式, 与 sort 配合使用
    rules?: Array<RulesItem>, // 筛选规则
  },
  // 资源下筛选业务功能相关的 prop
  bizFilter?: FilterType,
}

export const useTable = (props: IProp) => {
  const { isBusinessPage, isResourcePage } = useWhereAmI();
  const resourceStore = useResourceStore();
  const accountStore = useAccountStore();
  const businessMapStore = useBusinessMapStore();
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
    pagination.start = 0;
    getListData();
  };
  const handlePageValueCHange = (v: number) => {
    pagination.start = (v - 1) * pagination.limit;
    getListData();
  };
  const getListData = async (customRules: Array<{
    op: QueryRuleOPEnum,
    field: string,
    value: string | number,
  }> = []) => {
    // 预览
    if (props.tableOptions.reviewData) {
      dataList.value = props.tableOptions.reviewData;
      return;
    }
    isLoading.value = true;
    const [detailsRes, countRes] = await Promise.all([false, true].map(isCount => resourceStore.list({
      page: {
        limit: isCount ? 0 : pagination.limit,
        start: isCount ? 0 : pagination.start,
        sort: isCount ? null : (props.requestOption.sort || ''),
        order: isCount ? null : (props.requestOption.order || ''),
        count: isCount,
      },
      filter: {
        op: filter.op,
        rules: [...filter.rules, ...customRules],
      },
    }, props.requestOption.type)));
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
            {!props.searchOptions.disabled && (
              <SearchSelect
                class='w500'
                v-model={searchVal.value}
                data={props.searchOptions.searchData}
                {...(props.searchOptions.extra || {})}
              />
            )}
          </section>
          <Loading loading={isLoading.value} class='loading-table-container'>
            <Table
              class='table-container'
              data={dataList.value}
              columns={props.tableOptions.columns}
              pagination={pagination}
              remotePagination
              showOverflowTooltip
              {...(props.tableOptions.extra || {})}
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
    [
      () => searchVal.value,
      () => accountStore.bizs,
    ],
    ([searchVal, bizs]) => {
      if (isBusinessPage && !bizs) return;
      filter.rules = Array.isArray(searchVal) ? searchVal.map((val: any) => {
        const field = val?.id;
        const op = val?.id === 'domain' ? QueryRuleOPEnum.JSON_CONTAINS : QueryRuleOPEnum.EQ;
        const value = field === 'bk_biz_id'
          ? (businessMapStore.businessNameToIDMap.get(val?.values?.[0]?.id) || Number(val?.values?.[0]?.id))
          : val?.values?.[0]?.id;
        return { field, op, value };
      }) : [];
      // 页码重置
      pagination.start = 0;
      getListData();
    },
    {
      immediate: true,
    },
  );

  // 分配业务筛选
  watch(() => props.bizFilter, (val) => {
    if (isResourcePage) searchVal.value = '';
    const idx = filter.rules.findIndex(rule => rule.field === 'bk_biz_id');
    const bizFilter = val.rules[0];
    if (bizFilter) {
      if (idx !== -1) {
        filter.rules[idx] = bizFilter;
      } else {
        filter.rules.push(val.rules[0]);
      }
    } else {
      filter.rules.splice(idx, 1);
    }
    getListData();
  }, { deep: true });

  watch(() => props.requestOption.rules, (val) => {
    if (!val) return;
    const idx = filter.rules.findIndex(rule => rule.field === 'res_type');
    if (idx === -1) {
      filter.rules.push(...val);
    } else {
      const rule = val[0];
      if (!rule.value) {
        filter.rules.splice(idx, 1);
      } else {
        filter.rules[idx] = rule;
      }
    }
    getListData();
  }, {
    deep: true,
  });

  return {
    CommonTable,
    getListData,
  };
};
