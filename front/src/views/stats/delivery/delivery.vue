<script setup lang="ts">
import { ref, reactive } from 'vue';
import dayjs from 'dayjs';
import { Message } from 'bkui-vue';
import { getPeriodRange, getRecentMonths } from '../utils';
import ChartApplyHostTop from './children/chart-apply-host-top.vue';
import ChartApplyCoreTop from './children/chart-apply-core-top.vue';
import DetailApplyHostTop from './children/detail-apply-host-top.vue';
import DetailApplyCoreTop from './children/detail-apply-core-top.vue';
import DetailApplyAvgTime from './children/detail-apply-avg-time.vue';
import DetailApplyPercentileTime from './children/detail-apply-percentile-time.vue';
import DetailApplyOrderTime from './children/detail-apply-order-time.vue';
import DetailApplyProdTime from './children/detail-apply-prod-time.vue';
import DetailApplyCompletionRate from './children/detail-apply-completion-rate.vue';
import DetailApplyDeliveryRate from './children/detail-apply-delivery-rate.vue';
import ChartApplyAvgTime from './children/chart-apply-avg-time.vue';
import ChartApplyPercentileTime from './children/chart-apply-percentile-time.vue';
import ChartApplyOrderTime from './children/chart-apply-order-time.vue';
import ChartApplyProdTime from './children/chart-apply-prod-time.vue';
import ChartApplyCompletionRate from './children/chart-apply-completion-rate.vue';
import ChartApplyDeliveryRate from './children/chart-apply-delivery-rate.vue';
import ComparePicker from './children/compare-picker.vue';
import OrderConfigSlider from './children/order-config/order-config-slider.vue';
import type { IComparePickerModel } from '../typings';

const detailSlider = {
  applyHostTop: {
    title: '申请主机数',
    content: DetailApplyHostTop,
  },
  applyCoreTop: {
    title: '申请核心数',
    content: DetailApplyCoreTop,
  },
  applyCompletionRate: {
    title: '【申请单据】结单率统计',
    content: DetailApplyCompletionRate,
  },
  applyDeliveryRate: {
    title: '【申请单据】需求交付率统计',
    content: DetailApplyDeliveryRate,
  },
  applyAvgTime: {
    title: '【交付耗时】平均耗时统计',
    content: DetailApplyAvgTime,
  },
  applyPercentileTime: {
    title: '【交付耗时】P95 耗时统计',
    content: DetailApplyPercentileTime,
  },
  applyOrderTime: {
    title: '【交付耗时】剔除审批阶段耗时统计',
    content: DetailApplyOrderTime,
  },
  applyProdTime: {
    title: '【交付耗时】生产阶段耗时统计',
    content: DetailApplyProdTime,
  },
};

const detailSideSliderState = reactive({
  isShow: false,
  title: '',
  content: null,
});

const recentSixMonths = getRecentMonths(6);
const periodRange = getPeriodRange(new Date(), new Date(), 'mom');
const topDateRange = ref([recentSixMonths.startDate, recentSixMonths.endDate]);
const detailCurrentDateTime = ref(periodRange.currentRange.start);
const detailCompareDateTime = ref(periodRange.comparisonRange.start);

const completionRateModel = ref<IComparePickerModel>({
  compareType: 'yoy',
  daterange: topDateRange.value,
});

const handleDetailSideSliderShow = (key: keyof typeof detailSlider) => {
  const { title, content } = detailSlider[key];
  detailSideSliderState.title = title;
  detailSideSliderState.content = content;
  detailSideSliderState.isShow = true;
};

const handleTopDateRangeChange = (date: Date[]) => {
  if (!Array.isArray(date)) {
    return;
  }
  const start = dayjs(date[0]);
  const end = dayjs(date[1]);
  const months = end.diff(start, 'month');
  if (months > 11) {
    Message({
      theme: 'error',
      message: '只支持选择12个月以内的时间范围',
    });
    return;
  }
  topDateRange.value = date;
  completionRateModel.value.daterange = date;
};

