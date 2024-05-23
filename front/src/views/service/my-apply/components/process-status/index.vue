<template>
  <div class="process-status-wrapper">
    <div class="title">{{ $t('申请状态') }}</div>
    <div class="item">
      <label class="label">{{ t('步骤') }}：</label>
      <div :class="['status', statusItem[data.status]?.value || '']">
        {{ statusItem[data.status]?.label || '' }}
      </div>
    </div>
    <div class="item mt10">
      <label class="label">{{ t('状态') }}：</label>
      <div class="content">{{ statusItem[data.status]?.desc || '' }}</div>
      <div class="error-text ml10" v-if="data.status === 'deliver_error'">
        {{ JSON.parse(data.delivery_detail).error.message }}
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { computed, defineComponent } from 'vue';
import { useI18n } from 'vue-i18n';

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
  setup(props) {
    const { t } = useI18n();
    const curLanguageIsCn = computed(() => {
      return true;
    });

    const statusItem = Object.freeze({
      pending: {
        label: t('审批中'),
        value: 'pending',
        desc: t('您的资源申请单已提交，在单据审批通过后，会创建相关资源'),
      },
      rejected: {
        label: t('审批驳回'),
        value: 'rejected',
        desc: t('您的资源申请单未通过审批，如有疑问，请联系审批人，或请重新提单'),
      },
      pass: {
        label: t('审批通过'),
        value: 'pass',
        desc: t('您的资源申请单已通过审批'),
      },
      cancelled: {
        label: t('已撤销'),
        value: 'cancelled',
        desc: t('您的资源申请单已撤销'),
      },
      delivering: {
        label: t('交付中'),
        value: 'delivering',
        desc: t('您的资源申请单已通过审批，正在等待资源从云上生产'),
      },
      completed: {
        label: t('已完成'),
        value: 'completed',
        desc: t('您的资源申请单已通过审批，并成功交付资源'),
      },
      deliver_error: {
        label: t('交付异常'),
        value: 'deliver_error',
        desc: JSON.parse(props.data.delivery_detail)?.error,
      },
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
      curLanguageIsCn,
      getApplyTypeDisplay,
      statusItem,
    };
  },
});
</script>

<style lang="scss" scoped>
@import './index.scss';
</style>
