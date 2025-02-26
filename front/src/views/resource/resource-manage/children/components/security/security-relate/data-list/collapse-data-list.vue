<script setup lang="ts">
import { computed, onBeforeMount, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import usePage from '@/hooks/use-page';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import {
  type ISecurityGroupDetail,
  type SecurityGroupRelResourceByBizItem,
  type SecurityGroupRelatedResourceName,
  useSecurityGroupStore,
} from '@/store/security-group';
import { useBusinessGlobalStore } from '@/store/business-global';
import { transformSimpleCondition } from '@/utils/search';
import { RELATED_RES_NAME_MAP, RELATED_RES_PROPERTIES_MAP } from '@/constants/security-group';

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
}>();

const { t } = useI18n();
const { getBizsId } = useWhereAmI();
const securityGroupStore = useSecurityGroupStore();
const { getBusinessNames } = useBusinessGlobalStore();

const isExpand = ref(props.bkBizId === getBizsId());
const iconClass = computed(() => (isExpand.value ? 'bkhcm-icon-angle-up-fill' : 'bkhcm-icon-right-shape'));
const businessName = computed(() => getBusinessNames(props.bkBizId)?.[0]);
const isCurrentBusiness = computed(() => getBizsId() === props.bkBizId);

const relResList = ref<SecurityGroupRelResourceByBizItem[]>([]);
const { pagination, getPageParams } = usePage();

const condition = ref<Record<string, any>>({});

const handleToggle = async () => {
  isExpand.value = !isExpand.value;
  if (isExpand.value) {
    await getList();
  }
};

const loading = ref(false);
const getList = async (sort = 'created_at', order = 'DESC') => {
  loading.value = true;
  try {
    const api =
      props.tabActive === 'CVM' ? securityGroupStore.queryRelCvmByBiz : securityGroupStore.queryRelLoadBalancerByBiz;

    const res = await api(props.detail.id, props.bkBizId, {
      filter: transformSimpleCondition(condition.value, RELATED_RES_PROPERTIES_MAP[props.tabActive]),
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
const handleBind = async (ids: string[]) => {
  // TODO：当前只支持CVM
  await securityGroupStore.batchAssociateCvms({ security_group_id: props.detail.id, cvm_ids: ids });
  Message({ theme: 'success', message: t('绑定成功') });
  getList();
};
const handleBatchUnbind = async (ids: string[]) => {
  // TODO：当前只支持CVM
  await securityGroupStore.batchDisassociateCvms({ security_group_id: props.detail.id, cvm_ids: ids });
  Message({ theme: 'success', message: t('解绑成功') });
  getList();
};

onBeforeMount(() => {
  if (isCurrentBusiness.value) getList();
});
</script>

<template>
  <div class="collapse-wrap">
    <div class="tools">
      <i class="hcm-icon" :class="iconClass" @click="handleToggle"></i>
      <span class="name">{{ businessName }}</span>
      <!-- 只允许对本业务的实例进行绑定和解绑 -->
      <template v-if="isCurrentBusiness">
        <bk-tag v-if="isCurrentBusiness" class="tag" theme="success" type="filled">{{ t('当前业务') }}</bk-tag>
        <bind :tab-active="tabActive" :detail="detail" text @confirm="handleBind">
          <template #icon>
            <i class="hcm-icon bkhcm-icon-plus-circle-shape mr2"></i>
          </template>
        </bind>
        <batch-unbind
          class="unbind-btn"
          theme="primary"
          text
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
      <template v-if="tabActive === 'CVM' && isCurrentBusiness" #operate>
        <bk-table-column :label="'操作'">
          <template #default="{ row }">
            <single-unbind :row="row" :tab-active="tabActive" @confirm="handleBatchUnbind([row.id])" />
          </template>
        </bk-table-column>
      </template>
    </data-list>
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
