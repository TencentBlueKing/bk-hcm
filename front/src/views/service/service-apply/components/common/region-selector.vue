<script setup lang="ts">
import { computed, ref, useAttrs, watchEffect } from 'vue';
import { IRegionItem, useRegionStore } from '@/store/region';
import { ResourceTypeEnum } from '@/common/resource-constant';

defineOptions({ name: 'hcm-form-region' });

const props = defineProps<{
  vendor: string;
  resourceType?: ResourceTypeEnum.CVM | ResourceTypeEnum.VPC | ResourceTypeEnum.DISK | ResourceTypeEnum.SUBNET;
  multiple?: boolean;
  clearable?: boolean;
  disabled?: boolean;
}>();
const model = defineModel<string | string[]>();
const attrs = useAttrs();

const localModel = computed({
  get() {
    if (props.multiple && model.value && !Array.isArray(model.value)) {
      return [model.value];
    }
    return model.value;
  },
  set(val) {
    model.value = val;
  },
});

const regionStore = useRegionStore();
const list = ref<IRegionItem[]>([]);
watchEffect(async () => {
  if (props.vendor) {
    list.value = await regionStore.getRegionList({ vendor: props.vendor, resourceType: props.resourceType });
  } else {
    list.value = [];
  }
});
</script>

<template>
  <bk-select
    v-model="localModel"
    :list="list"
    :clearable="clearable"
    :disabled="disabled"
    :multiple="multiple"
    :multiple-mode="multiple ? 'tag' : 'default'"
    :id-key="'id'"
    :display-key="'name'"
    :loading="regionStore.regionListLoading"
    v-bind="attrs"
  />
</template>
