<script lang="ts" setup>
import { ref } from 'vue';
import { Button, Dropdown } from 'bkui-vue';
import { BkDropdownItem, BkDropdownMenu } from 'bkui-vue/lib/dropdown';
import { AngleDown } from 'bkui-vue/lib/icon';
import { DropDownPopover } from '@/typings';
import { useI18n } from 'vue-i18n';
import CopyToClipboard from '@/components/copy-to-clipboard/index.vue';

export interface DropDownMenuProps {
  isShow: boolean;
  disabled: boolean;
  btnText: string;
  popoverOptions: DropDownPopover;
  type: string;
  itemValue: any;
  className: string;
}

defineOptions({ name: 'DropDownMenu' });

const { t } = useI18n();
const emit = defineEmits<(e: 'clickItemVal', val: string | number) => void>();

const props = withDefaults(defineProps<DropDownMenuProps>(), {
  isShow: false,
  btnText: '批量操作',
  type: 'default',
  className: 'host_operations_container',
  popoverOptions: () => ({
    trigger: 'manual',
    forceClickoutside: true,
  }),
});

const show = ref<boolean>(props.isShow);
let popHideTimerId: any = undefined;
const popShowTimerId: any = undefined;
let isMouseenter = false;

const showPopover = () => {
  show.value = true;
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
const handleClickItem = (type: string | number) => {
  hidePopover();
  emit('clickItemVal', type);
};
const handleContentEnter = () => {
  if (popHideTimerId) {
    isMouseenter = true;
    clearTimeout(popHideTimerId);
    popHideTimerId = undefined;
  }
};
const handleContentLeave = () => {
  if (isMouseenter) {
    hidePopover();
    isMouseenter = false;
  }
};
</script>

<template>
  <div :class="props.className" @mouseenter="handleMouseEnter" @mouseleave="handleMouseLeave">
    <Dropdown :disabled="props.disabled" :is-show="show" :popover-options="props.popoverOptions" @hide="hidePopover">
      <Button :disabled="props.disabled" @click="showPopover">
        {{ t(props.btnText) }}
        <AngleDown class="f26"></AngleDown>
      </Button>
      <template #content>
        <BkDropdownMenu @mouseenter="handleContentEnter" @mouseleave="handleContentLeave">
          <template v-if="props.type === 'default'">
            <BkDropdownItem
              v-for="(item, value, index) in props.itemValue"
              :key="index"
              @click="handleClickItem(value)"
            >
              批量{{ item }}
            </BkDropdownItem>
          </template>
          <template v-else>
            <CopyToClipboard
              v-for="(item, index) in props.itemValue"
              :key="index"
              type="dropdown-item"
              :text="t(item.name)"
              :content="item.value"
            />
          </template>
        </BkDropdownMenu>
      </template>
    </Dropdown>
  </div>
</template>
