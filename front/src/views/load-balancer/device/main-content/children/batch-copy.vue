<script setup lang="ts">
import { useTemplateRef } from 'vue';
import { useI18n } from 'vue-i18n';
import { AngleDown } from 'bkui-vue/lib/icon';

import HcmDropdown from '@/components/hcm-dropdown/index.vue';
import CopyToClipboard from '@/components/copy-to-clipboard/index.vue';

export interface IBatchCopyItem {
  text: string;
  content: string;
}

withDefaults(
  defineProps<{
    items: IBatchCopyItem[];
    disabled?: boolean;
  }>(),
  {
    disabled: false,
  },
);

const { t } = useI18n();
const dropdownRef = useTemplateRef<typeof HcmDropdown>('dropdown');

const handleSuccess = () => {
  dropdownRef.value?.hidePopover();
};
</script>

<template>
  <HcmDropdown ref="dropdown" :disabled="disabled || !items.length" class="batch-copy">
    {{ t('复制') }}
    <AngleDown class="dropdown-icon" />
    <template #menus>
      <CopyToClipboard
        v-for="item in items"
        :key="item.text"
        type="dropdown-item"
        :text="item.text"
        :content="item.content"
        @success="handleSuccess"
      />
    </template>
  </HcmDropdown>
</template>

<style scoped lang="scss">
.batch-copy {
  .dropdown-icon {
    font-size: 26px;
  }
}
</style>
