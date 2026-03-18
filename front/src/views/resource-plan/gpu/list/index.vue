<script setup lang="ts">
import { computed, ref, watch, inject } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { Message } from 'bkui-vue';
import { Plus } from 'bkui-vue/lib/icon';
import useSearchQs from '@/hooks/use-search-qs';
import usePage from '@/hooks/use-page';
import { getModel } from '@/model/manager';
import { transformSimpleCondition } from '@/utils/search';
import { useGpuDemandStore } from '@/store/resource-plan/gpu-demand';
import { MENU_BUSINESS_RESOURCE_PLAN_GPU_DETAIL, MENU_SERVICE_RESOURCE_PLAN_GPU_DETAIL } from '@/constants/menu-symbol';
import type { IGpuDemandItem } from '@/store/resource-plan/gpu-demand';
import type { ISearchCondition } from '../types';
import { SearchCondition, SERVICE_ONLY_FIELDS } from './children/search/condition';
import { TableColumn, SERVICE_ONLY_COLUMNS } from './children/data-list/column';
import Search from './children/search/search.vue';
import DataList from './children/data-list/data-list.vue';
import CreateSlider from '../create/index.vue';
import { useTerminateConfirm } from '../hooks/use-terminate-confirm';

const route = useRoute();
const router = useRouter();

const isBusinessPage = inject('isBusinessPage', false);

const gpuDemandStore = useGpuDemandStore();
const { confirmTerminateOrder } = useTerminateConfirm();

const { pagination, getPageParams } = usePage();

const conditionModel = getModel(SearchCondition);
const conditionProperties = computed(() => conditionModel.getProperties());

const columnModel = getModel(TableColumn);
const columnProperties = computed(() => columnModel.getProperties());

const searchFields = computed(() => {
  const fields = conditionProperties.value;
  if (isBusinessPage) {
    return fields.filter((field) => !SERVICE_ONLY_FIELDS.includes(field.id));
  }
  return fields;
});
const dataListColumns = computed(() => {
  const cols = columnProperties.value;
  return isBusinessPage ? cols.filter((field) => !SERVICE_ONLY_COLUMNS.includes(field.id)) : cols;
});

const searchQs = useSearchQs({ key: 'filter', properties: conditionProperties });

const condition = ref<ISearchCondition>({});
const gpuDemandList = ref<IGpuDemandItem[]>([]);

const fetchList = async () => {
  const { query } = route;
  condition.value = searchQs.get(query, {});
  pagination.current = Number(query.page) || 1;
  pagination.limit = Number(query.limit) || pagination.limit;

  const filter = transformSimpleCondition(condition.value, conditionProperties.value);
  const page = getPageParams(pagination, {
    sort: (query.sort || 'created_at') as string,
    order: (query.order || 'DESC') as string,
  });

  try {
    const { list, count } = await gpuDemandStore.getGpuDemandList({ filter, page });
    gpuDemandList.value = list;
    pagination.count = count;
  } catch {
    gpuDemandList.value = [];
    pagination.count = 0;
  }
};

watch(() => route.query, fetchList, { immediate: true });

const handleSearch = (vals: ISearchCondition) => {
  searchQs.set(vals);
};

const handleReset = () => {
  searchQs.clear();
};

const isServicePage = inject('isServicePage', false);

const selectedRows = ref<IGpuDemandItem[]>([]);

const handleSelect = (selections: IGpuDemandItem[]) => {
  selectedRows.value = selections;
};

const handleBatchPending = async () => {
  const orderIds = selectedRows.value.map((row) => row.id);
  if (!orderIds.length) return;
  await gpuDemandStore.batchPendingOrders({ order_ids: orderIds });
  fetchList();
};

const isCreateSliderShow = ref(false);
const isCreateSliderHidden = ref(true);

const handleCreate = () => {
  isCreateSliderHidden.value = false;
  isCreateSliderShow.value = true;
};

const handleCreateSuccess = () => {
  fetchList();
};

const handleCreateHidden = () => {
  isCreateSliderHidden.value = true;
};

