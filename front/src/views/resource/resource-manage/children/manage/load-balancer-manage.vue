<template>
  <Loading :loading="isLoading" :opacity="1">
    <section
      class="flex-row align-items-center"
      :class="isResourcePage ? 'justify-content-end' : 'justify-content-between'"
    >
      <slot></slot>
      <BatchDistribution
        :selections="selections"
        :type="DResourceType.load_balancers"
        :get-data="
          () => {
            triggerApi();
            resetSelections();
          }
        "
      />
      <Button class="mw88" @click="handleClickBatchDelete" :disabled="selections.length === 0">
        {{ t('批量删除') }}
      </Button>
      <div class="flex-row align-items-center justify-content-arround search-selector-container">
        <bk-search-select class="w500" clearable :conditions="[]" :data="clbsSearchData" v-model="searchValue" />
        <slot name="recycleHistory"></slot>
      </div>
    </section>
    <Table
      :columns="renderColumns"
      :data="datas"
      :settings="settings"
      :pagination="pagination"
      remote-pagination
      :row-class="getTableNewRowClass()"
      show-overflow-tooltip
      :is-row-select-enable="isRowSelectEnable"
      @page-limit-change="handlePageSizeChange"
      @page-value-change="handlePageChange"
      @selection-change="(selections: any) => handleSelectionChange(selections, isCurRowSelectEnable)"
      @select-all="(selections: any) => handleSelectionChange(selections, isCurRowSelectEnable, true)"
      @column-sort="handleSort"
      row-key="id"
    />
  </Loading>
  <!-- 批量删除负载均衡 -->
  <BatchOperationDialog
    class="batch-delete-lb-dialog"
    v-model:is-show="isBatchDeleteDialogShow"
    :title="t('批量删除负载均衡')"
    theme="danger"
    confirm-text="删除"
    :is-submit-loading="isSubmitLoading"
    :is-submit-disabled="isSubmitDisabled"
    :table-props="tableProps"
    :list="computedListenersList"
    @handle-confirm="handleBatchDeleteSubmit"
  >
    <template #tips>
      已选择
      <span class="blue">{{ tableProps.data.length }}</span>
      个负载均衡，其中
      <span class="red">{{ tableProps.data.filter(({ listenerNum }) => listenerNum > 0).length }}</span>
      个存在监听器、
      <span class="red">
        <!-- eslint-disable-next-line vue/camelcase -->
        {{ tableProps.data.filter(({ delete_protect }) => delete_protect).length }}
      </span>
      个负载均衡开启了删除保护，不可删除。
    </template>
    <template #tab>
      <BkRadioGroup v-model="radioGroupValue">
        <BkRadioButton :label="true">{{ t('可删除') }}</BkRadioButton>
        <BkRadioButton :label="false">{{ t('不可删除') }}</BkRadioButton>
      </BkRadioGroup>
    </template>
  </BatchOperationDialog>

  <!-- 单个负载均衡分配业务 -->
  <bk-dialog
    :is-show="isDialogShow"
    title="负载均衡分配"
    :theme="'primary'"
    quick-close
    @closed="() => (isDialogShow = false)"
    @confirm="handleSingleDistributionConfirm"
    :is-loading="isDialogBtnLoading"
  >
    <p class="mb16">当前操作负载均衡为：{{ currentOperateItem.name }}</p>
    <p class="mb6">请选择所需分配的目标业务</p>
    <business-selector v-model="selectedBizId" :authed="true" class="mb32" :auto-select="true"></business-selector>
  </bk-dialog>
</template>

<script setup lang="ts">
import { PropType, computed, h, withDirectives, ref } from 'vue';
import { Loading, Table, Button, bkTooltips, Message } from 'bkui-vue';
import { BkRadioButton, BkRadioGroup } from 'bkui-vue/lib/radio';
import { BatchDistribution, DResourceType, DResourceTypeMap } from '../dialog/batch-distribution';
import BatchOperationDialog from '@/components/batch-operation-dialog';
import BusinessSelector from '@/components/business-selector/index.vue';
import Confirm from '@/components/confirm';
import { Senarios, useWhereAmI } from '@/hooks/useWhereAmI';
import type { DoublePlainObject, FilterType } from '@/typings/resource';
import useFilter from '@/views/resource/resource-manage/hooks/use-filter';
import useQueryList from '../../hooks/use-query-list';
import useSelection from '../../hooks/use-selection';
import useColumns from '../../hooks/use-columns';
import useBatchDeleteLB from '@/views/business/load-balancer/clb-view/all-clbs-manager/useBatchDeleteLB';
import { useI18n } from 'vue-i18n';
import { asyncGetListenerCount } from '@/utils';
import { getTableNewRowClass } from '@/common/util';
import { useResourceStore } from '@/store';
const props = defineProps({
  filter: {
    type: Object as PropType<FilterType>,
  },
  isResourcePage: {
    type: Boolean,
  },
  whereAmI: {
    type: String,
  },
});

