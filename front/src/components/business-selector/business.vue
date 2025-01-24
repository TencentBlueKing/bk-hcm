<script setup lang="ts">
import { computed, ref, watchEffect } from 'vue';
import { useBusinessGlobalStore, type IBusinessItem } from '@/store/business-global';
import { isEmpty } from '@/common/util';

export type BusinessScopeType = 'full' | 'auth';

export interface IBusinessSelectorProps {
  disabled?: boolean;
  multiple?: boolean;
  clearable?: boolean;
  filterable?: boolean;
  showAll?: boolean;
  showSelectAll?: boolean;
  collapseTags?: boolean;
  allOptionId?: number;
  emptySelectAll?: boolean;
  scope?: BusinessScopeType;
  data?: IBusinessItem[];
  optionDisabled?: (item: IBusinessItem) => boolean;
}

const props = withDefaults(defineProps<IBusinessSelectorProps>(), {
  disabled: false,
  multiple: false,
  clearable: true,
  filterable: true,
  showAll: false,
  showSelectAll: false,
  allOptionId: 0,
  emptySelectAll: false,
  scope: 'full',
});

const model = defineModel<number | number[]>();

const businessGlobalStore = useBusinessGlobalStore();

const list = ref<IBusinessItem[]>([]);
const loading = ref(false);

watchEffect(async () => {
  loading.value = true;
  if (props.data) {
    list.value = props.data.slice();
    loading.value = false;
  } else if (props.scope === 'full') {
    // businessFullList在preload时已获取，这里直接使用，如之后有不使用缓存数据需要则另处理
    list.value = businessGlobalStore.businessFullList;
    loading.value = businessGlobalStore.businessFullListLoading;
  } else if (props.scope === 'auth') {
    // businessAuthorizedList在preload时已获取
    list.value = businessGlobalStore.businessAuthorizedList;
    loading.value = businessGlobalStore.businessAuthorizedListLoading;
  }
});

const localModel = computed({
  get() {
    if (props.showAll && props.emptySelectAll && isEmpty(model.value)) {
      return [props.allOptionId];
    }
    if (props.multiple && model.value && !Array.isArray(model.value)) {
      return [model.value];
    }
    return model.value;
  },
  set(val) {
    model.value = val;
  },
});
</script>

<template>
  <bk-select
    v-model="localModel"
    :disabled="disabled"
    :multiple="multiple"
    :filterable="filterable"
    :loading="loading"
    :clearable="clearable"
    :collapse-tags="collapseTags"
    :show-all="showAll"
    :all-option-id="allOptionId"
    :show-select-all="showSelectAll"
    :multiple-mode="multiple ? 'tag' : 'default'"
  >
    <!-- fix “全部”回显 -->
    <template #tag v-if="showAll && localModel?.[0] === allOptionId">
      <span class="all-option-name">全部</span>
    </template>
    <bk-option
      v-for="item in list"
      :key="item.id"
      :value="item.id"
      :label="item.name"
      :disabled="optionDisabled?.(item) === true"
    />
  </bk-select>
</template>

<style lang="scss" scoped>
.all-option-name {
  font-size: 12px;
}
</style>
