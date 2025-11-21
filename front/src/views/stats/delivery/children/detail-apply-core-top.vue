<script setup lang="ts">
import { computed, reactive, ref, Ref, watch } from 'vue';
import http from '@/http';
import { PrimaryTable, type TableProps } from '@blueking/tdesign-ui';
import BusinessValue from '@/components/display-value/business-value.vue';
import TicketLinkButton from './ticket-link-button.vue';
import type { IDetailComponentProps } from '../../typings';
import { getMonthRange, mergeTableData, calculateGrowthRate } from '../../utils';

const props = defineProps<IDetailComponentProps>();

const detailData = ref<any[]>([]);

const currentMonth = computed(() => getMonthRange(props.currentDate, 'YYYY-MM').startTime);
const compareMonth = computed(() => getMonthRange(props.compareDate, 'YYYY-MM').startTime);

const loading = ref(false);

const fetchDetailData = async () => {
  loading.value = true;
  const [currentRes, compareRes] = await Promise.all([
    http.post('/api/v1/woa/task/apply/findmany/bizs_cpucores/statistics', {
      start_time: getMonthRange(props.currentDate).startTime,
      end_time: getMonthRange(props.currentDate).endTime,
    }),
    http.post('/api/v1/woa/task/apply/findmany/bizs_cpucores/statistics', {
      start_time: getMonthRange(props.compareDate).startTime,
      end_time: getMonthRange(props.compareDate).endTime,
    }),
  ]);

  const currentData = currentRes.data?.details || [];
  const compareData = compareRes.data?.details || [];
  detailData.value = mergeTableData(currentData, compareData, ['delivered_core_count', 'order_count']);
  pagination.total = detailData.value.length;
  loading.value = false;
};

const columns: Ref<TableProps['columns']> = computed(() => [
  {
    title: '业务名称',
    colKey: 'bk_biz_id',
    cell: (h, { row }) => {
      return h(
        TicketLinkButton,
        {
          bizId: row.bk_biz_id,
          filter: {
            create_at: [],
            bk_username: [],
          },
        },
        h(BusinessValue, {
          value: row.bk_biz_id,
        }),
      );
    },
  },
  {
    title: '核心数',
    colKey: 'delivered_core_count',
    align: 'center',
    children: [
      {
        title: currentMonth.value,
        colKey: 'current_delivered_core_count',
        align: 'right',
        sortType: 'all',
        sorter: (a, b) => a.current_delivered_core_count - b.current_delivered_core_count,
        cell: (h, { row }) => row?.current_delivered_core_count ?? '--',
      },
      {
        title: compareMonth.value,
        colKey: 'compare_delivered_core_count',
        align: 'right',
        sortType: 'all',
        sorter: (a, b) => a.compare_delivered_core_count - b.compare_delivered_core_count,
        cell: (h, { row }) => row?.compare_delivered_core_count ?? '--',
      },
      {
        title: '环比',
        colKey: 'growth_rate',
        align: 'left',
        sortType: 'all',
        sorter: (a, b) => {
          const aValue = calculateGrowthRate(a.current_delivered_core_count, a.compare_delivered_core_count).value;
          const bValue = calculateGrowthRate(b.current_delivered_core_count, b.compare_delivered_core_count).value;
          return aValue - bValue;
        },
        cell: (h, { row }) => {
          const a = calculateGrowthRate(row?.current_delivered_core_count, row?.compare_delivered_core_count);
          return h(
            'div',
            {
              class: a?.class,
            },
            a?.text,
          );
        },
      },
    ],
  },
  {
    title: '单据数量',
    colKey: 'order_count',
    align: 'center',
    children: [
      {
        title: currentMonth.value,
        colKey: 'current_order_count',
        align: 'right',
        sortType: 'all',
        sorter: (a, b) => a.current_order_count - b.current_order_count,
        cell: (h, { row }) => {
          if (row?.current_order_count === undefined) {
            return '--';
          }
          return h(
            TicketLinkButton,
            {
              bizId: row.bk_biz_id,
              filter: {
                create_at: [
                  getMonthRange(props.currentDate, 'YYYY-MM-DD').startTime,
                  getMonthRange(props.currentDate, 'YYYY-MM-DD').endTime,
                ],
                bk_username: [],
              },
            },
            row?.current_order_count,
          );
        },
      },
      {
        title: compareMonth.value,
        colKey: 'compare_order_count',
        align: 'right',
        sortType: 'all',
        sorter: (a, b) => a.compare_order_count - b.compare_order_count,
        cell: (h, { row }) => {
          if (row?.compare_order_count === undefined) {
            return '--';
          }
          return h(
            TicketLinkButton,
            {
              bizId: row.bk_biz_id,
              filter: {
                create_at: [
                  getMonthRange(props.compareDate, 'YYYY-MM-DD').startTime,
                  getMonthRange(props.compareDate, 'YYYY-MM-DD').endTime,
                ],
                bk_username: [],
              },
            },
            row?.compare_order_count,
          );
        },
      },
    ],
  },
]);

const pagination: TableProps['pagination'] = reactive({
  current: 1,
  pageSize: 20,
  total: detailData.value.length,
});

const onPageChange: TableProps['onPageChange'] = (pageInfo) => {
  pagination.current = pageInfo.current;
  pagination.pageSize = pageInfo.pageSize;
};

watch(
  [() => props.currentDate, () => props.compareDate],
  () => {
    fetchDetailData();
  },
  { immediate: true },
);
</script>

<template>
  <div class="detail-apply-table-container" v-bkloading="{ loading }">
    <primary-table
      class="detail-apply-table"
      max-height="calc(100vh - 200px)"
      :bordered="false"
      row-key="bk_biz_id"
      size="small"
      :hide-sort-tips="true"
      :stripe="true"
      :hover="true"
      :sort="sort"
      :data="detailData"
      :columns="columns"
      @sort-change="sortChange"
      @data-change="dataChange"
      @page-change="onPageChange"
      :pagination="pagination"
    ></primary-table>
  </div>
</template>
<style lang="scss" scoped></style>
