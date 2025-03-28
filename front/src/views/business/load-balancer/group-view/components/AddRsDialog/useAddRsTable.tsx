import { Ref, reactive, ref, watch } from 'vue';
// import hooks
import useSelection from '@/views/resource/resource-manage/hooks/use-selection';
// import stores
import { useBusinessStore } from '@/store';
// import types
import { QueryRuleOPEnum, RulesItem } from '@/typings';
import { getDifferenceSet } from '@/common/util';
import usePagination from '@/hooks/usePagination';

export default (
  rsSelections: Ref<any[]>,
  getTableRsList: () => any[],
  callback: () => { vpc_ids: string[]; account_id: string; isCorsVersion2: boolean },
) => {
  // use stores
  const businessStore = useBusinessStore();

  // 搜索相关
  const searchData = [
    {
      name: '内网IP',
      id: 'private_ipv4_addresses',
    },
    {
      name: '公网IP',
      id: 'public_ipv4_addresses',
    },
    {
      name: '名称',
      id: 'name',
    },
    {
      name: '资源类型',
      id: 'machine_type',
    },
  ];
  const searchVal = ref('');

  // 表格相关
  const tableRef = ref(null);
  const isTableLoading = ref(false);
  const columns = [
    { type: 'selection', width: 30, minWidth: 30 },
    {
      label: '内网IP',
      render({ data }: any) {
        return [...data.private_ipv4_addresses, ...data.private_ipv6_addresses].join(',');
      },
    },
    {
      label: '公网IP',
      render({ data }: any) {
        return [...data.public_ipv4_addresses, ...data.public_ipv6_addresses].join(',');
      },
    },
    {
      label: '名称',
      field: 'name',
    },
    {
      label: '资源类型',
      field: 'machine_type',
      filter: true,
    },
  ];
  const rsTableList = ref([]);
  const pagination = reactive({
    small: true,
    align: 'left',
    start: 0,
    limit: 10,
    count: 0,
    limitList: [10, 20, 50, 100],
  });
  const { handlePageLimitChange, handlePageValueChange } = usePagination(() => {
    const { account_id, vpc_ids, isCorsVersion2 } = callback();
    getRSTableList(account_id, vpc_ids, isCorsVersion2);
  }, pagination);

  const selectedCount = ref(0);
  const { selections, handleSelectionChange, resetSelections } = useSelection();

  const handleSelect = (selection: any) => {
    handleSelectionChange(selection, () => true, false);
    if (selection.checked) {
      selectedCount.value += 1;
    } else {
      selectedCount.value -= 1;
    }
  };
  const handleSelectAll = (selection: any) => {
    handleSelectionChange(
      selection,
      (row) => !getTableRsList().some((rs) => rs.id === row.id || rs.inst_id === row.id),
      true,
    );
    if (selection.checked) {
      selectedCount.value = selections.value.length;
    } else {
      selectedCount.value = 0;
    }
  };
  const handleClear = () => {
    tableRef.value.clearSelection();
    resetSelections();
    selectedCount.value = 0;
  };

  const filter = reactive({
    op: QueryRuleOPEnum.AND,
    rules: [],
  });

  // 获取 rs 列表
  const getRSTableList = async (accountId: string, vpcIds: string[], isCorsV2: boolean) => {
    if (!accountId) {
      rsTableList.value = [];
      return;
    }
    try {
      isTableLoading.value = true;
      const queryFilter = {
        op: QueryRuleOPEnum.AND,
        rules: [
          {
            field: 'account_id',
            op: QueryRuleOPEnum.EQ,
            value: accountId,
          },
          ...filter.rules,
        ],
      };
      // 如果没有开启跨域2.0, 则需要使用vpc过滤rs列表
      if (!isCorsV2) {
        queryFilter.rules.push({
          field: 'vpc_ids',
          op: QueryRuleOPEnum.JSON_OVERLAPS,
          value: vpcIds,
        });
      }
      const [detailRes, countRes] = await Promise.all(
        [false, true].map((isCount) =>
          businessStore.getAllRsList({
            filter: queryFilter,
            page: {
              count: isCount,
              start: isCount ? 0 : pagination.start,
              limit: isCount ? 0 : pagination.limit,
            },
          }),
        ),
      );
      rsTableList.value = detailRes.data.details;
      pagination.count = countRes.data.count;
    } finally {
      isTableLoading.value = false;
    }
  };

  watch(
    selections,
    (val) => {
      rsSelections.value = val;
    },
    {
      deep: true,
    },
  );

  /**
   * 构建请求筛选条件
   * @param options 配置对象
   */
  const buildFilter = (options: {
    rules: Array<RulesItem>; // 规则列表
    differenceFields?: string[]; // search-select 移除条件时的搜索字段差集(只用于 search-select 组件)
  }) => {
    const { rules, differenceFields } = options;
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
    const { id } = val;
    if (!id) return;
    // 如果是domain或者zones(数组类型), 则使用JSON_CONTAINS
    if (val?.id === 'private_ipv4_addresses' || val?.id === 'public_ipv4_addresses') op = QueryRuleOPEnum.JSON_CONTAINS;
    // 如果是名称或指定了模糊搜索, 则模糊搜索
    else if (val?.id === 'name') op = QueryRuleOPEnum.CIS;
    // 否则, 精确搜索
    else op = QueryRuleOPEnum.EQ;
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
            const value = val?.values?.[0]?.id;
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
      const { account_id, vpc_ids, isCorsVersion2 } = callback();
      getRSTableList(account_id, vpc_ids, isCorsVersion2);
    },
    {
      immediate: true,
    },
  );

  return {
    searchData,
    searchVal,
    isTableLoading,
    tableRef,
    columns,
    rsTableList,
    pagination,
    selectedCount,
    handleSelect,
    handleSelectAll,
    handleClear,
    getRSTableList,
    handlePageLimitChange,
    handlePageValueChange,
  };
};
