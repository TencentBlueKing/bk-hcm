<script lang="ts" setup>
import { ref } from 'vue';

export interface DropDownPopover {
  trigger: 'manual' | 'click' | 'hover';
}
export interface DropDownMenuProps {
  disabled: boolean;
  isShow?: boolean;
  popoverOptions?: DropDownPopover;
}

defineOptions({ name: 'hcm-dropdown' });

const props = withDefaults(defineProps<DropDownMenuProps>(), {
  isShow: false,
  popoverOptions: () => ({
    trigger: 'manual',
  }),
});

const show = ref<boolean>(props.isShow);
let popHideTimerId: ReturnType<typeof setTimeout> | null = null;
let popShowTimerId: ReturnType<typeof setTimeout> | null = null;
let isMouseenter = false;

const showPopover = () => {
  popShowTimerId = setTimeout(() => {
    if (popHideTimerId) {
      clearTimeout(popHideTimerId);
    }
    show.value = true;
  });
};
const hidePopover = () => {
  show.value = false;
};
const handleMouseEnter = () => {
  if (!props.disabled) showPopover();
};
const handleMouseLeave = () => {
  popHideTimerId = setTimeout(() => {
    popShowTimerId && clearTimeout(popShowTimerId);
    hidePopover();
  }, 300);
};
const handleContentEnter = () => {
  if (popHideTimerId) {
    isMouseenter = true;
    clearTimeout(popHideTimerId);
    popHideTimerId = null;
  }
};
const handleContentLeave = () => {
  if (isMouseenter) {
    hidePopover();
    isMouseenter = false;
  }
};

defineExpose({ hidePopover });
</script>

<template>
  <div @mouseenter="handleMouseEnter" @mouseleave="handleMouseLeave" class="hcm-dropdown">
    <bk-dropdown :disabled="disabled" :is-show="show" :popover-options="popoverOptions" @hide="hidePopover">
      <bk-button :disabled="disabled" @click="showPopover">
        <slot></slot>
      </bk-button>
      <template #content>
        <bk-dropdown-menu @mouseenter="handleContentEnter" @mouseleave="handleContentLeave">
          <slot name="menus"></slot>
        </bk-dropdown-menu>
      </template>
    </bk-dropdown>
  </div>
</template>

<style lang="scss" scoped>
.hcm-dropdown {
  :deep(.icon-angle-down) {
    font-size: 26px;
  }
}
</style>
