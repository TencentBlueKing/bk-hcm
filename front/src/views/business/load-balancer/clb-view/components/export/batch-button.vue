<script lang="ts" setup>
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';
import { AUTH_BIZ_UPDATE_CLB } from '@/constants/auth-symbols';
import { useExport } from './use-export';

const props = defineProps<{ selections: any[] }>();
const { t } = useI18n();

const ids = computed(() => props.selections.map((item) => item.id));
const vendorSet = computed(() => new Set(props.selections.map((item) => item.vendor)));
const vendor = computed(() => [...vendorSet.value][0]);

const handleExport = () => {
  const { invokeExport } = useExport({
    target: 'lb',
    vendor: vendor.value,
    listeners: ids.value.map((id) => ({ lb_id: id })),
  });
  invokeExport();
};
</script>

<template>
  <hcm-auth :sign="{ type: AUTH_BIZ_UPDATE_CLB }" v-slot="{ noPerm }">
    <bk-button
      :disabled="!selections.length || vendorSet.size > 1 || noPerm"
      v-bk-tooltips="{ content: '所选负载均衡需属于同一云厂商', disabled: !selections.length || vendorSet.size === 1 }"
      @click="handleExport"
    >
      {{ t('批量导出') }}
    </bk-button>
  </hcm-auth>
</template>
