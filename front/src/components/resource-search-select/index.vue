<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { SearchSelect } from 'bkui-vue';
import type { ICommonItem, ISearchItem, ISearchValue, ValidateValuesFunc } from 'bkui-vue/lib/search-select/utils';
import { ResourceTypeEnum } from '@/common/resource-constant';
import type { ModelPropertySearch } from '@/model/typings';
import optionFactory from './option-factory';
import { mapSearchFieldsToItems } from './map-search-fields';
import { useResourceAccountStore } from '@/store/useResourceAccountStore';
import { storeToRefs } from 'pinia';

defineOptions({ name: 'ResourceSearchSelect' });

const props = withDefaults(defineProps<IResourceSelectProps>(), {
  clearable: true,
  valueBehavior: 'need-key',
  exclude: () => [],
});

const emit = defineEmits(['update:modelValue']);

export interface IResourceSelectProps {
  modelValue: ISearchValue[];
  resourceType?: ResourceTypeEnum;
  fields?: ModelPropertySearch[];
  clearable?: boolean;
  valueBehavior?: 'all' | 'need-key';
  validateValues?: ValidateValuesFunc;
  // 使用方按场景剔除不需要的条件，例如业务视角去掉使用业务/管理业务
  exclude?: string[];
}

const resourceAccountStore = useResourceAccountStore();
const { selectedAccountId, vendorInResourcePage } = storeToRefs(resourceAccountStore);

const { getOptionData, getOptionMenu } = optionFactory();
const searchOptions = computed(() => {
  let data: ISearchItem[] = [];
  if (props.fields?.length) {
    data = mapSearchFieldsToItems(props.fields);
  } else if (props.resourceType) {
    data = getOptionData(props.resourceType) ?? [];
  }
  if (props.exclude.length) {
    const omit = new Set(props.exclude);
    data = data.filter((item) => !omit.has(item.id));
  }
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

// 从 URL 回填时只有 id，异步选项（业务、云账号等）的名称要用 getOptionMenu 现取
const menuCache = new Map<string, Promise<ISearchItem[]>>();
const loadMenu = (option: ISearchItem) => {
  if (!menuCache.has(option.id)) {
    menuCache.set(
      option.id,
      Promise.resolve(getOptionMenu(option, '')).catch(() => [] as ISearchItem[]),
    );
  }
  return menuCache.get(option.id);
};

const isIdAsName = (value: ICommonItem) => value.name === undefined || String(value.name) === String(value.id);
const findAsyncOption = (id: string) => searchOptions.value.find((option) => option.id === id && option.async);

const displayValue = ref<ISearchValue[]>([]);
let resolveToken = 0;

const resolveDisplayNames = async (val: ISearchValue[]) => {
  resolveToken += 1;
  const token = resolveToken;
  const pending = val.filter((item) => findAsyncOption(item.id) && item.values?.some(isIdAsName));
  if (!pending.length) return;

  const menus = new Map<string, ISearchItem[]>();
  await Promise.all(pending.map(async (item) => menus.set(item.id, await loadMenu(findAsyncOption(item.id)))));
  if (token !== resolveToken) return;

  displayValue.value = val.map((item) => {
    const menu = menus.get(item.id);
    if (!menu?.length) return item;
    return {
      ...item,
      values: item.values.map((value) => {
        if (!isIdAsName(value)) return value;
        const hit = menu.find((option) => String(option.id) === String(value.id));
        return hit ? { ...value, name: hit.name } : value;
      }),
    };
  });
};

watch(
  () => props.modelValue,
  (val) => {
    displayValue.value = val ?? [];
    resolveDisplayNames(displayValue.value);
  },
  { immediate: true },
);

const handleUpdate = (val: ISearchValue[]) => {
  displayValue.value = val;
  emit('update:modelValue', val);
};
</script>

<template>
  <SearchSelect
    :model-value="displayValue"
    :class="'resource-search-select'"
    :clearable="props.clearable"
    :conditions="[]"
    :data="searchOptions"
    :get-menu-list="getOptionMenu"
    :unique-select="true"
    :value-behavior="valueBehavior"
    :validate-values="validateValues"
    @update:model-value="handleUpdate"
  />
</template>

<style lang="scss" scoped>
.resource-search-select {
  width: 500px;

  // 容器比 500 窄时自适应收缩（需容器允许收缩，如 flex 项加 min-width: 0）
  max-width: 100%;
  min-width: 240px;
}
</style>
