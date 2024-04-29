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
import { getDifferenceSet } from '@/common/util';
import { get as lodash_get } from 'lodash-es';
export interface IProp {
  // search-select 相关字段
  searchOptions: {
    searchData?: Array<ISearchItem>; // search-select 可选项
    disabled?: boolean; // 是否禁用 search-select
    extra?: {
      searchSelectExtStyle?: Record<string, string>; // 搜索框样式
    }; // 其他 search-select 属性/自定义事件, 比如 placeholder, onSearch, searchSelectExtStyle...
  };
  // table 相关字段
  tableOptions: {
    columns: Array<Column>; // 表格字段
    reviewData?: Array<Record<string, any>>; // 用于预览效果的数据
    extra?: Object; // 其他 table 属性/自定义事件, 比如 settings, onSelectionChange...
  };
  // 模糊查询开关true开启，false关闭
  fuzzySwitch?: boolean;
  // 请求相关字段
  requestOption: {
    type: string; // 资源类型
    sortOption?: {
      sort: string; // 需要排序的字段
      order: 'ASC' | 'DESC'; // 排序方式
    }; // 排序参数
    filterOption?: {
      rules: Array<RulesItem>; // 规则
      deleteOption?: {
        field: string;
        flagValue: string; // 当 rule.value = flagValue 时, 删除该 rule
      }; // Tab 切换时选用项(如选中全部时, 删除对应的 rule)
    }; // 筛选参数
    extension?: Record<string, any>; // 请求需要的额外荷载数据
    callback?: (...args: any) => any; // 可以根据当前请求结果异步更新 dataList
    dataPath?: string; // 列表数据的路径，如 data.details
  };
  // 资源下筛选业务功能相关的 prop
  bizFilter?: FilterType;
}

