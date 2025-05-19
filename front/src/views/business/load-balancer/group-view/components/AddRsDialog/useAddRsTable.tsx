import { Ref, reactive, ref, watch } from 'vue';
// import hooks
import useSelection from '@/views/resource/resource-manage/hooks/use-selection';
// import stores
import { useBusinessStore } from '@/store';
// import types
import { QueryRuleOPEnum } from '@/typings';
import usePagination from '@/hooks/usePagination';
import useFilterHost from '@/views/resource/resource-manage/hooks/use-filter-host';

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
      id: 'private_ip',
    },
    {
      name: '公网IP',
      id: 'public_ip',
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

  // 表格相关
  const tableRef = ref(null);
  const isTableLoading = ref(false);
  const columns = [
    { type: 'selection', width: 30, minWidth: 30 },
    {
      label: '内网IP',
      render({ data }: any) {
        return [...(data.private_ipv4_addresses || []), ...(data.private_ipv6_addresses || [])].join(',') || '--';
      },
    },
    {
      label: '公网IP',
      render({ data }: any) {
        return [...(data.public_ipv4_addresses || []), ...(data.public_ipv6_addresses || [])].join(',') || '--';
      },
    },
    {
      label: '名称',
      field: 'name',
    },
    {
      label: '资源类型',
      field: 'machine_type',
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
    tableRef.value?.clearSelection();
    resetSelections();
    selectedCount.value = 0;
  };

  const filter = reactive({ op: QueryRuleOPEnum.AND, rules: [] });
  const { searchValue, filter: searchFilter } = useFilterHost({ filter });

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
        // 添加useFilterHost构建的条件
        rules: [{ field: 'account_id', op: QueryRuleOPEnum.EQ, value: accountId }, ...searchFilter.value.rules],
      };
      // 如果没有开启跨域2.0, 则需要使用vpc过滤rs列表
      if (!isCorsV2 && vpcIds.length) {
        queryFilter.rules.push({ field: 'vpc_ids', op: QueryRuleOPEnum.JSON_OVERLAPS, value: vpcIds });
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

  watch(
    searchFilter,
    () => {
      // 页码重置
      pagination.start = 0;
      const { account_id, vpc_ids, isCorsVersion2 } = callback();
      getRSTableList(account_id, vpc_ids, isCorsVersion2);
    },
    { deep: true },
  );

  watch(rsTableList, () => {
    // 当表格数据变更时，重置勾选项
    handleClear();
  });

  return {
    searchData,
    searchValue,
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
