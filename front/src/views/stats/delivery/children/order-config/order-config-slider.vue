<script setup lang="ts">
import { Plus } from 'bkui-vue/lib/icon';
import { ref, useTemplateRef, watch } from 'vue';
import OrderConfigBlock from './order-config-block.vue';
import orderConfigDialog from './order-config-dialog.vue';
import http from '@/http';
import { YearMonthsListData } from './typings';

const model = defineModel<boolean>();
const emit = defineEmits<{
  confirm: [config: any];
  cancel: [];
  hidden: [];
}>();

const handleClosed = () => {
  model.value = false;
  emit('hidden');
};

const yearMonths = ref([]);
const orderConfig = useTemplateRef('orderConfig');

const handelAddConfig = () => {
  orderConfig.value.add();
};
const handleEditConfig = (configs: any[], month: string) => {
  orderConfig.value.edit({
    configs,
    month,
  });
};

const fetchYearMonths = async () => {
  const res = await http.post<YearMonthsListData>(
    '/api/v1/woa/task/config/findmany/apply/order/statistics/year_months',
  );
  yearMonths.value = res.data.details;
};

watch(model, (val) => {
  if (val) {
    fetchYearMonths();
  }
});
</script>

<template>
  <bk-sideslider
    v-model:is-show="model"
    title="剔除统计配置"
    width="960"
    @closed="handleClosed"
    @hidden="emit('hidden')"
  >
    <div class="dialog-content">
      <bk-button theme="primary" @click="handelAddConfig">
        <plus class="f22" />
        新增
      </bk-button>

      <div class="order-list">
        <order-config-block
          v-for="item in yearMonths"
          :key="item"
          :year-month="item.stat_month"
          @edit="handleEditConfig"
        />
      </div>
    </div>
  </bk-sideslider>

  <order-config-dialog ref="orderConfig" @refresh="fetchYearMonths" />
</template>

<style scoped lang="scss">
.dialog-content {
  padding: 28px 40px 0;
  overflow-y: auto;
}
</style>
