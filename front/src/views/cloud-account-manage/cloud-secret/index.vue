<script setup lang="ts">
import { ref, computed, watch, inject, type Ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { InfoLine } from 'bkui-vue/lib/icon';
import { Message } from 'bkui-vue';
import usePage from '@/hooks/use-page';
import useSearchQs from '@/hooks/use-search-qs';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import { ModelPropertyColumn, ModelPropertySearch } from '@/model/typings';
import { transformSimpleCondition } from '@/utils/search';
import { type ISubAccountSecretParams, useCloudSecretStore } from '@/store/cloud-account-manage/cloud-secret';
import { VendorEnum } from '@/common/constant';

import Search from './children/search/search.vue';
import DataList from './children/data-list/data-list.vue';
import SecretDetailSlider from './children/secret-detail-sideslider/index.vue';
import SecretActionDialog from './children/secret-action-dialog/index.vue';
import { SearchConditionFactory } from './children/search/condition-factory';
import { TableColumnFactory } from './children/data-list/column-factory';
import type { ICloudSecretItem, ISearchCondition, SecretActionType } from './typings';

const currentVendor = inject<Ref<VendorEnum>>('currentVendor', ref(VendorEnum.TCLOUD));

const route = useRoute();
const router = useRouter();
const cloudSecretStore = useCloudSecretStore();
const { getBizsId } = useWhereAmI();

const searchModel = SearchConditionFactory.createModel();
const columnModel = TableColumnFactory.createModel();
const searchFields = computed<ModelPropertySearch[]>(() => searchModel.getProperties());
const columns = computed<ModelPropertyColumn[]>(() => columnModel.getProperties());
const condition = ref<ISearchCondition>({});
const tableData = ref<ICloudSecretItem[]>([]);
const { pagination, getPageParams } = usePage();
const searchQs = useSearchQs({ key: 'filter', properties: searchFields.value });
const showDetailSlider = ref(false);
const currentSecret = ref<ICloudSecretItem | null>(null);
const showActionDialog = ref(false);
const currentActionType = ref<SecretActionType>('disable');
const actionSecret = ref<ICloudSecretItem | null>(null);
const sortParams = ref<{ sort: string; order: string }>({ sort: 'cloud_created_at', order: 'DESC' });
const isLoading = computed(() => cloudSecretStore.subAccountSecretListLoading);

const fetchList = async () => {
  const bkBizId = getBizsId();
  const vendor = currentVendor.value;

  try {
    const requestParams: ISubAccountSecretParams = {
      page: getPageParams(pagination, sortParams.value),
    };
    condition.value = searchQs.get(route.query, {});

    const filterRules = transformSimpleCondition(condition.value, searchFields.value);
    if (filterRules && filterRules.rules && filterRules.rules.length > 0) {
      const extensionFields: Record<string, string> = {
        cloud_secret_id: 'cloud_secret_ids',
        cloud_sub_account_id: 'cloud_sub_account_ids',
        cloud_main_account_id: 'cloud_main_account_ids',
      };
      const topLevelFields: Record<string, string> = {
        status: 'status',
        sub_account_managers: 'sub_account_managers',
        account_managers: 'account_managers',
      };

      filterRules.rules.forEach((rule: any) => {
        if (rule.field && rule.value) {
          const extKey = extensionFields[rule.field] as keyof NonNullable<typeof requestParams.extension>;
          if (extKey) {
            if (!requestParams.extension) requestParams.extension = {};
            requestParams.extension[extKey] = Array.isArray(rule.value) ? rule.value : [rule.value];
          } else {
            const paramKey = topLevelFields[rule.field] as keyof Omit<typeof requestParams, 'page' | 'extension'>;
            if (paramKey) {
              if (paramKey === 'status') {
                requestParams[paramKey] = rule.value;
              } else {
                requestParams[paramKey] = Array.isArray(rule.value) ? rule.value : [rule.value];
              }
            }
          }
        }
      });
    }

    const { list, count } = await cloudSecretStore.getSubAccountSecretList(bkBizId, vendor, requestParams);
    tableData.value = list as ICloudSecretItem[];
    pagination.count = count;
  } catch (error) {
    console.error('获取云密钥列表失败:', error);
    tableData.value = [];
    pagination.count = 0;
  }
};

watch(
  () => route.query,
  async (query) => {
    // 设置分页
    pagination.current = Number(query.page) || 1;
    pagination.limit = Number(query.limit) || pagination.limit;

    // 排序参数
    sortParams.value = {
      sort: (query.sort || 'cloud_created_at') as string,
      order: (query.order || 'DESC') as string,
    };

    // 判断是否只是分页/排序变化（不需要重新拉取全量数据）
    const newCondition = searchQs.get(query, {});
    const conditionChanged = JSON.stringify(newCondition) !== JSON.stringify(condition.value);
    const isRefresh = query._t !== undefined;

    if (conditionChanged || tableData.value.length === 0 || isRefresh) {
      await fetchList();
    }
  },
  { immediate: true },
);

watch(
  () => currentVendor.value,
  () => {
    pagination.current = 1;
    const query = { ...route.query };
    delete query.page;
    query._t = String(Date.now());
    router.replace({ query });
  },
);

const handleSearch = (searchCondition: ISearchCondition) => {
  searchQs.set(searchCondition);
};

const handleReset = () => {
  searchQs.clear();
};

const handleViewDetails = (row: ICloudSecretItem) => {
  currentSecret.value = row;
  showDetailSlider.value = true;
  // 将 id 写入 URL query，支持分享/刷新/浏览器后退
  router.replace({ query: { ...route.query, id: row.id, _t: undefined } });
};

watch(showDetailSlider, (val) => {
  if (!val && route.query.id) {
    const query = { ...route.query };
    delete query.id;
    router.replace({ query });
  }
});

watch(
  () => route.query.id,
  async (id) => {
    if (!id) {
      showDetailSlider.value = false;
      currentSecret.value = null;
      return;
    }
    try {
      const detail = await cloudSecretStore.getSubAccountSecretDetail(getBizsId(), currentVendor.value, id as string);
      if (detail) {
        currentSecret.value = detail as ICloudSecretItem;
        showDetailSlider.value = true;
      } else {
        Message({ theme: 'warning', message: `未找到云密钥「${id}」的数据` });
        router.replace({ query: { ...route.query, id: undefined } });
      }
    } catch (error) {
      console.error('获取云密钥详情失败:', error);
      Message({ theme: 'error', message: '获取云密钥详情失败' });
      router.replace({ query: { ...route.query, id: undefined } });
    }
  },
  { immediate: true },
);

const handleEnableSecret = (row: ICloudSecretItem) => {
  actionSecret.value = row;
  currentActionType.value = 'enable';
  showActionDialog.value = true;
};

const handleDisableSecret = (row: ICloudSecretItem) => {
  actionSecret.value = row;
  currentActionType.value = 'disable';
  showActionDialog.value = true;
};

const handleDeleteSecret = (row: ICloudSecretItem) => {
  actionSecret.value = row;
  currentActionType.value = 'delete';
  showActionDialog.value = true;
};

const handleActionSuccess = () => {
  fetchList();
};

const handleDetailActionSuccess = () => {
  fetchList();
};
</script>

<template>
  <div class="cloud-secret-page">
    <Search :fields="searchFields" :condition="condition" @search="handleSearch" @reset="handleReset" />

    <div class="info-tip">
      <InfoLine class="tip-icon" />
      <span>本页面仅用于管理已创建的密钥，如需创建新密钥，请进入对应三级账号的详情页进行操作</span>
    </div>

    <div class="table-container">
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

    <SecretDetailSlider
      v-model="showDetailSlider"
      :secret-data="currentSecret"
      :vendor="currentVendor"
      @action-success="handleDetailActionSuccess"
    />

    <SecretActionDialog
      v-model="showActionDialog"
      :action-type="currentActionType"
      :secret-data="actionSecret"
      :vendor="currentVendor"
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
      flex-shrink: 0;
    }
  }
}
</style>
