<script lang="ts" setup>
import {
  watch,
  ref,
} from 'vue';

const props = defineProps({
  detail: {
    type: Object,
  },
});

const columns = ref([]);

watch(
  () => props.detail,
  () => {
    switch (props.detail.vendor) {
      case 'tcloud':
        columns.value = [
          {
            label: '地址类型',
            field: 'type',
          },
          {
            label: 'CIDR',
            field: 'cidr',
          },
          {
            label: '类别',
            field: 'category',
          },
        ];
        break;
      case 'aws':
        columns.value = [
          {
            label: '地址类型',
            field: 'type',
          },
          {
            label: 'CIDR',
            field: 'cidr',
          },
          {
            label: '地址池',
            field: 'address_pool',
          },
          {
            label: '状态',
            field: 'state',
          },
        ];
        break;
      case 'azure':
      case 'huawei':
        columns.value = [
          {
            label: '地址类型',
            field: 'type',
          },
          {
            label: 'CIDR',
            field: 'cidr',
          },
        ];
        break;
    }
  },
  {
    deep: true,
    immediate: true,
  },
);
</script>

<template>
  <bk-table
    row-hover="auto"
    :columns="columns"
    :data="props.detail.cidr"
    show-overflow-tooltip
  />
</template>
