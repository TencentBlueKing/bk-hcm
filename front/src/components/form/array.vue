<script setup lang="ts">
import { computed, ref } from 'vue';
import { ModelProperty } from '@/model/typings';
import { DisplayType } from './typings';
import { InputColumn } from '@blueking/ediatable';

defineOptions({ name: 'hcm-form-array' });

const props = withDefaults(
  defineProps<{ multiple: boolean; option: ModelProperty['option']; display?: DisplayType }>(),
  {
    multiple: false,
    option: () => ({}),
  },
);

const model = defineModel<string | string[]>();

const displayOn = computed(() => props.display?.on || 'cell');
const appearance = computed(() => props.display?.appearance);

const localModel = computed({
  get() {
    if (props.multiple && !Array.isArray(model.value)) {
      return [model.value];
    }
    return model.value;
  },
  set(val) {
    model.value = val;
  },
});

const defaultComp = computed(() => (props.display?.on === 'cell' ? InputColumn : 'bk-input'));

const appearanceComps: Record<string, any> = {};

const compColumnRef = ref();

defineExpose({
  getValue() {
    if (compColumnRef.value?.getValue) {
      return compColumnRef.value.getValue().then(() => model.value);
    }
    return model.value;
  },
});
</script>

<template>
  <component
    v-if="appearance"
    ref="compColumnRef"
    v-model="localModel"
    :is="appearanceComps[appearance]"
    :display-on="displayOn"
    :option="option"
  />
  <template v-else>
    <component :is="defaultComp" v-model="localModel" ref="compColumnRef" />
  </template>
</template>