const showOrderConfigSlider = ref(false);
const handleOrderConfig = () => {
  showOrderConfigSlider.value = true;
};
</script>

<template>
  <div class="delivery">
    <div class="group">
      <div class="group-title">申请 TOP 10</div>
      <div class="group-content">
        <div class="chart">
          <div class="chart-title">
            <div class="chart-title-text">申请主机数 TOP10 业务</div>
            <div class="chart-title-extra">
              <bk-button text theme="primary" @click="handleDetailSideSliderShow('applyHostTop')">详情 &gt;</bk-button>
            </div>
          </div>
          <div class="chart-content"><chart-apply-host-top :daterange="topDateRange" /></div>
        </div>
        <div class="chart">
          <div class="chart-title">
            <div class="chart-title-text">申请核心数 TOP10 业务</div>
            <div class="chart-title-extra">
              <bk-button text theme="primary" @click="handleDetailSideSliderShow('applyCoreTop')">详情 &gt;</bk-button>
            </div>
          </div>
          <div class="chart-content"><chart-apply-core-top :daterange="topDateRange" /></div>
        </div>
      </div>
    </div>
    <div class="group">
      <div class="group-title">申请单据比例</div>
      <div class="group-content">
        <div class="chart">
          <div class="chart-title">
            <div class="chart-title-text">结单率</div>
            <compare-picker v-model="completionRateModel" />
            <div class="chart-title-extra">
              <bk-button text theme="primary" @click="handleDetailSideSliderShow('applyCompletionRate')">
                详情 &gt;
              </bk-button>
            </div>
          </div>
          <div class="chart-content"><chart-apply-completion-rate :option="completionRateModel" /></div>
        </div>
        <div class="chart">
          <div class="chart-title">
            <div class="chart-title-text">需求交付率</div>
            <div class="chart-title-extra">
              <bk-button text theme="primary" @click="handleDetailSideSliderShow('applyDeliveryRate')">
                详情 &gt;
              </bk-button>
            </div>
          </div>
          <div class="chart-content"><chart-apply-delivery-rate :daterange="topDateRange" /></div>
        </div>
      </div>
    </div>
    <div class="group">
      <div class="group-title">交付耗时情况</div>
      <div class="group-content">
        <div class="chart">
          <div class="chart-title">
            <div class="chart-title-text">平均耗时</div>
            <div class="chart-title-extra">
              <bk-button text theme="primary" @click="handleDetailSideSliderShow('applyAvgTime')">详情 &gt;</bk-button>
            </div>
          </div>
          <div class="chart-content"><chart-apply-avg-time :daterange="topDateRange" /></div>
        </div>
        <div class="chart">
          <div class="chart-title">
            <div class="chart-title-text">P95 耗时</div>
            <div class="chart-title-extra">
              <bk-button text theme="primary" @click="handleDetailSideSliderShow('applyPercentileTime')">
                详情 &gt;
              </bk-button>
            </div>
          </div>
          <div class="chart-content"><chart-apply-percentile-time :daterange="topDateRange" /></div>
        </div>
        <div class="chart">
          <div class="chart-title">
            <div class="chart-title-text">剔除审批阶段耗时</div>
            <div class="chart-title-extra">
              <bk-button text theme="primary" @click="handleDetailSideSliderShow('applyOrderTime')">
                详情 &gt;
              </bk-button>
            </div>
          </div>
          <div class="chart-content"><chart-apply-order-time :daterange="topDateRange" /></div>
        </div>
        <div class="chart">
          <div class="chart-title">
            <div class="chart-title-text">生产阶段耗时</div>
            <div class="chart-title-extra">
              <bk-button text theme="primary" @click="handleDetailSideSliderShow('applyProdTime')">详情 &gt;</bk-button>
            </div>
          </div>
          <div class="chart-content"><chart-apply-prod-time :daterange="topDateRange" /></div>
        </div>
      </div>
    </div>

    <bk-sideslider
      v-model:is-show="detailSideSliderState.isShow"
      :width="960"
      :title="detailSideSliderState.title"
      render-directive="if"
    >
      <div class="detail-toolbar">
        <div class="composed-date-picker">
          <span class="composed-title">当前月份</span>
          <bk-date-picker v-model="detailCurrentDateTime" type="month" :clearable="false" />
        </div>
        <div class="composed-date-picker">
          <span class="composed-title">对比月份</span>
          <bk-date-picker v-model="detailCompareDateTime" type="month" :clearable="false" />
        </div>
      </div>
      <div class="detail-content">
        <component
          :is="detailSideSliderState.content"
          :current-date="detailCurrentDateTime"
          :compare-date="detailCompareDateTime"
        />
      </div>
    </bk-sideslider>
    <teleport to="#breadcrumbHead">
      <bk-date-picker
        class="top-date-picker"
        type="monthrange"
        :model-value="topDateRange"
        :clearable="false"
        @change="handleTopDateRangeChange"
      />
    </teleport>
    <teleport to="#breadcrumbExtra">
      <bk-button class="config-button" @click="handleOrderConfig">
        <i class="icon hcm-icon bkhcm-icon-shezhi"></i>
        单据配置
      </bk-button>
    </teleport>
  </div>
  <order-config-slider v-model="showOrderConfigSlider" />
