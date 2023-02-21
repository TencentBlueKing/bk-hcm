<script lang="ts" setup>
import { computed, ref, watchEffect, defineExpose } from 'vue';
import {
  useAccountStore,
} from '@/store';

const emit = defineEmits(['input']);

const accountStore = useAccountStore();
const businessList = ref([]);
const loading = ref(null);

watchEffect(void (async () => {
  loading.value = true;
  const res = await accountStore.getBizList();
  loading.value = false;
  businessList.value = res?.data;
})());

const selectedValue = computed({
  get() {
    return '';
  },
  set(val) {
    emit('input', val);
  },
});

defineExpose({
  businessList,
});
</script>

<template>
  <bk-select
    v-model="selectedValue"
    filterable
    :loading="loading"
  >
    <bk-option
      v-for="(item, index) in businessList"
      :key="index"
      :value="item.id"
      :label="item.name"
    />
  </bk-select>
</template>
