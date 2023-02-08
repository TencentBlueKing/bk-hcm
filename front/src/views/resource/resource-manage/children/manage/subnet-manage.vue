<script setup lang="ts">
import type {
  FilterType,
} from '@/typings/resource';
import {
  PropType,
  ref,
} from 'vue';
import {
  useI18n,
} from 'vue-i18n';
import {
  useResourceStore,
} from '@/store/resource';

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

const isLoadingVpc = ref(false);

// use hooks
const {
  t,
} = useI18n();

const resourceStore = useResourceStore();

const {
  isShowDistribution,
  handleDistribution,
  ResourceDistribution,
} = useSteps();

const columns = useColumns('subnet');

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
  'subnets',
  t('删除子网'),
  true,
);

const {
  datas,
  pagination,
  isLoading,
  handlePageChange,
  handlePageSizeChange,
  handleSort,
} = useQueryList(props, 'subnets');

// 分配 subnet
const handleDistributionSubnet = () => {
  isLoadingVpc.value = true;
  // 获取 vpc 数据
  resourceStore
    .list(
      {
        page: {
          count: false,
          start: 0,
          limit: selections.value.length,
        },
        filter: {
          op: 'and',
          rules: [{
            field: 'id',
            op: 'in',
            value: selections.value.map(selection => selection.vpc_id),
          }],
        },
      },
      'vpcs',
    )
    .then((res: any) => {
      selections.value.forEach((item) => {
        const vpcData = res?.data?.details?.find((vpc: any) => vpc.id === item.vpc_id);
        item.bk_cloud_id = vpcData.bk_cloud_id;
      });
      handleDistribution();
    })
    .finally(() => {
      isLoadingVpc.value = false;
    });
};
</script>

<template>
  <bk-loading
    :loading="isLoading"
  >
    <section>
      <bk-button
        class="w100"
        theme="primary"
        :loading="isLoadingVpc"
        :disabled="selections.length <= 0"
        @click="handleDistributionSubnet"
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
    :data="selections"
    :title="t('子网分配')"
  />

  <delete-dialog>
    {{ t('请注意该子网包含一个或多个资源，在释放这些资源前，无法删除VPC') }}<br />
    {{ t('CVM：{count} 个', { count: 5 }) }}
  </delete-dialog>
</template>

<style lang="scss" scoped>
.w100 {
  width: 100px;
}
.w60 {
  width: 60px;
}
.mt20 {
  margin-top: 20px;
}
</style>
