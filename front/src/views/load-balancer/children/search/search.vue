<script setup lang="ts">
import { ModelPropertySearch } from '@/model/typings';
import { computed, ref, useAttrs, watch } from 'vue';
import { ISearchItem, ValidateValuesFunc } from 'bkui-vue/lib/search-select/utils';
import { ISearchCondition, ISearchSelectValue } from '@/typings';
import {
  buildSearchSelectValueBySearchQsCondition,
  getLocalFilterFnBySearchSelect,
  getSimpleConditionBySearchSelect,
} from '@/utils/search';

const props = withDefaults(
  defineProps<{
    fields: ModelPropertySearch[];
    condition?: ISearchCondition;
    flat?: boolean;
    localSearch?: boolean;
    options?: Array<{ field: string; formatter: Function }>;
    validateValues?: ValidateValuesFunc;
    searchDataConfig?: Record<string, Partial<ISearchItem>>;
  }>(),
  { flat: true },
);
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

const clear = () => {
  searchValue.value = [];
};

watch(searchValue, (val) => {
  if (props.flat) {
    emit('search', val, getSimpleConditionBySearchSelect(val, props.options));
  } else if (props.localSearch) {
    emit('search', val, getLocalFilterFnBySearchSelect(val, props.options));
  } else {
    emit('search', val);
  }
});

defineExpose({ clear });
</script>

<template>
  <bk-search-select v-model="searchValue" :data="searchData" :validate-values="validateValues" v-bind="attrs" />
</template>
