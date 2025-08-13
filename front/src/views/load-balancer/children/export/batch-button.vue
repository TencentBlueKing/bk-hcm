<script lang="ts" setup>
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';
import { AUTH_BIZ_UPDATE_CLB } from '@/constants/auth-symbols';
import { useExport } from './use-export';

const props = defineProps<{ selections: any[]; isInDropdown?: boolean }>();
const emit = defineEmits(['click']);

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

  emit('click');
};
</script>

<template>
  <hcm-auth :sign="{ type: AUTH_BIZ_UPDATE_CLB }" v-slot="{ noPerm }" :class="{ 'is-in-dropdown': isInDropdown }">
    <bk-button
      text
      class="button"
      :disabled="!selections.length || vendorSet.size > 1 || noPerm"
      v-bk-tooltips="{ content: '所选负载均衡需属于同一云厂商', disabled: !selections.length || vendorSet.size === 1 }"
      @click="handleExport"
    >
      {{ t('批量导出') }}
    </bk-button>
  </hcm-auth>
</template>

<style lang="less" scoped>
.is-in-dropdown {
  width: 100%;

  .button {
    padding: 0 16px;
    display: inline-block;
    width: 100%;
    text-align: left;
  }
}
</style>
