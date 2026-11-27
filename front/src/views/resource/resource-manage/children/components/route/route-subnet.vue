<script lang="ts" setup>
import { ref, watch } from 'vue';

import useQueryList from '../../../hooks/use-query-list';
import { useResourceDetailRoute } from '../../../hooks/use-resource-detail-route';
import { QueryRuleOPEnum } from '@/typings';

const props = defineProps({
  detail: {
    type: Object,
  },
});

const { detailId } = useResourceDetailRoute();

const columns = ref<any[]>([]);

watch(
  () => props.detail,
  () => {
    switch (props.detail.vendor) {
      case 'tcloud':
        columns.value.push(
          ...[
            {
              label: '子网ID',
              field: 'id',
            },
            {
              label: '名称',
              field: 'name',
            },
            {
              label: '可用区',
              field: 'cloud_gateway_id',
              render({ cell }: any) {
                return cell || '--';
              },
            },
            {
              label: 'CIDR',
              field: 'ipv4_cidr',
            },
          ],
        );
        break;
      case 'azure':
        columns.value.push(
          ...[
            {
              label: '名称',
              field: 'name',
            },
            {
              label: '地址范围',
              field: '',
              render({ cell }: { cell: string }) {
                return cell || '--';
              },
            },
            {
              label: 'CIDR',
              field: 'ipv4_cidr',
            },
            {
              label: '安全组',
              field: '',
              render({ cell }: { cell: string }) {
                return cell || '--';
              },
            },
          ],
        );
        break;
      case 'aws':
        columns.value.push(
          ...[
            {
              label: '子网ID',
              field: 'id',
            },
            {
              label: 'CIDR',
              field: 'ipv4_cidr',
            },
          ],
        );
        break;
      case 'huawei':
        columns.value.push(
          ...[
            {
              label: '子网ID',
              field: 'id',
            },
            {
              label: '名称',
              field: 'name',
            },
            {
              label: 'IPv4 CIDR',
              field: 'ipv4_cidr',
            },
            {
              label: 'IPv6 CIDR',
              field: 'ipv6_cidr',
            },
          ],
        );
        break;
    }
    // bk-table 对 columns prop 的 watch 仅监听引用变化，原地 push 不触发更新，需整体重新赋值
    columns.value = [...columns.value];
  },
  {
    deep: true,
    immediate: true,
  },
);

const { datas, pagination, isLoading, handlePageChange, handlePageSizeChange } = useQueryList(
  {
    filter: {
      op: 'and',
      rules: [
        {
          field: 'route_table_id',
          op: QueryRuleOPEnum.EQ,
          value: detailId.value,
        },
      ],
    },
  },
  'subnets',
);
</script>

<template>
  <bk-loading :loading="isLoading">
    <bk-table
      row-hover="auto"
      remote-pagination
      :pagination="pagination"
      :columns="columns"
      :data="datas"
      show-overflow-tooltip
      @page-limit-change="handlePageSizeChange"
      @page-value-change="handlePageChange"
    />
  </bk-loading>
</template>
