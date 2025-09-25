<script setup lang="ts">
import { computed, ComputedRef, h, inject, reactive, ref, useTemplateRef, watch } from 'vue';
import { useRoute } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { ILoadBalancerDetails } from '@/store/load-balancer/clb';
import { IListenerItem, useLoadBalancerListenerStore } from '@/store/load-balancer/listener';
import { ActiveQueryKey, ClbDetailsTabKey, ListenerActionType } from '../constants';
import { ActionItemType } from '../typing';
import { DisplayFieldType, DisplayFieldFactory } from '../children/display/field-factory';
import { ModelPropertyColumn } from '@/model/typings';
import { ConditionKeyType, SearchConditionFactory } from '../children/search/condition-factory';
import usePage from '@/hooks/use-page';
import useSearchQs from '@/hooks/use-search-qs';
import useTableSelection from '@/hooks/use-table-selection';
import { ISearchSelectValue } from '@/typings';
import { getSimpleConditionBySearchSelect, transformSimpleCondition } from '@/utils/search';
import { ResourceTypeEnum } from '@/common/constant';
import { IAuthSign } from '@/common/auth-service';
import routerAction from '@/router/utils/action';
import useTimeoutPoll from '@/hooks/use-timeout-poll';
import { getInstVip } from '@/utils';

import { Button, Message } from 'bkui-vue';
import { Plus } from 'bkui-vue/lib/icon';
import ActionItem from '../children/action-item.vue';
import Search from '../children/search/search.vue';
import DataList from '../children/display/data-list.vue';
import BindingStatus from './children/binding-status.vue';
import AddListenerSideslider from './add.vue';
import BatchDeleteDialog from './children/batch-delete-dialog.vue';
import ListenerBatchExportButton from '../children/export/listener-batch-button.vue';
import SyncAccountResourceDialog from '@/components/sync-account-resource/index.vue';
import Confirm from '@/components/confirm';
import DetailsSideslider from './details.vue';
import { ValidateValuesFunc } from 'bkui-vue/lib/search-select/utils';

//* 接口请求使用lbId，避免details暂无数据，导致接口报错
const props = defineProps<{ lbId: string; details: ILoadBalancerDetails }>();

let firstIn = true;
const route = useRoute();
const { t } = useI18n();
const loadBalancerListenerStore = useLoadBalancerListenerStore();

const currentGlobalBusinessId = inject<ComputedRef<number>>('currentGlobalBusinessId');
const clbOperationAuthSign = inject<ComputedRef<IAuthSign | IAuthSign[]>>('clbOperationAuthSign');

const actionConfig: Record<ListenerActionType, ActionItemType> = {
  [ListenerActionType.ADD]: {
    type: 'button',
    label: t('新增监听器'),
    value: ListenerActionType.ADD,
    displayProps: { theme: 'primary' },
    prefix: () => h(Plus),
    authSign: () => clbOperationAuthSign.value,
    handleClick: () => {
      addSidesliderState.isHidden = false;
      addSidesliderState.isShow = true;
    },
  },
  [ListenerActionType.BATCH_EXPORT]: {
    value: ListenerActionType.BATCH_EXPORT,
    render: () => h(ListenerBatchExportButton, { selections: selections.value }),
  },
  [ListenerActionType.REMOVE]: {
    type: 'button',
    label: t('批量删除'),
    value: ListenerActionType.REMOVE,
    disabled: () => selections.value.length === 0,
    authSign: () => clbOperationAuthSign.value,
    handleClick: () => {
      batchDeleteDialogState.isHidden = false;
      batchDeleteDialogState.isShow = true;
    },
  },
  [ListenerActionType.SYNC]: {
    type: 'button',
    label: t('同步'),
    value: ListenerActionType.SYNC,
    disabled: () => selections.value.length > 0,
    handleClick: () => {
      syncDialogState.isHidden = false;
      syncDialogState.isShow = true;
    },
  },
};
const listenerActionList = computed<ActionItemType[]>(() => {
  return [
    { value: ListenerActionType.ADD },
    { value: ListenerActionType.BATCH_EXPORT },
    { value: ListenerActionType.REMOVE },
    { value: ListenerActionType.SYNC },
  ];
});
const actionList = computed<ActionItemType[]>(() => {
  return listenerActionList.value.reduce((prev, curr) => {
    const config = actionConfig[curr.value as ListenerActionType];
    if (curr.children) {
      prev.push({
        ...config,
        ...curr,
        children: curr.children.map((childAction) => ({
          ...actionConfig[childAction.value as ListenerActionType],
          ...childAction,
        })),
      });
    } else {
      prev.push({ ...config, ...curr });
    }
    return prev;
  }, []);
});