const { t } = useI18n();
const { whereAmI } = useWhereAmI();
const { searchData, searchValue, filter } = useFilter(props);

const resourceStore = useResourceStore();

const { datas, pagination, isLoading, handlePageChange, handlePageSizeChange, handleSort, triggerApi } = useQueryList(
  { filter: filter.value },
  'load_balancers/with/delete_protection',
  null,
  'list',
  {},
  asyncGetListenerCount,
);
const { selections, handleSelectionChange, resetSelections } = useSelection();
const { columns, settings } = useColumns('lb');
const renderColumns = [
  ...columns,
  {
    label: '操作',
    width: 120,
    render: ({ data }: any) =>
      h('div', { class: 'operation-column' }, [
        withDirectives(
          h(
            Button,
            {
              class: 'mr10',
              text: true,
              theme: 'primary',
              disabled: data.bk_biz_id !== -1,
              onClick: () => handleSingleDistribution(data),
            },
            '分配',
          ),
          [[bkTooltips, { content: t('该负载均衡仅可在业务下操作'), disabled: !(data.bk_biz_id !== -1) }]],
        ),
        withDirectives(
          h(
            Button,
            {
              text: true,
              theme: 'primary',
              disabled: data.bk_biz_id !== -1 || data.listenerNum > 0 || data.delete_protect,
              onClick: () => handleDelete(data),
            },
            '删除',
          ),
          [
            [
              bkTooltips,
              (function () {
                if (data.bk_biz_id !== -1) {
                  return { content: t('该负载均衡仅可在业务下操作'), disabled: !(data.bk_biz_id !== -1) };
                }
                if (data.listenerNum > 0) {
                  return { content: t('该负载均衡已绑定监听器, 不可删除'), disabled: !(data.listenerNum > 0) };
                }
                if (data.delete_protect) {
                  return { content: t('该负载均衡已开启删除保护, 不可删除'), disabled: !data.delete_protect };
                }
              })(),
            ],
          ],
        ),
      ]),
  },
];

const clbsSearchData = computed(() => [
  {
    name: t('负载均衡ID'),
    id: 'cloud_id',
  },
  ...searchData.value,
]);

const isRowSelectEnable = ({ row, isCheckAll }: DoublePlainObject) => {
  if (isCheckAll) return true;
  return isCurRowSelectEnable(row);
};
const isCurRowSelectEnable = (row: any) => {
  if (whereAmI.value === Senarios.business) return true;
  if (row.id) {
    return row.bk_biz_id === -1;
  }
};
// 批量删除负载均衡
const {
  isBatchDeleteDialogShow,
  isSubmitLoading,
  isSubmitDisabled,
  radioGroupValue,
  tableProps,
  handleRemoveSelection,
  handleClickBatchDelete,
  handleBatchDeleteSubmit,
  computedListenersList,
} = useBatchDeleteLB(
  [
    ...columns.slice(1, 8),
    {
      label: '',
      width: 50,
      minWidth: 50,
      render: ({ data }: any) =>
        h(
          Button,
          { text: true, onClick: () => handleRemoveSelection(data.id) },
          h('i', { class: 'hcm-icon bkhcm-icon-minus-circle-shape' }),
        ),
    },
  ],
  selections,
  resetSelections,
  triggerApi,
);

// 删除单个负载均衡
const handleDelete = (data: any) => {
  Confirm('请确定删除负载均衡', `将删除负载均衡【${data.name}】`, async () => {
    await resourceStore.deleteBatch('load_balancers', {
      ids: [data.id],
    });
    Message({ message: '删除成功', theme: 'success' });
    triggerApi();
  });
};

// 分配单个负载均衡
const isDialogShow = ref(false);
const currentOperateItem = ref(null);
const isDialogBtnLoading = ref(false);
const selectedBizId = ref(0);
const handleSingleDistribution = (lb: any) => {
  isDialogShow.value = true;
  currentOperateItem.value = lb;
};
const handleSingleDistributionConfirm = async () => {
  isDialogBtnLoading.value = true;
  try {
    await resourceStore.assignBusiness(DResourceType.load_balancers, {
      [DResourceTypeMap[DResourceType.load_balancers].key]: [currentOperateItem.value.id],
      bk_biz_id: selectedBizId.value,
    });
    Message({ message: t('分配成功'), theme: 'success' });
    triggerApi();
  } finally {
    isDialogShow.value = false;
    isDialogBtnLoading.value = false;
  }
};
</script>

<style lang="scss" scoped>
.mr15 {
  margin-right: 15px;
}

.search-selector-container {
  margin-left: auto;
}
.batch-delete-lb-dialog {
  :deep(.bkhcm-icon-minus-circle-shape) {
    font-size: 14px;
    color: #c4c6cc;
  }
}
</style>
