<script setup lang="ts">
import { reactive, ref, watchEffect } from 'vue';
import { WeixinPro } from 'bkui-vue/lib/icon';
import CommonCard from '@/components/CommonCard';
import WName from '@/components/w-name';
import HappyFaceIcon from '@/assets/image/happy-face.svg';
import SadFaceIcon from '@/assets/image/sad-face.svg';
import ChartTrendMini from './chart-trend-mini.vue';
import ChartTrend from './chart-trend.vue';
import { useResourceUsageRate, type IDeviceLoadUsage } from './use-resource-usage-rate';

const props = defineProps<{
  bizId: number;
}>();

const { deviceLoadUsageLoading, getDeviceLoadUsage } = useResourceUsageRate();
const deviceLoadUsage = ref<IDeviceLoadUsage | null>(null);

const trendDialogState = reactive({
  show: false,
});

watchEffect(async () => {
  if (props.bizId) {
    const data = await getDeviceLoadUsage({ bk_biz_id: props.bizId });
    deviceLoadUsage.value = data;
  }
});

const numberToFixed = (value: number, digits = 2) => {
  if (isNaN(value) || !isFinite(value)) {
    return '--';
  }
  return value.toFixed(digits);
};

const numberCeil = (value: number) => {
  if (isNaN(value) || !isFinite(value)) {
    return '--';
  }
  return Math.ceil(value);
};

const handleViewTrend = () => {
  trendDialogState.show = true;
};
</script>

<template>
  <common-card class="resource-usage-rate-card">
    <template #title>
      <div class="title">
        <div class="title-text">资源利用率</div>
        <div class="title-info" v-if="deviceLoadUsage">
          <i class="hcm-icon bkhcm-icon-info-line" style="color: #979ba5"></i>
          <div class="title-info-content">
            根据公司政策，主机的CPU利用率需
            <em class="metric">>={{ deviceLoadUsage.threshold }}%</em>
            ，如未达标，新申领资源需评估合理性。如有疑问请
            <w-name name="ICR" alias="咨询ICR(IEG资源服务助手)">
              <template #icon>
                <weixin-pro width="16" height="16" />
              </template>
            </w-name>
          </div>
        </div>
      </div>
    </template>
    <div v-bkloading="{ loading: deviceLoadUsageLoading }" style="min-height: 80px">
      <div class="content" v-if="!deviceLoadUsageLoading && deviceLoadUsage">
        <img :src="HappyFaceIcon" alt="happy-face" class="face-icon" v-if="deviceLoadUsage.achieved_kpi" />
        <img :src="SadFaceIcon" alt="sad-face" class="face-icon" v-if="!deviceLoadUsage.achieved_kpi" />
        <dl class="content-item cpu-usage-rate">
          <dt class="content-item-title">
            CPU利用率
            <span :class="['rate-status', deviceLoadUsage.achieved_kpi ? 'qualified' : 'unqualified']">
              {{ deviceLoadUsage.achieved_kpi ? '达标' : '不达标' }}
            </span>
          </dt>
          <dd class="content-item-value">
            <span :class="['rate-num', deviceLoadUsage.achieved_kpi ? 'qualified' : 'unqualified']">
              {{ numberToFixed(deviceLoadUsage.cpu_usage) }}
            </span>
            %
          </dd>
        </dl>
        <dl class="content-item trend">
          <dt class="content-item-title">
            <span class="trend-text">趋势</span>
            <i class="hcm-icon bkhcm-icon-trend trend-icon" @click="handleViewTrend"></i>
          </dt>
          <dd class="content-item-value"><chart-trend-mini :biz-id="bizId" /></dd>
        </dl>
        <div class="divider-line"></div>
        <dl class="content-item current-empty-load">
          <dt class="content-item-title">当前空负载</dt>
          <dd class="content-item-value">
            <span class="num">{{ numberCeil(deviceLoadUsage.empty_load_cpu_core) }} 核</span>
            （{{ numberCeil(deviceLoadUsage.empty_load_os) }}个OS）
          </dd>
        </dl>
        <dl class="content-item current-low-load">
          <dt class="content-item-title">当前低负载</dt>
          <dd class="content-item-value">
            <span class="num">{{ numberCeil(deviceLoadUsage.low_load_cpu_core) }} 核</span>
            （{{ numberCeil(deviceLoadUsage.low_load_os) }}个OS）
          </dd>
        </dl>
        <a :href="`https://finops.woa.com/${bizId}/device-load-analysis`" target="_blank" class="view-detail-link">
          查看详情
          <i class="hcm-icon bkhcm-icon-jump-fill"></i>
        </a>
      </div>
      <div class="content" v-else-if="!deviceLoadUsageLoading">
        <bk-exception description="未查询到数据" scene="part" type="empty" />
      </div>
    </div>
    <bk-dialog v-model:is-show="trendDialogState.show" title="业务CPU利用率" dialog-type="show" width="960">
      <div class="trend-dialog-content"><chart-trend :height="'360px'" :width="'852px'" :biz-id="bizId" /></div>
    </bk-dialog>
  </common-card>
