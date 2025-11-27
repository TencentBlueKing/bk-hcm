<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue';
import dayjs from 'dayjs';
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
  height: string;
  width: string;
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

const { deviceCpuUsageTrendLoading, getDeviceCpuUsageTrend } = useResourceUsageRate();
const chartData = ref<IDeviceCpuUsageTrend[]>([]);
const chartRef = ref<HTMLElement | null>(null);
let chartInstance: echarts.ECharts | null = null;

const option: EChartsOption = {
  tooltip: {
    trigger: 'axis',
  },
  grid: {
    left: 0,
    right: 0,
    top: 0,
  },
  legend: {
    show: true,
    data: [
      {
        name: '利用率（%）',
        textStyle: {
          color: '#4D4F56',
        },
      },
    ],
  },
  dataset: {
    source: [],
  },
  xAxis: {
    type: 'category',
    axisLine: {
      show: false,
    },
  },
  yAxis: {
    type: 'value',
  },
  series: [
    {
      type: 'line',
      name: '利用率（%）',
      symbol: 'none',
      tooltip: {
        valueFormatter: (value: any) => `${value.toFixed(2)}%`,
      },
    },
  ],
};

const fetchChartData = async () => {
  const data = await getDeviceCpuUsageTrend({ bk_biz_id: props.bizId });
  chartData.value = data ?? [];
  chartInstance.setOption({
    dataset: {
      source: chartData.value.map((item: IDeviceCpuUsageTrend) => [dayjs(item.date).format('YYYY-MM'), item.cpu_usage]),
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
  <div
    class="chart-trend-container"
    :style="{ height: props.height, width: props.width }"
    v-bkloading="{ loading: deviceCpuUsageTrendLoading }"
  >
    <div
      class="chart-instance-container"
      ref="chartRef"
      v-show="chartData.length"
      :style="{ height: props.height, width: props.width }"
    ></div>
    <bk-exception
      class="empty-container"
      v-show="!chartData.length && !deviceCpuUsageTrendLoading"
      :style="{ height: props.height, width: props.width }"
      description="未查询到数据"
      scene="part"
      type="empty"
    />
  </div>
</template>

<style lang="scss" scoped>
.chart-trend-container {
  width: 100%;
  height: 100%;

  .chart-instance-container {
    width: 100%;
    height: 100%;
  }

  .empty-container {
    width: 100%;
    height: 100%;
    display: flex;
    justify-content: center;
    align-items: center;
    margin-top: -30px;
  }
}
</style>
