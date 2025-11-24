<script setup lang="ts">
import http from '@/http';
import { ref, watch } from 'vue';
import type { OrderStatisticsListData, IOrderStatistics } from './typings';
import { columns } from './column';

interface IProps {
  yearMonth: string;
}
const props = defineProps<IProps>();
const emit = defineEmits<{
  edit: [config: IOrderStatistics[], month: string];
}>();

const listData = ref<IOrderStatistics[]>([]);

const fetchOrderStatisticsData = async () => {
  const res = await http.post<OrderStatisticsListData>('/api/v1/woa/task/config/findmany/apply/order/statistics', {
    stat_month: props.yearMonth,
  });
  listData.value = res.data.details;
};
watch(
  () => props.yearMonth,
  () => {
    fetchOrderStatisticsData();
  },
  {
    immediate: true,
  },
);
</script>

<template>
  <div class="config-block">
    <div class="header">
      <span class="title">{{ yearMonth }}</span>
      <div class="btns">
        <bk-button text theme="primary" @click="emit('edit', listData, yearMonth)">编辑</bk-button>
      </div>
    </div>
    <data-list-table :list="listData" :columns="columns" :settings="false" stripe>
      <template #start_at="{ row }">
        {{ row?.start_at ? row?.start_at + ' ~ ' + row?.end_at : '--' }}
      </template>
    </data-list-table>
  </div>
</template>

<style scoped lang="scss">
.config-block {
  margin-top: 16px;
  width: 100%;
  border: 1px solid #dcdee5;
  padding-bottom: 10px;

  .header {
    height: 36px;
    padding: 6px 12px;
    background-color: #f0f1f5;
    display: flex;
    justify-content: space-between;
    align-items: center;
    font-size: 12px;

    .title {
      color: #4d4f56;
      font-weight: 700;
    }
  }
}
</style>
