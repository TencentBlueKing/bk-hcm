<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue';
import * as echarts from 'echarts/core';
import { LineChart } from 'echarts/charts';
import {
  LegendComponent,
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
  LineSeriesOption,
} from 'echarts/charts';
import type {
  LegendComponentOption,
  TitleComponentOption,
  TooltipComponentOption,
  GridComponentOption,
  DatasetComponentOption,
} from 'echarts/components';
import type { ComposeOption } from 'echarts/core';
import { type IDeviceCpuUsageTrend, useResourceUsageRate } from './use-resource-usage-rate';

// 通过 ComposeOption 来组合出一个只有必须组件和图表的 Option 类型
type EChartsOption = ComposeOption<
  | LineSeriesOption
  | TitleComponentOption
  | TooltipComponentOption
  | GridComponentOption
  | DatasetComponentOption
  | LegendComponentOption
>;

const props = defineProps<{
  bizId: number;
}>();

// 注册必须的组件
echarts.use([
  TitleComponent,
  LegendComponent,
  TooltipComponent,
  GridComponent,
  DatasetComponent,
  TransformComponent,
  LineChart,
  LabelLayout,
  UniversalTransition,
  CanvasRenderer,
]);

const { getDeviceCpuUsageTrend } = useResourceUsageRate();
const chartData = ref<IDeviceCpuUsageTrend[]>([]);
const chartRef = ref<HTMLElement | null>(null);
let chartInstance: echarts.ECharts | null = null;

const option: EChartsOption = {
  grid: {
    left: 0,
    right: 0,
    top: 0,
    bottom: 2,
  },
  dataset: {
    source: [],
  },
  xAxis: {
    type: 'category',
    show: false,
  },
  yAxis: {
    type: 'value',
    show: false,
  },
  series: [
    {
      type: 'line',
      symbol: 'none',
      lineStyle: {
        color: '#18c0a1',
        width: 1,
      },
      areaStyle: {
        color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
          { offset: 0, color: 'rgba(24, 192, 161, 0.4)' },
          { offset: 1, color: 'rgba(24, 192, 161, 0)' },
        ]),
      },
    },
  ],
};

const fetchChartData = async () => {
  const data = await getDeviceCpuUsageTrend({ bk_biz_id: props.bizId });
  chartData.value = data ?? [];
  chartInstance.setOption({
    dataset: {
      source: chartData.value.map((item: IDeviceCpuUsageTrend) => [item.date, item.cpu_usage]),
    },
  });
};

onMounted(() => {
  chartInstance = echarts.init(chartRef.value);
  chartInstance.setOption(option);
  fetchChartData();
});

onUnmounted(() => {
  chartInstance?.dispose();
  chartInstance = null;
});
</script>

<template>
  <div class="chart-trend-mini" ref="chartRef"></div>
</template>

<style lang="scss" scoped>
.chart-trend-mini {
  width: 100%;
  height: 100%;
}
</style>
