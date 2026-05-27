<script setup lang="ts">
import { ref, computed, watch, inject, type Ref, h } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { InfoBox, Message } from 'bkui-vue';
import { Plus } from 'bkui-vue/lib/icon';
import usePage from '@/hooks/use-page';
import useSearchQs from '@/hooks/use-search-qs';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import { ModelPropertyColumn, ModelPropertySearch } from '@/model/typings';
import { transformSimpleCondition, localPaginate, localSort } from '@/utils/search';
import { useSecondaryAccountStore, type ISecondaryAccountItem } from '@/store/cloud-account-manage/secondary-account';
import { VendorEnum } from '@/common/constant';
import { QueryFilterType, RulesItem } from '@/typings';

import Search from './children/search/search.vue';
import DataList from './children/data-list/data-list.vue';
import AccountDetailSideslider from './children/account-detail-sideslider/index.vue';
import AccountFormSideslider from './children/account-form-sideslider/index.vue';
import AccountSyncSideslider from './children/account-sync-sideslider/index.vue';
import { SearchConditionFactory } from './children/search/condition-factory';
import { TableColumnFactory } from './children/data-list/column-factory';

export type ISearchCondition = Record<string, any>;

const currentVendor = inject<Ref<VendorEnum>>('currentVendor', ref(VendorEnum.TCLOUD));

const route = useRoute();
const router = useRouter();
const secondaryAccountStore = useSecondaryAccountStore();
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
const fullList = ref<ISecondaryAccountItem[]>([]);

// 当前页展示的数据
const tableData = ref<ISecondaryAccountItem[]>([]);

// 排序参数（默认按创建时间倒序）
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

    // 使用 rollRequest 获取全量数据
    const list = await secondaryAccountStore.getSecondaryAccountFullList(
      getBizsId(),
      vendorFilter,
      (progressList, count) => {
        // 进度回调：每批次数据返回时更新
        fullList.value = progressList;
        pagination.count = count;
        updateTableData();
      },
    );

    fullList.value = list;
    pagination.count = list.length;
    updateTableData();
  } catch (error) {
    console.error('获取账号列表失败:', error);
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

    if (conditionChanged || fullList.value.length === 0) {
      await loadFullList();
    } else {
      // 仅分页/排序变化，前端处理
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
    fullList.value = []; // 清空全量数据，触发重新加载
    const query = { ...route.query };
    delete query.page;
    query._t = String(Date.now());
    router.replace({ query });
  },
);

// 加载状态
const isLoading = computed(() => secondaryAccountStore.accountListLoading);

// 详情侧栏状态
const showDetailSideslider = ref(false);
const currentAccount = ref<ISecondaryAccountItem | null>(null);

const handleViewDetails = (row: ISecondaryAccountItem) => {
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
      const detail = await secondaryAccountStore.getSecondaryAccountDetail(getBizsId(), id as string);
      if (detail) {
        currentAccount.value = detail;
        showDetailSideslider.value = true;
      } else {
        Message({ theme: 'warning', message: `未找到账号数据` });
        // 移除id
        router.replace({ query: { ...route.query, id: undefined } });
      }
    } catch (error) {
      console.error('获取账号详情失败:', error);
      Message({ theme: 'error', message: '获取账号详情失败' });
      router.replace({ query: { ...route.query, id: undefined } });
    }
  },
  { immediate: true },
);

const handleDetailUpdateSuccess = () => {
  refreshList();
};

const refreshList = () => {
  fullList.value = [];
  const query = { ...route.query };
  delete query._t;
  query._t = String(Date.now());
  router.replace({ query });
};

// 录入/编辑账号弹窗状态
const showAccountFormSideslider = ref(false);
const isEditMode = ref(false);
const editingAccount = ref<ISecondaryAccountItem | null>(null);

const handleAddAccount = () => {
  isEditMode.value = false;
  editingAccount.value = null;
  showAccountFormSideslider.value = true;
};

