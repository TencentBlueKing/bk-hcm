<script setup lang="ts">
import { nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue';
import http from '@/http';
import * as echarts from 'echarts/core';
import { LineChart } from 'echarts/charts';
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
  LineSeriesOption,
} from 'echarts/charts';
import type {
  // 组件类型的定义后缀都为 ComponentOption
  TitleComponentOption,
  TooltipComponentOption,
  GridComponentOption,
  DatasetComponentOption,
} from 'echarts/components';
import type { ComposeOption } from 'echarts/core';
import type { OptionDataValue } from 'echarts/types/src/util/types';
import { getPeriodRange } from '../../utils';
import type { IChartCompareProps } from '../../typings';

// 通过 ComposeOption 来组合出一个只有必须组件和图表的 Option 类型
type ECOption = ComposeOption<
  LineSeriesOption | TitleComponentOption | TooltipComponentOption | GridComponentOption | DatasetComponentOption
>;

const props = defineProps<IChartCompareProps>();

// 注册必须的组件
echarts.use([
  TitleComponent,
  TooltipComponent,
  GridComponent,
  DatasetComponent,
  TransformComponent,
  LineChart,
  LabelLayout,
  UniversalTransition,
  CanvasRenderer,
]);

const currentChartData = ref<any[]>([]);
const comparisonChartData = ref<any[]>([]);
const loading = ref(false);
let chartInstance: echarts.ECharts | null = null;
const chartRef = ref<HTMLElement | null>(null);
const option: ECOption = {
  tooltip: {
    trigger: 'item',
  },
  legend: {
    show: true,
  },
  grid: {
    left: 10,
    right: 10,
    top: 10,
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
    axisLabel: {
      formatter: '{value} %',
    },
  },
  series: [
    {
      type: 'line',
      name: '当前结单率',
      color: '#699DF4',
      tooltip: {
        valueFormatter: (value: OptionDataValue | OptionDataValue[]) => `${value}%`,
      },
    },
    {
      type: 'line',
      name: '对比结单率',
      color: '#F27051',
      tooltip: {
        valueFormatter: (value: OptionDataValue | OptionDataValue[]) => `${value}%`,
      },
    },
  ],
};

const fetchChartData = async () => {
  loading.value = true;
  const periodRange = getPeriodRange(
    props.option.daterange[0],
    props.option.daterange[1],
    props.option.compareType,
    'YYYY-MM-DD',
  );
  const [currentRes, comparisonRes] = await Promise.all([
    http.post('/api/v1/woa/task/apply/completion-rate/statistics', {
      start_time: periodRange.currentRange.start,
      end_time: periodRange.currentRange.end,
    }),
    http.post('/api/v1/woa/task/apply/completion-rate/statistics', {
      start_time: periodRange.comparisonRange.start,
      end_time: periodRange.comparisonRange.end,
    }),
  ]);
  currentChartData.value = currentRes.data?.details || [];
  comparisonChartData.value = comparisonRes.data?.details || [];

  chartInstance.setOption({
    dataset: {
      source: currentChartData.value.map((item: any, index: number) => [
        item.year_month,
        item.completion_rate,
        comparisonChartData.value[index]?.completion_rate,
      ]),
    },
  });
  loading.value = false;
  nextTick(() => {
    handleChartResize();
  });
};

watch(
  () => props.option,
  () => {
    fetchChartData();
  },
  { deep: true, immediate: true },
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
    <!-- 当前结单率数据为空则不显示图表，无论对比结单率数据是否为空 -->
    <div ref="chartRef" class="chart-instance-container" v-show="currentChartData.length"></div>
    <div class="empty-container" v-show="!currentChartData.length && !loading">
      <i class="hcm-icon hcm-icon bkhcm-icon-chart-empty-line empty-icon"></i>
      <div class="empty-text">暂无数据</div>
    </div>
  </div>
</template>

<style lang="scss" scoped></style>
