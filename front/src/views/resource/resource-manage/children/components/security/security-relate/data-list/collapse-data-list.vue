<script setup lang="ts">
import { computed, onBeforeMount, ref, useTemplateRef, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import usePage from '@/hooks/use-page';
import { Senarios, useWhereAmI } from '@/hooks/useWhereAmI';
import { useVerify } from '@/hooks';
import {
  type ISecurityGroupDetail,
  type SecurityGroupRelResourceByBizItem,
  type SecurityGroupRelatedResourceName,
  useSecurityGroupStore,
} from '@/store/security-group';
import { useBusinessGlobalStore } from '@/store/business-global';
import { transformSimpleCondition } from '@/utils/search';
import { RELATED_RES_KEY_MAP, RELATED_RES_NAME_MAP, RELATED_RES_PROPERTIES_MAP } from '@/constants/security-group';

import { Message } from 'bkui-vue';
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
const businessName = computed(() => getBusinessNames(props.bkBizId)?.[0]);
const isCurrentBusiness = computed(() => getBizsId() === props.bkBizId);

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

const bindRef = useTemplateRef('bind-comp');
const selected = ref<SecurityGroupRelResourceByBizItem[]>([]);
const handleBind = async (ids: string[]) => {
  // TODO：当前只支持CVM
  await securityGroupStore.batchAssociateCvms({ security_group_id: props.detail.id, cvm_ids: ids });
  Message({ theme: 'success', message: t('绑定成功') });
  bindRef.value.handleClosed();
  getList();
};
const handleBatchUnbind = async (ids: string[]) => {
  // TODO：当前只支持CVM
  await securityGroupStore.batchDisassociateCvms({ security_group_id: props.detail.id, cvm_ids: ids });
  Message({ theme: 'success', message: t('解绑成功') });
  getList();
};
const singleUnbindVisible = ref(false);
const singleUnbindOperateInfo = ref<SecurityGroupRelResourceByBizItem>(null);
const handleShowSingleUnbind = async (row: SecurityGroupRelResourceByBizItem) => {
  if (!authVerifyData.value?.permissionAction?.[authAction.value]) {
    handleAuth(authAction.value);
    return;
  }
  singleUnbindVisible.value = true;
  const res = await securityGroupStore.pullSecurityGroup(RELATED_RES_KEY_MAP[props.tabActive], [row]);
  [singleUnbindOperateInfo.value] = res;
};
const handleSingleUnbind = async () => {
  await handleBatchUnbind([singleUnbindOperateInfo.value.id]);
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
        <bk-tag v-if="isCurrentBusiness" class="tag" theme="success" type="filled">{{ t('当前业务') }}</bk-tag>
        <bind ref="bind-comp" :tab-active="tabActive" :detail="detail" text-button @confirm="handleBind">
          <template #icon>
            <i class="hcm-icon bkhcm-icon-plus-circle-shape mr2"></i>
          </template>
        </bind>
        <batch-unbind
          class="unbind-btn"
          theme="primary"
          text-button
          :selections="selected"
          :disabled="!selected.length"
          :tab-active="tabActive"
          :handle-confirm="handleBatchUnbind"
        />
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
      <template v-if="tabActive === 'CVM' && isCurrentBusiness" #operate="{ row }">
        <bk-button
          theme="primary"
          text
          :class="{ 'hcm-no-permision-text-btn': !authVerifyData?.permissionAction?.[authAction] }"
          @click="handleShowSingleUnbind(row)"
        >
          {{ t('解绑') }}
        </bk-button>
      </template>
    </data-list>

    <template v-if="singleUnbindVisible && isCurrentBusiness">
      <single-unbind
        v-model="singleUnbindVisible"
        :res-name="RELATED_RES_NAME_MAP[tabActive]"
        :info="singleUnbindOperateInfo"
        :loading="securityGroupStore.isBatchQuerySecurityGroupByResIdsLoading"
        :handle-confirm="handleSingleUnbind"
        :confirm-loading="securityGroupStore.isBatchDisassociateCvmsLoading"
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

    :deep(.unbind-btn) {
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
