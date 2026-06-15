<script setup lang="ts">
import { computed, ref, useAttrs, useSlots, watchEffect } from 'vue';
import { useBusinessGlobalStore, type IBusinessItem } from '@/store/business-global';
import { isEmpty } from '@/common/util';
import { SelectColumn } from '@blueking/ediatable';
import type { DisplayType } from '@/components/form/typings';

export type BusinessScopeType = 'full' | 'auth';

export interface IBusinessSelectorProps {
  disabled?: boolean;
  multiple?: boolean;
  clearable?: boolean;
  filterable?: boolean;
  showAll?: boolean;
  showSelectAll?: boolean;
  collapseTags?: boolean;
  multipleMode?: 'tag' | 'default';
  allOptionId?: number;
  emptySelectAll?: boolean;
  scope?: BusinessScopeType;
  data?: IBusinessItem[];
  optionDisabled?: (item: IBusinessItem) => boolean;
  tagClearable?: boolean;
  display?: DisplayType;
}

const model = defineModel<number | number[]>();

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
  tagClearable: true,
});

const emit = defineEmits(['change']);

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
    list.value = businessGlobalStore.businessFullList.map((item) => ({
      ...item,
      disabled: props.optionDisabled?.(item) === true,
    }));
    loading.value = businessGlobalStore.businessFullListLoading;
  } else if (props.scope === 'auth') {
    // businessAuthorizedList在preload时已获取
    list.value = businessGlobalStore.businessAuthorizedList.map((item) => ({
      ...item,
      disabled: props.optionDisabled?.(item) === true,
    }));
    loading.value = businessGlobalStore.businessAuthorizedListLoading;
  }
});

const comp = computed(() => (props.display?.on === 'cell' ? SelectColumn : 'bk-select'));

const localModel = computed({
  get() {
    if (props.showAll && props.emptySelectAll && isEmpty(model.value)) {
      return [props.allOptionId];
    }
    if (props.multiple && model.value && !Array.isArray(model.value)) {
      return [model.value];
    }
    return model.value || (props.multiple ? [] : undefined);
  },
  set(val) {
    model.value = val;
  },
});

const handleChange = (val: number | number[]) => {
  emit('change', val);
};

const attrs = useAttrs();
const slots = useSlots();
const selectColumnRef = ref();

defineExpose({
  getValue() {
    if (selectColumnRef.value?.getValue) {
      return selectColumnRef.value.getValue().then(() => model.value);
    }
    return model.value;
  },
});
</script>

<template>
  <component
    :is="comp"
    v-model="(localModel as any)"
    ref="selectColumnRef"
    :disabled="disabled"
    :multiple="multiple"
    :filterable="filterable"
    :loading="loading"
    :clearable="clearable"
    :collapse-tags="collapseTags"
    :show-all="showAll"
    :all-option-id="allOptionId"
    :show-select-all="showSelectAll"
    :multiple-mode="multipleMode ? multipleMode : multiple ? 'tag' : 'default'"
    :class="{ 'hide-tag-close': !tagClearable }"
    :list="list"
    id-key="id"
    display-key="name"
    v-bind="attrs"
    @change="handleChange"
  >
    <template v-for="(slotFn, slotName) in slots" :key="slotName" #[slotName]="slotProps">
      <component :is="slotFn" v-if="slotFn" v-bind="slotProps" />
    </template>
  </component>
</template>

<style lang="scss" scoped>
.all-option-name {
  font-size: 12px;
}

.hide-tag-close {
  :deep(.bk-select-trigger) {
    .bk-tag-closable {
      .bk-tag-close {
        display: none !important;
      }
    }
  }
}
</style>
