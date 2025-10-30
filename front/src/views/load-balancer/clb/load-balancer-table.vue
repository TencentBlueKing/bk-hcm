<script setup lang="ts">
import { computed, h, inject, reactive, ref, Ref, watch } from 'vue';
import { useRoute } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { useRegionStore } from '@/store/region';
import { ILoadBalancerWithDeleteProtectionItem, useLoadBalancerClbStore } from '@/store/load-balancer/clb';
import { LB_TYPE_NAME, LoadBalancerActionType, LoadBalancerType } from '../constants';
import { ActionItemType } from '../typing';
import {
  AUTH_BIZ_CREATE_CLB,
  AUTH_BIZ_DELETE_CLB,
  AUTH_BIZ_UPDATE_CLB,
  AUTH_CREATE_CLB,
  AUTH_DELETE_CLB,
  AUTH_UPDATE_CLB,
} from '@/constants/auth-symbols';
import { DisplayFieldFactory, DisplayFieldType } from '../children/display/field-factory';
import { ConditionKeyType, SearchConditionFactory } from '../children/search/condition-factory';
import usePage from '@/hooks/use-page';
import useSearchQs from '@/hooks/use-search-qs';
import useTableSelection from '@/hooks/use-table-selection';
import { ISearchItem, ValidateValuesFunc } from 'bkui-vue/lib/search-select/utils';
import { formatBandwidth, getAuthSignByBusinessId, getInstVip, parseIP } from '@/utils';
import { ISearchSelectValue } from '@/typings';
import {
  buildMultipleValueRulesItem,
  getSimpleConditionBySearchSelect,
  transformSimpleCondition,
} from '@/utils/search';
import { GLOBAL_BIZS_KEY, ResourceTypeEnum } from '@/common/constant';
import { ModelPropertyColumn, ModelPropertySearch } from '@/model/typings';
import routerAction from '@/router/utils/action';
import { MENU_BUSINESS_LOAD_BALANCER_APPLY, MENU_BUSINESS_LOAD_BALANCER_DETAILS } from '@/constants/menu-symbol';
import { parseTimeFromNow } from '@/common/util';

import { Button, Message, Tag } from 'bkui-vue';
import ActionItem from '../children/action-item.vue';
import BatchCopy from './children/batch-copy.vue';
import Search from '../children/search/search.vue';
import DataList from '../children/display/data-list.vue';
import BatchDeleteDialog from './children/batch-delete-dialog.vue';
import BatchImportSideslider from './children/batch-import/index.vue';
import BatchExportButton from '../children/export/batch-button.vue';
import SingleExportButton from '../children/export/single-button.vue';
import SyncAccountResourceDialog from '@/components/sync-account-resource/index.vue';
import Confirm from '@/components/confirm';
import HoverCopy from '@/components/copy-to-clipboard/hover-copy.vue';

defineOptions({ name: 'load-balancer-table' });

const emit = defineEmits<(e: 'delete-listener') => void>();

const route = useRoute();
const { t } = useI18n();
const { getAllVendorRegion } = useRegionStore();
const loadBalancerClbStore = useLoadBalancerClbStore();

const currentGlobalBusinessId = inject<Ref<number>>('currentGlobalBusinessId');
const isBusinessPage = computed(() => currentGlobalBusinessId.value);

