<script lang="ts" setup>
import { computed, ref, defineExpose, PropType, useAttrs, watch } from 'vue';
import {
  useAccountStore,
} from '@/store';

const props = defineProps({
  bizId: Number as PropType<number>,
});
const emit = defineEmits(['input']);

const attrs = useAttrs();

const accountStore = useAccountStore();
const accountList = ref([]);
const loading = ref(null);
const accountPage = ref(0);

const selectedValue = computed({
  get() {
    return '';
  },
  set(val) {
    emit('input', val);
  },
});

const getAccoutList = async () => {
  if (loading.value === true) {
    return;
  }

  loading.value = true;
  const res = await accountStore.getAccountList({
    filter: { op: 'and', rules: [] },
    page: {
      start: accountPage.value,
      limit: 100,
    },
  }, props.bizId);
  accountPage.value += 1;
  if (props.bizId > 0) {
    accountList.value.push(...res?.data || []);
  } else {
    accountList.value.push(...res?.data?.details || []);
  }
  loading.value = false;
};

getAccoutList();

watch(() => props.bizId, (bizId, old) => {
  if (bizId > 0) {
    accountList.value = [];
    getAccoutList();
  }
});

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
    v-bind="attrs"
  >
    <bk-option
      v-for="(item, index) in accountList"
      :key="index"
      :value="item.id"
      :label="item.name"
    />
  </bk-select>
</template>
