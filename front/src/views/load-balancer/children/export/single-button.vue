<script lang="ts" setup>
import { useI18n } from 'vue-i18n';
import { AUTH_BIZ_UPDATE_CLB } from '@/constants/auth-symbols';
import { useExport } from './use-export';

const props = defineProps<{ data: any }>();
const { t } = useI18n();

const handleExport = () => {
  const { invokeExport } = useExport({
    target: 'lb',
    vendor: props.data.vendor,
    listeners: [{ lb_id: props.data.id }],
    single: { name: props.data.name },
  });
  invokeExport();
};
</script>

<template>
  <hcm-auth :sign="{ type: AUTH_BIZ_UPDATE_CLB }" v-slot="{ noPerm }">
    <bk-button text theme="primary" :disabled="noPerm" @click="handleExport">
      {{ t('导出') }}
    </bk-button>
  </hcm-auth>
</template>
