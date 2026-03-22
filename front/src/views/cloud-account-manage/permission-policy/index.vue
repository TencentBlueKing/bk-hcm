<script setup lang="ts">
import { ref, computed, watch, inject, type Ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import usePage from '@/hooks/use-page';
import useSearchQs from '@/hooks/use-search-qs';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import { ModelPropertyColumn, ModelPropertySearch } from '@/model/typings';
import { transformSimpleCondition, localPaginate, localSort } from '@/utils/search';
import { VendorEnum } from '@/common/constant';
import { QueryFilterType, RulesItem } from '@/typings';

import Search from './children/search/search.vue';
import DataList from './children/data-list/data-list.vue';
import PolicyFormSideslider from './children/policy-form-sideslider/index.vue';
import PolicyInfoSideslider from './children/policy-form-sideslider/info.vue';
import ApplySideslider from './children/apply-sideslider/index.vue';
import LogSideslider from './children/log-sideslider/index.vue';
import { SearchConditionFactory } from './children/search/condition-factory';
import { TableColumnFactory } from './children/data-list/column-factory';
import type { IPermissionPolicyItem } from './typings';
import { ENABLE_MOCK, MOCK_PERMISSION_POLICY_LIST } from './constants';

export type ISearchCondition = Record<string, any>;

const currentVendor = inject<Ref<VendorEnum>>('currentVendor', ref(VendorEnum.TCLOUD));

const route = useRoute();
const router = useRouter();
const { getBizsId: _getBizsId } = useWhereAmI();

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
const fullList = ref<IPermissionPolicyItem[]>([]);

// 当前页展示的数据
const tableData = ref<IPermissionPolicyItem[]>([]);

// 排序参数
const sortParams = ref<{ sort: string; order: string }>({ sort: 'created_at', order: 'DESC' });

// 分页
const { pagination, getPageParams } = usePage();

// URL 查询参数处理
const searchQs = useSearchQs({ key: 'filter', properties: searchFields.value });

// 前端分页处理：根据全量数据计算当前页数据
const updateTableData = () => {
  let list = [...fullList.value];

  // 前端排序
  if (sortParams.value.sort) {
    list = localSort(list, {
      column: { field: sortParams.value.sort },
      type: sortParams.value.order,
    });
  }

  // 前端分页
  const pageParams = getPageParams(pagination, sortParams.value);
  tableData.value = localPaginate(list, pageParams);
};

// 加载全量数据
const loadFullList = async () => {
  try {
    // 从 URL 获取搜索条件
    condition.value = searchQs.get(route.query, {});

    // 构建 filter，加入云厂商 vendor 条件
    const baseFilter = transformSimpleCondition(condition.value, searchFields.value);
    const _vendorFilter: QueryFilterType = {
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

    if (ENABLE_MOCK) {
      // Mock 模式：使用本地模拟数据，并根据搜索条件前端过滤
      let mockData = MOCK_PERMISSION_POLICY_LIST.filter((item) => item.vendor === currentVendor.value);

      // 前端搜索过滤
      const cond = condition.value;
      if (cond.name) {
        mockData = mockData.filter((item) => item.name.toLowerCase().includes(cond.name.toLowerCase()));
      }
      if (cond.description) {
        mockData = mockData.filter((item) => item.description.includes(cond.description));
      }
      if (cond.creator) {
        mockData = mockData.filter((item) => item.creator === cond.creator);
      }
      if (cond.reviser) {
        mockData = mockData.filter((item) => item.reviser === cond.reviser);
      }

      fullList.value = mockData;
      pagination.count = mockData.length;
      updateTableData();
      return;
    }

    // TODO: 替换为真实API调用
    // const list = await permissionPolicyStore.getPermissionPolicyFullList(getBizsId(), vendorFilter);
    fullList.value = [];
    pagination.count = 0;
    updateTableData();
  } catch (error) {
    console.error('获取权限策略库列表失败:', error);
    fullList.value = [];
    tableData.value = [];
    pagination.count = 0;
  }
};

// 监听路由变化，获取列表数据
watch(
  () => route.query,
  async (query) => {
    // 设置分页
    pagination.current = Number(query.page) || 1;
    pagination.limit = Number(query.limit) || pagination.limit;

    // 排序参数
    sortParams.value = {
      sort: (query.sort || 'created_at') as string,
      order: (query.order || 'DESC') as string,
    };

    // 判断是否只是分页/排序变化（不需要重新拉取全量数据）
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

// 监听云厂商变化，刷新列表
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

// 加载状态
const isLoading = ref(false);

const showApplySideslider = ref(false);
const showPolicyInfoSideslider = ref(false);
const currentApplyPolicy = ref<IPermissionPolicyItem | null>(null);

const handleApplyToAccount = (row: IPermissionPolicyItem) => {
  currentApplyPolicy.value = row;
  showApplySideslider.value = true;
};

// 查看详情
const handleViewDetails = (row: IPermissionPolicyItem) => {
  currentApplyPolicy.value = row;
  showPolicyInfoSideslider.value = true;
};

// 应用成功回调
const handleApplySuccess = () => {
  refreshList();
  showLogSideslider.value = true;
};

// 应用成功查看日志弹窗
const showLogSideslider = ref(false);

const refreshList = () => {
  const query = { ...route.query };
  query._t = String(Date.now());
  router.replace({ query });
};

// 新建/编辑权限策略库状态
const showPolicyFormSideslider = ref(false);
const isEditMode = ref(false);
const editingAccount = ref<IPermissionPolicyItem | null>(null);

const handleAddPolicy = () => {
  isEditMode.value = false;
  editingAccount.value = null;
  showPolicyFormSideslider.value = true;
};

// 编辑权限策略（从列表操作列触发）
const handleEditAccount = (row: IPermissionPolicyItem) => {
  isEditMode.value = true;
  editingAccount.value = row;
  showPolicyFormSideslider.value = true;
};

const handlePolicyFormSuccess = () => {
  refreshList();
};

const handleSearch = (searchCondition: ISearchCondition) => {
  searchQs.set(searchCondition);
};

const handleReset = () => {
  searchQs.clear();
};
</script>

<template>
  <div class="permission-policy-page">
    <!-- 搜索区域 -->
    <Search :fields="searchFields" :condition="condition" @search="handleSearch" @reset="handleReset" />

    <!-- 表格区域 -->
    <div class="table-container">
      <!-- 操作按钮区域 -->
      <div class="action-btns">
        <bk-button theme="primary" @click="handleAddPolicy">
          <plus style="font-size: 22px" />
          新增权限策略库
        </bk-button>
      </div>

      <!-- 数据列表 -->
      <DataList
        :columns="columns"
        :list="tableData"
        :pagination="pagination"
        :loading="isLoading"
        @view-details="handleViewDetails"
        @apply-to-account="handleApplyToAccount"
        @edit-account="handleEditAccount"
      />
    </div>

    <!-- 应用策略库到二级账号弹窗 -->
    <ApplySideslider v-model="showApplySideslider" :policy-data="currentApplyPolicy" @success="handleApplySuccess" />

    <!-- 应用成功后弹出得账号列表查看应用日志 -->
    <LogSideslider v-model="showLogSideslider" />

    <!-- 详情侧栏 -->
    <PolicyInfoSideslider
      v-model="showPolicyInfoSideslider"
      :policy-data="currentApplyPolicy"
      @apply-to-account="handleApplyToAccount"
    />

    <!-- 新建/编辑权限策略库 -->
    <PolicyFormSideslider
      v-model="showPolicyFormSideslider"
      :is-edit="isEditMode"
      :account-data="editingAccount"
      @success="handlePolicyFormSuccess"
    />
  </div>
</template>

<style lang="scss" scoped>
.permission-policy-page {
  height: 100%;

  .table-container {
    background: #fff;
    border-radius: 2px;
    margin: 24px;
    padding: 16px 24px;
  }

  .action-btns {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-bottom: 16px;
  }
}
</style>
