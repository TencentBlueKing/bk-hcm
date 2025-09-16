<script setup lang="ts">
import { ModelProperty } from '@/model/typings';
defineOptions({ name: 'hcm-search-enum' });

const model = defineModel<string | string[]>();
withDefaults(defineProps<{ multiple: boolean; option: ModelProperty['option'] }>(), {
  multiple: true,
  option: () => ({}),
});
</script>

<template>
  <bk-select v-model="model" :multiple="multiple" :multiple-mode="multiple ? 'tag' : 'default'" :collapse-tags="true">
    <bk-option v-for="(name, id) in option" :key="id" :id="id" :name="name"></bk-option>
  </bk-select>
</template>
