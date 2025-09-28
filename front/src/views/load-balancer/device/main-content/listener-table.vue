<script setup lang="ts">
import { computed, ComputedRef, h, inject, watch, reactive, ref } from 'vue';
import { useRoute } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { ILoadBalancerDetails, useLoadBalancerClbStore } from '@/store/load-balancer/clb';
import { IListenerItem, useLoadBalancerListenerStore } from '@/store/load-balancer/listener';
import { ListenerDeviceType } from '@/views/load-balancer/constants';
import { ActionItemType } from '@/views/load-balancer/typing';
import { DisplayFieldType, DisplayFieldFactory } from '@/views/load-balancer/children/display/field-factory';
import { ModelPropertyColumn } from '@/model/typings';
import usePage from '@/hooks/use-page';
import { LB_TYPE_MAP, ResourceTypeEnum } from '@/common/constant';
import { IAuthSign } from '@/common/auth-service';
import routerAction from '@/router/utils/action';

import { Button, Message } from 'bkui-vue';
import ActionItem from '@/views/load-balancer/children/action-item.vue';
import AddListenerSideslider from '@/views/load-balancer/listener/add.vue';
import BatchDeleteDialog from '@/views/load-balancer/listener/children/batch-delete-dialog.vue';
import ListenerBatchExportButton from '@/views/load-balancer/children/export/listener-batch-button.vue';
import Confirm from '@/components/confirm';
import DetailsSideslider from '@/views/load-balancer/listener/details.vue';
import BatchCopy from '@/views/load-balancer/device/main-content/children/batch-copy.vue';
import { MENU_BUSINESS_TASK_MANAGEMENT_DETAILS } from '@/constants/menu-symbol';
import { ILoadBalanceDeviceCondition, IDeviceListDataLoadedEvent, DeviceTabEnum } from '../typing';
import { useLoadBalancerDeviceSearchStore } from '@/store/load-balancer/device-search';
import { PrimaryTable, TableColumn } from '@blueking/tdesign-ui';

const props = defineProps<{ condition: ILoadBalanceDeviceCondition }>();
const emit = defineEmits<IDeviceListDataLoadedEvent>();
const details = ref<ILoadBalancerDetails>();
const route = useRoute();
const { t } = useI18n();
const loadBalancerListenerStore = useLoadBalancerListenerStore();
const loadBalancerClbStore = useLoadBalancerClbStore();
const loadBalancerDeviceSearchStore = useLoadBalancerDeviceSearchStore();

const headCheckOptions = [
  {
    id: 'across',
    name: t('跨页全选'),
  },
  {
    id: 'current',
    name: t('当页全选'),
  },
];
const max = 1000;
const LISTENER_ROW_KEY = 'id';

// t组件分页属性
const tablePageProps = ref<{
  pageSize: number;
  current: number;
  total: number;
  showPageNumber: boolean;
  showPageSize: boolean;
  showPreviousAndNextBtn: boolean;
  totalContent: boolean;
}>({
  pageSize: 20,
  current: 1,
  total: 0,
  showPageNumber: false,
  showPageSize: false,
  showPreviousAndNextBtn: false,
  totalContent: false,
});
// 表头checkbox选择框状态
const headCheckBox = reactive({
  checked: false,
  indeterminate: false,
  onChange: (val: boolean) => handleHeadCheckBoxChange(val),
});
// 所有列表的复选框状态 id: boolean
const checkStatus = ref<{ [key: string]: boolean }>({});

const selections = computed(() => {
  const values: IListenerItem[] = [];
  Object.entries(checkStatus.value).forEach(([id, isCheck]) => {
    if (isCheck) {
      const item = listenerList.value.find((item) => item[LISTENER_ROW_KEY] === id);
      if (item) {
        values.push(item);
      }
    }
  });
  return values;
});
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
const moreData = computed(() => selections.value.length > max);

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
  cloud_lb_id: 'lb_cloud_id',
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

const { pagination, getPageParams } = usePage(false);
const listenerList = ref<IListenerItem[]>([]);

const asyncSetRsWeightStat = async (list: IListenerItem[]) => {
  list.forEach((item) => {
    const { non_zero_weight_target_count, target_count } = item;
    Object.assign(item, {
      non_zero_weight_count: non_zero_weight_target_count,
      zero_weight_count: target_count - non_zero_weight_target_count,
      rs_num: target_count,
    });
  });
};

const handleSingleDelete = (row: any) => {
  Confirm('请确定删除监听器', `将删除监听器【${row.name}】`, async () => {
    const res = await loadBalancerListenerStore.batchDeleteListener(
      { ids: [row.id], account_id: row.account_id },
      currentGlobalBusinessId.value,
    );
    if (res.data?.task_management_id) {
      routerAction.open({
        name: MENU_BUSINESS_TASK_MANAGEMENT_DETAILS,
        query: { bizs: currentGlobalBusinessId.value },
        params: { resourceType: ResourceTypeEnum.CLB, id: res.data.task_management_id },
      });
    } else {
      Message({ theme: 'success', message: t('删除成功') });
    }
    getList(props.condition);
  });
};

watch(
  () => route.query,
  () => {
    getList(props.condition);
  },
);

