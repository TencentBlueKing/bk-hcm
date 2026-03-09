<script setup lang="ts">
import { computed } from 'vue';
import { SearchSelect } from 'bkui-vue';
import type { ISearchValue, ValidateValuesFunc } from 'bkui-vue/lib/search-select/utils';
import { ResourceTypeEnum } from '@/common/resource-constant';
import optionFactory from './option-factory';
import { useRoute } from 'vue-router';
import { useAccountSelectorStore } from '@/store/account-selector';
import { VendorEnum } from '@/common/constant';
import { searchSelectValueToSearchQsCondition } from '@/utils/search';

defineOptions({ name: 'ResourceSearchSelect' });

const props = withDefaults(defineProps<IResourceSelectProps>(), {
  clearable: true,
  valueBehavior: 'need-key',
});

const emit = defineEmits<{
  (e: 'update:modelValue', value: ISearchValue[]): void;
  (e: 'change', condition: Record<string, string | number | string[] | number[]>): void;
}>();

export interface IResourceSelectProps {
  modelValue: ISearchValue[];
  resourceType: ResourceTypeEnum;
  subType?: string;
  clearable?: boolean;
  valueBehavior?: 'all' | 'need-key';
  validateValues?: ValidateValuesFunc;
}

const route = useRoute();
const accountSelectorStore = useAccountSelectorStore();

// 从 route.query 获取 accountId 和 vendor
const selectedAccountId = computed(() => (route.query.accountId as string) || '');
const currentVendor = computed(() => {
  const queryVendor = route.query.vendor as VendorEnum;
  if (queryVendor) return queryVendor;
  const accountIdVal = selectedAccountId.value;
  if (accountIdVal) {
    const account = accountSelectorStore.authorizedResourceAccountList.find(
      (a: { id: string }) => a.id === accountIdVal,
    );
    return account?.vendor || null;
  }
  return null;
});

// 根据 resourceType + subType 获取对应的 option 模块
const optionModule = computed(() => optionFactory(props.resourceType, props.subType));

const searchOptions = computed(() => {
  const data = optionModule.value.getOptionData();
  if (!data?.length) return [];
  // 如果当前选定了某个云账号筛选条件就剔除云厂商
  if (currentVendor.value) {
    return data.filter((item) => {
      if (item.id === 'vendor') return false;
      if (selectedAccountId.value && item.id === 'account_id') return false;
      return true;
    });
  }
  return data;
});

const selectValue = computed({
  get() {
    return props.modelValue;
  },
  set(val) {
    emit('update:modelValue', val);
    // 仅在用户交互时 emit change（setter 由 bk-search-select 触发）
    // 不能放在 watch(modelValue) 中，否则从 route.query 回写 searchValue 时也会触发，造成死循环
    const condition = searchSelectValueToSearchQsCondition((val || []) as any);
    emit('change', condition);
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
    :get-menu-list="optionModule.getOptionMenu"
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
