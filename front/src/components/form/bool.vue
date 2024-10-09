<script setup lang="ts">
import { computed, ref, useAttrs } from 'vue';
import { ModelProperty } from '@/model/typings';
import { DisplayType } from './typings';
import { SelectColumn } from '@blueking/ediatable';

defineOptions({ name: 'hcm-form-bool' });

const props = withDefaults(defineProps<{ option: ModelProperty['option']; display: DisplayType }>(), {
  option: () => ({}),
  display: () => ({
    on: 'default',
  }),
});

const model = defineModel<boolean | string>();
const attrs = useAttrs();

const trueText = computed(() => props.option.trueText as string);
const falseText = computed(() => props.option.falseText as string);
const dipslayOn = computed(() => props.display.on);
const appearance = computed(() => props.display.appearance);

// select组件目前不支持false或0的value，这里使用字符串值替代
const selectList = computed(() => [
  { value: 'true', label: trueText.value },
  { value: 'false', label: falseText.value },
]);

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
  <template v-if="dipslayOn === 'cell' && appearance === 'select'">
    <select-column :list="selectList" v-model="model" ref="selectColumnRef" v-bind="attrs" />
  </template>
  <template v-else-if="appearance === 'select'">
    <bk-select v-model="model" :list="selectList" v-bind="attrs" />
  </template>
  <template v-else>
    <bk-switcher v-model="model" :on-text="trueText" :off-text="falseText" v-bind="attrs" />
  </template>
</template>
