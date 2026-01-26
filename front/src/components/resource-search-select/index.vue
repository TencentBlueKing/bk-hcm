<script setup lang="ts">
import { computed } from 'vue';
import { SearchSelect } from 'bkui-vue';
import type { ISearchValue, ValidateValuesFunc } from 'bkui-vue/lib/search-select/utils';
import { ResourceTypeEnum } from '@/common/resource-constant';
import optionFactory from './option-factory';
import { useResourceAccountStore } from '@/store/useResourceAccountStore';
import { storeToRefs } from 'pinia';

defineOptions({ name: 'ResourceSearchSelect' });

const props = withDefaults(defineProps<IResourceSelectProps>(), {
  clearable: true,
  valueBehavior: 'need-key',
});

const emit = defineEmits(['update:modelValue']);

export interface IResourceSelectProps {
  modelValue: ISearchValue[];
  resourceType: ResourceTypeEnum;
  clearable?: boolean;
  valueBehavior?: 'all' | 'need-key';
  validateValues?: ValidateValuesFunc;
}

const resourceAccountStore = useResourceAccountStore();
const { selectedAccountId, vendorInResourcePage } = storeToRefs(resourceAccountStore);

const { getOptionData, getOptionMenu } = optionFactory();
const searchOptions = computed(() => {
  let data = getOptionData(props.resourceType);
  // 如果当前选定了某个云账号筛选条件就剔除云厂商
  if (vendorInResourcePage.value) {
    data = data.filter((item) => item.id !== 'vendor');
    if (selectedAccountId.value) {
      // 如果选中了某个账号ID筛选条件就剔除云账号ID
      data = data.filter((item) => item.id !== 'account_id');
    }
  }
  return data;
});

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
    :value-behavior="valueBehavior"
    :validate-values="validateValues"
  />
</template>

<style lang="scss" scoped>
.resource-search-select {
  width: 500px;
}
</style>
