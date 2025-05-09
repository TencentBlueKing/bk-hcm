<script setup lang="ts">
import { computed, ref, useAttrs, watch } from 'vue';
import { useBusinessGlobalStore } from '@/store/business-global';
import { useWhereAmI, Senarios } from '@/hooks/useWhereAmI';
import conditionFactory from './condition-factory';
import { ISearchSelectValue } from '@/typings';
import { ISearchItem } from 'bkui-vue/lib/search-select/utils';
import { SecurityGroupRelatedResourceName } from '@/constants/security-group';

const props = defineProps<{ resourceName: SecurityGroupRelatedResourceName; operation: string }>();
const emit = defineEmits(['search']);
const attrs: any = useAttrs();
const { whereAmI } = useWhereAmI();
const businessGlobalStore = useBusinessGlobalStore();

const searchValue = ref<ISearchSelectValue>([]);

const { getConditionField } = conditionFactory();

// 业务下的配置
const filedExtraConfig: Record<string, Partial<ISearchItem>> = {
  bk_biz_id: {
    children: businessGlobalStore.businessFullList.map(({ id, name }) => ({ id, name })),
  },
};

const fields = computed(() => {
  const fields = getConditionField(props.resourceName, props.operation);
  if (whereAmI.value === Senarios.business) {
    return fields.filter((field) => field.id !== 'bk_biz_id');
  }

  return fields.map((field) => {
    if (filedExtraConfig[field.id]) {
      return { ...field, ...filedExtraConfig[field.id] };
    }
    return field;
  });
});
const searchData = computed(() => fields.value.map(({ id, name, children }) => ({ id, name, children })));

const clear = () => {
  searchValue.value = [];
};

watch(searchValue, (val) => {
  emit('search', val);
});

defineExpose({ clear });
</script>

<template>
  <bk-search-select v-model="searchValue" :data="searchData" v-bind="attrs" />
</template>

<style scoped lang="scss"></style>
