<script lang="ts" setup>
import {
  ref,
  watch,
} from 'vue';

import useQueryList from '../../../hooks/use-query-list';

const props = defineProps({
  detail: {
    type: Object,
  },
});

const columns = ref<any[]>([]);

watch(
  () => props.detail,
  () => {
    switch (props.detail.vendor) {
      case 'tcloud':
        columns.value.push(...[
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
          },
          {
            label: 'CIDR',
            field: 'ipv4_cidr',
          },
        ]);
        break;
      case 'azure':
        columns.value.push(...[
          {
            label: '名称',
            field: 'name',
          },
          {
            label: '地址范围',
            field: '',
          },
          {
            label: 'CIDR',
            field: 'ipv4_cidr',
          },
          {
            label: '安全组',
            field: '',
          },
        ]);
        break;
      case 'aws':
        columns.value.push(...[
          {
            label: '子网ID',
            field: 'id',
          },
          {
            label: 'CIDR',
            field: 'ipv4_cidr',
          },
        ]);
        break;
      case 'huawei':
        columns.value.push(...[
          {
            label: '子网ID',
            field: 'id',
          },
          {
            label: '名称',
            field: 'name',
          },
          {
            label: 'IPv4网段',
            field: 'ipv4_cidr',
          },
          {
            label: 'ipv6_cidr',
            field: 'ipv4_cidr',
          },
          {
            label: '状态',
            field: '',
          },
        ]);
        break;
    }
  },
  {
    deep: true,
    immediate: true
  }
)

const {
  datas,
  pagination,
  isLoading,
  handlePageChange,
  handlePageSizeChange,
} = useQueryList(
  {
    filter: {
      op: 'and',
      rules: [{
        field: 'route_table_id',
        op: 'eq',
        value: props.detail.id
      }]
    }
  },
  'subnets'
);
</script>

<template>
  <bk-loading
    :loading="isLoading"
  >
    <bk-table
      class="mt20"
      row-hover="auto"
      :pagination="pagination"
      :columns="columns"
      :data="datas"
      @page-limit-change="handlePageSizeChange"
      @page-value-change="handlePageChange"
    />
  </bk-loading>
</template>
