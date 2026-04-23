<script setup lang="ts">
import { ref, computed, watch, inject, type Ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { Message } from 'bkui-vue';
import { Plus } from 'bkui-vue/lib/icon';
import usePage from '@/hooks/use-page';
import useSearchQs from '@/hooks/use-search-qs';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import { ModelPropertyColumn, ModelPropertySearch } from '@/model/typings';
import { transformSimpleCondition, localPaginate, localSort } from '@/utils/search';
import { useTertiaryAccountStore, type ISubAccountItem } from '@/store/cloud-account-manage/tertiary-account';
import { VendorEnum } from '@/common/constant';
import { QueryFilterType, RulesItem } from '@/typings';

import Search from './children/search/search.vue';
import DataList from './children/data-list/data-list.vue';
import AccountCreateSideslider from './children/account-create-sideslider/index.vue';
import AccountBatchUpdateSideslider from './children/account-batch-update-sideslider/index.vue';
import AccountDetailSideslider from './children/account-detail-sideslider/index.vue';
import AccountEditSideslider from './children/account-edit-sideslider/index.vue';
import AccountDeleteDialog from './children/account-delete-dialog/index.vue';
import { SearchConditionFactory } from './children/search/condition-factory';
import { TableColumnFactory } from './children/data-list/column-factory';
import { AUTH_BIZ_CREATE_SUB_ACCOUNT, AUTH_BIZ_UPDATE_SUB_ACCOUNT } from '@/constants/auth-symbols';

export type ISearchCondition = Record<string, any>;

const currentVendor = inject<Ref<VendorEnum>>('currentVendor', ref(VendorEnum.TCLOUD));

const route = useRoute();
const router = useRouter();
const tertiaryAccountStore = useTertiaryAccountStore();
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
const totalCount = computed(() => fullList.value.length);
const isPendingItem = (item: ISubAccountItem) => {
  if (item.operable === false) return false;
  const managers = item.managers ?? item.extension?.managers;
  const bizIds = item.bk_biz_ids ?? item.extension?.bk_biz_ids;
  const emptyManagers = !managers || (Array.isArray(managers) && managers.length === 0);
  const emptyBizIds = !bizIds || (Array.isArray(bizIds) && bizIds.length === 0);
  return emptyManagers || emptyBizIds;
};
const pendingList = computed(() => fullList.value.filter(isPendingItem));
const pendingCount = computed(() => pendingList.value.length);

const enrichList = (list: ISubAccountItem[]) =>
  list.map((item) => ({
    ...item,
    permission_template_count: item.permission_templates?.length ?? 0, // 额外字段用于排序
  }));

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

const showDetailSideslider = ref(false);
const currentAccount = ref<ISubAccountItem | null>(null);
const handleViewDetails = (row: ISubAccountItem) => {
  currentAccount.value = row;
  showDetailSideslider.value = true;
  // 将 id 写入 URL query，支持分享/刷新/浏览器后退
  router.replace({ query: { ...route.query, id: row.id, _t: undefined } });
};

// 弹窗被用户手动关闭（点击 X / quick-close）时，同步移除 URL 中的 id
watch(showDetailSideslider, (val) => {
  if (!val && route.query.id) {
    const query = { ...route.query };
    delete query.id;
    router.replace({ query });
  }
});

// 监听 route.query.id，使用详情接口加载并打开弹窗
watch(
  () => route.query.id,
  async (id) => {
    if (!id) {
      showDetailSideslider.value = false;
      currentAccount.value = null;
      return;
    }
    try {
      const detail = await tertiaryAccountStore.getSubAccountDetail(getBizsId(), currentVendor.value, id as string);
      if (detail) {
        currentAccount.value = detail;
        showDetailSideslider.value = true;
      } else {
        Message({ theme: 'warning', message: `未找到账号数据` });
        router.replace({ query: { ...route.query, id: undefined } });
      }
    } catch (error) {
      console.error('获取三级账号详情失败:', error);
      Message({ theme: 'error', message: '获取三级账号详情失败' });
      router.replace({ query: { ...route.query, id: undefined } });
    }
  },
  { immediate: true },
);

const loadFullList = async () => {
  try {
    // 从 URL 获取搜索条件
    condition.value = searchQs.get(route.query, {});
    urlCondition.value = { ...condition.value };

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

    const list = await tertiaryAccountStore.getSubAccountFullList(
      getBizsId(),
      currentVendor.value,
      vendorFilter,
      (progressList, count) => {
        fullList.value = enrichList(progressList);
        pagination.count = count;
        updateTableData();
      },
    );

    fullList.value = enrichList(list);
    pagination.count = list.length;
    updateTableData();
  } catch (error) {
    console.error('获取三级账号列表失败:', error);
    fullList.value = [];
    tableData.value = [];
    pagination.count = 0;
  }
};

// 记录从 URL 解析出的纯搜索条件，用于 conditionChanged 判断
const urlCondition = ref<ISearchCondition>({});

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
    const conditionChanged = JSON.stringify(newCondition) !== JSON.stringify(urlCondition.value);
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

const isLoading = computed(() => tertiaryAccountStore.subAccountListLoading);

const showCreateSideslider = ref(false);
const handleCreateAccount = () => {
  showCreateSideslider.value = true;
};

const showBatchUpdateSideslider = ref(false);
const batchUpdateRows = ref<ISubAccountItem[]>([]);
const handleBatchUpdate = () => {
  batchUpdateRows.value = [...selectedRows.value];
  showBatchUpdateSideslider.value = true;
};

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
};

const handleGoToPending = () => {
  batchUpdateRows.value = [...pendingList.value];
  showBatchUpdateSideslider.value = true;
};
</script>

<template>
  <div class="tertiary-account-page">
    <Search :fields="searchFields" :condition="condition" @search="handleSearch" @reset="handleReset" />

    <div class="table-container">
      <div class="tertiary-action-bar">
        <div class="action-btns">
          <hcm-auth
            v-if="getBizsId()"
            :sign="{ type: AUTH_BIZ_CREATE_SUB_ACCOUNT, relation: [getBizsId()] }"
            v-slot="{ noPerm }"
          >
            <bk-button theme="primary" :disabled="noPerm" @click="handleCreateAccount">
              <plus style="font-size: 22px" />
              创建账号
            </bk-button>
          </hcm-auth>

          <hcm-auth
            v-if="getBizsId()"
            :sign="{ type: AUTH_BIZ_UPDATE_SUB_ACCOUNT, relation: [getBizsId()] }"
            v-slot="{ noPerm }"
          >
            <bk-button :disabled="noPerm || selectedRows.length === 0" @click="handleBatchUpdate">批量更新</bk-button>
          </hcm-auth>
        </div>
        <bk-alert v-if="pendingCount > 0" theme="warning" class="info-alert">
          <template #title>
            当前有
            <strong>{{ totalCount }}</strong>
            个账号，其中待补充信息账号有
            <strong>{{ pendingCount }}</strong>
            个
            <hcm-auth
              v-if="getBizsId()"
              :sign="{ type: AUTH_BIZ_UPDATE_SUB_ACCOUNT, relation: [getBizsId()] }"
              v-slot="{ noPerm }"
            >
              <bk-button :disabled="noPerm" text theme="primary" style="margin-left: 8px" @click="handleGoToPending">
                去处理
              </bk-button>
            </hcm-auth>
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
      :selected-rows="batchUpdateRows"
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
