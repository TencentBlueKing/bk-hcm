<script setup lang="ts">
import { computed, inject, Ref, ref, useTemplateRef, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import {
  type ISecurityGroupDetail,
  type ISecurityGroupRelBusiness,
  type ISecurityGroupRelResCountItem,
  type SecurityGroupRelResourceByBizItem,
  SecurityGroupRelatedResourceName,
  useSecurityGroupStore,
} from '@/store/security-group';
import { useBusinessGlobalStore } from '@/store/business-global';
import { useRegionsStore } from '@/store/useRegionsStore';
import { Senarios, useWhereAmI } from '@/hooks/useWhereAmI';
import usePage from '@/hooks/use-page';
import { useVerify } from '@/hooks';
import { getSimpleConditionBySearchSelect, transformSimpleCondition } from '@/utils/search';
import {
  RELATED_RES_KEY_MAP,
  RELATED_RES_NAME_MAP,
  RELATED_RES_OPERATE_DISABLED_TIPS_MAP,
  RELATED_RES_OPERATE_TYPE,
  RELATED_RES_PROPERTIES_MAP,
} from '@/constants/security-group';
import { ISearchSelectValue } from '@/typings';

import { Plus } from 'bkui-vue/lib/icon';
import tab from './tab/index.vue';
import bind from './bind/index.vue';
import batchUnbind from './unbind/batch.vue';
import search from './search/index.vue';
import dataList from './data-list/index.vue';
import singleUnbind from './unbind/single.vue';

const props = defineProps<{
  detail: ISecurityGroupDetail;
  relatedResourcesCountList: ISecurityGroupRelResCountItem[];
  relatedBiz: ISecurityGroupRelBusiness;
  getRelatedInfo: () => Promise<void>;
}>();

const { t } = useI18n();
const { getBizsId, whereAmI } = useWhereAmI();
const { getBusinessNames, getBusinessIds } = useBusinessGlobalStore();
const securityGroupStore = useSecurityGroupStore();
const regionStore = useRegionsStore();

const isBusinessPage = computed(() => whereAmI.value === Senarios.business);

// 预鉴权
const { handleAuth, authVerifyData } = useVerify();
const authAction = computed(() => {
  return isBusinessPage.value ? 'biz_iaas_resource_operate' : 'iaas_resource_operate';
});

const tabActive = ref<SecurityGroupRelatedResourceName>(SecurityGroupRelatedResourceName.CVM);
// 当前业务所关联资源
const currentBizRelatedResources = computed(
  () =>
    props.relatedBiz?.[RELATED_RES_KEY_MAP[tabActive.value]]?.find((item) => item.bk_biz_id === getBizsId()) || {
      res_count: 0,
    },
);
// 其他业务数
const otherBusinessCount = computed(
  () =>
    props.relatedBiz?.[RELATED_RES_KEY_MAP[tabActive.value]]?.filter((item) => item.bk_biz_id !== getBizsId()).length ||
    0,
);

// 关联资源table
const list = ref<SecurityGroupRelResourceByBizItem[]>([]);
const { pagination, getPageParams } = usePage();
const condition = ref<Record<string, any>>({});

// 业务下的平台管理：只拉取当前业务所关联的实例列表；其他业务只展示业务数量。
// 账号下的平台管理：拉取所有业务所关联的实例列表
const loading = ref(false);
const getList = async (sort = 'created_at', order = 'DESC') => {
  loading.value = true;
  try {
    const { id } = props.detail;
    const api =
      tabActive.value === 'CVM' ? securityGroupStore.queryRelCvmByBiz : securityGroupStore.queryRelLoadBalancerByBiz;
    const bizIds = isBusinessPage.value
      ? [getBizsId()]
      : props.relatedBiz[RELATED_RES_KEY_MAP[tabActive.value]].map(({ bk_biz_id }) => bk_biz_id);

    const res = await Promise.all(
      bizIds.map((bk_biz_id) =>
        api(id, bk_biz_id, {
          filter: transformSimpleCondition(condition.value, RELATED_RES_PROPERTIES_MAP[tabActive.value]),
          page: getPageParams(pagination, { sort, order }),
        }),
      ),
    );

    list.value = res.flatMap((item) => item.list);
    // 设置页码总条数
    pagination.count = isBusinessPage.value ? res[0].count : res.reduce((acc, cur) => acc + cur.count, 0);
  } finally {
    loading.value = false;
  }
};

const selected = ref<SecurityGroupRelResourceByBizItem[]>([]);
const isAssigned = inject<Ref<boolean>>('isAssigned');
const isClb = computed(() => {
  // 暂不支持负载均衡相关的操作
  return tabActive.value === SecurityGroupRelatedResourceName.CLB;
});
const bindDisabledTooltipsOption = computed(() => {
  if (!isBusinessPage.value && isAssigned.value) {
    return { content: t('安全组已分配，请到业务下操作'), disabled: isBusinessPage.value || !isAssigned.value };
  }
  if (isClb.value) {
    return {
      content: RELATED_RES_OPERATE_DISABLED_TIPS_MAP[RELATED_RES_OPERATE_TYPE.BIND],
      disabled: !isClb.value,
    };
  }
  return { disabled: true };
});
const unbindDisabledTooltipsOption = computed(() => {
  if (!isBusinessPage.value && isAssigned.value) {
    return { content: t('安全组已分配，请到业务下操作'), disabled: isBusinessPage.value || !isAssigned.value };
  }
  if (isClb.value) {
    return {
      content: RELATED_RES_OPERATE_DISABLED_TIPS_MAP[RELATED_RES_OPERATE_TYPE.UNBIND],
      disabled: !isClb.value,
    };
  }
  return { disabled: true };
});

const bindVisible = ref(false);
const batchUnbindVisible = ref(false);
const singleUnbindVisible = ref(false);
const singleUnbindOperateRow = ref<SecurityGroupRelResourceByBizItem>(null);
const handleShowOperateDialog = (
  operate: 'bind' | 'single-unbind' | 'batch-unbind',
  row?: SecurityGroupRelResourceByBizItem,
) => {
  if (!authVerifyData.value?.permissionAction?.[authAction.value]) {
    handleAuth(authAction.value);
    return;
  }
  switch (operate) {
    case 'bind':
      bindVisible.value = true;
      break;
    case 'single-unbind':
      singleUnbindVisible.value = true;
      singleUnbindOperateRow.value = row;
      break;
    case 'batch-unbind':
      batchUnbindVisible.value = true;
      break;
  }
};
const handleOperateSuccess = () => {
  // 重新拉取安全组所关联的资源信息
  props.getRelatedInfo();
  handleClear();
};

const searchRef = useTemplateRef('relate-resource-search');
const handleSearch = (searchValue: ISearchSelectValue) => {
  // 搜索条件变更后，重置勾选
  handleClear();

  condition.value = getSimpleConditionBySearchSelect(searchValue, [
    { field: 'region', formatter: (val: string) => regionStore.getRegionNameEN(val) },
    { field: 'bk_biz_id', formatter: (name: string) => getBusinessIds(name) },
  ]);

  if (pagination.current === 1) {
    getList();
  } else {
    pagination.current = 1;
  }
};

const dataListRef = useTemplateRef('data-list');
const handleClear = () => {
  dataListRef.value.handleClear();
};

watch(tabActive, () => {
  // 切换tab时，清空搜索条件，触发搜索
  searchRef.value?.clear();
});

watch(
  [() => pagination.current, () => pagination.limit, () => props.relatedBiz],
  () => {
    if (props.relatedBiz) {
      getList();
    }
  },
  { immediate: true },
);
</script>

<template>
  <div class="platform-manage-module">
    <div class="tools-bar">
      <tab v-model="tabActive" :detail="detail" :related-resources-count-list="relatedResourcesCountList" />

      <!-- TODO：目前只支持CVM -->
      <div class="operate-btn-wrap">
        <bk-button
          theme="primary"
          :class="{ 'hcm-no-permision-btn': !authVerifyData?.permissionAction?.[authAction] }"
          :disabled="(!isBusinessPage && isAssigned) || isClb"
          v-bk-tooltips="bindDisabledTooltipsOption"
          @click="handleShowOperateDialog('bind')"
        >
          <plus width="26" height="26" />
          {{ t('新增绑定') }}
        </bk-button>
        <bk-button
          :class="{ 'hcm-no-permision-btn': !authVerifyData?.permissionAction?.[authAction] }"
          :disabled="!selected.length || (!isBusinessPage && isAssigned) || isClb"
          v-bk-tooltips="unbindDisabledTooltipsOption"
          @click="handleShowOperateDialog('batch-unbind')"
        >
          {{ t('批量解绑') }}
        </bk-button>
      </div>

      <search
        class="search"
        ref="relate-resource-search"
        :resource-name="tabActive"
        operation="base"
        @search="handleSearch"
      />
    </div>

    <div v-if="isBusinessPage" class="overview">
      {{ t(`当前业务（${getBusinessNames(getBizsId())}）下共有`) }}
      <span class="number">{{ currentBizRelatedResources?.res_count }}</span>
      {{ t(`台${RELATED_RES_NAME_MAP[tabActive]}，还有`) }}
      <span class="number">{{ otherBusinessCount }}</span>
      {{ t(`个业务也在使用`) }}
    </div>

    <div class="rel-res-display-wrap">
      <data-list
        v-bkloading="{ loading }"
        ref="data-list"
        :resource-name="tabActive"
        operation="base"
        :list="list"
        :pagination="pagination"
        :is-row-select-enable="() => true"
        @select="(selections) => (selected = selections)"
      >
        <template #operate="{ row }">
          <bk-button
            :class="{ 'hcm-no-permision-text-btn': !authVerifyData?.permissionAction?.[authAction] }"
            theme="primary"
            text
            :disabled="(!isBusinessPage && isAssigned) || isClb"
            v-bk-tooltips="unbindDisabledTooltipsOption"
            @click="handleShowOperateDialog('single-unbind', row)"
          >
            {{ t('解绑') }}
          </bk-button>
        </template>
      </data-list>
    </div>

    <template v-if="bindVisible">
      <bind v-model="bindVisible" :tab-active="tabActive" :detail="detail" @success="handleOperateSuccess" />
    </template>

    <template v-if="batchUnbindVisible">
      <batch-unbind
        v-model="batchUnbindVisible"
        :selections="selected"
        :tab-active="tabActive"
        :detail="detail"
        @success="handleOperateSuccess"
      />
    </template>

    <template v-if="singleUnbindVisible">
      <single-unbind
        v-model="singleUnbindVisible"
        :row="singleUnbindOperateRow"
        :tab-active="tabActive"
        :detail="detail"
        @success="handleOperateSuccess"
      />
    </template>
  </div>
</template>

<style scoped lang="scss">
.tools-bar {
  display: flex;
  align-items: center;

  .operate-btn-wrap {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  .search {
    margin-left: auto;
    width: 320px;
  }
}

.overview {
  margin-top: 12px;
  font-size: 12px;
  color: #4d4f56;

  .number {
    font-weight: 700;
  }
}

.rel-res-display-wrap {
  margin-top: 12px;
}
</style>
