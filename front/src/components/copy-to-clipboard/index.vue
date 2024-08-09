<script setup lang="ts">
import { useSlots } from 'vue';
import { Button, Message } from 'bkui-vue';
import { Copy } from 'bkui-vue/lib/icon';
import { BkDropdownItem } from 'bkui-vue/lib/dropdown';
import useClipboard from 'vue-clipboard3';

defineOptions({ name: 'CopyToClipboard' });

export interface ICopyToClipboardProps {
  content: string;
  type?: 'icon' | 'dropdown-item';
  text?: string;
  disabled?: boolean;
  disabledTips?: string;
  successMsg?: string;
  errorMsg?: string;
}

const props = withDefaults(defineProps<ICopyToClipboardProps>(), {
  type: 'icon',
  text: '复制',
  disabled: false,
  disabledTips: '',
  successMsg: '复制成功',
  errorMsg: '复制失败',
});

const emit = defineEmits(['success', 'error']);

const slots = useSlots();

const { toClipboard } = useClipboard();

const handleCopy = async () => {
  if (props.disabled) {
    return;
  }
  try {
    const { content } = props;
    await toClipboard(content);
    Message({ theme: 'success', message: props.successMsg });
    emit('success', content);
  } catch (e) {
    Message({ theme: 'error', message: props.errorMsg });
    emit('error', e);
  }
};
</script>

<template>
  <template v-if="slots.default">
    <div @click.stop="handleCopy" class="copy-to-clipboard">
      <slot v-bind="{ disabled: props.disabled, disabledTips: props.disabledTips }"></slot>
    </div>
  </template>
  <Button
    v-else-if="props.type === 'icon'"
    class="copy-to-clipboard-button"
    :text="true"
    :disabled="props.disabled"
    @click.stop="handleCopy"
    v-bk-tooltips="{
      content: props.disabled ? props.disabledTips : props.text,
      disabled: (props.disabled && !props.disabledTips) || (!props.disabled && !props.text),
    }"
  >
    <Copy />
  </Button>
  <BkDropdownItem
    v-else-if="props.type === 'dropdown-item'"
    :class="['copy-to-clipboard-dropdown', { disabled: props.disabled }]"
    @click.stop="handleCopy"
    v-bk-tooltips="{ content: props.disabledTips, disabled: !props.disabledTips || !props.disabled }"
  >
    {{ props.text }}
  </BkDropdownItem>
</template>

<style lang="scss" scoped>
.copy-to-clipboard {
  display: inline-flex;
  align-items: center;
}
.copy-to-clipboard-button {
  vertical-align: text-top;
  &:not([disabled]):hover {
    color: #3a84ff;
  }
}
.copy-to-clipboard-dropdown {
  &.disabled {
    color: #dcdee5;
    cursor: not-allowed;
  }
}
</style>
