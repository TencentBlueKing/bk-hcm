<script setup lang="ts">
import { ref, useAttrs, watch } from 'vue';
import { SecurityGroupRelatedResourceName } from '@/store/security-group';
import conditionFactory from './condition-factory';
import { ISearchSelectValue } from '@/typings';

const props = defineProps<{ resourceName: SecurityGroupRelatedResourceName; operation: string }>();
const emit = defineEmits(['search']);
const attrs: any = useAttrs();

const searchValue = ref<ISearchSelectValue>([]);

const { getConditionField } = conditionFactory();
const fields = getConditionField(props.resourceName, props.operation);
const searchData = fields.map(({ id, name }) => ({ id, name }));

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
