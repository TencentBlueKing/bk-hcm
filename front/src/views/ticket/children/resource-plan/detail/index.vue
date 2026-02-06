<script setup lang="ts">
import { ref, onBeforeMount, computed, useTemplateRef, reactive } from 'vue';
import { RouteLocationRaw, useRoute, useRouter } from 'vue-router';
import { useResourcePlanStore } from '@/store';
import { useI18n } from 'vue-i18n';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import useTimeoutPoll from '@/hooks/use-timeout-poll';
import Approval from '@/components/resource-plan/applications/detail/approval';
import ResourcePlanTicketAudit from '@/components/resource-plan/applications/detail/ticket-audit/index.vue';
import Basic from '@/components/resource-plan/applications/detail/basic/index.vue';
import ResourcePlanList from '@/components/resource-plan/applications/detail/list/index.vue';
import SubTicketList from '../sub-ticket/sub-ticket-list.vue';
import DetailHeader from '@/views/resource/resource-manage/common/header/detail-header';
import { TicketByIdResult } from '@/typings/resourcePlan';
import { SubTicketAudit } from '@/store/ticket/res-sub-ticket';
import { MENU_BUSINESS_TICKET_MANAGEMENT, MENU_SERVICE_TICKET_MANAGEMENT } from '@/constants/menu-symbol';
import { GLOBAL_BIZS_KEY } from '@/common/constant';

// 路由、状态管理、工具函数
const route = useRoute();
const router = useRouter();
const { t } = useI18n();
const resourcePlanStore = useResourcePlanStore();
const { getBizsId, isBusinessPage } = useWhereAmI();

// 跳转至单据列表
const navigateTo: RouteLocationRaw = reactive({
  name: computed(() => (isBusinessPage ? MENU_BUSINESS_TICKET_MANAGEMENT : MENU_SERVICE_TICKET_MANAGEMENT)),
  query: computed(() => ({ type: 'resource_plan', [GLOBAL_BIZS_KEY]: route.query[GLOBAL_BIZS_KEY] })),
});

// 响应式数据
const ticketDetail = ref<TicketByIdResult>();
const ticketAuditDetail = ref<SubTicketAudit>();
const isLoading = ref(false);
const errorMessage = ref<string>();
const active = ref('approval');
const subTicketListRef = useTemplateRef('subTicketList');

// 计算属性：详情标题
const detailTitle = computed(() => `${t('申请单详情')} - ${ticketDetail.value?.id || ''}`);

// 计算属性：是否显示审批详情组件
const isTicketAuditDetailShow = computed(() => {
  return ticketAuditDetail.value?.itsm_audit.status !== 'init';
});

// 获取数据的逻辑
const getResultData = async () => {
  try {
    isLoading.value = true;
    let promise = null;
    // 判断是否业务页面
    if (isBusinessPage) {
      promise = Promise.all([
        resourcePlanStore.getBizResourcesTicketsById(getBizsId(), route.query?.id as string),
        resourcePlanStore.getBizResourcesTicketsAuditById(getBizsId(), route.query?.id as string),
      ]);
    } else {
      // 服务页面
      promise = Promise.all([
        resourcePlanStore.getOpResourcesTicketsById(route.query?.id as string),
        resourcePlanStore.getOpResourcesTicketsAuditById(route.query?.id as string),
      ]);
    }

    const [ticketRes, ticketAuditRes] = await promise;

    // 错误信息处理
    if (ticketRes.data?.status_info?.status === 'failed') {
      errorMessage.value = ticketRes.data?.status_info?.message;
    } else {
      errorMessage.value = '';
    }

    // 响应式数据赋值
    ticketDetail.value = ticketRes?.data;
    ticketAuditDetail.value = ticketAuditRes?.data as SubTicketAudit;

    // 获取子单列表数据
    subTicketListRef.value && subTicketListRef.value?.getData();

    // 轮询逻辑：init 或 auditing 状态时自动刷新
    if (ticketRes?.data?.status_info?.status === 'init' || ticketRes?.data?.status_info?.status === 'auditing') {
      autoFlushTask.resume();
    } else {
      autoFlushTask.reset();
    }
  } catch (error) {
    console.error('error', error); // eslint-disable-line no-console
  } finally {
    isLoading.value = false;
  }
};
const handelUpdate = () => {
  // 修改路由参数 tab 参数
  router.replace({
    query: {
      ...route.query,
      tab: active.value,
    },
  });
};

// 轮询任务：30 秒自动刷新
const autoFlushTask = useTimeoutPoll(() => {
  getResultData();
}, 30000);

// 组件挂载时执行一次数据拉取
onBeforeMount(() => {
  getResultData();
  active.value = (route.query?.tab as string) || 'approval';
});
</script>

<template>
  <bk-loading :loading="isLoading">
    <DetailHeader :to="navigateTo">
      <span>{{ detailTitle }}</span>
    </DetailHeader>
    <section class="home">
      <!-- 当前审批节点信息 -->
      <Approval
        class="mb-16"
        :status-info="ticketDetail?.status_info"
        :is-biz="isBusinessPage"
        :error-message="errorMessage"
        :ticket-audit-detail="ticketAuditDetail"
      />

      <bk-tab type="card-grid" v-model:active="active" class="header-tab" @update:active="handelUpdate">
        <bk-tab-panel name="approval" label="审批信息">
          <!-- 审批信息 -->
          <ResourcePlanTicketAudit
            v-if="isTicketAuditDetailShow"
            class="no-shadow"
            :detail="ticketAuditDetail"
            :fetch-data="getResultData"
            :timeout-poll-action="autoFlushTask"
            :is-business-page="isBusinessPage"
          />
          <div class="divider">
            <bk-divider color="#dcdee5"></bk-divider>
          </div>
          <!-- 子单信息列表 -->
          <SubTicketList
            ref="subTicketList"
            :ticket-status="ticketDetail?.status_info?.status"
            @retry-ticket="getResultData"
          />
        </bk-tab-panel>
        <bk-tab-panel render-directive="if" name="application" label="申请单信息">
          <!-- 基本信息 -->
          <Basic :base-info="ticketDetail?.base_info" class="mb-16 no-shadow" :is-biz="isBusinessPage" />

          <div class="divider">
            <bk-divider color="#dcdee5"></bk-divider>
          </div>

          <!-- 资源预测列表 -->
          <ResourcePlanList
            :demands="ticketDetail?.demands"
            :ticket-type="ticketDetail?.base_info?.type"
            :is-biz="isBusinessPage"
          />
        </bk-tab-panel>
      </bk-tab>
    </section>
  </bk-loading>
</template>

<style scoped lang="scss">
.home {
  padding: 24px;
  margin-top: 52px;
}

.mb-16 {
  margin-bottom: 16px;
}

:deep(.bk-tab-content) {
  padding: 0;
}

.divider {
  margin: 0 24px;
}

.no-shadow {
  box-shadow: none;
}
</style>
