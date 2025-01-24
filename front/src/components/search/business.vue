<script setup lang="ts">
import { computed, useAttrs } from 'vue';
import BusinessSelector from '@/components/business-selector/business.vue';
import { isEmpty } from '@/common/util';

defineOptions({ name: 'hcm-search-business' });

const props = withDefaults(
  defineProps<{
    multiple?: boolean;
    clearable?: boolean;
    filterable?: boolean;
    collapseTags?: boolean;
    showAll?: boolean;
    cacheKey?: string;
  }>(),
  {
    multiple: true,
    clearable: true,
    filterable: true,
    collapseTags: true,
    showAll: false,
  },
);

const model = defineModel<number | number[]>();

const localModel = computed({
  get() {
    if (props.multiple && model.value && !Array.isArray(model.value)) {
      return [model.value];
    }
    return model.value;
  },
  set(val) {
    // 操作本地缓存，如果是空数据则删除缓存
    if (props.cacheKey) {
      if (isEmpty(val)) {
        localStorage.removeItem(props.cacheKey);
      } else {
        localStorage.setItem(props.cacheKey, JSON.stringify(val));
      }
    }
    model.value = val;
  },
});

const attrs = useAttrs();
</script>

<template>
  <business-selector
    v-model="localModel"
    :multiple="multiple"
    :clearable="clearable"
    :filterable="filterable"
    :collapse-tags="collapseTags"
    :show-all="showAll"
    v-bind="attrs"
  />
</template>
