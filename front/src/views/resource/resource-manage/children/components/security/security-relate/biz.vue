<script setup lang="ts">
import { ref, useTemplateRef, watch } from 'vue';
import { useRegionsStore } from '@/store/useRegionsStore';
import {
  type ISecurityGroupDetail,
  type ISecurityGroupRelBusiness,
  type ISecurityGroupRelResCountItem,
} from '@/store/security-group';
import { useBusinessGlobalStore } from '@/store/business-global';
import { RELATED_RES_KEY_MAP, SecurityGroupRelatedResourceName } from '@/constants/security-group';
import { ISearchSelectValue } from '@/typings';

import tab from './tab/index.vue';
import collapseDataList from './data-list/collapse-data-list.vue';
import search from './search/index.vue';

const props = defineProps<{
  detail: ISecurityGroupDetail;
  relatedResourcesCountList: ISecurityGroupRelResCountItem[];
  relatedBiz: ISecurityGroupRelBusiness;
  getRelatedInfo: () => Promise<void>;
  relBizLoading: boolean;
}>();

const regionStore = useRegionsStore();
const { getBusinessIds } = useBusinessGlobalStore();

const tabActive = ref(SecurityGroupRelatedResourceName.CVM);

const handleOperateSuccess = () => {
  props.getRelatedInfo();
  searchRef.value?.clear();
};

const searchRef = useTemplateRef('relate-resource-search');
const collapseDataListRef = useTemplateRef('collapse-data-list');
const condition = ref<Record<string, any>>({});
const formatterOptions = [
  { field: 'region', formatter: (val: string) => regionStore.getRegionNameEN(val) },
  { field: 'bk_biz_id', formatter: (name: string) => getBusinessIds(name) },
];
const handleSearch = (_: ISearchSelectValue, flatCondition: Record<string, any>) => {
  condition.value = flatCondition;

  collapseDataListRef.value?.forEach((compRef) => {
    if (compRef.isExpand) {
      compRef.reload(tabActive.value, condition.value);
    }
  });
};

watch(tabActive, () => {
  // 切换tab时，清空搜索条件，触发搜索
  searchRef.value?.clear();
});
</script>

<template>
  <div class="business-manage-module">
    <div class="tools-bar">
      <tab v-model="tabActive" :detail="detail" :related-resources-count-list="relatedResourcesCountList" />
      <search
        class="search"
        ref="relate-resource-search"
        :resource-name="tabActive"
        operation="base"
        flat
        :options="formatterOptions"
        @search="handleSearch"
      />
    </div>

    <div v-bkloading="{ loading: relBizLoading }" class="rel-res-display-wrap">
      <collapse-data-list
        v-for="{ bk_biz_id: bkBizId, res_count: resCount } in relatedBiz?.[RELATED_RES_KEY_MAP[tabActive]]"
        ref="collapse-data-list"
        :key="bkBizId"
        :detail="detail"
        :bk-biz-id="bkBizId"
        :tab-active="tabActive"
        :res-count="resCount"
        :condition="condition"
        @operate-success="handleOperateSuccess"
      />
    </div>
  </div>
</template>

<style scoped lang="scss">
.tools-bar {
  display: flex;
  align-items: center;

  .search {
    margin-left: auto;
    width: 320px;
  }
}

.rel-res-display-wrap {
  min-height: 300px;
  margin-top: 12px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}
</style>
