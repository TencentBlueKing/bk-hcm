<script lang="ts" setup>
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';
import { AUTH_BIZ_UPDATE_CLB } from '@/constants/auth-symbols';
import { useExport } from './use-export';

const props = defineProps<{ selections: any[]; onlyExportListener?: boolean }>();
const { t } = useI18n();

const vendor = computed(() => props.selections[0].vendor);
const listeners = computed(() =>
  props.selections.reduce((acc, cur) => {
    const found = acc.find((item: any) => item.lb_id === cur.lb_id);
    if (found) {
      found.lbl_ids.push(cur.id);
    } else {
      acc.push({
        lb_id: cur.lb_id,
        lbl_ids: [cur.id],
      });
    }
    return acc;
  }, []),
);

const handleExport = () => {
  const { invokeExport } = useExport({
    target: 'listener',
    vendor: vendor.value,
    listeners: listeners.value,
    only_export_listener: props.onlyExportListener,
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
