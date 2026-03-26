<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { Message } from 'bkui-vue';
import usageOrderViewProperties from '@/model/rolling-server/usage-order.view';
import { RollingServerRecordItem, useRollingServerUsageStore } from '@/store';
import { ModelPropertyDisplay } from '@/model/typings';
import { timeFormatter } from '@/common/util';
import GridContainer from '@/components/layout/grid-container/grid-container.vue';
import GridItem from '@/components/layout/grid-container/grid-item.vue';
import GridItemFormElement from '@/components/layout/grid-container/grid-item-form-element.vue';

const model = defineModel<boolean>();
const props = defineProps<{ details: RollingServerRecordItem }>();
const emit = defineEmits<{
  success: [];
}>();

const rollingServerUsageStore = useRollingServerUsageStore();

const fieldIds = [
  'suborder_id',
  'bk_biz_id',
  'roll_date',
  'applied_core',
  'delivered_core',
  'returned_core',
  'not_returned_core',
];
const columnConfig: Record<string, Partial<ModelPropertyDisplay>> = {
  roll_date: { render: (value) => timeFormatter(String(value), 'YYYY-MM-DD') },
};

const fields: ModelPropertyDisplay[] = fieldIds.map((id) => ({
  ...usageOrderViewProperties.find((view) => view.id === id),
  ...columnConfig[id],
}));

// 减免核心数输入值
const exemptedCore = ref<number>(0);

// 监听 details 变化，初始化输入值
watch(
  () => props.details,
  (newDetails) => {
    if (newDetails) {
      exemptedCore.value = newDetails.exempted_returned_core || 0;
    }
  },
  { immediate: true },
);

// 计算减免后的未退还核心数
const notReturnedCoreAfterExempted = computed(() => {
  const notReturned = props.details?.not_returned_core || 0;
  const exempted = Number(exemptedCore.value) || 0;
  return Math.max(0, notReturned - exempted);
});

// 验证输入是否有效
const isValidInput = computed(() => {
  const value = Number(exemptedCore.value);
  const maxValue = props.details?.not_returned_core || 0;
  return !isNaN(value) && value >= 0 && value <= maxValue;
});

const isAgree = ref(false);

const handleConfirm = async () => {
  await rollingServerUsageStore.updateExemptedReturnedCore([props.details.id], Number(exemptedCore.value));
  Message({ theme: 'success', message: '提交成功' });
  handleClosed();
  emit('success');
};

const handleClosed = () => {
  model.value = false;
  isAgree.value = false;
};
</script>

<template>
  <bk-dialog v-model:is-show="model" title="单据配置" width="640">
    <div class="section">
      <div class="section-title">基本信息</div>
      <grid-container :column="1" label-width="170" label-align="right" content-min-width="auto" size="small">
        <grid-item v-for="field in fields" :key="field.id" :label="field.name as string">
          <display-value :property="field" :value="details[field.id]" :display="field" />
        </grid-item>
      </grid-container>
    </div>

    <div class="section">
      <div class="section-title">参数设置</div>
      <grid-container :column="1" label-width="170" label-align="right" content-min-width="120" content-max-width="120">
        <grid-item-form-element label="减免退还核数">
          <bk-input
            v-model="exemptedCore"
            type="number"
            :min="0"
            :max="details.not_returned_core"
            suffix="核"
          ></bk-input>
        </grid-item-form-element>
        <grid-item label="未退还核数（减免前）">
          <span class="calc-value">{{ details.not_returned_core }} 核</span>
        </grid-item>
        <grid-item label="未退还核数（减免后）">
          <span class="calc-value highlight">{{ notReturnedCoreAfterExempted }} 核</span>
        </grid-item>
      </grid-container>
    </div>

    <bk-alert theme="warning">
      <template #title>
        <span>配置减免核数，将会影响滚服核数，请确认该操作的影响。仅支持当月操作</span>
      </template>
    </bk-alert>

    <template #footer>
      <div class="footer">
        <bk-checkbox v-model="isAgree" class="agree-checkbox">已知晓变更影响，仍需变更</bk-checkbox>
        <bk-button
          theme="primary"
          :disabled="!isAgree || !isValidInput"
          :loading="rollingServerUsageStore.updateExemptedReturnedCoreLoading"
          @click="handleConfirm"
        >
          提交
        </bk-button>
        <bk-button
          class="ml8"
          :disabled="rollingServerUsageStore.updateExemptedReturnedCoreLoading"
          @click="handleClosed"
        >
          取消
        </bk-button>
      </div>
    </template>
  </bk-dialog>
</template>

<style scoped lang="scss">
.section {
  margin-bottom: 16px;

  .section-title {
    font-size: 14px;
    font-weight: 700;
    color: #313238;
    margin-bottom: 12px;
  }
}

.calc-value {
  color: #313238;

  &.highlight {
    font-weight: 700;
  }
}

.footer {
  display: flex;
  align-items: center;

  .agree-checkbox {
    margin-right: auto;
  }
}
</style>
