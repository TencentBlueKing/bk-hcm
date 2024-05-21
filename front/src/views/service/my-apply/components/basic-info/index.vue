<template>
  <div :class="['iam-apply-basic-info-wrapper']">
    <div class="title">{{ $t('基本信息') }}</div>
    <div class="item">
      <label class="label">{{ t('单号') }}：</label>
      <div class="content">{{ data.sn }}</div>
    </div>
    <div class="item">
      <label class="label">{{ t('类型') }}：</label>
      <div class="content">{{ ServiceAccountType[data.type] }}</div>
    </div>
    <div class="item">
      <label class="label">{{ t('申请人') }}：</label>
      <div class="content">{{ data.creator }}</div>
    </div>
    <div class="item">
      <label class="label">{{ $t('备注') }}：</label>
      <div class="content" :title="data.memo || ''">
        {{ data.memo || '--' }}
      </div>
    </div>
    <div class="item">
      <label class="label">{{ t('申请时间') }}：</label>
      <div class="content">{{ timeFormatter(data.created_at) }}</div>
    </div>
  </div>
</template>

<script lang="ts">
import { computed, defineComponent } from 'vue';
import { useI18n } from 'vue-i18n';
import { ServiceAccountType } from '@/typings';
import { timeFormatter } from '@/common/util';

export default defineComponent({
  name: 'MyApplyBasicInfo',
  props: {
    data: {
      type: Object,
      default: {} as any,
    },
    isShowExpired: {
      type: Boolean,
      default: false,
    },
  },
  setup() {
    const { t } = useI18n();
    const curLanguageIsCn = computed(() => {
      return true;
    });
    const getApplyTypeDisplay = (payload: string) => {
      const formatApplyType = {
        account_apply: () => {
          return t('账号申请');
        },
        service_apply: () => {
          return t('服务申请');
        },
      };
      const result = formatApplyType[payload] ? formatApplyType[payload]() : '-';
      return result;
    };
    return {
      t,
      timeFormatter,
      curLanguageIsCn,
      getApplyTypeDisplay,
      ServiceAccountType,
    };
  },
});
</script>

<style lang="scss" scoped>
@import './index.scss';
</style>
