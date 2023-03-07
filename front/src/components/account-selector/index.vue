<script lang="ts" setup>
import { computed, ref, watchEffect, defineExpose } from 'vue';
import {
  useAccountStore,
} from '@/store';

const emit = defineEmits(['input']);

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
  loading.value = true;
  const res = await accountStore.getAccountList({
    filter: { op: 'and', rules: [] },
    page: {
      start: accountPage.value,
      limit: 100,
    },
  });
  accountPage.value += 1;
  accountList.value.push(...res?.data?.details || []);
  loading.value = false;
};

watchEffect(void (async () => {
  getAccoutList();
})());

defineExpose({
  accountList,
});
</script>

<template>
  <bk-select
    v-model="selectedValue"
    filterable
    @scroll-end="getAccoutList"
  >
    <bk-option
      v-for="(item, index) in accountList"
      :key="index"
      :value="item.id"
      :label="item.name"
    />
  </bk-select>
</template>
