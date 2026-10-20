<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import { LOAD_BALANCER_SUBTYPE, type LoadBalancerSubtype } from '@/views/load-balancer/constants';

defineOptions({ name: 'ResourceSubtypeSwitch' });

const activeType = defineModel<LoadBalancerSubtype>({ required: true });
const { t } = useI18n();
</script>

<template>
  <bk-radio-group v-model="activeType" class="resource-subtype-switch">
    <bk-radio-button :label="LOAD_BALANCER_SUBTYPE.CLB">{{ t('负载均衡') }}</bk-radio-button>
    <bk-radio-button :label="LOAD_BALANCER_SUBTYPE.EXCLUSIVE_CLUSTER">{{ t('独占集群') }}</bk-radio-button>
  </bk-radio-group>
</template>

<style lang="scss" scoped>
.resource-subtype-switch {
  flex-shrink: 0;
}
</style>
