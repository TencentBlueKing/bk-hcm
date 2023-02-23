<script lang="ts" setup>
import { h, ref, watchEffect } from 'vue';

const props = defineProps({
  detail: {
    type: Object,
  },
});
const columns = ref([
  {
    label: '接口名称',
    field: 'name',
  },
  {
    label: '内网IP',
    field: 'internal_ip',
  },
  {
    label: '公网IP',
    field: 'public_ip',
  },
  {
    label: '所属网络(VPC)',
    field: 'cloud_vpc_id',
    showOverflowTooltip: true,
    render({ cell }: { cell: string }) {
      if (!cell) {
        return '--';
      }
      return h('div', { class: 'cell-content-list' }, cell?.split(';')
        .map(item => h('p', { class: 'cell-content-item' }, item)));
    },
  },
  {
    label: '所属子网',
    field: 'cloud_subnet_id',
    showOverflowTooltip: true,
    render({ cell }: { cell: string }) {
      if (!cell) {
        return '--';
      }
      return h('div', { class: 'cell-content-list' }, cell?.split(';')
        .map(item => h('p', { class: 'cell-content-item' }, item)));
    },
  },
  {
    label: '网络层级',
    field: 'networkTier',
    render({ cell }: { cell: number }) {
      const vals = { PREMIUM: '高级', STANDARD: '标准' };
      return vals[cell];
    },
  },
  {
    label: 'IP转发',
    field: 'can_ip_forward',
    render({ cell }: { cell: number }) {
      return cell ? '开启' : '关闭';
    },
  },
]);

const data = ref([]);

watchEffect(() => {
  data.value = [{
    ...props.detail,
    networkTier: props.detail?.access_configs?.[0]?.network_tier,
  }];

  console.log(data.value);
});
</script>

<template>
  <bk-table
    class="table-list mt20"
    row-hover="auto"
    :columns="columns"
    :data="data"
  />
</template>

<style lang="scss" scoped>
.table-list {
  :deep(.cell-content-list) {
    .cell-content-item {
      overflow: hidden;
      text-overflow: ellipsis;
    }
  }
}
</style>
