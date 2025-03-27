<script setup lang="ts">
import { computed, onBeforeMount, ref, useTemplateRef, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import usePage from '@/hooks/use-page';
import { Senarios, useWhereAmI } from '@/hooks/useWhereAmI';
import { useVerify } from '@/hooks';
import {
  type ISecurityGroupDetail,
  type SecurityGroupRelResourceByBizItem,
  SecurityGroupRelatedResourceName,
  useSecurityGroupStore,
} from '@/store/security-group';
import { useBusinessGlobalStore } from '@/store/business-global';
import { transformSimpleCondition } from '@/utils/search';
import {
  RELATED_RES_NAME_MAP,
  RELATED_RES_OPERATE_DISABLED_TIPS_MAP,
  RELATED_RES_OPERATE_TYPE,
  RELATED_RES_PROPERTIES_MAP,
} from '@/constants/security-group';

import dataList from './index.vue';
import bind from '../bind/index.vue';
import batchUnbind from '../unbind/batch.vue';
import singleUnbind from '../unbind/single.vue';

const props = defineProps<{
  detail: ISecurityGroupDetail;
  bkBizId: number;
  tabActive: SecurityGroupRelatedResourceName;
  resCount: number;
  condition: Record<string, any>;
}>();
const emit = defineEmits(['operate-success']);

const { t } = useI18n();
const { getBizsId, whereAmI } = useWhereAmI();
const securityGroupStore = useSecurityGroupStore();
const { getBusinessNames } = useBusinessGlobalStore();

// 预鉴权
const { handleAuth, authVerifyData } = useVerify();
const authAction = computed(() => {
  return whereAmI.value === Senarios.business ? 'biz_iaas_resource_operate' : 'iaas_resource_operate';
});

const isExpand = ref(props.bkBizId === getBizsId());
const iconClass = computed(() => (isExpand.value ? 'bkhcm-icon-angle-up-fill' : 'bkhcm-icon-right-shape'));
const businessName = computed(() => {
  if (props.bkBizId === -1) return t('未分配');
  return getBusinessNames(props.bkBizId)?.[0];
});
const isCurrentBusiness = computed(() => getBizsId() === props.bkBizId);
const isOperateDisabled = computed(() => {
  // 暂不支持负载均衡相关的操作
  return props.tabActive === SecurityGroupRelatedResourceName.CLB;
});

const relResList = ref<SecurityGroupRelResourceByBizItem[]>([]);
const { pagination, getPageParams } = usePage();

const handleToggle = async () => {
  isExpand.value = !isExpand.value;
  if (isExpand.value) {
    await getList();
  }
};

const loading = ref(false);
const getList = async (
  tabActive = props.tabActive,
  condition = props.condition,
  sort = 'created_at',
  order = 'DESC',
) => {
  loading.value = true;
  try {
    const api =
      tabActive === 'CVM' ? securityGroupStore.queryRelCvmByBiz : securityGroupStore.queryRelLoadBalancerByBiz;

    const res = await api(props.detail.id, props.bkBizId, {
      filter: transformSimpleCondition(condition, RELATED_RES_PROPERTIES_MAP[props.tabActive]),
      page: getPageParams(pagination, { sort, order }),
    });

    relResList.value = res.list;
    // 设置页码总条数
    pagination.count = res.count;
  } finally {
    loading.value = false;
  }
};

const selected = ref<SecurityGroupRelResourceByBizItem[]>([]);
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
  emit('operate-success');
};

const datalistRef = useTemplateRef('data-list');
const reload = (tabActive: SecurityGroupRelatedResourceName, condition: Record<string, any>) => {
  datalistRef.value.handleClear();
  if (pagination.current === 1) {
    getList(tabActive, condition);
  } else {
    pagination.current = 1;
  }
};

watch([() => pagination.current, () => pagination.limit], () => {
  getList();
});

onBeforeMount(() => {
  if (isCurrentBusiness.value) getList();
});

defineExpose({ isExpand, reload });
</script>

<template>
  <div class="collapse-wrap">
    <div class="tools">
      <i class="hcm-icon" :class="iconClass" @click="handleToggle"></i>
      <span class="name">{{ businessName }}</span>
      <!-- 只允许对本业务的实例进行绑定和解绑 -->
      <template v-if="isCurrentBusiness">
        <bk-tag class="tag" theme="success" type="filled">{{ t('当前业务') }}</bk-tag>
        <bk-button
          theme="primary"
          text
          :class="{ 'hcm-no-permision-text-btn': !authVerifyData?.permissionAction?.[authAction] }"
          :disabled="isOperateDisabled"
          v-bk-tooltips="{
            content: RELATED_RES_OPERATE_DISABLED_TIPS_MAP[RELATED_RES_OPERATE_TYPE.BIND],
            disabled: !isOperateDisabled,
          }"
          @click="handleShowOperateDialog('bind')"
        >
          <i class="hcm-icon bkhcm-icon-plus-circle-shape mr2"></i>
          {{ t('新增绑定') }}
        </bk-button>
        <bk-button
          theme="primary"
          text
          class="unbind-btn"
          :class="{ 'hcm-no-permision-text-btn': !authVerifyData?.permissionAction?.[authAction] }"
          :disabled="!selected.length || isOperateDisabled"
          v-bk-tooltips="{
            content: RELATED_RES_OPERATE_DISABLED_TIPS_MAP[RELATED_RES_OPERATE_TYPE.UNBIND],
            disabled: !isOperateDisabled,
          }"
          @click="handleShowOperateDialog('batch-unbind')"
        >
          {{ t('批量解绑') }}
        </bk-button>
      </template>
      <!-- 其他业务的实例，在当前业务只读，不可以操作 -->
      <template v-else>
        <span class="overview">
          {{ RELATED_RES_NAME_MAP[tabActive] }}：
          <span class="number">{{ resCount }}</span>
        </span>
      </template>
    </div>
    <data-list
      ref="data-list"
      v-show="isExpand"
      v-bkloading="{ loading }"
      :resource-name="tabActive"
      operation="base"
      :list="relResList"
      :pagination="pagination"
      :has-selections="isCurrentBusiness"
      :has-settings="isCurrentBusiness"
      :is-row-select-enable="() => true"
      @select="(selections) => (selected = selections)"
    >
      <template v-if="isCurrentBusiness" #operate="{ row }">
        <bk-button
          theme="primary"
          text
          :class="{ 'hcm-no-permision-text-btn': !authVerifyData?.permissionAction?.[authAction] }"
          :disabled="isOperateDisabled"
          v-bk-tooltips="{
            content: RELATED_RES_OPERATE_DISABLED_TIPS_MAP[RELATED_RES_OPERATE_TYPE.UNBIND],
            disabled: !isOperateDisabled,
          }"
          @click="handleShowOperateDialog('single-unbind', row)"
        >
          {{ t('解绑') }}
        </bk-button>
      </template>
    </data-list>

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
.collapse-wrap {
  border: 1px solid #dcdee5;

  .tools {
    padding: 0 24px 0 8px;
    display: flex;
    align-items: center;
    height: 32px;
    background: #f0f1f5;
    font-size: 12px;

    .name {
      margin: 0 8px;
      color: #313238;
    }

    .tag {
      margin-right: 16px;
      height: 16px;
    }

    .unbind-btn {
      margin-left: auto;
    }

    .overview {
      margin-left: 50px;
      .number {
        color: #313238;
      }
    }
  }
}
</style>
