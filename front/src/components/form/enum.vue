<script setup lang="ts">
import { computed, ref, useAttrs } from 'vue';
import { ModelProperty } from '@/model/typings';
import { SelectColumn } from '@blueking/ediatable';
import { DisplayType } from './typings';

defineOptions({ name: 'hcm-form-enum' });

const props = withDefaults(
  defineProps<{ multiple: boolean; option: ModelProperty['option']; display?: DisplayType }>(),
  {
    multiple: false,
    option: () => ({}),
  },
);
const model = defineModel<string | string[]>();
const attrs = useAttrs();

const comp = computed(() => (props.display?.on === 'cell' ? SelectColumn : 'bk-select'));

const selectList = computed(() => Object.entries(props.option).map(([value, label]) => ({ value, label })));

const selectColumnRef = ref();

defineExpose({
  getValue() {
    if (selectColumnRef.value?.getValue) {
      return selectColumnRef.value.getValue().then(() => model.value);
    }
    return model.value;
  },
});
</script>

<template>
  <component
    :is="comp"
    v-model="model"
    ref="selectColumnRef"
    :list="selectList"
    :multiple="multiple"
    :multiple-mode="multiple ? 'tag' : 'default'"
    v-bind="attrs"
  />
</template>
