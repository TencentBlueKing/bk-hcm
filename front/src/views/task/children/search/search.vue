<script setup lang="ts">
import { ref, watch } from 'vue';
import { ModelProperty } from '@/model/typings';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import GridContainer from '@/components/layout/grid-container/grid-container.vue';
import GridItemFormElement from '@/components/layout/grid-container/grid-item-form-element.vue';
import GridItem from '@/components/layout/grid-container/grid-item.vue';
import type { ISearchProps, ISearchConditon } from '@/views/task/typings';
import conditionFactory from './condition-factory';

const props = withDefaults(defineProps<ISearchProps>(), {});

const emit = defineEmits<{
  (e: 'search', condition: ISearchConditon): void;
  (e: 'reset'): void;
}>();

const { getBizsId } = useWhereAmI();
const { getConditionField } = conditionFactory();

const formValues = ref<ISearchConditon>({});
let conditionInitValues: ISearchConditon;

const fields = getConditionField(props.resource);

const getSearchCompProps = (field: ModelProperty) => {
  if (field.type === 'account') {
    return {
      bizId: getBizsId(),
      resourceType: props.resource,
    };
  }
  return {
    option: field.option,
  };
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
  <div class="task-search">
    <grid-container layout="vertical" :column="4" :content-min-width="300" :gap="[16, 60]">
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
.task-search {
  background: #fff;
  box-shadow: 0 2px 4px 0 #1919290d;
  border-radius: 2px;
  padding: 16px 24px;
  margin-bottom: 16px;
  position: relative;
  z-index: 3; // fix被bk-table-head遮挡

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
