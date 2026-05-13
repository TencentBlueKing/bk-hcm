<script setup lang="ts">
import { ref, computed, watch, inject, type Ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import usePage from '@/hooks/use-page';
import useSearchQs from '@/hooks/use-search-qs';
import { ModelPropertyColumn, ModelPropertySearch } from '@/model/typings';
import { transformSimpleCondition } from '@/utils/search';
import { VendorEnum } from '@/common/constant';
import { QueryFilterType, RulesItem } from '@/typings';
import { usePermissionPolicyStore, type IApplyResultItem } from '@/store/cloud-account-manage/permission-policy';
import routerAction from '@/router/utils/action';
import { MENU_BUSINESS_TICKET_MANAGEMENT, MENU_BUSINESS_TICKET_DETAILS } from '@/constants/menu-symbol';
import Search from './children/search/search.vue';
import DataList from './children/data-list/data-list.vue';
import PolicyFormSideslider from './children/policy-form-sideslider/index.vue';
import PolicyInfoSideslider from './children/policy-form-sideslider/info.vue';
import ApplySideslider from './children/apply-sideslider/index.vue';
import LogSideslider from './children/log-sideslider/index.vue';
import { SearchConditionFactory } from './children/search/condition-factory';
import { TableColumnFactory } from './children/data-list/column-factory';
import type { IPermissionPolicyItem } from './typings';
import { ApplyOperationType } from './typings';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import {
  AUTH_CREATE_PERMISSION_POLICY_LIBRARY,
  AUTH_BIZ_CREATE_PERMISSION_POLICY_LIBRARY,
} from '@/constants/auth-symbols';
import { getAuthSignByBusinessId } from '@/utils';

export type ISearchCondition = Record<string, any>;

const currentVendor = inject<Ref<VendorEnum>>('currentVendor', ref(VendorEnum.TCLOUD));

const route = useRoute();
const router = useRouter();
const { isBusinessPage, getBizsId } = useWhereAmI();

const bizId = computed(() => (isBusinessPage ? getBizsId() : 0));

const permissionPolicyStore = usePermissionPolicyStore();
// 创建模型实例
const searchModel = SearchConditionFactory.createModel();
const columnModel = TableColumnFactory.createModel();

// 搜索字段
const searchFields = computed<ModelPropertySearch[]>(() => searchModel.getProperties());

// 表格列
const columns = computed<ModelPropertyColumn[]>(() => columnModel.getProperties());

// 搜索条件
const condition = ref<ISearchCondition>({});

// 当前页展示的数据
const tableData = ref<IPermissionPolicyItem[]>([]);

// 排序参数
const sortParams = ref<{ sort: string; order: string }>({ sort: 'created_at', order: 'DESC' });

// 分页
const { pagination, getPageParams } = usePage();

// URL 查询参数处理
const searchQs = useSearchQs({ key: 'filter', properties: searchFields.value });

// 加载数据
const loadList = async () => {
  try {
    const pageParams = getPageParams(pagination, sortParams.value);
    // 从 URL 获取搜索条件
    condition.value = searchQs.get(route.query, {});

    // 构建 filter，加入云厂商 vendor 条件
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

    // 获取数据
    const data = await permissionPolicyStore.getPermissionPolicyList(bizId.value, currentVendor.value, {
      page: { ...pageParams },
      filter: { ...vendorFilter },
    });
    // 获取关联的账号列表
    tableData.value = await getAssociationAccountList(data.list);
    pagination.count = data.count;
  } catch (error) {
    console.error('获取权限策略库列表失败:', error);
    tableData.value = [];
    pagination.count = 0;
  }
};

const getAssociationAccountList = async (list: IPermissionPolicyItem[]) => {
  const data = [...list];
  const res = await Promise.allSettled(
    data.map((item) =>
      permissionPolicyStore.getPermissionAssoAccountList(
        bizId.value,
        currentVendor.value,
        item.id,
        item.associated_account_count,
      ),
    ),
  );
  res.forEach((item: any, index) => {
    data[index]['related_accounts'] = item?.value || [];
  });
  return data;
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

    await loadList();
  },
  { immediate: true },
);

// 监听云厂商变化，刷新列表
watch(
  () => currentVendor.value,
  () => {
    pagination.current = 1;
    tableData.value = [];
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
const currentApplyPolicy = ref<IPermissionPolicyItem | undefined>(undefined);

// 应用日志相关
const applyReason = ref<IApplyResultItem[]>([]);
const operationType = ref<ApplyOperationType>(ApplyOperationType.APPLY_NEW);

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
const handleApplySuccess = (data: IApplyResultItem[] | string[], type: ApplyOperationType) => {
  if (isBusinessPage) {
    const ticketIds = data as string[];
    routerAction.redirect({
      name: ticketIds?.length > 1 ? MENU_BUSINESS_TICKET_MANAGEMENT : MENU_BUSINESS_TICKET_DETAILS,
      query: {
        id: ticketIds?.length > 1 ? undefined : ticketIds[0],
        type: 'account',
        bizs: bizId.value,
      },
    });
    return;
  }
  refreshList();
  showLogSideslider.value = true;
  applyReason.value = [...(data as IApplyResultItem[])];
  operationType.value = type;
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
const editingData = ref<IPermissionPolicyItem | null>(null);

const handleAddPolicy = (row: IPermissionPolicyItem) => {
  isEditMode.value = false;
  editingData.value = row;
  showPolicyFormSideslider.value = true;
};

// 编辑权限策略（从列表操作列触发）
const handleEditAccount = (row: IPermissionPolicyItem) => {
  isEditMode.value = true;
  editingData.value = row;
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
      <div class="action-btns" v-if="!isBusinessPage">
        <hcm-auth
          :sign="
            getAuthSignByBusinessId(
              bizId,
              AUTH_CREATE_PERMISSION_POLICY_LIBRARY,
              AUTH_BIZ_CREATE_PERMISSION_POLICY_LIBRARY,
            )
          "
          v-slot="{ noPerm }"
        >
          <bk-button theme="primary" :disabled="noPerm" @click="handleAddPolicy">
            <plus style="font-size: 22px" />
            新增权限策略库
          </bk-button>
        </hcm-auth>
      </div>

      <!-- 数据列表 -->
      <DataList
        v-bkloading="{ loading: permissionPolicyStore.permissionPolicyListLoading }"
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
    <LogSideslider v-model="showLogSideslider" :data="applyReason" :type="operationType" :biz-id="bizId" />

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
      :permission-policy-data="editingData"
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
