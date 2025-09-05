<script setup lang="ts">
import { computed, ComputedRef, h, inject, onMounted, reactive, ref, watch } from 'vue';
import { useRoute } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { ILoadBalancerDetails, useLoadBalancerClbStore } from '@/store/load-balancer/clb';
import { IListenerItem, useLoadBalancerListenerStore } from '@/store/load-balancer/listener';
import { ListenerDeviceType } from '@/views/load-balancer/constants';
import { ActionItemType } from '@/views/load-balancer/typing';
import { DisplayFieldType, DisplayFieldFactory } from '@/views/load-balancer/children/display/field-factory';
import { ModelPropertyColumn } from '@/model/typings';
import usePage from '@/hooks/use-page';
import useTableSelection from '@/hooks/use-table-selection';
import { LB_TYPE_MAP } from '@/common/constant';
import { IAuthSign } from '@/common/auth-service';
import routerAction from '@/router/utils/action';

import { Button, Message } from 'bkui-vue';
import ActionItem from '@/views/load-balancer/children/action-item.vue';
import DataList from '@/views/load-balancer/children/display/data-list.vue';
import AddListenerSideslider from '@/views/load-balancer/listener/add.vue';
import BatchDeleteDialog from '@/views/load-balancer/listener/children/batch-delete-dialog.vue';
import ListenerBatchExportButton from '@/views/load-balancer/children/export/listener-batch-button.vue';
import Confirm from '@/components/confirm';
import DetailsSideslider from '@/views/load-balancer/listener/details.vue';
import BatchCopy from '@/views/load-balancer/device/main/children/batch-copy.vue';
import { ILoadBalanceDeviceCondition } from '../../common';

const props = defineProps<{ condition: ILoadBalanceDeviceCondition }>();
const emit = defineEmits(['getList']);
const details = ref<ILoadBalancerDetails>();
const route = useRoute();
const { t } = useI18n();
const loadBalancerListenerStore = useLoadBalancerListenerStore();
const loadBalancerClbStore = useLoadBalancerClbStore();

const dataListRef = ref(null);
const max = 1000;

const currentGlobalBusinessId = inject<ComputedRef<number>>('currentGlobalBusinessId');
const clbOperationAuthSign = inject<ComputedRef<IAuthSign | IAuthSign[]>>('clbOperationAuthSign');

const actionConfig: Record<ListenerDeviceType, ActionItemType> = {
  [ListenerDeviceType.BATCH_EXPORT]: {
    value: ListenerDeviceType.BATCH_EXPORT,
    render: () =>
      h(ListenerBatchExportButton, { selections: moreData.value ? [] : selections.value, onlyExportListener: true }),
  },
  [ListenerDeviceType.REMOVE]: {
    type: 'button',
    label: t('批量删除'),
    value: ListenerDeviceType.REMOVE,
    disabled: () => selections.value.length === 0 || moreData.value,
    authSign: () => clbOperationAuthSign.value,
    handleClick: () => {
      batchDeleteDialogState.isHidden = false;
      batchDeleteDialogState.isShow = true;
    },
  },
  [ListenerDeviceType.COPY]: {
    label: t('复制'),
    value: ListenerDeviceType.COPY,
    render: () => h(BatchCopy, { selections: moreData.value ? [] : selections.value }),
  },
};
const listenerActionList = computed<ActionItemType[]>(() => {
  return [
    { value: ListenerDeviceType.BATCH_EXPORT },
    { value: ListenerDeviceType.REMOVE },
    { value: ListenerDeviceType.COPY },
  ];
});
const actionList = computed<ActionItemType[]>(() => {
  return listenerActionList.value.reduce((prev, curr) => {
    const config = actionConfig[curr.value as ListenerDeviceType];
    if (curr.children) {
      prev.push({
        ...config,
        ...curr,
        children: curr.children.map((childAction) => ({
          ...actionConfig[childAction.value as ListenerDeviceType],
          ...childAction,
        })),
      });
    } else {
      prev.push({ ...config, ...curr });
    }
    return prev;
  }, []);
});
const moreData = computed(() => dataListRef.value?.getSelection()?.length > max);

