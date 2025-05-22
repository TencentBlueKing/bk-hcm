<template>
  <div class="operation-record-module" :class="{ 'is-resource-page': isResourcePage }">
    <panel class="panel">
      <div class="tools">
        <bk-radio-group v-model="activeResourceType" type="capsule">
          <bk-radio-button v-for="{ label, text } in resourceTypes" :key="label" :label="label">
            {{ text }}
          </bk-radio-button>
        </bk-radio-group>
        <bk-search-select
          class="search"
          v-model="searchValue"
          :data="searchData"
          value-behavior="need-key"
          @update:modelValue="handleSearch"
        />
      </div>
      <data-list
        v-bkloading="{ loading: auditStore.isAuditListLoading }"
        :is-biz-page="isBusinessPage"
        :properties="operationRecordViewProperties"
        :list="operationRecordList"
        :pagination="pagination"
        @view-detail="handleViewDetail"
      />
    </panel>
    <template v-if="!detailSliderOption.isHidden">
      <detail-slider
        v-model="detailSliderOption.isShow"
        :properties="operationRecordViewProperties"
        :info="detailInfo"
        @closed="detailSliderOption.isHidden = true"
      />
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue';
import { useRoute } from 'vue-router';
import routerAction from '@/router/utils/action';
import { ISearchItem } from 'bkui-vue/lib/search-select/utils';
import { IAuditItem, useAuditStore } from '@/store/audit';
import { useBusinessGlobalStore } from '@/store/business-global';
import { useResourceAccountStore } from '@/store/useResourceAccountStore';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import useSearchQs from '@/hooks/use-search-qs';
import usePage from '@/hooks/use-page';
import accountProperties from '@/model/account/properties';
import operationRecordProperties from '@/model/operation-record/properties';
import { buildSearchValue, getSimpleConditionBySearchSelect, transformSimpleCondition } from '@/utils/search';
import { GLOBAL_BIZS_KEY } from '@/common/constant';
import { ISearchSelectValue } from '@/typings';

import panel from '@/components/panel';
import dataList from './data-list/index.vue';
import detailSlider from './detail-slider/index.vue';

const auditStore = useAuditStore();
const { businessFullList } = useBusinessGlobalStore();
const resourceAccountStore = useResourceAccountStore();
const route = useRoute();
const { isBusinessPage, isResourcePage } = useWhereAmI();

// 资源类型 tab 选项
const resourceTypes = [
  { label: 'all', text: '全部' },
  { label: 'security_group', text: '安全组' },
  { label: 'load_balancer', text: '负载均衡' },
];
// 当前选中的资源类型，默认为全部
const activeResourceType = computed({
  get() {
    return searchQs.get(route.query)?.res_type || 'all';
  },
  set(val) {
    // tab为'all'时，置空res_type条件
    searchQs.set({ ...condition.value, res_type: val === 'all' ? undefined : val });
  },
});

const { pagination, getPageParams } = usePage();
const operationRecordViewProperties = [...accountProperties, ...operationRecordProperties];
const searchQs = useSearchQs({ key: 'filter', properties: operationRecordViewProperties });

const condition = ref<Record<string, any>>({});
const operationRecordList = ref<IAuditItem[]>([]);

const searchValue = ref([]);
const searchData = computed(() => {
  const base = [
    { name: '资源名称', id: 'res_name' },
    { name: '操作方式', id: 'action' },
    { name: '操作来源', id: 'source' },
    { name: '所属业务', id: 'bk_biz_id', children: businessFullList.map(({ id, name }) => ({ id, name })) },
    { name: '云账号', id: 'account_id' },
    { name: '操作人', id: 'operator' },
  ] as ISearchItem[];
  // 业务下, 不展示所属业务选项
  isBusinessPage && base.splice(3, 1);
  // 如果当前 tab 为负载均衡, 则展示任务类型选项(异步任务详情入口)
  activeResourceType.value === 'load_balancer' &&
    base.push({ name: '任务类型', id: 'detail.data.res_flow.flow_id', children: [{ name: '异步任务', id: '' }] });
  return base;
});
const handleSearch = (val: ISearchSelectValue) => {
  searchQs.set({
    res_type: activeResourceType.value === 'all' ? undefined : activeResourceType.value,
    ...getSimpleConditionBySearchSelect(val),
  });
};

const detailSliderOption = reactive({ isShow: false, isHidden: true });
const detailInfo = ref<IAuditItem>(null);
const handleViewDetail = (row: IAuditItem) => {
  const { id, res_type, res_name, res_id, bk_biz_id } = row;
  const flowId = row.detail?.data?.res_flow?.flow_id;

  // 负载均衡相关资源需要跳转路由
  const clbResTypes = ['load_balancer', 'url_rule', 'listener', 'url_rule_domain', 'target_group'];

  if (clbResTypes.includes(res_type) && flowId) {
    routerAction.redirect(
      {
        path: `/${isResourcePage ? 'resource' : 'business'}/record/detail`,
        query: {
          record_id: id,
          name: res_name,
          flow: flowId,
          res_id,
          [GLOBAL_BIZS_KEY]: isBusinessPage ? bk_biz_id : undefined,
        },
      },
      {
        history: true,
      },
    );
    return;
  }

  detailInfo.value = row;
  detailSliderOption.isHidden = false;
  detailSliderOption.isShow = true;
};

watch(
  () => route.query,
  async (query) => {
    condition.value = searchQs.get(query, {
      vendor: resourceAccountStore.vendorInResourcePage,
    });

    pagination.current = Number(query.page) || 1;
    pagination.limit = Number(query.limit) || pagination.limit;

    const sort = (query.sort || 'created_at') as string;
    const order = (query.order || 'DESC') as string;

    const { list, count } = await auditStore.getAuditList({
      filter: transformSimpleCondition(condition.value, operationRecordViewProperties),
      page: getPageParams(pagination, { sort, order }),
    });

    operationRecordList.value = list;
    pagination.count = count;
  },
  { immediate: true },
);

onMounted(() => {
  const condition = searchQs.get(route.query);
  searchValue.value = buildSearchValue(searchData.value, condition);
});
</script>

<style scoped lang="scss">
.operation-record-module {
  height: 100%;
  padding: 24px;

  &.is-resource-page {
    padding: 0;
    height: calc(100% - 100px);
  }

  .panel {
    height: 100%;

    .tools {
      margin-bottom: 16px;
      display: flex;
      align-items: center;
      justify-content: space-between;

      .search {
        width: 500px;
      }
    }
  }
}
</style>
