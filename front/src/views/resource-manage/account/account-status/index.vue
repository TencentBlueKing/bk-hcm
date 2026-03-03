<script setup lang="ts">
import { ref, watch } from 'vue';
import { Message } from 'bkui-vue';
import { useI18n } from 'vue-i18n';
import { useAccountStore } from '@/store';
import { CloudType } from '@/typings';
import GridContainer from '@/components/layout/grid-container/grid-container.vue';
import GridItem from '@/components/layout/grid-container/grid-item.vue';
import ResourceStatus from './resource-status';

const props = defineProps<{ isShow: boolean; accountId: string }>();
const emit = defineEmits(['update:isShow']);

const { t } = useI18n();
const accountStore = useAccountStore();

const accountDetail = ref<Record<string, any> | null>(null);
const isLoading = ref(false);
const isSyncLoading = ref(false);

watch(
  () => props.accountId,
  async (id) => {
    if (!id) {
      accountDetail.value = null;
      return;
    }
    isLoading.value = true;
    try {
      const res = await accountStore.getAccountDetail(id);
      accountDetail.value = res?.data || res;
    } catch {
      accountDetail.value = null;
    } finally {
      isLoading.value = false;
    }
  },
  { immediate: true },
);

const handleSync = async () => {
  if (!props.accountId) return;
  isSyncLoading.value = true;
  try {
    await accountStore.accountSync(props.accountId as any);
    Message({ message: t('本次同步任务触发成功。如需再次同步，请在20分钟后重试'), theme: 'success' });
  } catch {
    // error handled by http interceptor
  } finally {
    isSyncLoading.value = false;
  }
};
</script>

<template>
  <bk-sideslider
    class="account-status-slider"
    :is-show="isShow"
    :width="800"
    title="资源账号状态"
    :quick-close="false"
    @update:is-show="emit('update:isShow', $event)"
  >
    <template #header>
      <div class="slider-header">
        <span class="slider-title">{{ t('资源账号状态') }}</span>
        <bk-pop-confirm
          :content="t('同步该账号下的资源，点击确定后，立即触发同步任务')"
          trigger="click"
          @confirm="handleSync"
        >
          <bk-button theme="primary" :loading="isSyncLoading">{{ t('同步资源') }}</bk-button>
        </bk-pop-confirm>
      </div>
    </template>
    <div class="slider-content">
      <bk-loading :loading="isLoading">
        <template v-if="accountDetail">
          <div class="section-title">{{ t('基本信息') }}</div>
          <GridContainer :column="2" :label-width="120" :content-min-width="200" :gap="[0, 12]">
            <GridItem :label="t('账号名称')">{{ accountDetail.name || '--' }}</GridItem>
            <GridItem :label="t('云厂商')">
              {{ CloudType[accountDetail.vendor as keyof typeof CloudType] || '--' }}
            </GridItem>
            <GridItem :label="t('账号ID')">{{ accountDetail.id || '--' }}</GridItem>
            <GridItem v-if="accountDetail.extension?.cloud_main_account_id" :label="t('主账号ID')">
              {{ accountDetail.extension.cloud_main_account_id }}
            </GridItem>
            <GridItem v-if="accountDetail.extension?.cloud_sub_account_id" :label="t('子账号ID')">
              {{ accountDetail.extension.cloud_sub_account_id }}
            </GridItem>
            <GridItem v-if="accountDetail.extension?.cloud_account_id" :label="t('云账号ID')">
              {{ accountDetail.extension.cloud_account_id }}
            </GridItem>
            <GridItem v-if="accountDetail.extension?.cloud_tenant_id" :label="t('云租户ID')">
              {{ accountDetail.extension.cloud_tenant_id }}
            </GridItem>
            <GridItem v-if="accountDetail.extension?.cloud_subscription_name" :label="t('云订阅名称')">
              {{ accountDetail.extension.cloud_subscription_name }}
            </GridItem>
            <GridItem v-if="accountDetail.extension?.cloud_project_id" :label="t('云项目ID')">
              {{ accountDetail.extension.cloud_project_id }}
            </GridItem>
            <GridItem v-if="accountDetail.extension?.cloud_project_name" :label="t('云项目名称')">
              {{ accountDetail.extension.cloud_project_name }}
            </GridItem>
          </GridContainer>

          <div class="section-title" style="margin-top: 24px">{{ t('资源状态') }}</div>
          <ResourceStatus :account-id="accountId" />
        </template>
      </bk-loading>
    </div>
    <template #footer>
      <div class="slider-footer">
        <bk-button @click="emit('update:isShow', false)">{{ t('关闭') }}</bk-button>
      </div>
    </template>
  </bk-sideslider>
</template>

<style lang="scss" scoped>
.slider-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  padding-right: 20px;
}

.slider-title {
  font-size: 16px;
  color: #313238;
  font-weight: normal;
}

.slider-content {
  height: calc(100vh - 52px - 48px);
  padding: 20px 24px;
  overflow-y: auto;
}

.slider-footer {
  display: flex;
  justify-content: flex-end;
  width: 100%;
}

.section-title {
  font-size: 14px;
  font-weight: 700;
  color: #313238;
  margin-bottom: 12px;
}
</style>
<style lang="scss">
.account-status-slider {
  .bk-modal-content {
    overflow: hidden !important;
  }
}
</style>
