<script lang="ts" setup>
import { computed, ref, watchEffect, defineExpose, PropType, reactive } from 'vue';
import { useAccountStore } from '@/store';
import { SelectColumn } from '@blueking/ediatable';

const props = defineProps({
  modelValue: Number as PropType<number>,
  authed: Boolean as PropType<boolean>,
  autoSelect: Boolean as PropType<boolean>,
  isAudit: Boolean as PropType<boolean>,
  isEditable: Boolean as PropType<boolean>,
});
const emit = defineEmits(['update:modelValue']);

const accountStore = useAccountStore();
const businessList = ref([]);
const loading = ref(null);

watchEffect(async () => {
  loading.value = true;
  let req = props.authed ? accountStore.getBizListWithAuth : accountStore.getBizList;
  if (props.isAudit) {
    req = accountStore.getBizAuditListWithAuth;
  }

  const res = await req();
  loading.value = false;
  businessList.value = res?.data;
});

const computedBuinessList = computed(() => {
  return businessList.value.map(({ name, id }) => ({
    value: id,
    label: name,
  }));
});

const selectedValue = computed({
  get() {
    if (props.modelValue) {
      return props.modelValue;
    }
    if (props.autoSelect) {
      const val = businessList.value.at(0)?.id ?? null;
      emit('update:modelValue', val);
      return val;
    }
    return null;
  },
  set(val) {
    emit('update:modelValue', val);
  },
});

const rules = reactive([
  {
    validator: (value: string) => Boolean(value),
    message: '请选择业务',
  },
]);

defineExpose({
  businessList,
  computedBuinessList,
  rules,
});
</script>

<template>
  <select-column
    v-model="selectedValue"
    v-if="isEditable"
    :list="computedBuinessList"
    filterable
    :loading="loading"
    :rules="rules"
  ></select-column>

  <bk-select v-model="selectedValue" filterable :loading="loading" v-else>
    <bk-option v-for="(item, index) in businessList" :key="index" :value="item.id" :label="item.name" />
  </bk-select>
</template>
