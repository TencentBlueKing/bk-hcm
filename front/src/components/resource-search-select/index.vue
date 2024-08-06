<script setup lang="ts">
import { computed } from 'vue';
import { SearchSelect } from 'bkui-vue';
import type { ISearchValue } from 'bkui-vue/lib/search-select/utils';
import { ResourceTypeEnum } from '@/common/resource-constant';
import optionFactory from './option-factory';

defineOptions({ name: 'ResourceSearchSelect' });

export interface IResourceSelectProps {
  modelValue: ISearchValue;
  resourceType: ResourceTypeEnum;
  clearable?: boolean;
}

const props = withDefaults(defineProps<IResourceSelectProps>(), {
  clearable: true,
  searchOptions: () => [],
});

const emit = defineEmits(['update:modelValue']);

const { getOptionData, getOptionMenu } = optionFactory();
const searchOptions = getOptionData(props.resourceType);

const selectValue = computed({
  get() {
    return props.modelValue;
  },
  set(val) {
    emit('update:modelValue', val);
  },
});
</script>

<template>
  <SearchSelect
    v-model="selectValue"
    :class="'resource-search-select'"
    :clearable="props.clearable"
    :conditions="[]"
    :data="searchOptions"
    :get-menu-list="getOptionMenu"
    :unique-select="true"
  />
</template>

<style lang="scss" scoped>
.resource-search-select {
  width: 500px;
}
</style>
