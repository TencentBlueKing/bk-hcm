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

const buildSearchValue = (condition?: ISearchCondition) => {
  if (!condition) return [];
  const fieldIdSet = new Set(props.fields.map((field) => field.id));
  const filteredCondition = Object.fromEntries(Object.entries(condition).filter(([id]) => fieldIdSet.has(id)));
  return buildSearchSelectValueBySearchQsCondition(filteredCondition, props.fields);
};

const searchValue = ref<ISearchSelectValue>(buildSearchValue(props.condition));

const searchData = computed(() =>
  props.fields.map(({ id, name, option }) => ({
    id,
    name,
    children: option ? Object.entries(option).map(([key, value]) => ({ id: key, name: value })) : [],
    ...(props.searchDataConfig?.[id] || {}),
  })),
);

watch(
  () => props.condition,
  (condition) => {
    if (condition === undefined) return;
    searchValue.value = buildSearchValue(condition);
  },
);

const handleUpdateModelValue = (val: ISearchSelectValue) => {
  searchValue.value = val;
  emit('search', val);
};

const clear = (trigger = true) => {
  searchValue.value = [];
  if (trigger) {
    emit('search', []);
  }
};

defineExpose({ clear });
</script>

<template>
  <bk-search-select
    :model-value="searchValue"
    :data="searchData"
    :validate-values="validateValues"
    v-bind="attrs"
    value-behavior="need-key"
    @update:model-value="handleUpdateModelValue"
  />
</template>
