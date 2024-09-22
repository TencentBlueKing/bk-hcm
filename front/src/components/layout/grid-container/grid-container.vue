<template>
  <div
    :class="{
      'grid-container': true,
      [layout]: true,
      bordered,
      fixed,
      [`grid-container-label-${labelAlign}`]: labelAlign,
    }"
    :style="{
      '--col-gap': colGap,
      '--row-gap': rowGap,
      '--content-min-width': contentMinWidth,
      '--content-max-width': contentMaxWidth,
      '--column': column,
      '--label-width': labelContainerWidth,
    }"
  >
    <slot></slot>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue';

export interface IGridContainerProps {
  bordered?: boolean;
  fixed?: boolean;
  column?: number;
  contentMaxWidth?: number | string;
  contentMinWidth?: number | string;
  gap?: (number | string)[] | number | string;
  labelAlign?: 'center' | 'left' | 'right';
  labelWidth?: number | string; // 上下布局时重置为100%
  layout: 'horizontal' | 'vertical';
}

defineOptions({
  name: 'GridContainer',
});

const props = withDefaults(defineProps<IGridContainerProps>(), {
  layout: 'horizontal',
  column: 4,
});

const localGap = computed(() => (Array.isArray(props.gap) ? props.gap : [props.gap, props.gap]));
const rowGap = computed(() => (isNaN(Number(localGap.value?.[0])) ? localGap.value?.[0] : `${localGap.value?.[0]}px`));
const colGap = computed(() => (isNaN(Number(localGap.value?.[1])) ? localGap.value?.[1] : `${localGap.value?.[1]}px`));

const contentMinWidth = computed(() =>
  isNaN(Number(props.contentMinWidth)) ? props.contentMinWidth : `${props.contentMinWidth}px`,
);
const contentMaxWidth = computed(() =>
  isNaN(Number(props.contentMaxWidth)) ? props.contentMaxWidth : `${props.contentMaxWidth}px`,
);

const labelContainerWidth = computed(() => {
  if (props.layout === 'vertical') return undefined;

  let width = props.labelWidth;
  if (!width) {
    width = props.layout === 'horizontal' ? 160 : undefined;
  }

  return isNaN(Number(width)) ? width : `${width}px`;
});
</script>

<style lang="scss" scoped>
.grid-container {
  display: grid;
  column-gap: var(--col-gap, 60px);

  // 默认方式，自动填满容器
  grid-template-columns: repeat(var(--column, auto-fill), 1fr);

  // 固定方式，根据内容宽度，更可控
  &.fixed {
    grid-template-columns: repeat(var(--column, auto-fill), max-content);
  }

  &.horizontal {
    // grid-auto-rows: minmax(var(--min-height, 32px), max-content);
    row-gap: var(--row-gap, 0px);
    :deep(.grid-item) {
      // grid-template-columns: var(--label-width) 1fr;
      // 以下样式提供一个最小宽度，确保在切换为编辑态时不会因为宽度的变化导致布局跳动
      grid-template-columns: var(--label-width) minmax(var(--content-min-width, 312px), var(--content-max-width, 1fr));
      gap: 8px;

      // span时，内容区域宽度填满不限最小最大宽度限制
      &.span {
        grid-template-columns: var(--label-width) 1fr;
      }

      .item-label {
        justify-content: flex-end;
        text-align: right;
        &::after {
          content: '：';
        }
      }

      .item-content {
        .form-element {
          &.element-checkbox,
          &.element-bool {
            // 水平排版时，设置固定高度使其与label对齐
            height: 32px;
          }
          .bk-switcher,
          .bk-checkbox-group {
            & + .action-button {
              margin-left: 10px;
            }
          }
        }
      }
    }
  }

  &.vertical {
    row-gap: var(--row-gap, 24px);
    :deep(.grid-item) {
      grid-template-columns: minmax(var(--content-min-width, 312px), var(--content-max-width, 1fr));
      // grid-template-rows: max-content;
      grid-auto-rows: auto minmax(var(--min-height, 32px), max-content);

      &.span {
        grid-template-columns: 1fr;
      }

      .item-label {
        padding: 2px 0;

        & + .item-content {
          padding: 7px 0;
        }
      }
    }
  }

  &.bordered {
    gap: 0;
    :deep(.grid-item) {
      margin-left: -1px;
      margin-top: -1px;
      gap: 0;
      .item-label,
      .item-content {
        color: #63656e;
      }
      .item-label {
        background: #fafbfd;
        border: 1px solid #dcdee5;
      }
      .item-content {
        background: #fff;
        border: 1px solid #dcdee5;

        .form-element {
          height: 100% !important;

          .bk-textarea,
          .bk-input,
          .bk-select {
            height: 100%;
          }
          .bk-select {
            .bk-select-trigger {
              height: 100%;
            }
          }

          .bk-textarea > textarea {
            height: 100%;
          }

          &.element-checkbox,
          &.element-bool {
            margin-left: 16px;
          }
        }
      }
    }

    &.horizontal {
      :deep(.grid-item) {
        .item-label,
        .item-content {
          margin: 0;
          padding: 11px 16px;
        }
        .item-label {
          &::after {
            content: '';
          }
        }
        .item-content {
          border-left: none;
        }
      }
    }

    &.vertical {
      :deep(.grid-item) {
        // 确保有折行时表格态边框能对齐
        grid-template-rows: 1fr 1fr;
        .item-label,
        .item-content {
          margin: 0;
          padding: 11px 16px;
        }

        .item-label {
          border-bottom: none;
        }
      }
    }
  }

  &.grid-container-label-left {
    :deep(.grid-item) {
      .item-label {
        justify-content: flex-start;
        text-align: left;
      }
    }
  }
  &.grid-container-label-right {
    :deep(.grid-item) {
      .item-label {
        justify-content: flex-end;
        text-align: right;
      }
    }
  }
  &.grid-container-label-center {
    :deep(.grid-item) {
      .item-label {
        justify-content: center;
        text-align: center;
      }
    }
  }
}
</style>
