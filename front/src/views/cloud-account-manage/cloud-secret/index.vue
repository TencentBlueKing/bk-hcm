<script setup lang="ts">
import { ref, computed, watch, inject } from 'vue';
import { useRoute } from 'vue-router';
import { InfoLine } from 'bkui-vue/lib/icon';
import usePage from '@/hooks/use-page';
import useSearchQs from '@/hooks/use-search-qs';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import { ModelPropertyColumn, ModelPropertySearch } from '@/model/typings';
import { transformSimpleCondition } from '@/utils/search';
import { useCloudAccountStore } from '@/store/cloud-account';
import { VendorEnum } from '@/common/constant';

import Search from './children/search/search.vue';
import DataList from './children/data-list/data-list.vue';
import SecretDetailSlider from './components/secret-detail-slider.vue';
import SecretActionDialog from './components/secret-action-dialog.vue';
import { SearchConditionFactory } from './children/search/condition-factory';
import { TableColumnFactory } from './children/data-list/column-factory';
import type { ICloudSecretItem, ISearchCondition, SecretActionType } from './typings';

// 获取当前云厂商
const currentVendor = inject<{ value: VendorEnum }>('currentVendor', { value: VendorEnum.TCLOUD });

const route = useRoute();
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

// 表格数据
const tableData = ref<ICloudSecretItem[]>([]);

// 分页
const { pagination, getPageParams } = usePage();

// URL 查询参数处理
const searchQs = useSearchQs({ key: 'secretFilter', properties: searchFields.value });

// 详情侧栏状态
const showDetailSlider = ref(false);
const currentSecret = ref<ICloudSecretItem | null>(null);

// 操作弹窗状态
const showActionDialog = ref(false);
const currentActionType = ref<SecretActionType>('disable');
const actionSecret = ref<ICloudSecretItem | null>(null);

// 加载状态
const isLoading = computed(() => cloudAccountStore.subAccountSecretListLoading);

// 获取列表数据
const fetchList = async () => {
  const bkBizId = getBizsId();
  const vendor = currentVendor?.value || VendorEnum.TCLOUD;

  // 排序参数
  const sort = (route.query.sort || 'cloud_created_at') as string;
  const order = (route.query.order || 'DESC') as string;

  try {
    // 构建请求参数
    const requestParams: { filter?: any; page: any } & Record<string, any> = {
      page: getPageParams(pagination, { sort, order }),
    };

    // 处理搜索条件
    const filterRules = transformSimpleCondition(condition.value, searchFields.value);
    if (filterRules && filterRules.rules && filterRules.rules.length > 0) {
      // 将 filter 条件转换为接口需要的参数格式
      filterRules.rules.forEach((rule: any) => {
        if (rule.field && rule.value) {
          // 根据字段名称映射到接口参数
          const fieldMapping: Record<string, string> = {
            cloud_secret_id: 'cloud_secret_ids',
            status: 'status',
            cloud_sub_account_id: 'cloud_sub_account_ids',
            cloud_main_account_id: 'cloud_main_account_ids',
            sub_account_managers: 'sub_account_managers',
            account_managers: 'account_managers',
          };
          const paramKey = fieldMapping[rule.field];
          if (paramKey) {
            if (paramKey === 'status') {
              requestParams[paramKey] = rule.value;
            } else {
              requestParams[paramKey] = Array.isArray(rule.value) ? rule.value : [rule.value];
            }
          }
        }
      });
    }

    const { list, count } = await cloudAccountStore.getSubAccountSecretList(bkBizId, vendor, requestParams);
    tableData.value = list as ICloudSecretItem[];
    pagination.count = count;
  } catch (error) {
    console.error('获取云密钥列表失败:', error);
    tableData.value = [];
    pagination.count = 0;
  }
};

// 监听路由变化，获取列表数据
watch(
  () => route.query,
  async (query) => {
    // 从 URL 获取搜索条件
    condition.value = searchQs.get(query, {});

    // 设置分页
    pagination.current = Number(query.page) || 1;
    pagination.limit = Number(query.limit) || pagination.limit;

    await fetchList();
  },
  { immediate: true },
);

// 搜索
const handleSearch = (searchCondition: ISearchCondition) => {
  searchQs.set(searchCondition);
};

// 重置
const handleReset = () => {
  searchQs.clear();
};

// 查看详情
const handleViewDetails = (row: ICloudSecretItem) => {
  currentSecret.value = row;
  showDetailSlider.value = true;
};

// 启用密钥
const handleEnableSecret = (row: ICloudSecretItem) => {
  actionSecret.value = row;
  currentActionType.value = 'enable';
  showActionDialog.value = true;
};

// 禁用密钥
const handleDisableSecret = (row: ICloudSecretItem) => {
  actionSecret.value = row;
  currentActionType.value = 'disable';
  showActionDialog.value = true;
};

// 删除密钥
const handleDeleteSecret = (row: ICloudSecretItem) => {
  actionSecret.value = row;
  currentActionType.value = 'delete';
  showActionDialog.value = true;
};

// 操作成功回调
const handleActionSuccess = () => {
  fetchList();
};

// 详情侧栏操作成功回调
const handleDetailActionSuccess = () => {
  fetchList();
};
</script>

<template>
  <div class="cloud-secret-page">
    <!-- 搜索区域 -->
    <Search :fields="searchFields" :condition="condition" @search="handleSearch" @reset="handleReset" />
    <!-- 提示信息 -->
    <div class="info-tip">
      <InfoLine class="tip-icon" />
      <span>本页面仅用于管理已创建的密钥，如需创建新密钥，请进入对应三级账号的详情页进行操作</span>
    </div>

    <!-- 表格区域 -->
    <div class="table-container">
      <!-- 数据列表 -->
      <DataList
        :columns="columns"
        :list="tableData"
        :pagination="pagination"
        :loading="isLoading"
        @view-details="handleViewDetails"
        @enable="handleEnableSecret"
        @disable="handleDisableSecret"
        @delete="handleDeleteSecret"
      />
    </div>

    <!-- 详情侧栏 -->
    <SecretDetailSlider
      v-model="showDetailSlider"
      :secret-data="currentSecret"
      :vendor="currentVendor?.value || 'tcloud'"
      @action-success="handleDetailActionSuccess"
    />

    <!-- 操作弹窗 -->
    <SecretActionDialog
      v-model="showActionDialog"
      :action-type="currentActionType"
      :secret-data="actionSecret"
      :vendor="currentVendor?.value || 'tcloud'"
      @success="handleActionSuccess"
    />
  </div>
</template>

<style lang="scss" scoped>
.cloud-secret-page {
  height: 100%;

  .table-container {
    background: #fff;
    border-radius: 2px;
    margin: 0 24px;
    padding: 16px 24px;
  }

  .info-tip {
    display: flex;
    align-items: center;
    padding: 8px 16px;
    background: #f0f5ff;
    border: 1px solid #a3c5fd;
    border-radius: 2px;
    font-size: 12px;
    color: #63656e;
    margin: 16px 24px;

    .tip-icon {
      color: #3a84ff;
      margin-right: 8px;
      font-size: 14px;
    }
  }
}
</style>
