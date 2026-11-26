<script setup lang="ts">
import { computed, ComputedRef, h, inject, ref, useTemplateRef, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute } from 'vue-router';
import { DisplayFieldType, DisplayFieldFactory } from '@/views/load-balancer/children/display/field-factory';
import { ModelPropertyColumn } from '@/model/typings';
import usePage from '@/hooks/use-page';
import useTableSelection from '@/hooks/use-table-selection';
import DataList from '@/views/load-balancer/children/display/data-list.vue';
import { ILoadBalanceDeviceCondition, DeviceTabEnum, IDeviceListDataLoadedEvent } from '../typing';
import { Share } from 'bkui-vue/lib/icon';
import routerAction from '@/router/utils/action';
import { MENU_BUSINESS_LOAD_BALANCER_DETAILS } from '@/constants/menu-symbol';
import qs from 'qs';
import { useLoadBalancerDeviceSearchStore, type IUrlRuleItem } from '@/store/load-balancer/device-search';
import BatchCopy from '@/views/load-balancer/device/main-content/children/batch-copy.vue';

const props = defineProps<{ condition: ILoadBalanceDeviceCondition }>();
const emit = defineEmits<IDeviceListDataLoadedEvent>();

const route = useRoute();
const { t } = useI18n();
const loadBalancerDeviceSearchStore = useLoadBalancerDeviceSearchStore();
const currentGlobalBusinessId = inject<ComputedRef<number>>('currentGlobalBusinessId');
const dataListRef = useTemplateRef<typeof DataList>('dataList');
const ruleUrlList = ref<IUrlRuleItem[]>([]);

const { selections, resetSelections, handleSelectAll, handleSelectChange } = useTableSelection({
  isRowSelectable: () => true,
});

const getArrayDisplayValue = (value: string[]) => {
  const values = Array.isArray(value) ? value : [value];
  return values.join(', ') || '--';
};
const urlRuleCopyItems = computed(() => [
  {
    text: t('URL'),
    content: selections.value.map((item: IUrlRuleItem) => item.rule_url || '--').join('\n'),
  },
  {
    text: t('监听器域名'),
    content: selections.value.map((item: IUrlRuleItem) => getArrayDisplayValue(item.rule_domain)).join('\n'),
  },
]);

// data-list
const displayFieldIds = ['lb_vips', 'lbl_protocol', 'lbl_port', 'rule_url', 'rule_domain', 'target_count'];
const displayProperties = DisplayFieldFactory.createModel(DisplayFieldType.URL).getProperties();
const displayConfig: Record<string, Partial<ModelPropertyColumn>> = {
  lbl_port: {
    render: ({ row, cell }) => {
      const handleClick = () => {
        const filter = qs.stringify(
          {
            cloud_id: row.cloud_lbl_id,
          },
          {
            arrayFormat: 'comma',
            encode: false,
            allowEmptyArrays: true,
          },
        );
        routerAction.open({
          name: MENU_BUSINESS_LOAD_BALANCER_DETAILS,
          params: {
            bizId: currentGlobalBusinessId.value,
            id: row.lb_id,
          },
          query: {
            filter,
            detailShow: true,
          },
        });
      };
      return h('div', { onClick: handleClick, class: 'port' }, [h('span', {}, cell), h(Share, { class: 'share' })]);
    },
  },
};
const dataListColumns = displayFieldIds.map((id) => {
  const property = displayProperties.find((field) => field.id === id);
  return { ...property, ...displayConfig[id] };
});

const { pagination, getPageParams } = usePage();
pagination['limit-list'] = [10, 20, 50, 100, 500];

const handleClearSelection = () => {
  resetSelections();
  dataListRef.value?.clearSelection();
};
const handleUrlRuleSelectAll = (selection: Parameters<typeof handleSelectAll>[0]) => {
  handleSelectAll({
    ...selection,
    data: dataListRef.value?.getSelection() ?? [],
  });
};
const getList = async (condition: ILoadBalanceDeviceCondition, pageParams = { sort: 'created_at', order: 'DESC' }) => {
  if (!condition.account_id) return;
  handleClearSelection();
  try {
    const { list, count } = await loadBalancerDeviceSearchStore.getUrlRuleList(
      condition,
      getPageParams(pagination, pageParams),
      currentGlobalBusinessId.value,
    );
    ruleUrlList.value = list;
    pagination.count = count;
  } catch (error) {
    console.error(error);
    ruleUrlList.value = [];
    pagination.count = 0;
  } finally {
    emit('list-data-loaded', DeviceTabEnum.URL, {
      type: 'urlCount',
      data: {
        count: pagination.count,
      },
    });
  }
};

watch(
  () => route.query,
  async (query) => {
    pagination.current = Number(query.page) || 1;
    pagination.limit = Number(query.limit) || pagination.limit;

    const sort = (query.sort || 'created_at') as string;
    const order = (query.order || 'DESC') as string;

    getList(props.condition, { sort, order });
  },
);
</script>

<template>
  <div class="url-table-container">
    <div class="toolbar">
      <BatchCopy :items="urlRuleCopyItems" :disabled="selections.length === 0" />
    </div>
    <DataList
      ref="dataList"
      class="data-list"
      v-bkloading="{ loading: loadBalancerDeviceSearchStore.urlRuleListLoading }"
      :columns="dataListColumns"
      :list="ruleUrlList"
      :pagination="pagination"
      :max-height="`calc(100% - 48px)`"
      :enable-query="false"
      :remote-pagination="false"
      has-selection
      @page-limit-change="handleClearSelection"
      @page-value-change="handleClearSelection"
      @select-all="handleUrlRuleSelectAll"
      @selection-change="handleSelectChange"
    />
  </div>
</template>

<style scoped lang="scss">
.url-table-container {
  height: 100%;

  .toolbar {
    display: flex;
    align-items: center;
    margin-bottom: 16px;
  }

  :deep(.port) {
    .share {
      color: #3a84ff;
      margin-left: 6px;
      vertical-align: middle;
      display: none !important;
      cursor: pointer;
    }

    &:hover {
      .share {
        display: inline-flex !important;
      }
    }
  }
}
</style>
