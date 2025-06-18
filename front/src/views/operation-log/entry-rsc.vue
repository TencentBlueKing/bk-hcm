<script setup lang="ts">
import { computed, watch, ref, reactive, provide } from 'vue';
import { useRouter, useRoute } from 'vue-router';
import { ResourceTypeEnum } from '@/common/resource-constant';
import useSearchQs from '@/hooks/use-search-qs';
import usePage from '@/hooks/use-page';
import { transformSimpleCondition, getDateRange } from '@/utils/search';
import { IAuditItem, useAuditStore } from '@/store/audit';
import routerAction from '@/router/utils/action';
import { MENU_RESOURCE_OPERATION_LOG_DETAILS } from '@/constants/menu-symbol';
import { GLOBAL_BIZS_KEY } from '@/common/constant';
import { type ISearchCondition } from './typings';
import { SearchConditionFactory } from './children/search/condition-factory';
import { TableColumnFactory } from './children/data-list/column-factory';
import { properties as detailFields } from './details/field';
import Search from './children/search/search.vue';
import DataList from './children/data-list/data-list.vue';
import DetailSlider from './details/slider/index.vue';

const router = useRouter();
const route = useRoute();
const auditStore = useAuditStore();

const { pagination, getPageParams } = usePage();

const tabPanels = [
  { name: 'all', label: '全部' },
  { name: ResourceTypeEnum.SECURITY_GROUP, label: '安全组' },
  { name: ResourceTypeEnum.CLB, label: '负载均衡' },
];

const tabActive = computed({
  get() {
    return (route.query.tab || tabPanels[0].name) as ResourceTypeEnum & 'all';
  },
  set(value) {
    router.push({
      query: { [GLOBAL_BIZS_KEY]: route.query?.[GLOBAL_BIZS_KEY], tab: value === 'all' ? undefined : value },
    });
  },
});

const conditionProperties = computed(() => {
  const conditionModel = SearchConditionFactory.createModel(tabActive.value);
  return conditionModel.getProperties();
});

const columnProperties = computed(() => {
  const columnModel = TableColumnFactory.createModel(tabActive.value);
  return columnModel.getProperties();
});

const searchFields = computed(() => conditionProperties.value.filter((field) => !field.apiOnly));
const dataListColumns = computed(() => columnProperties.value);

const searchQs = useSearchQs({ key: 'filter', properties: conditionProperties });

const condition = ref<Record<string, any>>({});
const operationLogList = ref<IAuditItem[]>([]);

const detailSliderState = reactive({ isShow: false, isHidden: true });
const detailInfo = ref<IAuditItem>(null);

watch(
  () => route.query,
  async (query) => {
    condition.value = searchQs.get(query, {
      res_type: tabActive.value === 'all' ? undefined : tabActive.value,
      created_at: getDateRange('last30d'),
    });

    pagination.current = Number(query.page) || 1;
    pagination.limit = Number(query.limit) || pagination.limit;

    const sort = (query.sort || 'created_at') as string;
    const order = (query.order || 'DESC') as string;

    const { list, count } = await auditStore.getAuditList({
      filter: transformSimpleCondition(condition.value, conditionProperties.value),
      page: getPageParams(pagination, { sort, order }),
    });

    operationLogList.value = list;
    pagination.count = count;
  },
  { immediate: true },
);

const handleSearch = (vals: ISearchCondition) => {
  searchQs.set(vals);
};

const handleReset = () => {
  searchQs.clear();
};

const handleViewDetails = (row: IAuditItem) => {
  const { id, res_type, res_name, res_id } = row;
  const flowId = row.detail?.data?.res_flow?.flow_id;

  // 负载均衡相关资源需要跳转路由
  const clbResTypes = ['load_balancer', 'url_rule', 'listener', 'url_rule_domain', 'target_group'];

  if (clbResTypes.includes(res_type) && flowId) {
    routerAction.redirect(
      {
        name: MENU_RESOURCE_OPERATION_LOG_DETAILS,
        query: {
          record_id: id,
          name: res_name,
          flow: flowId,
          res_id,
        },
      },
      {
        history: true,
      },
    );
    return;
  }

  detailInfo.value = row;
  detailSliderState.isHidden = false;
  detailSliderState.isShow = true;
};

provide('isResourcePage', true);
</script>

<template>
  <div class="entry-rsc">
    <bk-tab type="card-grid" v-model:active="tabActive">
      <bk-tab-panel v-for="panel in tabPanels" :key="panel.name" :name="panel.name" :label="panel.label">
        <!-- 这里多个tab共用同一个内容 v-if 确保只渲染一个 -->
        <template v-if="tabActive === panel.name">
          <search :fields="searchFields" :condition="condition" @search="handleSearch" @reset="handleReset" />
          <data-list
            v-bkloading="{ loading: auditStore.isAuditListLoading }"
            :columns="dataListColumns"
            :list="operationLogList"
            :pagination="pagination"
            @view-details="handleViewDetails"
          />
        </template>
      </bk-tab-panel>
    </bk-tab>
  </div>
  <template v-if="!detailSliderState.isHidden">
    <detail-slider
      v-model="detailSliderState.isShow"
      :fields="detailFields"
      :info="detailInfo"
      @hidden="detailSliderState.isHidden = true"
    />
  </template>
</template>

<style lang="scss" scoped></style>
