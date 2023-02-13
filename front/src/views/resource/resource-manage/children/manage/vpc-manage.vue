<script setup lang="ts">
import type {
  FilterType,
} from '@/typings/resource';

import {
  PropType,
} from 'vue';
import {
  useI18n,
} from 'vue-i18n';
import useSteps from '../../hooks/use-steps';
import useColumns from '../../hooks/use-columns';
import useDelete from '../../hooks/use-delete';
import useQueryList from '../../hooks/use-query-list';
import useSelection from '../../hooks/use-selection';

const props = defineProps({
  filter: {
    type: Object as PropType<FilterType>,
  },
});

// use hooks
const {
  t,
} = useI18n();

const {
  isShowDistribution,
  handleDistribution,
  ResourceDistribution,
} = useSteps();

const columns = useColumns('vpc');

const {
  selections,
  handleSelectionChange,
} = useSelection();

const {
  handleShowDelete,
  DeleteDialog,
} = useDelete(
  columns,
  selections,
  'vpcs',
  t('删除 VPC'),
  true,
);

const {
  datas,
  pagination,
  isLoading,
  handlePageChange,
  handlePageSizeChange,
  handleSort,
} = useQueryList(props, 'vpcs');
</script>

<template>
  <bk-loading
    :loading="isLoading"
  >
    <section>
      <bk-button
        class="w100"
        theme="primary"
        :disabled="selections.length <= 0"
        @click="handleDistribution"
      >
        {{ t('分配') }}
      </bk-button>
      <bk-button
        class="w100 ml10"
        theme="primary"
        :disabled="selections.length <= 0"
        @click="handleShowDelete"
      >
        {{ t('删除') }}
      </bk-button>
    </section>

    <bk-table
      class="mt20"
      row-hover="auto"
      remote-pagination
      :pagination="pagination"
      :columns="columns"
      :data="datas"
      @page-limit-change="handlePageSizeChange"
      @page-value-change="handlePageChange"
      @column-sort="handleSort"
      @selection-change="handleSelectionChange"
    />
  </bk-loading>

  <resource-distribution
    v-model:is-show="isShowDistribution"
    :title="t('VPC 分配')"
    :data="selections"
  />

  <delete-dialog>
    {{ t('请注意该VPC包含一个或多个资源，在释放这些资源前，无法删除VPC') }}<br />
    {{ t('子网：{count} 个', { count: 5 }) }}<br />
    {{ t('CVM：{count} 个', { count: 5 }) }}
  </delete-dialog>
</template>

<style lang="scss" scoped>
.w100 {
  width: 100px;
}
</style>