// data-list
const displayFieldIds = [
  'name',
  'protocol',
  'port',
  'lb_vip',
  'lb_cloud_id',
  'lb_network_type',
  'domain_num',
  'url_num',
  'rs_num',
];
const convertFieldIds = {
  lb_vips: 'lb_vip',
  lb_id: 'lb_cloud_id',
  rule_domain_count: 'domain_num',
  url_count: 'url_num',
  target_count: 'rs_num',
};
const displayProperties = DisplayFieldFactory.createModel(DisplayFieldType.LISTENER).getProperties();
const displayConfig: Record<string, Partial<ModelPropertyColumn>> = {
  name: {
    render: ({ data, row }) => {
      const handleClick = async () => {
        details.value = await loadBalancerClbStore.getLoadBalancerDetails(row.lb_id, currentGlobalBusinessId.value);
        detailsSidesliderState.isHidden = false;
        detailsSidesliderState.isShow = true;
        detailsSidesliderState.rowData = data;
      };
      return h(Button, { theme: 'primary', text: true, onClick: handleClick }, row.name);
    },
  },
  port: {
    render: ({ row, cell }) => `${cell}${row.end_port ? `-${row.end_port}` : ''}`,
  },
  lb_network_type: {
    render: ({ cell }) => LB_TYPE_MAP[cell],
  },
};
const dataListColumns = displayFieldIds.map((id) => {
  const property = displayProperties.find((field) => field.id === id);
  return { ...property, ...displayConfig[id] };
});

const { pagination } = usePage();
const listenerList = ref<IListenerItem[]>([]);

const isCurRowSelectEnable = (row: any) => {
  if (currentGlobalBusinessId.value) return true;
  if (row.id) return row.bk_biz_id === -1;
};
const isRowSelectEnable = ({ row, isCheckAll }: any) => {
  if (isCheckAll) return true;
  return isCurRowSelectEnable(row);
};
const { selections, handleSelectAll, handleSelectChange } = useTableSelection({
  isRowSelectable: isRowSelectEnable,
});

const asyncSetRsWeightStat = async (list: IListenerItem[]) => {
  const ids = list.map((item) => item.id);
  const map = await loadBalancerListenerStore.getListenersRsWeightStat(ids, currentGlobalBusinessId.value);
  listenerList.value.forEach((item) => {
    const { non_zero_weight_count, zero_weight_count, total_count: totalCount } = map[item.id];
    Object.assign(item, { non_zero_weight_count, zero_weight_count, rs_num: totalCount });
  });
};

const handleSingleDelete = (row: any) => {
  Confirm('请确定删除监听器', `将删除监听器【${row.name}】`, async () => {
    await loadBalancerListenerStore.batchDeleteListener(
      { ids: [row.id], account_id: row.account_id },
      currentGlobalBusinessId.value,
    );
    Message({ theme: 'success', message: '删除成功' });
    getList(props.condition);
  });
};

const loading = ref(false);

watch(
  () => props.condition,
  async (condition) => {
    // 条件变化了调用列表接口
    getList(condition);
  },
  {
    deep: true,
  },
);

onMounted(() => {
  getList(props.condition);
});

const getList = async (condition: ILoadBalanceDeviceCondition) => {
  if (!condition.account_id) return;
  try {
    loading.value = true;
    const { list } = await loadBalancerListenerStore.getDeviceListenerList(condition, currentGlobalBusinessId.value);
    list.forEach((item) => {
      Object.entries(convertFieldIds).forEach(([key, oldKey]) => {
        item[oldKey] = item[key];
      });
    });

    if (list.length > 0) {
      asyncSetRsWeightStat(list);
    }
    listenerList.value = list;
  } catch (e) {
    listenerList.value = [];
  } finally {
    loading.value = false;
    emit('getList');
  }
};
// 新增/编辑监听器
const addSidesliderState = reactive({ isShow: false, isHidden: true, isEdit: false, initialModel: null });
const handleEditListener = async (row: IListenerItem) => {
  details.value = await loadBalancerClbStore.getLoadBalancerDetails(row.lb_id, currentGlobalBusinessId.value);
  Object.assign(addSidesliderState, { isShow: true, isHidden: false, isEdit: true });
  addSidesliderState.initialModel = await loadBalancerListenerStore.getListenerDetails(
    row.id,
    currentGlobalBusinessId.value,
  );
};
const handleAddSidesliderConfirmSuccess = (id?: string) => {
  if (id) {
    handleUpdateListenerSuccess();
    return;
  }
  routerAction.redirect({ query: { ...route.query, _t: Date.now() } });
};
const handleAddSidesliderHidden = () => {
  Object.assign(addSidesliderState, { isShow: false, isHidden: true, isEdit: false, initialModel: null });
};