// 编辑账号（从列表操作列触发）
const handleEditAccount = (row: ISecondaryAccountItem) => {
  isEditMode.value = true;
  editingAccount.value = row;
  showAccountFormSideslider.value = true;
};

const handleAccountFormSuccess = () => {
  refreshList();
};

const handleSearch = (searchCondition: ISearchCondition) => {
  searchQs.set(searchCondition);
};

const handleReset = () => {
  searchQs.clear();
};

// 同步账号功能
const showSyncSideslider = ref(false);
const syncingAccounts = ref<ISecondaryAccountItem[]>([]);

const handleSyncAccount = () => {
  const SyncContent = () =>
    h('div', { class: 'sync-info-content' }, [
      h('p', { class: 'sync-info-title' }, '同步信息包含：'),
      h('ul', { class: 'sync-info-list' }, [
        h('li', '二级账号本身的信息（邮箱、保护状态、MFA等）'),
        h('li', '二级账号下的三级账号'),
        h('li', '二级账号下的权限模板'),
      ]),
      h('p', { class: 'sync-info-tip' }, '同步操作可能需要几分钟，请耐心等待'),
    ]);

  InfoBox({
    title: '确定同步本业务下所有二级账号信息',
    type: 'warning',
    subTitle: SyncContent,
    width: 480,
    contentAlign: 'left',
    confirmText: '确定',
    cancelText: '取消',
    onConfirm: () => {
      const accounts = [...tableData.value];
      if (accounts.length === 0) {
        Message({ theme: 'warning', message: '当前没有可同步的账号' });
        return;
      }
      syncingAccounts.value = accounts;
      showSyncSideslider.value = true;
    },
  });
};

const handleSyncFinished = () => {
  syncingAccounts.value = [];
  refreshList();
};
</script>

<template>
  <div class="secondary-account-page">
    <!-- 搜索区域 -->
    <Search :fields="searchFields" :condition="condition" @search="handleSearch" @reset="handleReset" />

    <!-- 表格区域 -->
    <div class="table-container">
      <!-- 操作按钮区域 -->
      <div class="action-btns">
        <bk-button theme="primary" @click="handleAddAccount">
          <plus style="font-size: 22px" />
          录入账号
        </bk-button>

        <bk-button @click="handleSyncAccount">
          <i class="hcm-icon bkhcm-icon-update mr6"></i>
          同步账号
        </bk-button>
      </div>

      <!-- 数据列表 -->
      <DataList
        :columns="columns"
        :list="tableData"
        :pagination="pagination"
        :loading="isLoading"
        @view-details="handleViewDetails"
        @edit-account="handleEditAccount"
      />
    </div>

    <!-- 详情侧栏 -->
    <AccountDetailSideslider
      v-model="showDetailSideslider"
      :row-data="currentAccount"
      @update-success="handleDetailUpdateSuccess"
    />

    <!-- 录入/编辑账号弹窗 -->
    <AccountFormSideslider
      v-model="showAccountFormSideslider"
      :is-edit="isEditMode"
      :account-data="editingAccount"
      @success="handleAccountFormSuccess"
    />

    <!-- 同步账号侧滑 -->
    <AccountSyncSideslider
      v-model="showSyncSideslider"
      :accounts="syncingAccounts"
      :biz-id="getBizsId()"
      :vendor="currentVendor"
      @finished="handleSyncFinished"
    />
  </div>
</template>

<style lang="scss" scoped>
.secondary-account-page {
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

<!-- 同步账号弹窗全局样式 -->
<style lang="scss">
.sync-info-content {
  text-align: left;
  padding: 12px 16px;
  background-color: rgb(245 247 250);

  .sync-info-title {
    font-size: 12px;
    color: #4d4f56;
    line-height: 20px;
    font-weight: 700;
  }

  .sync-info-list {
    margin: 0;
    padding-left: 22px;

    li {
      font-size: 12px;
      color: #4d4f56;
      line-height: 20px;
      list-style-type: disc;
    }
  }

  .sync-info-tip {
    font-size: 12px;
    margin-top: 22px;
    line-height: 20px;
  }
}
</style>
