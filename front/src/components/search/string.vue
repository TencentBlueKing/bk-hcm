<script setup lang="ts">
import { computed, useAttrs } from 'vue';
import type { ModelProperty } from '@/model/typings';
import type { DisplayType } from '../form/typings';
defineOptions({ name: 'hcm-search-string' });

const model = defineModel<string | string[]>();
const props = withDefaults(
  defineProps<{ multiple: boolean; option: ModelProperty['option']; display?: DisplayType }>(),
  {
    multiple: true,
    option: () => ({}),
    display: () => ({
      appearance: 'tag-input',
    }),
  },
);
const attr = useAttrs();

const appearance = computed(() => props.display?.appearance);

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
</script>

<template>
  <template v-if="multiple">
    <!-- select -->
    <bk-select
      v-if="appearance === 'select'"
      v-model="localModel"
      v-bind="attr"
      :multiple="multiple"
      :multiple-mode="multiple ? 'tag' : 'default'"
      :collapse-tags="true"
    >
      <bk-option v-for="(name, id) in option" :key="id" :id="id" :name="name"></bk-option>
    </bk-select>
    <!-- tag-input -->
    <bk-tag-input
      v-else-if="appearance === 'tag-input'"
      v-model="localModel"
      v-bind="attr"
      clearable
      allow-create
      allow-auto-match
    ></bk-tag-input>
  </template>
  <!-- input -->
  <bk-input v-else v-model="localModel" clearable v-bind="attr"></bk-input>
</template>

<style scoped></style>
