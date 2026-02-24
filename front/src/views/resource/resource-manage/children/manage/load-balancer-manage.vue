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
      <bk-button :disabled="selections.length > 0" @click="() => handleSync(false, currentAccountForSync)">
        {{ t('同步负载均衡') }}
      </bk-button>
      <div class="flex-row align-items-center justify-content-arround search-selector-container">
        <resource-search-select
          v-model="searchValue"
          :resource-type="ResourceTypeEnum.CLB"
          @change="(condition) => searchQs.set(condition)"
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

  <template v-if="!batchDeleteDialogState.isHidden">
    <batch-delete-dialog
      v-model="batchDeleteDialogState.isShow"
      :selections="selections"
      @confirm-success="handleBatchDeleteSuccess"
      @hidden="batchDeleteDialogState.isHidden = true"
    />
  </template>

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
    <hcm-form-business :data="accountBizList" v-model="selectedBizId" />
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
import { h, withDirectives, ref, reactive, computed } from 'vue';
import { Loading, Table, Button, bkTooltips, Message } from 'bkui-vue';
import { BatchDistribution, DResourceType, DResourceTypeMap } from '../dialog/batch-distribution';
import BatchDeleteDialog from '@/views/load-balancer/clb/children/batch-delete-dialog.vue';
import Confirm from '@/components/confirm';
import { Senarios, useWhereAmI } from '@/hooks/useWhereAmI';
import type { DoublePlainObject } from '@/typings/resource';
import useFilterFromRoute from '@/views/resource-manage/hooks/use-filter-from-route';
import ResourceSearchSelect from '@/components/resource-search-select/index.vue';
import useQueryList from '../../hooks/use-query-list';
import useSelection from '../../hooks/use-selection';
import useColumns from '../../hooks/use-columns';
import { useI18n } from 'vue-i18n';
import { getTableNewRowClass } from '@/common/util';
import { useResourceStore } from '@/store';
import { useAccountSelectorStore } from '@/store/account-selector';
import { ResourceTypeEnum, VendorEnum } from '@/common/constant';
import SyncAccountResource from '@/components/sync-account-resource/index.vue';
import { useAccountBusiness } from '@/views/resource/resource-manage/hooks/use-account-business';
import { useRoute } from 'vue-router';
import { ILoadBalancerWithDeleteProtectionItem, useLoadBalancerClbStore } from '@/store/load-balancer/clb';

defineProps({
  isResourcePage: {
    type: Boolean,
  },
});

const { t } = useI18n();
const route = useRoute();
// eslint-disable-next-line vue/no-dupe-keys
const { whereAmI } = useWhereAmI();

const { searchValue, filter, searchQs } = useFilterFromRoute(ResourceTypeEnum.CLB);

const resourceStore = useResourceStore();
const accountSelectorStore = useAccountSelectorStore();
const currentAccountForSync = computed(() => {
  const accountId = route.query.accountId as string;
  if (accountId) {
    return accountSelectorStore.authorizedResourceAccountList.find((a: { id: string }) => a.id === accountId) || null;
  }
  return null;
});
const loadBalancerClbStore = useLoadBalancerClbStore();

const { datas, pagination, isLoading, handlePageChange, handlePageSizeChange, handleSort, triggerApi } = useQueryList(
  { filter: filter.value },
  'load_balancers/with/delete_protection',
  null,
  'list',
  {},
  (dataList: any) => asyncQueryListenerCount(dataList),
);
const asyncQueryListenerCount = async (list: ILoadBalancerWithDeleteProtectionItem[]) => {
  if (!list || list.length === 0) return;
  const ids = list.map((item) => item.id);
  const listenerCountDetails = await loadBalancerClbStore.getListenerCountByLoadBalancerIds(ids);
  return list.map((lb) => {
    const listenerCountDetail = listenerCountDetails.find((item) => item.lb_id === lb.id);
    if (listenerCountDetail) {
      lb.listener_count = listenerCountDetail.num;
    } else {
      lb.listener_count = 0;
    }
    return lb;
  });
};

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
              disabled: data.bk_biz_id !== -1 || data.listener_count > 0 || data.delete_protect,
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
                if (data.listener_count > 0) {
                  return { content: t('该负载均衡已绑定监听器, 不可删除'), disabled: !(data.listener_count > 0) };
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
const batchDeleteDialogState = reactive({ isShow: false, isHidden: true });
const handleClickBatchDelete = () => {
  batchDeleteDialogState.isShow = true;
  batchDeleteDialogState.isHidden = false;
};
const handleBatchDeleteSuccess = () => {
  triggerApi();
};

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

const accountId = computed(() => currentOperateItem.value?.account_id);

const { accountBizList } = useAccountBusiness(accountId);

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
    const { name, account_id: accountId, vendor, region, cloud_id: cloudId } = data;
    // TODO: azure支持负载均衡后，需要补充resource_group_names
    syncDialogState.initialModel = { name, account_id: accountId, vendor, regions: region, cloud_ids: [cloudId] };
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
</style>
