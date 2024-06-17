/* eslint-disable no-nested-ternary */
import { QueryRuleOPEnum, RulesItem } from '@/typings/common';
import { FilterType } from '@/typings';
import { Loading, SearchSelect, Table } from 'bkui-vue';
import type { Column } from 'bkui-vue/lib/table/props';
import { ISearchItem } from 'bkui-vue/lib/search-select/utils';
import { computed, defineComponent, reactive, ref, watch } from 'vue';
import './index.scss';
import Empty from '@/components/empty';
import { useResourceStore, useBusinessStore } from '@/store';
import { useBusinessMapStore } from '@/store/useBusinessMap';
import { useRegionsStore } from '@/store/useRegionsStore';
import { useWhereAmI, Senarios } from '../useWhereAmI';
import { getDifferenceSet } from '@/common/util';
import { get as lodash_get } from 'lodash-es';
import { VendorReverseMap } from '@/common/constant';
import { LB_NETWORK_TYPE_REVERSE_MAP, LISTENER_BINDING_STATUS_REVERSE_MAP, SCHEDULER_REVERSE_MAP } from '@/constants';
import usePagination from '../usePagination';
import { defaults } from 'lodash';
import { fetchData } from '@pluginHandler/useTable';

export interface IProp {
  // search-select 配置项
  searchOptions: {
    // search-select 可选项
    searchData?: Array<ISearchItem> | (() => Array<ISearchItem>);
    // 是否禁用 search-select
    disabled?: boolean;
    // 其他 search-select 属性/自定义事件, 比如 placeholder, onSearch, searchSelectExtStyle...
    extra?: {
      searchSelectExtStyle?: Record<string, string>; // 搜索框样式
    };
  };
  // table 配置项
  tableOptions: {
    // 表格字段
    columns: Array<Column> | (() => Array<Column>);
    // 用于预览效果的数据
    reviewData?: Array<Record<string, any>>;
    // 其他 table 属性/自定义事件, 比如 settings, onSelectionChange...
    extra?: Object;
  };
  // 请求相关字段
  requestOption: {
    // 资源类型
    type?: string;
    // 排序参数
    sortOption?: {
      sort: string; // 需要排序的字段
      order: 'ASC' | 'DESC'; // 排序方式
    };
    // 筛选参数
    filterOption?: {
      // 规则
      rules: Array<RulesItem>;
      // Tab 切换时选用项(如选中全部时, 删除对应的 rule)
      deleteOption?: {
        field: string;
        flagValue: string; // 当 rule.value = flagValue 时, 删除该 rule
      };
      // 模糊查询开关true开启，false关闭
      fuzzySwitch?: boolean;
    };
    // 请求需要的额外荷载数据
    extension?: Record<string, any>;
    // 钩子 - 可以根据当前请求结果异步更新 dataList
    resolveDataListCb?: (...args: any) => Promise<any>;
    // 钩子 - 可以根据当前请求结果异步更新 pagination.count
    resolvePaginationCountCb?: (...args: any) => Promise<any>;
    // 列表数据的路径，如 data.details
    dataPath?: string;
    // 是否为全量数据
    full?: boolean;
  };
  // 资源下筛选业务功能相关的 prop
  bizFilter?: FilterType;
}

