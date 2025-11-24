<script setup lang="ts">
import type { IComparePickerModel } from '../../typings';

const model = defineModel<IComparePickerModel>({ default: () => ({ compareType: 'yoy', daterange: [] }) });
</script>

<template>
  <div class="compare-picker">
    <bk-radio-group size="small" v-model="model.compareType">
      <bk-radio-button label="yoy">同比</bk-radio-button>
      <bk-radio-button label="mom">环比</bk-radio-button>
    </bk-radio-group>
    <bk-date-picker class="date-picker" size="small" type="monthrange" v-model="model.daterange" :clearable="false" />
  </div>
</template>

<style lang="scss" scoped>
.compare-picker {
  display: flex;
  align-items: center;
  gap: 8px;

  .date-picker {
    width: 170px;

    :deep(.bk-date-picker-rel) {
      .icon-wrapper {
        height: 26px;
        line-height: 26px;
      }

      .bk-date-picker-editor {
        height: 26px;
        line-height: 26px;
      }
    }
  }
}
</style>