const batchDeleteDialogState = reactive({ isShow: false, isHidden: true });
const handleBatchDeleteSuccess = () => {
  getList(props.condition);
};

// 详情
const detailsSidesliderState = reactive({ isShow: false, isHidden: true, rowData: null });
const handleUpdateListenerSuccess = () => {
  getList(props.condition);
};

const handleClearSelection = () => {
  dataListRef.value?.clearSelection();
};
</script>

<template>
  <div class="listener-table-container">
    <div class="toolbar">
      <div class="action-container">
        <template v-for="action in actionList" :key="action.value">
          <hcm-auth v-if="action.authSign" :sign="action.authSign()" v-slot="{ noPerm }">
            <action-item :action="action" :disabled="noPerm || action.disabled?.()" />
          </hcm-auth>
          <action-item v-else :action="action" :disabled="action.disabled?.()" />
        </template>
      </div>
    </div>
    <bk-alert class="mb16" theme="warning" closable v-if="moreData">
      <template #title>
        <span class="mr5">{{ t(`当前操作的监听器数量超过${max}个，批量变更时间可能较长，建议减少操作的数量`) }}</span>
        <bk-button text theme="primary" @click="handleClearSelection">{{ t('一键清空') }}</bk-button>
      </template>
    </bk-alert>
    <data-list
      class="data-list"
      ref="dataListRef"
      v-bkloading="{ loading }"
      :columns="dataListColumns"
      :list="listenerList"
      :pagination="pagination"
      :remote-pagination="false"
      has-selection
      :across-page="true"
      :max-height="`calc(100% - ${moreData ? '96px' : '48px'})`"
      @select-all="handleSelectAll"
      @selection-change="handleSelectChange"
    >
      <template #action>
        <bk-table-column :label="t('操作')" width="120" fixed="right">
          <template #default="{ row }">
            <hcm-auth :sign="clbOperationAuthSign" v-slot="{ noPerm }">
              <bk-button theme="primary" text :disabled="noPerm" @click="handleEditListener(row)">
                {{ t('编辑') }}
              </bk-button>
            </hcm-auth>
            <hcm-auth :sign="clbOperationAuthSign" v-slot="{ noPerm }">
              <bk-button
                class="ml8"
                theme="primary"
                text
                :disabled="noPerm || row.non_zero_weight_count !== 0"
                v-bk-tooltips="{
                  content: t('监听器RS的权重不为0，不可删除'),
                  disabled: row.non_zero_weight_count === 0,
                }"
                @click="handleSingleDelete(row)"
              >
                {{ t('删除') }}
              </bk-button>
            </hcm-auth>
          </template>
        </bk-table-column>
      </template>
    </data-list>

    <template v-if="!addSidesliderState.isHidden">
      <add-listener-sideslider
        v-model="addSidesliderState.isShow"
        :load-balancer-details="details"
        :is-edit="addSidesliderState.isEdit"
        :initial-model="addSidesliderState.initialModel"
        @confirm-success="handleAddSidesliderConfirmSuccess"
        @hidden="handleAddSidesliderHidden"
      />
    </template>

    <template v-if="!batchDeleteDialogState.isHidden">
      <batch-delete-dialog
        v-model="batchDeleteDialogState.isShow"
        :selections="selections"
        @confirm-success="handleBatchDeleteSuccess"
        @hidden="batchDeleteDialogState.isHidden = true"
      />
    </template>

    <template v-if="!detailsSidesliderState.isHidden">
      <details-sideslider
        v-model="detailsSidesliderState.isShow"
        :row-data="detailsSidesliderState.rowData"
        :load-balancer-details="details"
        @update-success="handleUpdateListenerSuccess"
        @hidden="detailsSidesliderState.isHidden = true"
      />
    </template>
  </div>
</template>

<style scoped lang="scss">
.listener-table-container {
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

    .search {
      margin-left: auto;
      width: 500px;
    }
  }

  :deep(.t-table) {
    height: calc(100% - 50px);

    .t-table__content {
      height: calc(100% - 50px);
      overflow-y: auto;
    }
  }
}
</style>
