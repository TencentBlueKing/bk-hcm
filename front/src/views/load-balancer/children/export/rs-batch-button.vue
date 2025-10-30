<script lang="ts" setup>
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';
import { AUTH_UPDATE_CLB, AUTH_BIZ_UPDATE_CLB } from '@/constants/auth-symbols';
import { getAuthSignByBusinessId } from '@/utils';
import { useExport } from './use-export';
import { VendorEnum } from '@/common/constant';

const props = defineProps<{ selections: any[]; vendor: VendorEnum; bizId?: number }>();
const { t } = useI18n();

const targetIds = computed(() =>
  props.selections.reduce((prev, cur) => {
    prev.push(...cur.targets.map((item) => item.id));
    return prev;
  }, []),
);

const authSign = computed(() => getAuthSignByBusinessId(props.bizId, AUTH_UPDATE_CLB, AUTH_BIZ_UPDATE_CLB));

const handleExport = () => {
  const { invokeExport } = useExport({
    target: 'rs',
    vendor: props.vendor,
    listeners: [],
    targetIds: targetIds.value,
    check: false,
  });
  invokeExport();
};
</script>

<template>
  <hcm-auth :sign="authSign" v-slot="{ noPerm }">
    <bk-button :disabled="!selections.length || noPerm" @click="handleExport">
      {{ t('批量导出') }}
    </bk-button>
  </hcm-auth>
</template>
