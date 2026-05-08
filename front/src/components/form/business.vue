<script setup lang="ts">
import { computed, ref, useAttrs } from 'vue';
import BusinessSelector from '@/components/business-selector/business.vue';
import { type IBusinessItem } from '@/store/business-global';
import { DisplayType } from './typings';
import type { Rules } from '@blueking/ediatable';

defineOptions({ name: 'hcm-form-business' });

const model = defineModel<number | number[]>();

const props = withDefaults(
  defineProps<{
    multiple?: boolean;
    clearable?: boolean;
    filterable?: boolean;
    collapseTags?: boolean;
    optionDisabled?: (item: IBusinessItem) => boolean;
    display?: DisplayType;
    rules?: Rules;
  }>(),
  {
    multiple: false,
    clearable: false,
    filterable: true,
    collapseTags: true,
  },
);

const emit = defineEmits(['change']);
const attrs = useAttrs();

const businessSelectorRef = ref<InstanceType<typeof BusinessSelector>>();

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

const handleChange = (val: number | number[]) => {
  emit('change', val);
};

defineExpose({
  getValue() {
    if (businessSelectorRef.value?.getValue) {
      return businessSelectorRef.value.getValue();
    }
    return model.value;
  },
});
</script>

<template>
  <business-selector
    ref="businessSelectorRef"
    v-model="localModel"
    :multiple="multiple"
    :clearable="clearable"
    :filterable="filterable"
    :collapse-tags="collapseTags"
    :option-disabled="optionDisabled"
    @change="handleChange"
    :display="display"
    :rules="rules"
    v-bind="attrs"
  />
</template>
