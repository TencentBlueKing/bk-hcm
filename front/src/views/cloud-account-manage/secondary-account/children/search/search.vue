<script setup lang="ts">
import { ref, watch } from 'vue';
import { ModelPropertySearch } from '@/model/typings';
import GridContainer from '@/components/layout/grid-container/grid-container.vue';
import GridItemFormElement from '@/components/layout/grid-container/grid-item-form-element.vue';
import GridItem from '@/components/layout/grid-container/grid-item.vue';

export interface ISearchCondition {
  [key: string]: any;
}

interface ISearchProps {
  fields: ModelPropertySearch[];
  condition: ISearchCondition;
}

const props = withDefaults(defineProps<ISearchProps>(), {});

const emit = defineEmits<{
  (e: 'search', condition: ISearchCondition): void;
  (e: 'reset'): void;
}>();

const formValues = ref<ISearchCondition>({});
let conditionInitValues: ISearchCondition;

const getSearchCompProps = (field: ModelPropertySearch) => {
  const searchProps = field?.props || {};
  const baseProps: Record<string, any> = {
    option: field.option,
    ...searchProps,
  };
  if (['extension.cloud_main_account_id', 'name'].includes(field.id)) {
    baseProps.pasteFn = (value: string) => value.split(/\r\n|\n|\r/).map((tag) => ({ id: tag, name: tag }));
  }
  return baseProps;
};

const handleSearch = () => {
  emit('search', formValues.value);
};

const handleReset = () => {
  formValues.value = { ...conditionInitValues };
  emit('reset');
};

watch(
  () => props.condition,
  (condition) => {
    formValues.value = { ...condition };
    // 只记录第一次的condition值，重置时回到最开始的默认值
    if (!conditionInitValues) {
      conditionInitValues = { ...formValues.value };
    }
  },
  { deep: true, immediate: true },
);
</script>

<template>
  <div class="search">
    <grid-container layout="vertical" :column="4" :content-min-width="'1fr'" :gap="[16, 60]">
      <grid-item-form-element v-for="field in fields" :key="field.id" :label="field.name">
        <component :is="`hcm-search-${field.type}`" v-bind="getSearchCompProps(field)" v-model="formValues[field.id]" />
      </grid-item-form-element>
      <grid-item :span="4" class="row-action">
        <bk-button theme="primary" @click="handleSearch">查询</bk-button>
        <bk-button @click="handleReset">重置</bk-button>
      </grid-item>
    </grid-container>
  </div>
</template>

<style lang="scss" scoped>
.search {
  background: #fff;
  box-shadow: 0 2px 4px 0 #1919290d;
  border-radius: 2px;
  padding: 16px 24px;

  // :deep(.grid-item .item-content .form-element) {
  //   position: relative;
  //   top: 0;
  // }

  .row-action {
    padding: 4px 0;

    :deep(.item-content) {
      gap: 10px;
    }

    .bk-button {
      min-width: 86px;
    }
  }
}
</style>
