<script setup lang="ts">
import { ModelProperty } from '@/model/typings';
import { ref, watchEffect } from 'vue';
defineOptions({ name: 'hcm-search-enum' });

const model = defineModel<string | string[]>();

const props = withDefaults(defineProps<{ multiple: boolean; option: ModelProperty['option'] }>(), {
  multiple: true,
  option: () => ({}),
});

const localOption = ref<ModelProperty['option']>({});

watchEffect(async () => {
  if (typeof props.option === 'function') {
    localOption.value = await props.option();
  } else {
    localOption.value = props.option;
  }
});
</script>

<template>
  <bk-select v-model="model" :multiple="multiple" :multiple-mode="multiple ? 'tag' : 'default'" :collapse-tags="true">
    <bk-option v-for="(name, id) in localOption" :key="id" :id="id" :name="name"></bk-option>
  </bk-select>
</template>
