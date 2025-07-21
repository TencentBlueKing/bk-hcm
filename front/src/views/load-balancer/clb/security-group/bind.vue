<script setup lang="ts">
import { watch } from 'vue';
import { useTable } from '@/hooks/useTable/useTable';
import { ISearchItem } from 'bkui-vue/lib/search-select/utils';
import { QueryRuleOPEnum } from '@/typings';
import { ILoadBalancerDetails } from '@/store/load-balancer/clb';

const props = defineProps<{
  details: ILoadBalancerDetails;
  show: boolean;
  bindedSecurityGroups: any;
  selectedSecuirtyGroupsSet: any;
  handleSelectionChange: any;
  resetSelections: any;
}>();

const tableColumns = [
  { type: 'selection', width: 30, minWidth: 30 },
  {
    label: '安全组名称',
    field: 'name',
  },
  {
    label: 'ID',
    field: 'cloud_id',
  },
  {
    label: '备注',
    field: 'memo',
  },
];
const searchData: ISearchItem[] = [
  {
    id: 'name',
    name: '安全组名称',
  },
  {
    id: 'cloud_id',
    name: 'ID',
  },
];

const isRowSelectEnable = ({ row, isCheckAll }: any) => {
  if (isCheckAll) return true;
  return isCurRowSelectEnable(row);
};

const isCurRowSelectEnable = (row: any) => {
  return !props.bindedSecurityGroups.map((v) => v.id).includes(row.id) && !props.selectedSecuirtyGroupsSet.has(row.id);
};

const { CommonTable, getListData } = useTable({
  searchOptions: {
    searchData,
    extra: {
      searchSelectExtStyle: {
        width: '100%',
      },
    },
  },
  tableOptions: {
    columns: tableColumns,
    extra: {
      isRowSelectEnable,
      onSelectionChange: (selections: any) => props.handleSelectionChange(selections, isCurRowSelectEnable),
      onSelectAll: (selections: any) => props.handleSelectionChange(selections, isCurRowSelectEnable, true),
      selectionKey: 'cloud_id',
    },
  },
  requestOption: {
    type: 'security_groups',
    filterOption: {
      rules: [
        {
          field: 'vendor',
          op: QueryRuleOPEnum.EQ,
          value: props.details.vendor,
        },
        {
          field: 'region',
          op: QueryRuleOPEnum.EQ,
          value: props.details.region,
        },
      ],
      // 属性里传入一个配置，选择是不是要模糊查询
      fuzzySwitch: true,
    },
  },
});

watch(
  () => props.show,
  (val) => {
    if (!val) return;
    getListData();
    props.resetSelections();
  },
);
</script>
<template>
  <CommonTable />
</template>
