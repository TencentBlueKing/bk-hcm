<script setup lang="ts">
import { computed } from 'vue';
import { SelectColumn } from '@blueking/ediatable';
import type { ISecondaryAccountItem } from '@/store/cloud-account-manage/secondary-account';

defineOptions({ name: 'SecondaryAccountSelector' });

const model = defineModel<string>();

const props = defineProps<{
  accountList: ISecondaryAccountItem[];
  loading?: boolean;
}>();

const selectList = computed(() =>
  props.accountList.map((item) => ({
    value: item.id,
    label: `${item.name}(${item.id})`,
  })),
);
</script>

<template>
  <SelectColumn v-model="model" :list="selectList" :loading="loading" filterable placeholder="请选择" />
</template>
