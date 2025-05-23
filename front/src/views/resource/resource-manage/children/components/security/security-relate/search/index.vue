<script setup lang="ts">
import { computed, ref, useAttrs, watch } from 'vue';
import { SecurityGroupRelatedResourceName } from '@/store/security-group';
import { useWhereAmI, Senarios } from '@/hooks/useWhereAmI';
import conditionFactory from './condition-factory';
import { ISearchSelectValue } from '@/typings';
import { getLocalFilterFnBySearchSelect, getSimpleConditionBySearchSelect } from '@/utils/search';
import { ValidateValuesFunc } from 'bkui-vue/lib/search-select/utils';
import { parseIP } from '@/utils';

const props = defineProps<{
  resourceName: SecurityGroupRelatedResourceName;
  operation: string;
  flat?: boolean;
  localSearch?: boolean;
  options?: Array<{ field: string; formatter: Function }>;
}>();
const emit = defineEmits<{
  search: [value: ISearchSelectValue, result?: any];
}>();
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

const validateValues: ValidateValuesFunc = async (item, values) => {
  if (!item) return '请选择条件';
  // IP值为单选，这里可以简单处理（即便是多IP搜索，粘贴上去也是一个值）
  if (['private_ip', 'public_ip'].includes(item.id)) {
    const { IPv4List, IPv6List } = parseIP(values[0].id);
    return Boolean(IPv4List.length || IPv6List.length) ? true : 'IP格式有误';
  }
  return true;
};

const clear = () => {
  searchValue.value = [];
};

watch(searchValue, (val) => {
  if (props.flat) {
    emit('search', val, getSimpleConditionBySearchSelect(val, props.options));
  } else if (props.localSearch) {
    emit('search', val, getLocalFilterFnBySearchSelect(val, props.options));
  } else {
    emit('search', val);
  }
});

defineExpose({ clear });
</script>

<template>
  <bk-search-select v-model="searchValue" :data="searchData" :validate-values="validateValues" v-bind="attrs" />
</template>

<style scoped lang="scss"></style>
