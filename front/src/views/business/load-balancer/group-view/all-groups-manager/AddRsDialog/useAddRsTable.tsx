import { Ref, reactive, ref, watch } from 'vue';
import useSelection from '@/views/resource/resource-manage/hooks/use-selection';

export default (rsSelections: Ref<any[]>) => {
  // 搜索相关
  const searchData = [
    {
      name: '内网IP',
      id: 'privateIp',
    },
    {
      name: '公网IP',
      id: 'publicIp',
    },
    {
      name: '名称',
      id: 'name',
    },
    {
      name: '资源类型',
      id: 'resourceType',
    },
  ];
  const searchValue = ref('');

  // 表格相关
  const tableRef = ref(null);
  const isTableLoading = ref(false);
  const columns = [
    { type: 'selection', width: '100' },
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
  const pagination = reactive({ small: true, align: 'left', limit: 10, limitList: [10, 20, 50, 100] });

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
    handleSelectionChange(selection, () => true, true);
    if (selection.checked) {
      selectedCount.value = selection.data.length > pagination.limit ? pagination.limit : selection.data.length;
    } else {
      selectedCount.value = 0;
    }
  };
  const handleClear = () => {
    tableRef.value.clearSelection();
    resetSelections();
    selectedCount.value = 0;
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

  return {
    searchData,
    searchValue,
    isTableLoading,
    tableRef,
    columns,
    pagination,
    selectedCount,
    handleSelect,
    handleSelectAll,
    handleClear,
  };
};