const handleViewDetails = (row: IGpuDemandItem) => {
  const detailName = isBusinessPage ? MENU_BUSINESS_RESOURCE_PLAN_GPU_DETAIL : MENU_SERVICE_RESOURCE_PLAN_GPU_DETAIL;
  router.push({ name: detailName, query: { ...route.query, id: row.id } });
};

const handleAdjust = (row: IGpuDemandItem) => {
  const detailName = isBusinessPage ? MENU_BUSINESS_RESOURCE_PLAN_GPU_DETAIL : MENU_SERVICE_RESOURCE_PLAN_GPU_DETAIL;
  router.push({ name: detailName, query: { ...route.query, id: row.id, action: 'adjust' } });
};

// 驳回确认弹窗
const isRejectDialogShow = ref(false);
const rejectRow = ref<IGpuDemandItem | null>(null);
const isRejectLoading = ref(false);

const handleReject = (row: IGpuDemandItem) => {
  rejectRow.value = row;
  isRejectDialogShow.value = true;
};

const handleRejectConfirm = async () => {
  if (!rejectRow.value) return;
  isRejectLoading.value = true;
  try {
    await gpuDemandStore.batchRejectOrders({ order_ids: [rejectRow.value.id] });
    Message({ theme: 'success', message: '驳回成功' });
    isRejectDialogShow.value = false;
    fetchList();
  } catch {
    Message({ theme: 'error', message: '驳回失败' });
  } finally {
    isRejectLoading.value = false;
  }
};

const handleRejectCancel = () => {
  isRejectDialogShow.value = false;
  rejectRow.value = null;
};

// 终止确认（InfoBox）
const handleTerminate = (row: IGpuDemandItem) => {
  confirmTerminateOrder(row.id, fetchList);
};

const handleStartReview = async (row: IGpuDemandItem) => {
  await gpuDemandStore.batchPendingOrders({ order_ids: [row.id] });
  fetchList();
};
</script>

<template>
  <div class="gpu-demand-list">
    <search :fields="searchFields" :condition="condition" @search="handleSearch" @reset="handleReset" />
    <div class="table-panel">
      <div class="toolbar">
        <bk-button v-if="isBusinessPage" theme="primary" @click="handleCreate">
          <plus style="font-size: 22px" />
          新增需求
        </bk-button>
        <bk-button v-if="isServicePage" theme="primary" :disabled="!selectedRows.length" @click="handleBatchPending">
          批量转为评审中
        </bk-button>
      </div>
      <data-list
        v-bkloading="{ loading: gpuDemandStore.listLoading }"
        :columns="dataListColumns"
        :list="gpuDemandList"
        :pagination="pagination"
        @view-details="handleViewDetails"
        @adjust="handleAdjust"
        @reject="handleReject"
        @terminate="handleTerminate"
        @start-review="handleStartReview"
        @select="handleSelect"
      />
    </div>
  </div>
  <template v-if="!isCreateSliderHidden">
    <create-slider v-model="isCreateSliderShow" @hidden="handleCreateHidden" @success="handleCreateSuccess" />
  </template>

  <!-- 驳回确认弹窗 -->
  <bk-dialog v-model:is-show="isRejectDialogShow" title="驳回确认" :quick-close="false" @closed="handleRejectCancel">
    <div class="reject-confirm-content">
      驳回需求
      <span class="reject-order-id">{{ rejectRow?.id }}</span>
      后，该需求单下所有关联记录（含已评审记录）将统一变更为驳回状态，是否确认？
    </div>
    <template #footer>
      <bk-button @click="handleRejectCancel">取消</bk-button>
      <bk-button theme="danger" :loading="isRejectLoading" @click="handleRejectConfirm">确认驳回</bk-button>
    </template>
  </bk-dialog>
</template>

<style lang="scss" scoped>
.gpu-demand-list {
  height: 100%;

  .table-panel {
    background: #fff;
    border-radius: 2px;
    box-shadow: 0 2px 4px 0 #1919290d;
    padding: 16px;
  }

  .toolbar {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-bottom: 16px;
  }
}

.reject-confirm-content {
  font-size: 14px;
  line-height: 22px;
  color: #63656e;

  .reject-order-id {
    color: #3a84ff;
    font-weight: 700;
  }
}
</style>
