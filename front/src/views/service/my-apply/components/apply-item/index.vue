<template>
  <div
    :class="['item', { 'has-bottom-border': hasBottomBorder }, { 'is-active': isActive }]"
    @click.stop="handleClick">
    <div class="header-info">
      <span class="title" :title="getApplyTitle(data)">{{ getApplyTitle(data) }}</span>
      <div :class="['status', statusItem[data.status]?.value || '']">
        {{ statusItem[data.status]?.label || '' }}
      </div>
    </div>
    <div class="bottom-info">
      <section>
        <label class="label">{{ fieldMap[0].field }}</label>
        <span class="value" :title="data[fieldMap[0].value]">{{ data[fieldMap[0].value] || '--' }}</span>
      </section>
    </div>
    <div class="bottom-info">
      <section>
        <label class="label">{{ fieldMap[1].field }}</label>
        <span class="value" :title="data[fieldMap[1].value]">{{ data[fieldMap[1].value] || '--' }}</span>
      </section>
      <section>
        {{ getComputedCreateTime(data.created_time) }}
      </section>
    </div>
  </div>
</template>
<script lang="ts">
import { defineComponent, reactive, computed, toRefs } from 'vue';
import { useI18n } from 'vue-i18n';
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
      reject: {
        label: t('已拒绝'),
        value: 'reject',
      },
      pass: {
        label: t('已通过'),
        value: 'pass',
      },
      cancelled: {
        label: t('已撤销'),
        value: 'cancelled',
      },
    });
    const state = reactive({
      fieldMap: [{ field: `${t('资源')}：`, value: 'resource' }, { field: `${t('备注')}：`, value: 'remarks' }],
    });

    const isActive = computed(() => {
      const { active, data } = props;
      return active === data?.id;
    });

    const handleClick = () => {
      emit('on-change', props.data);
    };

    const getApplyTitle = (data: Record<string, any>) => {
      const formatApplyTitle = {
        account_apply: () => {
          return `${data.extra_info.system_name}账号申请`;
        },
        service_apply: () => {
          return `${data.extra_info.system_name}服务申请`;
        },
      };
      const result = formatApplyTitle[data.type] ?  formatApplyTitle[data.type]() : '';
      return result;
    };

    const getComputedCreateTime = (payload: string) => {
      if (!payload) {
        return '--';
      }
      const date = payload.split(' ')[0];
      const time = date.split('-');
      return `${time[1]}-${time[2]}`;
    };

    return {
      ...toRefs(state),
      isActive,
      statusItem,
      handleClick,
      getApplyTitle,
      getComputedCreateTime,
    };
  },
});
</script>
<style lang="scss" scoped>
@import './index.scss';
</style>
