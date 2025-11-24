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
  const res = await http.post('/api/v1/woa/task/apply/analysis/percentile_time_consumption/compare', {
    current_date: getMonthRange(props.currentDate, 'YYYY-MM').startTime,
    compare_date: getMonthRange(props.compareDate, 'YYYY-MM').endTime,
  });
  const currentData = res.data?.current || [];
  const compareData = res.data?.compare || [];
  detailData.value = mergeTableData(currentData, compareData, ['p95_hours', 'done_orders']);
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
    title: 'P95耗时',
    colKey: 'p95_hours',
    align: 'center',
    children: [
      {
        title: currentMonth.value,
        colKey: 'current_p95_hours',
        sortType: 'all',
        sorter: (a, b) => a.current_p95_hours - b.current_p95_hours,
        cell: (h, { row }) => (row?.current_p95_hours !== undefined ? `${row?.current_p95_hours}小时` : '--'),
      },
      {
        title: compareMonth.value,
        colKey: 'compare_p95_hours',
        sortType: 'all',
        sorter: (a, b) => a.compare_p95_hours - b.compare_p95_hours,
        cell: (h, { row }) => (row?.compare_p95_hours !== undefined ? `${row?.compare_p95_hours}小时` : '--'),
      },
      {
        title: '环比',
        colKey: 'growth_rate',
        align: 'left',
        sortType: 'all',
        sorter: (a, b) => {
          const aValue = calculateGrowthRate(a.current_p95_hours, a.compare_p95_hours).value;
          const bValue = calculateGrowthRate(b.current_p95_hours, b.compare_p95_hours).value;
          return aValue - bValue;
        },
        cell: (h, { row }) => {
          const rate = calculateGrowthRate(row?.current_p95_hours, row?.compare_p95_hours);
          return h(
            'div',
            {
              class: rate?.class,
            },
            rate?.text,
          );
        },
      },
    ],
  },
  {
    title: '单据数量',
    colKey: 'done_orders',
    align: 'center',
    children: [
      {
        title: currentMonth.value,
        colKey: 'current_done_orders',
        align: 'right',
        sortType: 'all',
        sorter: (a, b) => a.current_done_orders - b.current_done_orders,
        cell: (h, { row }) => {
          if (row?.current_done_orders === undefined) {
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
            row?.current_done_orders,
          );
        },
      },
      {
        title: compareMonth.value,
        colKey: 'compare_done_orders',
        align: 'right',
        sortType: 'all',
        sorter: (a, b) => a.compare_done_orders - b.compare_done_orders,
        cell: (h, { row }) => {
          if (row?.compare_done_orders === undefined) {
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
            row?.compare_done_orders,
          );
        },
      },
    ],
  },
]);

const sort = ref<TableProps['sort']>();

const sortChange: TableProps['onSortChange'] = (sortVal, options) => {
  sort.value = sortVal;
  detailData.value = options.currentDataSource;
};
const dataChange: TableProps['onDataChange'] = (newData) => {
  detailData.value = newData;
};

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
