<script lang="ts" setup>
import { computed, ref, defineExpose, PropType, useAttrs, watch, onMounted } from 'vue';
import { useAccountStore } from '@/store';

import type {
  // PlainObject,
  FilterType,
} from '@/typings/resource';
import { Senarios, useWhereAmI } from '@/hooks/useWhereAmI';

const props = defineProps({
  bizId: Number as PropType<number>,
  filter: {
    type: Object as PropType<FilterType>,
    default() {
      return { op: 'and', rules: [] };
    },
  },
  type: String as PropType<string>,
  mustBiz: Boolean as PropType<boolean>,
  isResourcePage: Boolean as PropType<boolean>,
  disabled: Boolean,
});
const emit = defineEmits(['input', 'change']);

const attrs = useAttrs();

const accountStore = useAccountStore();
const accountList = ref([]);
const loading = ref(null);
const accountPage = ref(0);
const { whereAmI } = useWhereAmI();

const selectedValue = computed({
  get() {
    return '';
  },
  set(val) {
    emit('input', val);
  },
});

const getAccoutList = async () => {
  if (props.mustBiz && !props.bizId) {
    return;
  }

  if (loading.value === true) {
    return;
  }

  loading.value = true;
  const data = {
    filter: props.filter,
    page: {
      start: accountPage.value * 100,
      limit: 100,
    },
  };
  if (props.type) {
    data.params = { account_type: props.type };
  }
  const isResource = whereAmI.value === Senarios.resource;
  const res = await accountStore.getAccountList(data, props.bizId, isResource);

  accountPage.value += 1;

  if (!isResource) {
    accountList.value.push(...(res?.data || []));
  } else {
    accountList.value.push(...(res?.data?.details || []));
  }
  loading.value = false;
};

onMounted(() => {
  getAccoutList();
});

watch(
  () => whereAmI.value,
  (whereAmI) => {
    if (whereAmI === Senarios.business) {
      accountList.value = [];
      getAccoutList();
    }
  },
);

const handleChange = (val: string) => {
  const data = accountList.value.find(item => item.id === val);
  emit('change', data);
};

defineExpose({
  accountList,
});
</script>

<template>
  <bk-select
    v-model="selectedValue"
    filterable
    @scroll-end="getAccoutList"
    :loading="loading"
    @change="handleChange"
    v-bind="attrs"
    :disabled="props.disabled"
  >
    <bk-option v-for="(item, index) in accountList" :key="index" :value="item.id" :label="item.name" />
  </bk-select>
</template>
