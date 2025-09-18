<script setup lang="ts">
import { computed, useAttrs } from 'vue';
import { ModelProperty } from '@/model/typings';
import FormList from '@/components/form/list.vue';
defineOptions({ name: 'hcm-search-list' });

const model = defineModel<string | number | (string | number)[]>();

const props = withDefaults(defineProps<{ multiple: boolean; list: ModelProperty['list'] }>(), {
  multiple: true,
});

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

const attrs = useAttrs();
</script>

<template>
  <form-list v-model="localModel" :list="list" :multiple="multiple" :collapse-tags="true" v-bind="attrs" />
</template>
