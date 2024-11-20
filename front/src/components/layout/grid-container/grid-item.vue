<template>
  <div
    :class="{
      'grid-item': true,
      span,
    }"
    :style="{
      '--span': span,
    }"
  >
    <div class="item-label" v-if="$slots.label || label">
      <template v-if="$slots.label">
        <slot name="label" />
      </template>
      <component :is="label()" v-else-if="typeof label === 'function'" />
      <template v-else>{{ label }}</template>
    </div>
    <div class="item-content">
      <slot />
    </div>
  </div>
</template>
<script setup lang="ts">
import { type VNode, PropType } from 'vue';

export interface IGridItemProps {
  label?: (() => string | VNode) | string;
  span?: number;
}

defineProps({
  label: {
    type: [String, Function] as PropType<IGridItemProps['label']>,
  },
  span: {
    type: Number as PropType<IGridItemProps['span']>,
  },
});
</script>

<style lang="scss" scoped>
.grid-item {
  display: grid;
  grid-column: var(--span) span;

  .item-label,
  .item-content {
    display: inline-flex;
    line-height: 1.5;
    font-size: 14px;
    color: #63656e;
    box-sizing: border-box;
  }

  .item-label {
    padding: 9.5px 0;
    word-break: break-word;

    & + .item-content {
      padding: 9.5px 0;
    }
  }

  .item-content {
    position: relative;
    color: #313238;

    :deep(.form-element) {
      position: absolute;
      display: flex;
      align-items: center;
      top: 4px;
      left: 0;
      z-index: 1;
      width: 100%; // 表单控件宽度铺满
      gap: 4px;

      .bk-textarea,
      .bk-input,
      .bk-select,
      .bk-date-picker,
      .bk-tag-input {
        flex: 1;
      }

      .action-button {
        display: flex;
        gap: 4px;

        .button-item {
          width: 32px;

          &:hover {
            color: #3a84ff;
          }
        }
        .save-button {
          font-size: 28px;
        }
        .cancel-button {
          font-size: 18px;
        }
      }
    }

    :deep(.form-text) {
      position: relative;
      padding-right: 16px;

      .edit-button {
        position: absolute;
        font-size: 12px;
        top: 4px;
        right: 0;
        height: 12px;
        color: #979ba5;
        cursor: pointer;

        &:hover {
          color: #3a84ff;
        }
      }
    }
  }
}
</style>
