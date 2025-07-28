<script setup lang="ts">
import { ITargetGroupDetails } from '@/store/load-balancer/target-group';
import { DisplayFieldFactory, DisplayFieldType } from '../../children/display/field-factory';
import usePage from '@/hooks/use-page';

import DataList from '../../children/display/data-list.vue';

interface IProps {
  loading?: boolean;
  list: ITargetGroupDetails['target_list'];
  smallPagination?: boolean;
}

const props = withDefaults(defineProps<IProps>(), {
  smallPagination: true,
});

const columnProperties = DisplayFieldFactory.createModel(DisplayFieldType.RS).getProperties();

const { pagination } = usePage(false);
Object.assign(pagination, { small: props.smallPagination, layout: ['total', 'limit', 'list'] });
</script>

<template>
  <data-list
    v-bkloading="{ loading }"
    :columns="columnProperties"
    :list="list"
    :enable-query="false"
    :pagination="pagination"
    :remote-pagination="false"
  />
</template>

<style scoped lang="scss"></style>
