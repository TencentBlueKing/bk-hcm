<script setup lang="ts">
import { computed, ComputedRef, h, inject, reactive, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { IRsItem, useLoadBalancerRsStore } from '@/store/load-balancer/rs';
import { RsDeviceType } from '@/views/load-balancer/constants';
import { ActionItemType } from '@/views/load-balancer/typing';
import usePage from '@/hooks/use-page';
import { ILoadBalanceDeviceCondition, IDeviceListDataLoadedEvent, DeviceTabEnum } from '../typing';
import routeQuery from '@/router/utils/query';
import { IAuthSign } from '@/common/auth-service';
import { useRoute } from 'vue-router';
import expand from './children/expand.vue';
import ActionItem from '@/views/load-balancer/children/action-item.vue';
import BatchRsOperationDialog from '@/views/load-balancer/device/main-content/children/batch-rs-operation-dialog.vue';
import RsBatchExportButton from '@/views/load-balancer/children/export/rs-batch-button.vue';

const props = defineProps<{ condition: ILoadBalanceDeviceCondition }>();
const emit = defineEmits<IDeviceListDataLoadedEvent>();

const route = useRoute();
const { t } = useI18n();
const loadBalancerRsStore = useLoadBalancerRsStore();

const currentGlobalBusinessId = inject<ComputedRef<number>>('currentGlobalBusinessId');
const clbOperationAuthSign = inject<ComputedRef<IAuthSign | IAuthSign[]>>('clbOperationAuthSign');

const { pagination, getPageParams } = usePage();

const rsList = ref<IRsItem[]>([]);
const loading = ref(false);
const expandRef = ref(null);

const selections = computed(() => expandRef.value?.selections ?? []);

const batchOperationDialog = reactive({ isShow: false, isHidden: true, type: RsDeviceType.INFO });
const actionConfig: Record<RsDeviceType, ActionItemType> = {
  [RsDeviceType.ADJUST]: {
    type: 'button',
    label: t('批量调整RS权重'),
    value: RsDeviceType.ADJUST,
    disabled: () => selections.value.length === 0,
    authSign: () => clbOperationAuthSign.value,
    handleClick: () => {
      batchOperationDialog.isHidden = false;
      batchOperationDialog.isShow = true;
      batchOperationDialog.type = RsDeviceType.ADJUST;
    },
  },
  [RsDeviceType.UNBIND]: {
    type: 'button',
    label: t('批量解绑RS'),
    value: RsDeviceType.UNBIND,
    disabled: () => selections.value.length === 0,
    authSign: () => clbOperationAuthSign.value,
    handleClick: () => {
      batchOperationDialog.isHidden = false;
      batchOperationDialog.isShow = true;
      batchOperationDialog.type = RsDeviceType.UNBIND;
    },
  },
  [RsDeviceType.BATCH_EXPORT]: {
    value: RsDeviceType.BATCH_EXPORT,
    render: () =>
      h(RsBatchExportButton, {
        selections: selections.value,
        vendor: props.condition.vendor,
      }),
  },
};
const listenerActionList = computed<ActionItemType[]>(() => {
  return [{ value: RsDeviceType.ADJUST }, { value: RsDeviceType.UNBIND }, { value: RsDeviceType.BATCH_EXPORT }];
});
const actionList = computed<ActionItemType[]>(() => {
  return listenerActionList.value.reduce((prev, curr) => {
    const config = actionConfig[curr.value as RsDeviceType];
    if (curr.children) {
      prev.push({
        ...config,
        ...curr,
        children: curr.children.map((childAction) => ({
          ...actionConfig[childAction.value as RsDeviceType],
          ...childAction,
        })),
      });
    } else {
      prev.push({ ...config, ...curr });
    }
    return prev;
  }, []);
});

const getList = async (condition: ILoadBalanceDeviceCondition, pageParams = { sort: '', order: 'DESC' }) => {
  if (!condition.account_id) return;
  try {
    loading.value = true;
    const { list, count } = await loadBalancerRsStore.getRsList(
      condition,
      getPageParams(pagination, pageParams),
      currentGlobalBusinessId.value,
    );
    pagination.count = count;
    rsList.value = list;
  } catch (error) {
    console.error(error);
    rsList.value = [];
    pagination.count = 0;
  } finally {
    loading.value = false;
    emit('list-data-loaded', DeviceTabEnum.RS, {
      type: 'rsIPCount',
      data: {
        count: pagination.count,
      },
    });
  }
};
const handlePageChange = (page: number) => {
  routeQuery.set({
    page,
    _t: Date.now(),
  });
};
const handleLimitChange = (limit: number) => {
  routeQuery.set({
    limit,
    page: 1,
    _t: Date.now(),
  });
};

watch(
  () => route.query,
  async (query) => {
    pagination.current = Number(query.page) || 1;
    pagination.limit = Number(query.limit) || pagination.limit;

    const sort = (query.sort || '') as string;
    const order = (query.order || 'DESC') as string;

    getList(props.condition, { sort, order });
  },
);
</script>

<template>
  <div class="rs-table-container" v-bkloading="{ loading }">
    <div class="toolbar">
      <div class="action-container">
        <template v-for="action in actionList" :key="action.value">
          <hcm-auth v-if="action.authSign" :sign="action.authSign()" v-slot="{ noPerm }">
            <action-item :action="action" :disabled="noPerm || action.disabled?.()" :loading="action.loading?.()" />
          </hcm-auth>
          <action-item v-else :action="action" :disabled="action.disabled?.()" :loading="action.loading?.()" />
        </template>
      </div>
    </div>
    <expand
      :rs-list="rsList"
      :vendor="condition.vendor"
      :type="RsDeviceType.INFO"
      class="expand-table"
      ref="expandRef"
    />
    <bk-pagination
      class="expand-pagination"
      v-model="pagination.current"
      :count="pagination.count"
      :limit="pagination.limit"
      size="small"
      :layout="['total', 'limit', 'list']"
      align="right"
      :limit-list="[10, 20, 50, 100, 500]"
      @change="handlePageChange"
      @limit-change="handleLimitChange"
    />

    <template v-if="!batchOperationDialog.isHidden">
      <batch-rs-operation-dialog
        v-model="batchOperationDialog.isShow"
        :selections="selections"
        :vendor="condition.vendor"
        :type="batchOperationDialog.type"
        @hidden="batchOperationDialog.isHidden = true"
      />
    </template>
  </div>
</template>

<style scoped lang="scss">
.rs-table-container {
  height: 100%;

  .toolbar {
    margin-bottom: 16px;
    display: flex;
    align-items: center;

    .action-container {
      display: flex;
      align-items: center;
      gap: 8px;
    }
  }

  .expand-table {
    height: calc(100% - 100px);
    overflow-y: auto;
    margin-bottom: 10px;
  }
  :deep(.expand-pagination) {
    .is-last {
      margin-left: auto;
    }
  }
}
</style>
