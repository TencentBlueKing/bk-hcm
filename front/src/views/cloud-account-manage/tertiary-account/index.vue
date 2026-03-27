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

// 创建模型实例
const searchModel = SearchConditionFactory.createModel();
const columnModel = TableColumnFactory.createModel();

// 搜索字段
const searchFields = computed<ModelPropertySearch[]>(() => searchModel.getProperties());

// 表格列
const columns = computed<ModelPropertyColumn[]>(() => columnModel.getProperties());

// 搜索条件
const condition = ref<ISearchCondition>({});

// 全量数据（用于前端分页）
const fullList = ref<ISubAccountItem[]>([]);

// 当前页展示的数据
const tableData = ref<ISubAccountItem[]>([]);

// 排序参数
const sortParams = ref<{ sort: string; order: string }>({ sort: 'created_at', order: 'DESC' });

// 分页
const { pagination, getPageParams } = usePage();

// URL 查询参数处理
const searchQs = useSearchQs({ key: 'filter', properties: searchFields.value });

// 选中的行（用于批量更新）
const selectedRows = ref<ISubAccountItem[]>([]);

// 统计数据
const totalCount = ref(0);
const pendingCount = ref(0);

// 前端分页处理
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

// 构建基础厂商过滤条件（不含搜索条件，用于全量统计）
const buildVendorFilter = (): QueryFilterType => ({
  op: 'and',
  rules: [{ field: 'vendor', op: QueryRuleOPEnum.EQ, value: currentVendor.value }],
});

// 加载统计数据（通过接口请求全量总数和待补充信息账号数）
const loadStatistics = async () => {
  const bizId = getBizsId();
  const vendor = currentVendor.value;
  const vendorFilter = buildVendorFilter();

  // 待补充信息条件：负责人为空 OR 业务为空
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

// 加载全量数据
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
    // 加载统计数据（不依赖搜索条件，独立请求全量统计）
    loadStatistics();
  } catch (error) {
    console.error('获取三级账号列表失败:', error);
    fullList.value = [];
    tableData.value = [];
    pagination.count = 0;
  }
};

// 监听路由变化
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

// 监听云厂商变化
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

// 创建三级账号侧栏
const showCreateSideslider = ref(false);
const handleCreateAccount = () => {
  showCreateSideslider.value = true;
};

// 批量更新侧栏
const showBatchUpdateSideslider = ref(false);
const handleBatchUpdate = () => {
  showBatchUpdateSideslider.value = true;
};

// 详情侧栏
const showDetailSideslider = ref(false);
const currentAccount = ref<ISubAccountItem | null>(null);
const handleViewDetails = (row: ISubAccountItem) => {
  currentAccount.value = row;
  showDetailSideslider.value = true;
};

// 编辑侧栏
const showEditSideslider = ref(false);
const editingAccount = ref<ISubAccountItem | null>(null);
const handleEditAccount = (row: ISubAccountItem) => {
  editingAccount.value = row;
  showEditSideslider.value = true;
};

// 删除弹窗
const showDeleteDialog = ref(false);
const deletingAccount = ref<ISubAccountItem | null>(null);
const handleDeleteAccount = (row: ISubAccountItem) => {
  deletingAccount.value = row;
  showDeleteDialog.value = true;
};

// 选中变化
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
  // 筛选待补充信息的账号（暂时简单刷新列表）
  refreshList();
};
</script>

<template>
  <div class="tertiary-account-page">
    <!-- 搜索区域 -->
    <Search :fields="searchFields" :condition="condition" @search="handleSearch" @reset="handleReset" />

    <!-- 表格区域 -->
    <div class="table-container">
      <!-- 操作按钮区域 + 提示条 -->
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

      <!-- 数据列表 -->
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

    <!-- 创建三级账号侧栏 -->
    <AccountCreateSideslider v-model="showCreateSideslider" @success="handleFormSuccess" />

    <!-- 批量更新侧栏 -->
    <AccountBatchUpdateSideslider
      v-model="showBatchUpdateSideslider"
      :selected-rows="selectedRows"
      @success="handleFormSuccess"
    />

    <!-- 详情侧栏 -->
    <AccountDetailSideslider
      v-model="showDetailSideslider"
      :row-data="currentAccount"
      @update-success="handleFormSuccess"
      @edit="handleEditAccount"
      @delete="handleDeleteAccount"
    />

    <!-- 编辑侧栏 -->
    <AccountEditSideslider v-model="showEditSideslider" :account-data="editingAccount" @success="handleFormSuccess" />

    <!-- 删除确认弹窗 -->
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
