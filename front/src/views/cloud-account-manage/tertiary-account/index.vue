<script setup lang="ts">
import { ref, computed, watch, inject, type Ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { Plus } from 'bkui-vue/lib/icon';
import usePage from '@/hooks/use-page';
import useSearchQs from '@/hooks/use-search-qs';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import { ModelPropertyColumn, ModelPropertySearch } from '@/model/typings';
import { transformSimpleCondition, localPaginate, localSort } from '@/utils/search';
import { useCloudAccountStore, type ISubAccountItem } from '@/store/cloud-account';
import { VendorEnum } from '@/common/constant';
import { QueryFilterType, QueryRuleOPEnum, RulesItem } from '@/typings';

import Search from './children/search/search.vue';
import DataList from './children/data-list/data-list.vue';
import AccountCreateSideslider from './children/account-create-sideslider/index.vue';
import AccountBatchUpdateSideslider from './children/account-batch-update-sideslider/index.vue';
import AccountDetailSideslider from './children/account-detail-sideslider/index.vue';
import AccountEditSideslider from './children/account-edit-sideslider/index.vue';
import AccountDeleteDialog from './children/account-delete-dialog/index.vue';
import { SearchConditionFactory } from './children/search/condition-factory';
import { TableColumnFactory } from './children/data-list/column-factory';

export type ISearchCondition = Record<string, any>;

const currentVendor = inject<Ref<VendorEnum>>('currentVendor', ref(VendorEnum.TCLOUD));

const route = useRoute();
const router = useRouter();
const cloudAccountStore = useCloudAccountStore();
const { getBizsId } = useWhereAmI();

const searchModel = SearchConditionFactory.createModel();
const columnModel = TableColumnFactory.createModel();

const searchFields = computed<ModelPropertySearch[]>(() => searchModel.getProperties());
const columns = computed<ModelPropertyColumn[]>(() => columnModel.getProperties());
const condition = ref<ISearchCondition>({});
const fullList = ref<ISubAccountItem[]>([]);
const tableData = ref<ISubAccountItem[]>([]);
const sortParams = ref<{ sort: string; order: string }>({ sort: 'created_at', order: 'DESC' });
const { pagination, getPageParams } = usePage();
const searchQs = useSearchQs({ key: 'filter', properties: searchFields.value });
const selectedRows = ref<ISubAccountItem[]>([]);
const totalCount = ref(0);
const pendingCount = ref(0);

const updateTableData = () => {
  let list = [...fullList.value];
  if (sortParams.value.sort) {
    list = localSort(list, {
      column: { field: sortParams.value.sort },
      type: sortParams.value.order,
    });
  }
  const pageParams = getPageParams(pagination, sortParams.value);
  tableData.value = localPaginate(list, pageParams);
};

const buildVendorFilter = (): QueryFilterType => ({
  op: 'and',
  rules: [{ field: 'vendor', op: QueryRuleOPEnum.EQ, value: currentVendor.value }],
});

const loadStatistics = async () => {
  const bizId = getBizsId();
  const vendor = currentVendor.value;
  const vendorFilter = buildVendorFilter();

  const pendingFilter: QueryFilterType = {
    op: 'and',
    rules: [
      { field: 'vendor', op: QueryRuleOPEnum.EQ, value: vendor },
      {
        op: QueryRuleOPEnum.OR,
        rules: [
          { field: 'managers', op: QueryRuleOPEnum.JSON_EQ, value: '[]' as any },
          { field: 'bk_biz_ids', op: QueryRuleOPEnum.JSON_EQ, value: '[]' as any },
        ],
      },
    ],
  };

  const [total, pending] = await Promise.all([
    cloudAccountStore.getSubAccountCount(bizId, vendor, vendorFilter),
    cloudAccountStore.getSubAccountCount(bizId, vendor, pendingFilter),
  ]);

  totalCount.value = total;
  pendingCount.value = pending;
};

const loadFullList = async () => {
  try {
    condition.value = searchQs.get(route.query, {});
    const baseFilter = transformSimpleCondition(condition.value, searchFields.value);
    const vendorFilter: QueryFilterType = {
      op: 'and',
      rules: [
        ...((baseFilter?.rules || []) as RulesItem[]),
        {
          field: 'vendor',
          op: 'eq' as any,
          value: currentVendor.value,
        },
      ],
    };

    const list = await cloudAccountStore.getSubAccountFullList(
      getBizsId(),
      currentVendor.value,
      vendorFilter,
      (progressList, count) => {
        fullList.value = progressList;
        pagination.count = count;
        updateTableData();
      },
    );

    fullList.value = list;
    pagination.count = list.length;
    updateTableData();
    loadStatistics();
  } catch (error) {
    console.error('获取三级账号列表失败:', error);
    fullList.value = [];
    tableData.value = [];
    pagination.count = 0;
  }
};

watch(
  () => route.query,
  async (query) => {
    pagination.current = Number(query.page) || 1;
    pagination.limit = Number(query.limit) || pagination.limit;
    sortParams.value = {
      sort: (query.sort || 'created_at') as string,
      order: (query.order || 'DESC') as string,
    };
    const newCondition = searchQs.get(query, {});
    const conditionChanged = JSON.stringify(newCondition) !== JSON.stringify(condition.value);
    const isRefresh = query._t !== undefined;
    if (conditionChanged || fullList.value.length === 0 || isRefresh) {
      await loadFullList();
    } else {
      updateTableData();
    }
  },
  { immediate: true },
);

watch(
  () => currentVendor.value,
  () => {
    pagination.current = 1;
    fullList.value = [];
    const query = { ...route.query };
    delete query.page;
    query._t = String(Date.now());
    router.replace({ query });
  },
);

const isLoading = computed(() => cloudAccountStore.subAccountListLoading);

const showCreateSideslider = ref(false);
const handleCreateAccount = () => {
  showCreateSideslider.value = true;
};

const showBatchUpdateSideslider = ref(false);
const handleBatchUpdate = () => {
  showBatchUpdateSideslider.value = true;
};

const showDetailSideslider = ref(false);
const currentAccount = ref<ISubAccountItem | null>(null);
const handleViewDetails = (row: ISubAccountItem) => {
  currentAccount.value = row;
  showDetailSideslider.value = true;
};

// 监听 URL 中 detailCloudId 参数，自动打开对应账号详情弹窗
watch(
  () => route.query.detailCloudId,
  (detailCloudId) => {
    if (!detailCloudId || typeof detailCloudId !== 'string') return;
    const tryOpenDetail = () => {
      const target = fullList.value.find((item) => item.cloud_id === detailCloudId || item.id === detailCloudId);
      if (target) {
        handleViewDetails(target);
        const query = { ...route.query };
        delete query.detailCloudId;
        router.replace({ query });
      }
    };
    if (fullList.value.length > 0) {
      tryOpenDetail();
    } else {
      const unwatch = watch(
        () => fullList.value.length,
        (len) => {
          if (len > 0) {
            tryOpenDetail();
            unwatch();
          }
        },
      );
    }
  },
  { immediate: true },
);

const showEditSideslider = ref(false);
const editingAccount = ref<ISubAccountItem | null>(null);
const handleEditAccount = (row: ISubAccountItem) => {
  editingAccount.value = row;
  showEditSideslider.value = true;
};

const showDeleteDialog = ref(false);
const deletingAccount = ref<ISubAccountItem | null>(null);
const handleDeleteAccount = (row: ISubAccountItem) => {
  deletingAccount.value = row;
  showDeleteDialog.value = true;
};

const handleSelectionChange = (selection: ISubAccountItem[]) => {
  selectedRows.value = selection;
};

const refreshList = () => {
  const query = { ...route.query };
  query._t = String(Date.now());
  router.replace({ query });
};

const handleSearch = (searchCondition: ISearchCondition) => {
  searchQs.set(searchCondition);
};

const handleReset = () => {
  searchQs.clear();
};

const handleFormSuccess = () => {
  refreshList();
  loadStatistics();
};

const handleGoToPending = () => {
  refreshList();
};
</script>

<template>
  <div class="tertiary-account-page">
    <Search :fields="searchFields" :condition="condition" @search="handleSearch" @reset="handleReset" />

    <div class="table-container">
      <div class="tertiary-action-bar">
        <div class="action-btns">
          <bk-button theme="primary" @click="handleCreateAccount">
            <plus style="font-size: 22px" />
            创建账号
          </bk-button>
          <bk-button :disabled="selectedRows.length === 0" @click="handleBatchUpdate">批量更新</bk-button>
        </div>
        <bk-alert v-if="pendingCount > 0" theme="warning" class="info-alert">
          <template #title>
            当前有
            <strong>{{ totalCount }}</strong>
            个账号，其中待补充信息账号有
            <strong>{{ pendingCount }}</strong>
            个
            <bk-button text theme="primary" style="margin-left: 8px" @click="handleGoToPending">去处理</bk-button>
          </template>
        </bk-alert>
      </div>

      <DataList
        :columns="columns"
        :list="tableData"
        :pagination="pagination"
        :loading="isLoading"
        @view-details="handleViewDetails"
        @edit-account="handleEditAccount"
        @delete-account="handleDeleteAccount"
        @selection-change="handleSelectionChange"
      />
    </div>

    <AccountCreateSideslider v-model="showCreateSideslider" @success="handleFormSuccess" />

    <AccountBatchUpdateSideslider
      v-model="showBatchUpdateSideslider"
      :selected-rows="selectedRows"
      @success="handleFormSuccess"
    />

    <AccountDetailSideslider
      v-model="showDetailSideslider"
      :row-data="currentAccount"
      @update-success="handleFormSuccess"
      @edit="handleEditAccount"
      @delete="handleDeleteAccount"
    />

    <AccountEditSideslider v-model="showEditSideslider" :account-data="editingAccount" @success="handleFormSuccess" />

    <AccountDeleteDialog v-model="showDeleteDialog" :account-data="deletingAccount" @success="handleFormSuccess" />
  </div>
</template>

<style lang="scss" scoped>
.tertiary-account-page {
  height: 100%;

  .table-container {
    background: #fff;
    border-radius: 2px;
    margin: 24px;
    padding: 16px 24px;
  }

  .tertiary-action-bar {
    display: flex;
    align-items: center;
    gap: 16px;
    margin-bottom: 16px;

    .action-btns {
      display: flex;
      align-items: center;
      gap: 8px;
      flex-shrink: 0;
    }

    .info-alert {
      flex: 1;
    }
  }
}
</style>
