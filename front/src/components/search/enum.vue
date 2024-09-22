<script setup lang="ts">
import { type ResourceProperty } from '@/common/resource-constant';
defineOptions({ name: 'hcm-search-enum' });

withDefaults(defineProps<{ multiple: boolean; option: ResourceProperty['option'] }>(), {
  multiple: true,
  option: () => ({}),
});
const model = defineModel<string>();
</script>

<template>
  <bk-select v-model="model" :multiple="multiple" :multiple-mode="multiple ? 'tag' : 'default'" :collapse-tags="true">
    <bk-option v-for="(name, id) in option" :key="id" :id="id" :name="name"></bk-option>
  </bk-select>
</template>
