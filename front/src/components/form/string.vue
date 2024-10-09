<script setup lang="ts">
import { computed, ref, useAttrs } from 'vue';
import { DisplayType } from './typings';
import { ModelProperty } from '@/model/typings';
import { InputColumn } from '@blueking/ediatable';

defineOptions({ name: 'hcm-form-string' });

const props = withDefaults(defineProps<{ option: ModelProperty['option']; display?: DisplayType }>(), {
  option: () => ({}),
});

const model = defineModel<string>();
const attrs = useAttrs();

const comp = computed(() => (props.display?.on === 'cell' ? InputColumn : 'bk-select'));

const inputColumnRef = ref();

defineExpose({
  getValue() {
    if (inputColumnRef.value?.getValue) {
      return inputColumnRef.value.getValue().then(() => model.value);
    }
    return model.value;
  },
});
</script>

<template>
  <component :is="comp" v-model="model" ref="inputColumnRef" v-bind="attrs" />
</template>
