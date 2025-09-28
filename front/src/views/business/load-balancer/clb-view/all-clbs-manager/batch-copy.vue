<template>
  <hcm-dropdown
    v-bk-tooltips="{ content: '复制当前勾选数据' }"
    ref="dropdown"
    :disabled="!selections.length"
    class="dropdown-container"
  >
    {{ t('复制') }}
    <angle-down class="dropdown-icon" />
    <template #menus>
      <copy-to-clipboard
        type="dropdown-item"
        :text="t('负载均衡ID')"
        :content="selectedLoadBalancerCloudIDs"
        @success="handleSuccess"
      />
      <copy-to-clipboard
        type="dropdown-item"
        :text="t('负载均衡VIP')"
        :content="selectedLoadBalancerVIPs"
        @success="handleSuccess"
      />
      <copy-to-clipboard
        type="dropdown-item"
        :text="t('负载均衡域名')"
        :content="selectedLoadBalancerDomains"
        @success="handleSuccess"
      />
    </template>
  </hcm-dropdown>
</template>

<script setup lang="ts">
import { computed, useTemplateRef } from 'vue';
import { useI18n } from 'vue-i18n';
import { getInstVip } from '@/utils';

import { AngleDown } from 'bkui-vue/lib/icon';
import HcmDropdown from '@/components/hcm-dropdown/index.vue';
import CopyToClipboard from '@/components/copy-to-clipboard/index.vue';

const props = defineProps<{ selections: any[] }>();
const { t } = useI18n();
const dropdownRef = useTemplateRef<typeof HcmDropdown>('dropdown');

const selectedLoadBalancerCloudIDs = computed(() => props.selections?.map((item) => item.cloud_id)?.join('\n'));
const selectedLoadBalancerVIPs = computed(() => props.selections?.map((item) => getInstVip(item))?.join('\n'));
const selectedLoadBalancerDomains = computed(() => props.selections?.map((item) => item.domain)?.join('\n'));

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
