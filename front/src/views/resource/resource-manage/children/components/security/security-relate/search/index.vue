<script setup lang="ts">
import { computed, ref, useAttrs, watch } from 'vue';
import { SecurityGroupRelatedResourceName } from '@/store/security-group';
import { useWhereAmI, Senarios } from '@/hooks/useWhereAmI';
import conditionFactory from './condition-factory';
import { ISearchSelectValue } from '@/typings';

const props = defineProps<{ resourceName: SecurityGroupRelatedResourceName; operation: string }>();
const emit = defineEmits(['search']);
const attrs: any = useAttrs();
const { whereAmI } = useWhereAmI();

const searchValue = ref<ISearchSelectValue>([]);

const { getConditionField } = conditionFactory();
const fields = computed(() => {
  const fields = getConditionField(props.resourceName, props.operation);
  if (whereAmI.value === Senarios.business) {
    return fields.filter((field) => field.id !== 'bk_biz_id');
  }
  return fields;
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
