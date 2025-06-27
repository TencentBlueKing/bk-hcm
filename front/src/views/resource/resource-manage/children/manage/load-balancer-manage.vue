<template>
  <Loading :loading="isLoading" :opacity="1">
    <section class="toolbar" :class="isResourcePage ? 'justify-content-end' : 'justify-content-between'">
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
      <bk-button class="mw88" @click="handleClickBatchDelete" :disabled="selections.length === 0">
        {{ t('批量删除') }}
      </bk-button>
      <bk-button
        :disabled="selections.length > 0"
        @click="() => handleSync(false, resourceAccountStore.resourceAccount)"
      >
        {{ t('同步负载均衡') }}
      </bk-button>
      <div class="flex-row align-items-center justify-content-arround search-selector-container">
        <bk-search-select
          class="w500"
          clearable
          :conditions="[]"
          :data="clbsSearchData"
          v-model="searchValue"
          value-behavior="need-key"
        />
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

  <template v-if="!syncDialogState.isHidden">
    <sync-account-resource
      v-model="syncDialogState.isShow"
      title="同步负载均衡"
      desc="从云上同步负载均衡数据，包括负载均衡，监听器等"
      :resource-type="ResourceTypeEnum.CLB"
      resource-name="load_balancer"
      :initial-model="syncDialogState.initialModel"
      @hidden="
        () => {
          syncDialogState.isHidden = true;
          syncDialogState.initialModel = null;
        }
      "
    />
  </template>
</template>

<script setup lang="ts">
import { PropType, h, withDirectives, ref, reactive } from 'vue';
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
import { asyncGetListenerCount, buildVIPFilterRules } from '@/utils';
import { getTableNewRowClass } from '@/common/util';
import { useResourceStore, useBusinessStore } from '@/store';
import { useResourceAccountStore } from '@/store/useResourceAccountStore';
import { ResourceTypeEnum, VendorEnum, VendorMap } from '@/common/constant';
import SyncAccountResource from '@/components/sync-account-resource/index.vue';
import { CLB_STATUS_MAP, LB_NETWORK_TYPE_MAP } from '@/constants';

const props = defineProps({
  filter: {
    type: Object as PropType<FilterType>,
  },
  isResourcePage: {
    type: Boolean,
  },
});

const { t } = useI18n();
const { whereAmI } = useWhereAmI();
const { searchValue, filter } = useFilter(props, {
  conditionFormatterMapper: {
    lb_vip: (value: string) => buildVIPFilterRules(value),
  },
});

const resourceStore = useResourceStore();
const businessStore = useBusinessStore();
const resourceAccountStore = useResourceAccountStore();

const { datas, pagination, isLoading, handlePageChange, handlePageSizeChange, handleSort, triggerApi } = useQueryList(
  { filter: filter.value },
  'load_balancers/with/delete_protection',
  null,
  'list',
  {},
  (dataList: any) => asyncGetListenerCount(businessStore.asyncGetListenerCount, dataList),
);
const { selections, handleSelectionChange, resetSelections } = useSelection();
const { columns, settings } = useColumns('lb');
const renderColumns = [
  ...columns,
  {
    label: '操作',
    width: 150,
    fixed: 'right',
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
              class: 'mr10',
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
                return { disabled: true };
              })(),
            ],
          ],
        ),
        h(
          Button,
          {
            text: true,
            theme: 'primary',
            disabled: data.vendor !== VendorEnum.TCLOUD,
            onClick: () => handleSync(true, data),
          },
          '同步',
        ),
      ]),
  },
];

const clbsSearchData = [
  { id: 'name', name: '负载均衡名称' },
  { id: 'cloud_id', name: '负载均衡ID' },
  { id: 'domain', name: '负载均衡域名' },
  { id: 'lb_vip', name: '负载均衡VIP' },
  {
    id: 'lb_type',
    name: '网络类型',
    children: Object.keys(LB_NETWORK_TYPE_MAP).map((lbType) => ({
      id: lbType,
      name: LB_NETWORK_TYPE_MAP[lbType as keyof typeof LB_NETWORK_TYPE_MAP],
    })),
  },
  {
    id: 'ip_version',
    name: t('IP版本'),
    children: [
      { id: 'ipv4', name: 'IPv4' },
      { id: 'ipv6', name: 'IPv6' },
      { id: 'ipv6_dual_stack', name: 'IPv6DualStack' },
      { id: 'ipv6_nat64', name: 'IPv6Nat64' },
    ],
  },
  {
    id: 'vendor',
    name: t('云厂商'),
    children: [{ id: VendorEnum.TCLOUD, name: VendorMap[VendorEnum.TCLOUD] }],
  },
  { id: 'zones', name: '可用区域' },
  {
    id: 'status',
    name: '状态',
    children: Object.keys(CLB_STATUS_MAP).map((key) => ({ id: key, name: CLB_STATUS_MAP[key] })),
  },
  { id: 'cloud_vpc_id', name: '所属VPC' },
];

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

const syncDialogState = reactive({ isShow: false, isHidden: true, initialModel: null });
const handleSync = (inTable: boolean, data?: any) => {
  syncDialogState.isShow = true;
  syncDialogState.isHidden = false;
  if (inTable) {
    const { account_id: accountId, vendor, region, cloud_id: cloudId } = data;
    // TODO: azure支持负载均衡后，需要补充resource_group_names
    syncDialogState.initialModel = { account_id: accountId, vendor, regions: region, cloud_ids: [cloudId] };
  } else {
    const { id, vendor } = data;
    syncDialogState.initialModel = { account_id: id, vendor };
  }
};
</script>

<style lang="scss" scoped>
.toolbar {
  display: flex;
  align-items: center;
  gap: 10px;
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