</template>

<style lang="scss" scoped>
.delivery {
  padding: 16px 24px;
}

.group {
  margin-bottom: 24px;

  .group-title {
    display: flex;
    align-items: center;
    height: 32px;
    background: #eaebf0;
    border-radius: 2px;
    padding: 0 16px;
    font-size: 14px;
    color: #313238;
    margin-bottom: 16px;
  }

  .group-content {
    display: grid;
    grid-template-columns: repeat(2, minmax(500px, 1fr));
    gap: 16px;
  }
}

.chart {
  background: #fff;
  box-shadow: 0 2px 4px 0 #1919290d;
  border-radius: 2px;

  .chart-title {
    padding: 0 16px;
    margin: 16px 0;
    display: flex;
    align-items: center;
    gap: 10px;

    .chart-title-text {
      font-size: 14px;
      font-weight: 700;
      color: #4d4f56;
    }

    .chart-title-extra {
      font-size: 12px;
      margin-left: auto;
    }
  }

  .chart-content {
    height: 300px;
    padding: 0 16px;
    margin: 16px 0;

    :deep(.chart-container) {
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
        flex-direction: column;
        justify-content: center;
        align-items: center;
        gap: 12px;

        .empty-icon {
          font-size: 60px;
          color: #dee0e3;
        }

        .empty-text {
          font-size: 12px;
          color: #979ba5;
        }
      }
    }
  }
}

.top-date-picker {
  width: 180px;
}

.config-button {
  .icon {
    font-size: 16px;
    margin-right: 4px;
  }
}

.detail-toolbar {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-top: 28px;
  margin-bottom: 16px;
  padding: 0 40px;
}

.detail-content {
  padding: 0 40px;

  :deep(.detail-apply-table-container) {
    min-height: 300px;
  }

  :deep(.detail-apply-table) {
    .up {
      color: #ea3636;

      &::before {
        content: '↑';
        font-weight: 700;
        margin-right: 4px;
      }
    }

    .down {
      color: #18c0a1;

      &::before {
        content: '↓';
        font-weight: 700;
        margin-right: 4px;
      }
    }

    .flat {
      color: #4d4f56;

      &::before {
        content: '-';
        color: #c4c6cc;
        margin-right: 4px;
      }
    }
  }
}

.composed-date-picker {
  display: inline-flex;
  align-items: center;

  .composed-title {
    font-size: 12px;
    color: #4d4f56;
    height: 32px;
    line-height: 32px;
    padding: 0 10px;
    background: #fafbfd;
    border: 1px solid #c4c6cc;
    border-radius: 2px 0 0 2px;

    & + .bk-date-picker {
      width: 100px;
      margin-left: -1px;
    }
  }
}
</style>
