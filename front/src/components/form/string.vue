<script setup lang="ts">
import { computed, ref, useAttrs } from 'vue';
import { DisplayType } from './typings';
import { ModelProperty } from '@/model/typings';
import { InputColumn } from '@blueking/ediatable';

defineOptions({ name: 'hcm-form-string' });

const model = defineModel<string>();

const props = withDefaults(defineProps<{ option: ModelProperty['option']; display?: DisplayType }>(), {
  option: () => ({}),
});

const attrs = useAttrs();

const comp = computed(() => (props.display?.on === 'cell' ? InputColumn : 'bk-input'));
const appearance = computed(() => props.display?.appearance);

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
  <template v-if="appearance === 'radio'">
    <bk-radio-group v-model="model" ref="inputColumnRef" v-bind="attrs">
      <bk-radio v-for="(item, value) in option" :key="value" :label="value" :disabled="item.disabled">
        {{ item.label }}
      </bk-radio>
    </bk-radio-group>
  </template>
  <template v-else>
    <component :is="comp" v-model="model" ref="inputColumnRef" v-bind="attrs" />
  </template>
</template>
