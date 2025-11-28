<script setup lang="ts">
import { nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue';
import http from '@/http';
import * as echarts from 'echarts/core';
import { BarChart } from 'echarts/charts';
import {
  TitleComponent,
  TooltipComponent,
  GridComponent,
  // 数据集组件
  DatasetComponent,
  // 内置数据转换器组件 (filter, sort)
  TransformComponent,
} from 'echarts/components';
import { LabelLayout, UniversalTransition } from 'echarts/features';
import { CanvasRenderer } from 'echarts/renderers';
import type {
  // 系列类型的定义后缀都为 SeriesOption
  BarSeriesOption,
} from 'echarts/charts';
import type {
  // 组件类型的定义后缀都为 ComponentOption
  TitleComponentOption,
  TooltipComponentOption,
  GridComponentOption,
  DatasetComponentOption,
} from 'echarts/components';
import type { ComposeOption } from 'echarts/core';
import { useBusinessGlobalStore } from '@/store/business-global';

import { getMonthRange } from '../../utils';

// 通过 ComposeOption 来组合出一个只有必须组件和图表的 Option 类型
type ECOption = ComposeOption<
  BarSeriesOption | TitleComponentOption | TooltipComponentOption | GridComponentOption | DatasetComponentOption
>;

const props = defineProps<{
  daterange: Date[];
}>();

// 注册必须的组件
echarts.use([
  TitleComponent,
  TooltipComponent,
  GridComponent,
  DatasetComponent,
  TransformComponent,
  BarChart,
  LabelLayout,
  UniversalTransition,
  CanvasRenderer,
]);

const businessGlobalStore = useBusinessGlobalStore();
const chartData = ref<any[]>([]);
const loading = ref(false);
let chartInstance: echarts.ECharts | null = null;
const chartRef = ref<HTMLElement | null>(null);
const option: ECOption = {
  tooltip: {
    trigger: 'item',
  },
  grid: {
    left: 10,
    right: 10,
    bottom: 10,
    top: 10,
  },
  dataset: {
    source: [],
  },
  xAxis: {
    type: 'value',
  },
  yAxis: {
    type: 'category',
    axisLine: {
      show: false,
    },
  },
  series: [
    {
      type: 'bar',
      label: { show: true, position: 'insideRight', color: '#fff' },
      color: '#F27051',
      barMaxWidth: '20',
    },
  ],
};

const fetchChartData = async () => {
  loading.value = true;
  const res = await http.post('/api/v1/woa/task/apply/findmany/bizs_hosts/statistics', {
    start_time: getMonthRange(props.daterange[0]).startTime,
    end_time: getMonthRange(props.daterange[1]).endTime,
  });
  chartData.value = (res.data?.details || []).sort((a: any, b: any) => b.host_count - a.host_count).slice(0, 10);
  chartInstance.setOption({
    dataset: {
      source: chartData.value.map((item: any) => [
        businessGlobalStore.businessFullList.find((business) => business.id === item.bk_biz_id)?.name,
        item.host_count,
      ]),
    },
  });
  loading.value = false;
  nextTick(() => {
    handleChartResize();
  });
};

watch(
  () => props.daterange,
  () => {
    fetchChartData();
  },
  { immediate: true },
);

const handleChartResize = () => {
  chartInstance?.resize();
};

onMounted(() => {
  chartInstance = echarts.init(chartRef.value);
  chartInstance?.setOption(option);
  window.addEventListener('resize', handleChartResize);
  handleChartResize();
});

onBeforeUnmount(() => {
  window.removeEventListener('resize', handleChartResize);
  chartInstance?.dispose();
  chartInstance = null;
});
</script>

<template>
  <div class="chart-container" v-bkloading="{ loading }">
    <div ref="chartRef" class="chart-instance-container" v-show="chartData.length"></div>
    <div class="empty-container" v-show="!chartData.length && !loading">
      <i class="hcm-icon bkhcm-icon-chart-empty-bar-top empty-icon"></i>
      <div class="empty-text">暂无数据</div>
    </div>
  </div>
</template>

<style lang="scss" scoped></style>
