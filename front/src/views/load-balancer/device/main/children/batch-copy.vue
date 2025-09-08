<template>
  <hcm-dropdown ref="dropdown" :disabled="!selections.length" class="dropdown-container">
    {{ t('复制') }}
    <angle-down class="dropdown-icon" />
    <template #menus>
      <copy-to-clipboard
        type="dropdown-item"
        :text="t('监听器ID')"
        :content="selectedLoadBalancerIDs"
        @success="handleSuccess"
      />
      <copy-to-clipboard
        type="dropdown-item"
        :text="t('监听器端口')"
        :content="selectedLoadBalancerPorts"
        @success="handleSuccess"
      />
    </template>
  </hcm-dropdown>
</template>

<script setup lang="ts">
import { computed, useTemplateRef } from 'vue';
import { useI18n } from 'vue-i18n';

import { AngleDown } from 'bkui-vue/lib/icon';
import HcmDropdown from '@/components/hcm-dropdown/index.vue';
import CopyToClipboard from '@/components/copy-to-clipboard/index.vue';

const props = defineProps<{ selections: any[] }>();
const { t } = useI18n();
const dropdownRef = useTemplateRef<typeof HcmDropdown>('dropdown');

const selectedLoadBalancerIDs = computed(() => props.selections?.map((item) => item.id)?.join('\n'));
const selectedLoadBalancerPorts = computed(() => props.selections?.map((item) => item.port)?.join('\n'));

const handleSuccess = () => {
  dropdownRef.value?.hidePopover();
};
</script>

<style scoped lang="scss">
.dropdown-container {
  .dropdown-icon {
    font-size: 26px;
  }
}
</style>