// data-list
const displayFieldIds = [
  'name',
  'cloud_id',
  'protocol',
  'port',
  'scheduler',
  'rs_num',
  'domain_num',
  'url_num',
  'binding_status',
];
const displayProperties = DisplayFieldFactory.createModel(DisplayFieldType.LISTENER).getProperties();
const displayConfig: Record<string, Partial<ModelPropertyColumn>> = {
  name: {
    render: ({ data, row }) => {
      const handleClick = () => {
        showListenerDetail(data);
      };
      return h(Button, { theme: 'primary', text: true, onClick: handleClick }, row.name);
    },
  },
  port: {
    render: ({ row, cell }) => `${cell}${row.end_port ? `-${row.end_port}` : ''}`,
  },
  binding_status: {
    render: ({ row, cell }) => {
      return h(BindingStatus, { value: cell, type: 'listener', protocol: row.protocol });
    },
  },
};
const dataListColumns = displayFieldIds.map((id) => {
  const property = displayProperties.find((field) => field.id === id);
  return { ...property, ...displayConfig[id] };
});

const conditionIds = ['name', 'cloud_id', 'protocol', 'port'];
const conditionProperties = SearchConditionFactory.createModel(ConditionKeyType.LISTENER).getProperties();
const searchFields = conditionIds.map((id) => {
  const property = conditionProperties.find((field) => field.id === id);
  return { ...property };
});
const validateValues: ValidateValuesFunc = async (item, values) => {
  if (!item) return '请选择条件';
  if (item.id === 'port') {
    const port = parseInt(values[0].id, 10);
    return port >= 1 && port <= 65535 ? true : '端口范围为1-65535';
  }
  return true;
};

const { pagination, getPageParams } = usePage();
const searchQs = useSearchQs({ key: 'filter', properties: conditionProperties });

const condition = ref<Record<string, any>>({});
const listenerList = ref<IListenerItem[]>([]);

const showListenerDetail = (data: any) => {
  detailsSidesliderState.isHidden = false;
  detailsSidesliderState.isShow = true;
  detailsSidesliderState.rowData = data;
};
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

const taskPoll = useTimeoutPoll(
  () => {
    routerAction.redirect({ query: { ...route.query, _t: Date.now() } });
  },
  30000,
  { max: 10 },
);

const hasBindingStatus = (list: IListenerItem[]) => {
  return list.some((item) => item.binding_status === 'binding');
};

const asyncSetRsWeightStat = async (list: IListenerItem[]) => {
  const ids = list.map((item) => item.id);
  const map = await loadBalancerListenerStore.getListenersRsWeightStat(ids, currentGlobalBusinessId.value);
  listenerList.value.forEach((item) => {
    const { non_zero_weight_count, zero_weight_count, total_count: totalCount } = map[item.id];
    Object.assign(item, { non_zero_weight_count, zero_weight_count, rs_num: totalCount });
  });
};

const injectLoadBalancerFields = (list: IListenerItem[]) => {
  if (!props.details) return;
  list.forEach((item) => {
    Object.assign(item, { lb_cloud_id: props.details.cloud_id, lb_vip: getInstVip(props.details) });
  });
};

const handleSingleDelete = (row: any) => {
  Confirm('请确定删除监听器', `将删除监听器【${row.name}】`, async () => {
    await loadBalancerListenerStore.batchDeleteListener(
      { ids: [row.id], account_id: props.details.account_id },
      currentGlobalBusinessId.value,
    );
    Message({ theme: 'success', message: '删除成功' });
    routerAction.redirect({ query: { ...route.query, _t: Date.now() } });
  });
};

const searchRef = useTemplateRef<typeof Search>('search');
watch(
  () => props.lbId,
  () => searchRef.value?.clear(false),
);
const handleSearch = (val: ISearchSelectValue) => {
  searchQs.set(
    getSimpleConditionBySearchSelect(val, [
      {
        field: 'port',
        // 使用 buildSearchSelectValueBySearchQsCondition 处理后回填的值为数组 [1, 2]，在调用formatter时这里接收的是每一个元素的值，是number类型
        // 手动输入时这里接收的是原始值是string类型 '1 | 2'，所以这里根据不同的值类型来处理，目前整个数据流加工的数据如此这是局部的解决方式
        formatter: (value: string | number) => (typeof value === 'string' ? value?.split(' | ') : value),
      },
    ]),
  );
};