// 操作的基础配置，这里作打平处理，支持直接通过value索引访问
const actionConfig: Record<LoadBalancerActionType, ActionItemType> = {
  [LoadBalancerActionType.PURCHASE]: {
    type: 'button',
    label: t('购买'),
    value: LoadBalancerActionType.PURCHASE,
    displayProps: { theme: 'primary' },
    handleClick: () =>
      routerAction.redirect({
        name: MENU_BUSINESS_LOAD_BALANCER_APPLY,
        query: { [GLOBAL_BIZS_KEY]: currentGlobalBusinessId.value },
      }),
  },
  [LoadBalancerActionType.BATCH_OPERATION]: {
    type: 'dropdown',
    label: t('批量操作'),
    value: LoadBalancerActionType.BATCH_OPERATION,
  },
  [LoadBalancerActionType.CREATE_LISTENER_OR_RULES]: {
    label: t('批量导入-监听器及规则'),
    value: LoadBalancerActionType.CREATE_LISTENER_OR_RULES,
    handleClick: () => {
      batchImportSidesliderState.isShow = true;
      batchImportSidesliderState.action = LoadBalancerActionType.CREATE_LISTENER_OR_RULES;
    },
  },
  [LoadBalancerActionType.BIND_RS]: {
    label: t('批量导入-绑定RS'),
    value: LoadBalancerActionType.BIND_RS,
    handleClick: () => {
      batchImportSidesliderState.isShow = true;
      batchImportSidesliderState.action = LoadBalancerActionType.BIND_RS;
    },
  },
  [LoadBalancerActionType.REMOVE]: {
    label: t('批量删除'),
    value: LoadBalancerActionType.REMOVE,
    handleClick: () => {
      batchDeleteDialogState.isHidden = false;
      batchDeleteDialogState.isShow = true;
    },
    disabled: () => selections.value.length === 0,
  },
  [LoadBalancerActionType.SYNC]: {
    type: 'button',
    label: t('同步'),
    value: LoadBalancerActionType.SYNC,
    disabled: () => selections.value.length > 0,
    handleClick: () => {
      syncDialogState.isHidden = false;
      syncDialogState.isShow = true;
    },
  },
  [LoadBalancerActionType.COPY]: {
    label: t('复制'),
    value: LoadBalancerActionType.COPY,
    render: () => h(BatchCopy, { selections: selections.value }),
  },
  [LoadBalancerActionType.BATCH_EXPORT]: {
    value: LoadBalancerActionType.BATCH_EXPORT,
    render: () => h(BatchExportButton, { selections: selections.value, bizId: currentGlobalBusinessId.value }),
  },
};
const loadBalancerActionList = computed<ActionItemType[]>(() => {
  const list: ActionItemType[] = [
    {
      value: LoadBalancerActionType.PURCHASE,
      authSign: () => getAuthSignByBusinessId(currentGlobalBusinessId.value, AUTH_CREATE_CLB, AUTH_BIZ_CREATE_CLB),
      index: 0,
    },
    { value: LoadBalancerActionType.SYNC, index: 2 },
    { value: LoadBalancerActionType.COPY, index: 3 },
  ];

  // 业务下需要将批量操作收拢
  if (isBusinessPage.value) {
    list.push({
      value: LoadBalancerActionType.BATCH_OPERATION,
      children: [
        {
          value: LoadBalancerActionType.CREATE_LISTENER_OR_RULES,
          authSign: () => getAuthSignByBusinessId(currentGlobalBusinessId.value, AUTH_UPDATE_CLB, AUTH_BIZ_UPDATE_CLB),
        },
        {
          value: LoadBalancerActionType.BIND_RS,
          authSign: () => getAuthSignByBusinessId(currentGlobalBusinessId.value, AUTH_UPDATE_CLB, AUTH_BIZ_UPDATE_CLB),
        },
        {
          value: LoadBalancerActionType.BATCH_EXPORT,
        },
        {
          value: LoadBalancerActionType.REMOVE,
          authSign: () => getAuthSignByBusinessId(currentGlobalBusinessId.value, AUTH_DELETE_CLB, AUTH_BIZ_DELETE_CLB),
        },
      ],
      index: 1,
    });
  } else {
    list.push({
      value: LoadBalancerActionType.REMOVE,
      type: 'button',
      authSign: () => getAuthSignByBusinessId(currentGlobalBusinessId.value, AUTH_DELETE_CLB, AUTH_BIZ_DELETE_CLB),
      index: 1,
    });
  }

  return list.sort((a, b) => (a.index ?? 0) - (b.index ?? 0));
});
const actionList = computed<ActionItemType[]>(() => {
  return loadBalancerActionList.value.reduce((prev, curr) => {
    const config = actionConfig[curr.value as LoadBalancerActionType];
    if (curr.children) {
      prev.push({
        ...config,
        ...curr,
        children: curr.children.map((childAction) => ({
          ...actionConfig[childAction.value as LoadBalancerActionType],
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
const displayFieldProperties = DisplayFieldFactory.createModel(DisplayFieldType.CLB).getProperties();
const displayFieldConfig: Record<string, Partial<ModelPropertyColumn>> = {
  name: {
    render: ({ row }) => {
      const handleClick = () => {
        routerAction.redirect(
          {
            name: MENU_BUSINESS_LOAD_BALANCER_DETAILS,
            query: { [GLOBAL_BIZS_KEY]: currentGlobalBusinessId.value },
            params: { id: row.id },
          },
          { history: true },
        );
      };
      return h(
        HoverCopy,
        { content: row.name },
        h(Button, { theme: 'primary', text: true, onClick: handleClick }, row.name),
      );
    },
  },
  cloud_id: { render: ({ cell }) => h(HoverCopy, { content: cell }) },
  domain: { render: ({ cell }) => h(HoverCopy, { content: cell }) },
  lb_vip: {
    render: ({ row }) => h(HoverCopy, { content: getInstVip(row) }),
  },
  lb_type: {
    render: ({ cell }) => {
      return h(
        Tag,
        { radius: '11px', class: ['lb-type-tag', cell === LoadBalancerType.OPEN ? 'is-open' : 'is-internal'] },
        LB_TYPE_NAME[cell as LoadBalancerType],
      );
    },
    filter: {
      list: Object.entries(LB_TYPE_NAME).map(([value, label]) => ({ value, label, text: label })),
    },
  },
  delete_protect: {
    render: ({ cell }) => {
      return h(Tag, { theme: cell ? 'success' : undefined }, cell ? '开启' : '关闭');
    },
  },
  bandwidth: {
    render: ({ cell }) => formatBandwidth(cell),
  },
  sync_time: {
    render: ({ cell }) => parseTimeFromNow(cell),
  },
};
const dataListColumns = computed(() => {
  const properties = displayFieldProperties.map((field) => ({ ...field, ...displayFieldConfig[field.id] }));
  return isBusinessPage.value ? properties.filter((field) => field.id !== 'bk_biz_id') : properties;
});

const conditionConfig: Record<string, Partial<ModelPropertySearch>> = {
  cloud_id: {
    meta: { search: { filterRules: (value) => buildMultipleValueRulesItem('cloud_id', value) } },
  },
};
const conditionProperties = SearchConditionFactory.createModel(ConditionKeyType.CLB)
  .getProperties()
  .map((field) => ({ ...field, ...conditionConfig[field.id] }));
const searchFields = computed(() =>
  isBusinessPage.value ? conditionProperties.filter((field) => field.id !== 'bk_biz_id') : conditionProperties,
);
const getMenuList = (_item: any, values: any) => getAllVendorRegion(values);
const searchDataConfig = computed<Record<string, Partial<ISearchItem>>>(() => {
  return searchFields.value.reduce<Record<string, Partial<ISearchItem>>>((prev, curr) => {
    if (curr.id === 'region') {
      prev[curr.id] = { async: true };
    } else {
      prev[curr.id] = { async: false };
    }
    return prev;
  }, {});
});

const { pagination, getPageParams } = usePage();
const searchQs = useSearchQs({ key: 'filter', properties: conditionProperties });

const condition = ref<Record<string, any>>({});
const loadBalancerList = ref<ILoadBalancerWithDeleteProtectionItem[]>([]);

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

const getSingleDeleteDisabledTooltips = (row: any, noPerm: boolean) => {
  if (noPerm) {
    return { disabled: true };
  }
  if (row.listener_count > 0) {
    return { content: '该负载均衡已绑定监听器, 不可删除', disabled: !(row.listener_count > 0) };
  }
  return { content: t('该负载均衡已开启删除保护, 不可删除'), disabled: !row.delete_protect };
};

const handleSingleDelete = (row: any) => {
  Confirm('请确定删除负载均衡', `将删除负载均衡【${row.name}】`, async () => {
    await loadBalancerClbStore.batchDeleteLoadBalancer({ ids: [row.id] }, currentGlobalBusinessId.value);
    Message({ message: '删除成功', theme: 'success' });
    routerAction.redirect({ query: { ...route.query, _t: Date.now() } });
    emit('delete-listener');
  });
};

const asyncQueryListenerCount = async (list: ILoadBalancerWithDeleteProtectionItem[]) => {
  if (!list || list.length === 0) return;
  const ids = list.map((item) => item.id);
  const listenerCountDetails = await loadBalancerClbStore.getListenerCountByLoadBalancerIds(
    ids,
    currentGlobalBusinessId.value,
  );
  loadBalancerList.value.forEach((lb) => {
    const listenerCountDetail = listenerCountDetails.find((item) => item.lb_id === lb.id);
    if (listenerCountDetail) {
      lb.listener_count = listenerCountDetail.num;
    } else {
      lb.listener_count = 0;
    }
  });
};

const validateValues: ValidateValuesFunc = async (item, values) => {
  if (!item) return '请选择条件';
  if ('lb_vip' === item.id) {
    const { IPv4List, IPv6List } = parseIP(values[0].id);
    return Boolean(IPv4List.length || IPv6List.length) ? true : 'IP格式有误';
  }
  return true;
};

const handleSearch = (val: ISearchSelectValue) => {
  searchQs.set(getSimpleConditionBySearchSelect(val));
};

const isLoading = ref(false);
watch(
  () => route.query,
  async (query) => {
    condition.value = searchQs.get(query, {});

    pagination.current = Number(query.page) || 1;
    pagination.limit = Number(query.limit) || pagination.limit;

    const sort = (query.sort || 'created_at') as string;
    const order = (query.order || 'DESC') as string;

    isLoading.value = true;
    try {
      const { list, count } = await loadBalancerClbStore.getLoadBalancerListWithDeleteProtection(
        {
          filter: transformSimpleCondition(condition.value, conditionProperties),
          page: getPageParams(pagination, { sort, order }),
        },
        currentGlobalBusinessId.value,
        false,
      );

      loadBalancerList.value = list;
      pagination.count = count;

      // 清空选中
      selections.value = [];

      asyncQueryListenerCount(list);
    } catch (error) {
      console.error(error);
      loadBalancerList.value = [];
      pagination.count = 0;
    } finally {
      isLoading.value = false;
    }
  },
  { immediate: true },
);

const batchDeleteDialogState = reactive({ isShow: false, isHidden: true });
const handleBatchDeleteSuccess = () => {
  routerAction.redirect({ query: { ...route.query, _t: Date.now() } });
};

// 批量导入
const batchImportSidesliderState = reactive({ isShow: false, action: undefined });
// 同步
const syncDialogState = reactive({ isShow: false, isHidden: true });
</script>

<template>
  <div class="load-balancer-table-container">
    <section class="panel">
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
          :fields="searchFields"
          :condition="condition"
          :validate-values="validateValues"
          :get-menu-list="getMenuList"
          :search-data-config="searchDataConfig"
          @search="handleSearch"
        />
      </div>
      <data-list
        class="data-list"
        v-bkloading="{ loading: isLoading }"
        :columns="dataListColumns"
        :list="loadBalancerList"
        :pagination="pagination"
        has-selection
        :max-height="`calc(100% - 48px)`"
        @select-all="handleSelectAll"
        @selection-change="handleSelectChange"
      >
        <template #action>
          <bk-table-column :label="t('操作')" width="120" fixed="right" :show-overflow-tooltip="false">
            <template #default="{ row }">
              <div class="operation-cell">
                <single-export-button :data="row" :biz-id="currentGlobalBusinessId" />
                <hcm-auth
                  :sign="getAuthSignByBusinessId(currentGlobalBusinessId, AUTH_DELETE_CLB, AUTH_BIZ_DELETE_CLB)"
                  v-slot="{ noPerm }"
                >
                  <bk-button
                    theme="primary"
                    text
                    :disabled="noPerm || row.listener_count > 0 || row.delete_protect"
                    v-bk-tooltips="getSingleDeleteDisabledTooltips(row, noPerm)"
                    @click="handleSingleDelete(row)"
                  >
                    {{ t('删除') }}
                  </bk-button>
                </hcm-auth>
              </div>
            </template>
          </bk-table-column>
        </template>
      </data-list>
    </section>

    <template v-if="!batchDeleteDialogState.isHidden">
      <batch-delete-dialog
        v-model="batchDeleteDialogState.isShow"
        :selections="selections"
        @confirm-success="handleBatchDeleteSuccess"
        @hidden="batchDeleteDialogState.isHidden = true"
      />
    </template>

    <batch-import-sideslider
      v-model="batchImportSidesliderState.isShow"
      :active-action="batchImportSidesliderState.action"
    />

    <template v-if="!syncDialogState.isHidden">
      <sync-account-resource-dialog
        v-model="syncDialogState.isShow"
        :title="t('同步负载均衡')"
        :desc="t('从云上同步该业务的所有负载均衡数据，包括负载均衡，监听器等')"
        :resource-type="ResourceTypeEnum.CLB"
        :business-id="currentGlobalBusinessId"
        resource-name="load_balancer"
        @hidden="syncDialogState.isHidden = true"
      />
    </template>
  </div>
</template>

<style scoped lang="scss">
.load-balancer-table-container {
  height: 100%;
  padding: 24px;
  background: #f5f7fa;

  .panel {
    height: 100%;
    padding: 16px 24px;
    background: #fff;
    box-shadow: 0 2px 4px 0 #1919290d;
    border-radius: 2px;
  }

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

  .data-list {
    :deep(.lb-type-tag) {
      &.is-open {
        background-color: #d8edd9;
      }

      &.is-internal {
        background-color: #fff2c9;
      }
    }

    .operation-cell {
      display: flex;
      align-items: center;
      gap: 10px;
    }
  }
}
</style>
