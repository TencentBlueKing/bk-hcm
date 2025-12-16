<script setup lang="ts">
import { computed } from 'vue';
import Inventory from './children/inventory/index.vue';
import ResourcePlan from './children/resource-plan/index.vue';
import { RequirementType } from '@/store/config/requirement';
import { type IResourcesDemandItem } from '@/store/resource-plan';
import { type ICvmDeviceItem } from '@/store/cvm/device';
import { type ICondition } from './typings';

interface Props {
  isShow: boolean;
  requireType: number;
  bizId?: number;
  initialCondition?: ICondition;
}

const props = withDefaults(defineProps<Props>(), {});

const emit = defineEmits<{
  apply: [data: Partial<IResourcesDemandItem | ICvmDeviceItem>, show: boolean];
}>();

const isUseResourcePlan = computed(() =>
  [RequirementType.Regular, RequirementType.Spring, RequirementType.Dissolve, RequirementType.ShortRental].includes(
    props.requireType,
  ),
);

const view = computed(() => (isUseResourcePlan.value ? ResourcePlan : Inventory));

const handleApply = (data: IResourcesDemandItem | ICvmDeviceItem) => {
  if (isUseResourcePlan.value) {
    const { device_type, region_id: region, zone_id: zone } = data as IResourcesDemandItem;
    emit(
      'apply',
      {
        device_type,
        region,
        zone,
        zones: [zone],
        charge_type: data.plan_type === '预测内' ? 'PREPAID' : 'POSTPAID_BY_HOUR',
      },
      false,
    );
  } else {
    const { device_type, region, zone } = data as ICvmDeviceItem;
    emit('apply', { device_type, region, zone, zones: [zone] }, false);
  }
};
</script>

<template>
  <div class="view-container" v-if="isShow">
    <component :is="view" :require-type="requireType" :initial-condition="initialCondition" @apply="handleApply" />
  </div>
</template>

<style lang="scss" scoped>
.view-container {
  padding: 20px 40px;
}
</style>