export const useTable = (props: IProp) => {
  defaults(props, { requestOption: {} });
  defaults(props.requestOption, { dataPath: 'data.details' });

  const { whereAmI } = useWhereAmI();

  const regionsStore = useRegionsStore();
  const resourceStore = useResourceStore();
  const businessStore = useBusinessStore();
  const businessMapStore = useBusinessMapStore();

  const searchVal = ref('');
  const dataList = ref([]);
  const isLoading = ref(false);
  const sort = ref(props.requestOption.sortOption ? props.requestOption.sortOption.sort : 'created_at');
  const order = ref(props.requestOption.sortOption ? props.requestOption.sortOption.order : 'DESC');
  const getInitialRules = () => {
    const { filterOption } = props.requestOption;
    return filterOption && !filterOption.deleteOption ? filterOption.rules : [];
  };
  const filter = reactive({ op: QueryRuleOPEnum.AND, rules: getInitialRules() });

  const { pagination, handlePageLimitChange, handlePageValueChange } = usePagination(() => getListData());

  // 钩子 - 表头排序时
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

  /**
   * 请求表格数据
   * @param customRules 自定义规则
   * @param type 资源类型
   */
  const getListData = async (customRules: Array<RulesItem> = [], type?: string) => {
    buildFilter({ rules: customRules });
    // 预览
    if (props.tableOptions.reviewData) {
      dataList.value = props.tableOptions.reviewData;
      return;
    }
    isLoading.value = true;

    try {
      // 判断是业务下, 还是资源下
      const api = whereAmI.value === Senarios.business ? businessStore.list : resourceStore.list;
      const [detailsRes, countRes] = await fetchData({ api, pagination, sort, order, filter, props, type });

      // 更新数据
      dataList.value = lodash_get(detailsRes, props.requestOption.dataPath, []) || [];

      // 异步处理 dataList
      if (typeof props.requestOption.resolveDataListCb === 'function') {
        props.requestOption.resolveDataListCb(dataList.value, getListData).then((newDataList: any[]) => {
          dataList.value = newDataList;
        });
      }

      // 处理 pagination.count
      if (typeof props.requestOption.resolvePaginationCountCb === 'function') {
        props.requestOption.resolvePaginationCountCb(countRes?.data).then((newCount: number) => {
          pagination.count = newCount;
        });
      } else {
        pagination.count = countRes?.data?.count || 0;
      }
    } catch (error) {
      dataList.value = [];
      pagination.count = 0;
    } finally {
      isLoading.value = false;
    }
  };

  const CommonTable = defineComponent({
    setup(_props, { slots, expose }) {
      const searchData = computed(() => {
        return (
          (typeof props.searchOptions.searchData === 'function'
            ? props.searchOptions.searchData()
            : props.searchOptions.searchData) || []
        );
      });

      const tableRef = ref();

      expose({ tableRef });

      return () => (
        <div class={`remote-table-container${props.searchOptions.disabled ? ' no-search' : ''}`}>
          <section class='operation-wrap'>
            {slots.operation && <div class='operate-btn-groups'>{slots.operation?.()}</div>}
            {!props.searchOptions.disabled && (
              <SearchSelect
                class='table-search-selector'
                style={props.searchOptions?.extra?.searchSelectExtStyle}
                v-model={searchVal.value}
                data={searchData.value}
                valueBehavior='need-key'
                {...(props.searchOptions.extra || {})}
              />
            )}
          </section>
          <Loading loading={isLoading.value} class='loading-table-container'>
            <Table
              ref={tableRef}
              class='table-container'
              data={dataList.value}
              rowKey='id'
              columns={props.tableOptions.columns}
              pagination={pagination}
              remotePagination={!props.requestOption.full}
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
   * 处理搜索条件, 有需要映射的字段需要转换
   * @param rule 待添加的搜索条件
   */
  const resolveRule = (rule: RulesItem) => {
    const { field, op, value } = rule;
    switch (field) {
      case 'vendor':
        return { field, op, value: VendorReverseMap[value as string] || value };
      case 'region':
        return { field, op, value: regionsStore.getRegionNameEN(value as string) || value };
      case 'lb_type':
        return { field, op, value: LB_NETWORK_TYPE_REVERSE_MAP[value as string] || value };
      case 'scheduler':
        return { field, op, value: SCHEDULER_REVERSE_MAP[value as string] || value };
      case 'binding_status':
        return { field, op, value: LISTENER_BINDING_STATUS_REVERSE_MAP[value as string] || value };
      default:
        return { field, op, value };
    }
  };

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
      const newRule = resolveRule(rule);
      const tmpRule = filterMap.get(newRule.field);
      if (tmpRule) {
        if (Array.isArray(tmpRule.rules)) {
          filterMap.set(newRule.field, { op: QueryRuleOPEnum.OR, rules: [...tmpRule.rules, newRule] });
        } else {
          filterMap.set(newRule.field, { op: QueryRuleOPEnum.OR, rules: [tmpRule, newRule] });
        }
      } else {
        filterMap.set(newRule.field, JSON.parse(JSON.stringify(newRule)));
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

  /**
   * 处理字段的搜索模式
   */
  const resolveSearchFieldOp = (val: any) => {
    let op;
    const { id, name } = val;
    if (!id || !name) return;
    // 如果是domain或者zones(数组类型), 则使用JSON_CONTAINS
    if ((val?.id === 'domain' && val?.name !== '负载均衡域名') || val?.id === 'zones') {
      op = QueryRuleOPEnum.JSON_CONTAINS;
    }
    // 如果是名称或指定了模糊搜索, 则模糊搜索
    else if (
      props?.requestOption?.filterOption?.fuzzySwitch ||
      val?.id === 'name' ||
      (val?.id === 'domain' && val?.name === '负载均衡域名')
    ) {
      op = QueryRuleOPEnum.CIS;
    }
    // 如果是任务类型, 则使用 json_neq
    else if (val?.id === 'detail.data.res_flow.flow_id') {
      op = QueryRuleOPEnum.JSON_NEQ;
    } else if (val?.id === 'health_check.health_switch') {
      op = QueryRuleOPEnum.JSON_EQ;
    }
    // 否则, 精确搜索
    else {
      op = QueryRuleOPEnum.EQ;
    }
    return op;
  };

  watch(
    () => searchVal.value,
    (searchVal, oldSearchVal) => {
      // 记录上一次 search-select 的规则名
      const oldSearchFieldList: string[] =
        (Array.isArray(oldSearchVal) && oldSearchVal.reduce((prev: any, item: any) => [...prev, item.id], [])) || [];
      // 记录此次 search-select 规则名
      const searchFieldList: string[] = [];
      // 构建当前 search-select 规则
      const searchRules = Array.isArray(searchVal)
        ? searchVal.map((val: any) => {
            const field = val?.id;
            const op = resolveSearchFieldOp(val);
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
      getListData();
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
    dataList,
    getListData,
    pagination,
  };
};
