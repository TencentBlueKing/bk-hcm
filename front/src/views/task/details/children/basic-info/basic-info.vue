<script setup lang="ts">
import { ref, watch } from 'vue';
import { ResourceTypeEnum } from '@/common/resource-constant';
import GridContainer from '@/components/layout/grid-container/grid-container.vue';
import GridItem from '@/components/layout/grid-container/grid-item.vue';
import type { ITaskItem } from '@/store';
import fieldFactory from './field-factory';

const props = withDefaults(defineProps<{ resource: ResourceTypeEnum; data: Partial<ITaskItem> }>(), {});

const { getFields } = fieldFactory();

const detailValues = ref<Partial<ITaskItem>>();

const fields = getFields(props.resource);

watch(
  () => props.data,
  (data) => {
    detailValues.value = { ...data };
  },
  { deep: true, immediate: true },
);
</script>

<template>
  <grid-container fixed :column="3" :content-min-width="300" :label-width="110">
    <grid-item v-for="field in fields" :key="field.id" :label="field.name">
      <display-value
        :property="field"
        :value="detailValues[field.id]"
        :display="{ ...field.meta?.display, on: 'info' }"
      />
    </grid-item>
  </grid-container>
</template>

<style lang="scss" scoped></style>