</template>

<style lang="scss" scoped>
.resource-usage-rate-card {
  :deep(.common-card-content) {
    padding: 0 16px;
  }
}

.title {
  display: flex;
  align-items: center;
  gap: 16px;

  .title-text {
    font-size: 14px;
    color: #313238;
    font-weight: 700;
  }

  .title-info {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 12px;
    color: #4d4f56;
    font-weight: 400;
  }

  .title-info-content {
    display: flex;
    align-items: center;
    gap: 4px;

    .metric {
      color: #e71818;
      font-weight: 700;
      font-style: normal;
    }
  }
}

.content {
  display: flex;
  align-items: center;
  gap: 24px;
  margin: 4px 0 12px;

  .face-icon {
    width: 52px;
    height: 52px;
  }

  .divider-line {
    width: 1px;
    height: 52px;
    background: #dcdee5;
  }

  .rate-status {
    display: inline-flex;
    font-size: 12px;
    color: #299e56;
    height: 22px;
    line-height: 22px;
    text-align: center;
    background: #daf6e5;
    border-radius: 2px;
    padding: 0 8px;
    font-weight: 400;

    &.qualified {
      color: #299e56;
      background: #daf6e5;
    }

    &.unqualified {
      color: #e71818;
      background: #ffebeb;
    }
  }

  .content-item {
    display: flex;
    flex-direction: column;
    min-height: 70px;
    justify-content: space-around;

    .content-item-title {
      display: flex;
      align-items: center;
      gap: 8px;
      font-size: 14px;
      color: #313238;
    }

    .content-item-value {
      display: flex;
      align-items: center;
      gap: 4px;
    }

    &.cpu-usage-rate {
      .content-item-title {
        font-weight: 700;
      }

      .rate-status {
        margin-left: auto;
      }

      .content-item-value {
        color: #4d4f56;
        font-size: 12px;
        gap: 3.5px;
        align-items: baseline;
      }

      .rate-num {
        font-family: Arial-BoldMT, Arial, sans-serif;
        font-size: 24px;
        font-weight: 700;

        &.qualified {
          color: #2caf5e;
        }

        &.unqualified {
          color: #ea3636;
        }
      }
    }

    &.trend {
      margin: 0 12px;
      height: 72px;
      width: 114px;
      background: #f5f7fa;
      position: relative;

      .content-item-title {
        position: absolute;
        top: 6px;
        left: 0;
        width: 100%;
        padding: 0 8px;
        justify-content: space-between;

        .trend-icon {
          font-size: 16px;
          color: #3a84ff;
          cursor: pointer;
        }

        .trend-text {
          font-size: 12px;
          color: #313238;
        }
      }

      .content-item-value {
        position: absolute;
        bottom: 0;
        left: 0;
        width: 100%;
        height: 60%;
      }
    }

    &.current-empty-load {
      margin: 0 18px 0 12px;
    }

    &.current-low-load {
      margin: 0 12px 0 18px;
    }

    &.current-empty-load,
    &.current-low-load {
      .content-item-value {
        font-size: 12px;
        color: #979ba5;

        .num {
          color: #313238;
          font-size: 16px;
          font-weight: 700;
        }
      }
    }
  }

  .view-detail-link {
    display: flex;
    align-items: center;
    align-self: self-start;
    gap: 6px;
    font-size: 12px;
    color: #3a84ff;
    font-weight: 400;
    line-height: 32px;
  }
}

.trend-dialog-content {
  padding: 30px 30px 0;
}
</style>
