<script setup lang="ts">
import type {
  FilterType,
} from '@/typings/resource';

import {
  PropType,
  ref,
  defineExpose,
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

const isLoadingSubnets = ref(false);
const chooseVpcSubnetsNum = ref(0);
const chooseVpcCvmsNum = ref(0);
const chooseVpcEipsNum = ref(0);

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

// 抛出请求数据的方法，新增成功使用
const fetchComponentsData = () => {
  handlePageChange(1);
};
defineExpose({ fetchComponentsData });

const handleDeleteVpc = (vpcList: any) => {
  const vpcIds = vpcList.map((vpc: any) => vpc.id);
  isLoadingSubnets.value = true;
  const getRelateNum = (type: string) => {
    return resourceStore
      .list(
        {
          page: {
            count: true,
          },
          filter: {
            op: 'and',
            rules: [{
              field: 'vpc_id',
              op: 'in',
              value: vpcIds,
            }],
          },
        },
        type,
      )
  }
  Promise
    .all([
      getRelateNum('cvms'),
      getRelateNum('eips'),
      getRelateNum('subnets')
    ])
    .then(([cvmsResult, eipsResult, subnetsResult]: any) => {
      chooseVpcCvmsNum.value = cvmsResult?.data?.count || 0;
      chooseVpcEipsNum.value = eipsResult?.data?.count || 0;
      chooseVpcSubnetsNum.value = subnetsResult?.data?.count || 0;
      handleShowDelete(vpcIds);
    })
    .finally(() => {
      isLoadingSubnets.value = false;
    });
};
</script>

<template>
  <bk-loading
    :loading="isLoading"
  >
    <section>
      <slot>
        <bk-button
          class="w100"
          theme="primary"
          :disabled="selections.length <= 0"
          @click="handleDistribution"
        >
          {{ t('分配') }}
        </bk-button>
      </slot>
      <bk-button
        class="w100 ml10"
        theme="primary"
        :disabled="selections.length <= 0"
        :loading="isLoadingSubnets"
        @click="handleDeleteVpc(selections)"
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
    <template v-if="chooseVpcCvmsNum || chooseVpcEipsNum || chooseVpcSubnetsNum">
      {{ t('请注意该VPC包含一个或多个资源，在释放这些资源前，无法删除VPC') }}<br />
      {{ `子网${chooseVpcSubnetsNum}个` }}<br />
      {{ `弹性IP${chooseVpcEipsNum}个` }}<br />
      {{ `主机${chooseVpcCvmsNum}个` }}<br />
    </template>
  </delete-dialog>
</template>

<style lang="scss" scoped>
.w100 {
  width: 100px;
}
</style>
