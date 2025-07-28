<script setup lang="ts">
import { ModelPropertySearch } from '@/model/typings';
import { computed, ref, useAttrs, watch } from 'vue';
import { ISearchItem, ValidateValuesFunc } from 'bkui-vue/lib/search-select/utils';
import { ISearchCondition, ISearchSelectValue } from '@/typings';
import { buildSearchSelectValueBySearchQsCondition } from '@/utils/search';

const props = defineProps<{
  fields: ModelPropertySearch[];
  condition?: ISearchCondition;
  validateValues?: ValidateValuesFunc;
  searchDataConfig?: Record<string, Partial<ISearchItem>>;
}>();
const emit = defineEmits<{
  search: [value: ISearchSelectValue, result?: any];
}>();
const attrs: any = useAttrs();

const searchValue = ref<ISearchSelectValue>(buildSearchSelectValueBySearchQsCondition(props.condition, props.fields));

const searchData = computed(() =>
  props.fields.map(({ id, name, option }) => ({
    id,
    name,
    children: option ? Object.entries(option).map(([key, value]) => ({ id: key, name: value })) : [],
    ...(props.searchDataConfig?.[id] || {}),
  })),
);

const triggerQuery = ref(true); // 是否触发emit（查询），主要用于视觉交互层面上的searchValue，数据交互层面还是取决于watch query中的condition
const clear = (trigger = true) => {
  searchValue.value = [];
  triggerQuery.value = trigger;
};

watch(searchValue, (val) => {
  if (!triggerQuery.value) return;

  emit('search', val);
});

defineExpose({ clear });
</script>

<template>
  <bk-search-select v-model="searchValue" :data="searchData" :validate-values="validateValues" v-bind="attrs" />
</template>