// 全量获取列表数据，内部会根据是否获取过直接返回缓存数据，分页参数理论上不再会用到暂时保留
const getList = async (condition: ILoadBalanceDeviceCondition, pageParams = { sort: '', order: 'DESC' }) => {
  if (!condition.account_id) return;
  try {
    const { list, count } = await loadBalancerDeviceSearchStore.getListenerList(
      condition,
      getPageParams(pagination, pageParams),
      currentGlobalBusinessId.value,
    );
    list.forEach((item) => {
      Object.entries(convertFieldIds).forEach(([key, oldKey]) => {
        item[oldKey] = item[key];
      });
      checkStatus.value[item[LISTENER_ROW_KEY]] = false;
    });

    if (list.length > 0) {
      asyncSetRsWeightStat(list);
    }
    listenerList.value = list;
    tablePageProps.value.total = count;
    pagination.count = count;
  } catch (e) {
    listenerList.value = [];
  } finally {
    handleClearSelection();
    emit('list-data-loaded', DeviceTabEnum.LISTENER, {
      type: 'listenerCount',
      data: {
        count: pagination.count,
      },
    });
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

// 设置表头选择框状态
const setHeadCheckStatus = () => {
  const { current, pageSize, total } = tablePageProps.value;
  let checkLength = 0;
  for (let i = (current - 1) * pageSize, count = 0; count < pageSize; i++, count++) {
    if (!listenerList.value[i]) break;
    if (checkStatus.value[listenerList.value[i][LISTENER_ROW_KEY]]) {
      checkLength = checkLength + 1;
    }
  }
  if (checkLength === 0) {
    headCheckBox.checked = false;
    headCheckBox.indeterminate = false;
  } else {
    const lastPage = Math.ceil(total / pageSize);
    const limit = current === lastPage ? total - pageSize * (lastPage - 1) : pageSize;
    headCheckBox.checked = true;
    if (checkLength < limit) headCheckBox.indeterminate = true;
    else headCheckBox.indeterminate = false;
  }
};
const handleClearSelection = () => {
  listenerList.value.forEach((item) => {
    checkStatus.value[item[LISTENER_ROW_KEY]] = false;
  });
  headCheckBox.checked = false;
  headCheckBox.indeterminate = false;
};
const handlePageChange = (page: number) => {
  tablePageProps.value.current = page;
  setHeadCheckStatus();
};
const handleLimitChange = (limit: number) => {
  tablePageProps.value.pageSize = limit;
  setHeadCheckStatus();
};
// 选择跨页全选还是当页全选
const handleSelectAcross = (id: string, value = true) => {
  let list: IListenerItem[] = [...listenerList.value];
  initAllCheckStatus();
  if (id === 'current') {
    const { current, pageSize } = tablePageProps.value;
    list = list.splice((current - 1) * pageSize, pageSize);
  }
  list.forEach((item: IListenerItem) => {
    checkStatus.value[item[LISTENER_ROW_KEY]] = value;
  });
  setHeadCheckStatus();
};
// 表头复选框val变化
const handleHeadCheckBoxChange = (val: boolean) => {
  handleSelectAcross('current', val);
};
// 初始化所有选择框的状态为未选择
const initAllCheckStatus = () => {
  listenerList.value.forEach((item) => {
    checkStatus.value[item[LISTENER_ROW_KEY]] = false;
  });
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
    <primary-table :row-key="LISTENER_ROW_KEY" :data="listenerList" :pagination="{ ...tablePageProps }">
      <table-column width="65" col-key="row-select">
        <template #title>
          <bk-dropdown
            class="head-check"
            :popover-options="{
              clickContentAutoHide: true,
              hideIgnoreReference: true,
            }"
          >
            <bk-checkbox v-bind="{ ...headCheckBox }" :immediate-emit-change="false"></bk-checkbox>
            <i class="hcm-icon bkhcm-icon-down-shape arrow-icon" />
            <template #content>
              <bk-dropdown-menu>
                <bk-dropdown-item v-for="item in headCheckOptions" :key="item.id" @click="handleSelectAcross(item.id)">
                  {{ item.name }}
                </bk-dropdown-item>
              </bk-dropdown-menu>
            </template>
          </bk-dropdown>
        </template>
        <template #default="{ row }">
          <bk-checkbox v-model="checkStatus[row[LISTENER_ROW_KEY]]" @change="setHeadCheckStatus"></bk-checkbox>
        </template>
      </table-column>
      <template v-for="column in dataListColumns" :key="column.id">
        <table-column
          :col-key="column.id"
          :title="column.name"
          :sort="column.sort"
          :width="column.width"
          :fixed="column.fixed"
          :ellipsis="column.ellipsis"
        >
          <template #default="{ row }">
            <display-value :property="column" :value="row[column.id]" :display="column?.meta?.display" />
          </template>
        </table-column>
      </template>
      <table-column :title="t('操作')" width="120" fixed="right">
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
      </table-column>
    </primary-table>
    <bk-pagination
      class="listener-pagination"
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
  }

  :deep(.head-check) {
    .bk-dropdown-reference {
      display: flex;
      align-items: center;
      gap: 9px;
    }
  }

  :deep(.listener-pagination) {
    .is-last {
      margin-left: auto;
    }
  }

  :deep(.t-table) {
    height: calc(100% - 80px);

    .t-table__content {
      height: 100%;
      overflow-y: auto;
    }
    .t-table__pagination {
      padding: 0;
    }
  }
}
</style>
