<script lang="ts" setup>
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';
import { AUTH_BIZ_UPDATE_CLB } from '@/constants/auth-symbols';
import { useExport } from './use-export';
import { VendorEnum } from '@/common/constant';

const props = defineProps<{ selections: any[]; vendor: VendorEnum }>();
const { t } = useI18n();

const targetIds = computed(() => props.selections[0].targets.map((item) => item.id));

const handleExport = () => {
  const { invokeExport } = useExport({
    target: 'rs',
    vendor: props.vendor,
    listeners: [],
    target_ids: targetIds.value,
    check: false,
  });
  invokeExport();
};
</script>

<template>
  <hcm-auth :sign="{ type: AUTH_BIZ_UPDATE_CLB }" v-slot="{ noPerm }">
    <bk-button :disabled="!selections.length || noPerm" @click="handleExport">
      {{ t('批量导出') }}
    </bk-button>
  </hcm-auth>
</template>