export const useTable = (props: IProp) => {
  const { isBusinessPage } = useWhereAmI();
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
  const sort = ref(props.requestOption.sortOption ? props.requestOption.sortOption.sort : 'created_at');
  const order = ref(props.requestOption.sortOption ? props.requestOption.sortOption.order : 'DESC');
  const filter = reactive({
    op: QueryRuleOPEnum.AND,
    rules: [],
  });
  const handlePageLimitChange = (v: number) => {
    pagination.limit = v;
    pagination.start = 0;
    getListData();
  };
  const handlePageValueChange = (v: number) => {
    pagination.start = (v - 1) * pagination.limit;
    getListData();
  };
  const handleSort = ({ column, type }: any) => {
    pagination.start = 0;
    sort.value = column.field;
    order.value = type === 'asc' ? 'ASC' : 'DESC';
    // 如果type为null，则默认排序
    if (type === 'null') {
      sort.value = props.requestOption.sortOption ? props.requestOption.sortOption.sort : 'created_at';
      order.value = props.requestOption.sortOption ? props.requestOption.sortOption.order : 'DESC';
    }
    getListData();
  };
  const getListData = async (customRules: Array<RulesItem> = [], type?: string) => {
    // 预览
    if (props.tableOptions.reviewData) {
      dataList.value = props.tableOptions.reviewData;
      return;
    }
    isLoading.value = true;
    const [detailsRes, countRes] = await Promise.all(
      [false, true].map((isCount) =>
        resourceStore.list(
          {
            page: {
              limit: isCount ? 0 : pagination.limit,
              start: isCount ? 0 : pagination.start,
              sort: isCount ? undefined : sort.value,
              order: isCount ? undefined : order.value,
              count: isCount,
            },
            filter: {
              op: filter.op,
              rules: [...filter.rules, ...customRules],
            },
            ...props.requestOption.extension,
          },
          type ? type : props.requestOption.type,
        ),
      ),
    );
    dataList.value = props.requestOption.dataPath
      ? lodash_get(detailsRes, props.requestOption.dataPath)
      : detailsRes?.data?.details;
    // 如果需要, 可以根据当前结果异步更新 dataList
    if (props.requestOption.callback && typeof props.requestOption.callback === 'function') {
      props.requestOption.callback(detailsRes?.data?.details).then((newDataList: any[]) => {
        dataList.value = newDataList;
      });
    }
    pagination.count = countRes?.data?.count;
    isLoading.value = false;
  };
  const CommonTable = defineComponent({
    setup(_props, { slots }) {
      return () => (
        <div class={`remote-table-container${props.searchOptions.disabled ? ' no-search' : ''}`}>
          <section class='operation-wrap'>
            {slots.operation && <div class='operate-btn-groups'>{slots.operation?.()}</div>}
            {!props.searchOptions.disabled && (
              <SearchSelect
                class='table-search-selector'
                style={props.searchOptions?.extra?.searchSelectExtStyle}
                v-model={searchVal.value}
                data={props.searchOptions.searchData}
                valueBehavior='need-key'
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
              onPageValueChange={handlePageValueChange}
              onColumnSort={handleSort}
              onColumnFilter={() => {}}>
              {{
                empty: () => {
                  if (isLoading.value) return null;
                  return <Empty />;
                },
              }}
            </Table>
          </Loading>
        </div>
      );
    },
  });

  /**
   * 构建请求筛选条件
   * @param options 配置对象
   */
  const buildFilter = (options: {
    rules: Array<RulesItem>; // 规则列表
    deleteOption?: { field: string; flagValue: any }; // 删除选项(可选, 用于 tab 切换时, 删除规则)
    differenceFields?: string[]; // search-select 移除条件时的搜索字段差集(只用于 search-select 组件)
  }) => {
    const { rules, deleteOption, differenceFields } = options;
    const filterMap = new Map();
    // 先添加新的规则
    rules.forEach((rule) => {
      const tmpRule = filterMap.get(rule.field);
      if (tmpRule) {
        if (Array.isArray(tmpRule.rules)) {
          filterMap.set(rule.field, { op: QueryRuleOPEnum.OR, rules: [...tmpRule.rules, rule] });
        } else {
          filterMap.set(rule.field, { op: QueryRuleOPEnum.OR, rules: [tmpRule, rule] });
        }
      } else {
        filterMap.set(rule.field, JSON.parse(JSON.stringify(rule)));
      }
    });
    // 后添加 filter 的规则
    filter.rules.forEach((rule) => {
      if (!filterMap.get(rule.field) && !rule.rules) {
        filterMap.set(rule.field, rule);
      }
    });
    // 如果配置了 deleteOption, 则当符合条件时, 删除对应规则
    if (deleteOption) {
      const { field, flagValue } = deleteOption;
      const rule = filterMap.get(field);
      rule && rule.value === flagValue && filterMap.delete(field);
    }
    // 如果配置了 differenceFields, 则移除 differenceFields 中对应的规则
    if (differenceFields) {
      differenceFields.forEach((field) => {
        if (filterMap.has(field)) {
          filterMap.delete(field);
        }
      });
    }
    // 整合后的规则重新赋值给 filter.rules
    filter.rules = [...filterMap.values()];
  };

  watch(
    [() => searchVal.value, () => accountStore.bizs],
    ([searchVal, bizs], [oldSearchVal]) => {
      if (isBusinessPage && !bizs) return;
      // 记录上一次 search-select 的规则名
      const oldSearchFieldList: string[] =
        (Array.isArray(oldSearchVal) && oldSearchVal.reduce((prev: any, item: any) => [...prev, item.id], [])) || [];
      // 记录此次 search-select 规则名
      const searchFieldList: string[] = [];
      // 构建当前 search-select 规则
      const searchRules = Array.isArray(searchVal)
        ? searchVal.map((val: any) => {
            const field = val?.id;
            const op =
              // eslint-disable-next-line no-nested-ternary
              val?.id === 'domain'
                ? QueryRuleOPEnum.JSON_CONTAINS
                : val?.id === 'name'
                ? QueryRuleOPEnum.CIS
                : props?.fuzzySwitch
                ? QueryRuleOPEnum.CIS
                : QueryRuleOPEnum.EQ;
            const value =
              field === 'bk_biz_id'
                ? businessMapStore.businessNameToIDMap.get(val?.values?.[0]?.id) || Number(val?.values?.[0]?.id)
                : val?.values?.[0]?.id;
            searchFieldList.push(field);
            return { field, op, value };
          })
        : [];
      // 如果 search-select 的条件减少, 则移除差集中的规则
      if (oldSearchFieldList.length > searchFieldList.length) {
        buildFilter({ rules: searchRules, differenceFields: getDifferenceSet(oldSearchFieldList, searchFieldList) });
      } else {
        buildFilter({ rules: searchRules });
      }
      // 页码重置
      pagination.start = 0;
      // 如果有初始筛选条件, 则加入初始筛选条件
      const { rules, deleteOption } = props.requestOption.filterOption || {};
      getListData(deleteOption ? [] : rules);
    },
    {
      immediate: true,
    },
  );

  // 分配业务筛选
  watch(
    () => props.bizFilter,
    (val) => {
      const idx = filter.rules.findIndex((rule) => rule.field === 'bk_biz_id');
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
    },
    { deep: true },
  );

  watch(
    () => props.requestOption.filterOption,
    (val) => {
      if (!val) return;
      const { rules, deleteOption } = val;
      buildFilter({ rules, deleteOption });
      getListData();
    },
    {
      deep: true,
    },
  );

  return {
    CommonTable,
    getListData,
  };
};