const loading = ref(false);
watch(
  () => route.query,
  async (query) => {
    // 避免多次路由导致多次请求
    if (query[ActiveQueryKey.DETAILS] && ClbDetailsTabKey.LISTENER !== query[ActiveQueryKey.DETAILS]) return;
    const { detailShow } = query;
    condition.value = searchQs.get(query, {});
    pagination.current = Number(query.page) || 1;
    pagination.limit = Number(query.limit) || pagination.limit;

    const sort = (query.sort || 'created_at') as string;
    const order = (query.order || 'DESC') as string;

    loading.value = true;
    try {
      const { list, count } = await loadBalancerListenerStore.getListenerList(
        props.lbId,
        {
          filter: transformSimpleCondition(condition.value, conditionProperties),
          page: getPageParams(pagination, { sort, order }),
        },
        currentGlobalBusinessId.value,
      );

      listenerList.value = list;
      pagination.count = count;

      // 重置勾选项
      selections.value = [];

      if (list.length > 0) {
        asyncSetRsWeightStat(list);
        injectLoadBalancerFields(list);
        if (firstIn && detailShow) {
          showListenerDetail(list[0]);
          firstIn = false;
        }
      }

      if (hasBindingStatus(list)) {
        taskPoll.resume();
      } else {
        taskPoll.pause();
      }
    } finally {
      loading.value = false;
    }
  },
  { immediate: true },
);

// 新增/编辑监听器
const addSidesliderState = reactive({ isShow: false, isHidden: true, isEdit: false, initialModel: null });
const handleEditListener = async (row: IListenerItem) => {
  Object.assign(addSidesliderState, { isShow: true, isHidden: false, isEdit: true });
  addSidesliderState.initialModel = await loadBalancerListenerStore.getListenerDetails(
    row.id,
    currentGlobalBusinessId.value,
  );
};
const handleAddSidesliderConfirmSuccess = (id?: string) => {
  if (id) {
    handleUpdateListenerSuccess(id);
    return;
  }
  routerAction.redirect({ query: { ...route.query, _t: Date.now() } });
};
const handleAddSidesliderHidden = () => {
  Object.assign(addSidesliderState, { isShow: false, isHidden: true, isEdit: false, initialModel: null });
};

const batchDeleteDialogState = reactive({ isShow: false, isHidden: true });
const handleBatchDeleteSuccess = () => {
  routerAction.redirect({ query: { ...route.query, _t: Date.now() } });
};

// 同步
const syncDialogState = reactive({ isShow: false, isHidden: true });

// 详情
const detailsSidesliderState = reactive({ isShow: false, isHidden: true, rowData: null });
const handleUpdateListenerSuccess = async (id: string) => {
  const { query } = route;
  const sort = (query.sort || 'created_at') as string;
  const order = (query.order || 'DESC') as string;

  const { list } = await loadBalancerListenerStore.getListenerList(
    props.lbId,
    {
      filter: transformSimpleCondition({ ...condition.value, id }, conditionProperties),
      page: getPageParams(pagination, { sort, order }),
    },
    currentGlobalBusinessId.value,
  );

  const [newRow] = list;
  listenerList.value.forEach((item) => {
    if (item.id === id) {
      Object.assign(item, newRow);
    }
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
      <search
        class="search"
        ref="search"
        :fields="searchFields"
        :condition="condition"
        :validate-values="validateValues"
        @search="handleSearch"
      />
    </div>
    <data-list
      class="data-list"
      v-bkloading="{ loading }"
      :columns="dataListColumns"
      :list="listenerList"
      :pagination="pagination"
      has-selection
      :max-height="`calc(100% - 48px)`"
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

    <template v-if="!syncDialogState.isHidden">
      <sync-account-resource-dialog
        v-model="syncDialogState.isShow"
        :title="t('同步当前负载均衡')"
        :desc="t('从云上同步该负载均衡数据，包括负载均衡基本信息，监听器等')"
        :resource-type="ResourceTypeEnum.CLB"
        :business-id="currentGlobalBusinessId"
        resource-name="load_balancer"
        :initial-model="{
          name: details.name,
          account_id: details.account_id,
          vendor: details.vendor,
          regions: details.region,
          cloud_ids: [details.cloud_id],
        }"
        @hidden="syncDialogState.isHidden = true"
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
  padding: 16px 24px;

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
}
</style>
