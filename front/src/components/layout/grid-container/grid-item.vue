<template>
  <div
    :class="{
      'grid-item': true,
      'non-label': !($slots.label || label),
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
import { type VNode } from 'vue';

export interface IGridItemProps {
  label?: (() => string | VNode) | string;
  span?: number;
}

defineProps<IGridItemProps>();
</script>

<style lang="scss" scoped>
/* stylelint-disable */
.grid-item {
  display: grid;
  grid-column: var(--span) span;

  // 使用 1/-1 替代 span N，确保无论实际列数如何都能跨满整行
  &.span {
    grid-column: 1 / -1;
  }

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
      z-index: 3; // fix 被 bk-table header遮挡
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

    // 当前聚焦的元素 z-index 更高，解决被非聚焦元素遮挡的问题
    &:focus-within :deep(.form-element) {
      z-index: 10;
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
