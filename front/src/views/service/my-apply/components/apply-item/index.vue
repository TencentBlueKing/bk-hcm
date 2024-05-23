<template>
  <div :class="['item', { 'has-bottom-border': hasBottomBorder }, { 'is-active': isActive }]" @click.stop="handleClick">
    <div class="header-info">
      <span class="title" :title="data.sn">{{ data.sn }}</span>
      <div :class="['status', statusItem[data.status]?.value || '']">
        {{ statusItem[data.status]?.label || '' }}
      </div>
    </div>
    <div class="bottom-info">
      <section>
        <label class="label">类型：</label>
        <span class="value">{{ ServiceAccountType[data.type] }}</span>
      </section>
    </div>
    <div class="bottom-info">
      <!-- <section>
        <label class="label">备注：</label>
        <span class="value" :title="data.memo">{{ data.memo || '--' }}</span>
      </section> -->
      <section>
        {{ data.created_time }}
      </section>
    </div>
  </div>
</template>
<script lang="ts">
import { defineComponent, reactive, computed, toRefs } from 'vue';
import { useI18n } from 'vue-i18n';
import { ServiceAccountType } from '@/typings';
export default defineComponent({
  name: 'MyApplyItem',
  props: {
    data: {
      type: Object,
      default: () => {
        return {};
      },
    },
    active: {
      type: Number,
      required: true,
    },
    hasBottomBorder: {
      type: Boolean,
      default: true,
    },
  },
  emits: ['on-change'],
  setup(props, { emit }) {
    const { t } = useI18n();
    const statusItem = Object.freeze({
      pending: {
        label: t('审批中'),
        value: 'pending',
      },
      rejected: {
        label: t('审批驳回'),
        value: 'rejected',
      },
      pass: {
        label: t('审批通过'),
        value: 'pass',
      },
      cancelled: {
        label: t('已撤销'),
        value: 'cancelled',
      },
      delivering: {
        label: t('交付中'),
        value: 'delivering',
      },
      completed: {
        label: t('已完成'),
        value: 'completed',
      },
      deliver_error: {
        label: t('交付异常'),
        value: 'deliver_error',
      },
    });
    const state = reactive({
      fieldMap: [
        { field: `${t('资源')}：`, value: 'resource' },
        { field: `${t('备注')}：`, value: 'remarks' },
      ],
    });

    const isActive = computed(() => {
      const { active, data } = props;
      return active === data?.id;
    });

    const handleClick = () => {
      emit('on-change', props.data.id);
    };

    return {
      ...toRefs(state),
      isActive,
      statusItem,
      ServiceAccountType,
      handleClick,
    };
  },
});
</script>
<style lang="scss" scoped>
@import './index.scss';
</style>
