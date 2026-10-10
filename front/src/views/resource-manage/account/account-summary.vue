<script setup lang="ts">
import { ref, computed, watch } from 'vue';
import { useRoute } from 'vue-router';
import { VendorEnum, VendorMap } from '@/common/constant';
import { useAccountSelectorStore } from '@/store/account-selector';
import { useResourceAccountStore } from '@/store/useResourceAccountStore';
import routerAction from '@/router/utils/action';
import AccountStatusSlider from './account-status/index.vue';

const route = useRoute();
const accountSelectorStore = useAccountSelectorStore();
const resourceAccountStore = useResourceAccountStore();

const isShowAccountStatus = ref(false);

const selectedAccountId = computed(() => resourceAccountStore.selectedAccountId || '');
const isOtherVendor = computed(() => resourceAccountStore.vendorInResourcePage === VendorEnum.OTHER);

// 详情未返回或失败时，名称回退到云账号树里的简要信息
const accountName = computed(() => {
  const id = selectedAccountId.value;
  if (!id) return '';
  return (
    resourceAccountStore.resourceAccount?.name ||
    resourceAccountStore.currentAccountSimpleInfo?.name ||
    accountSelectorStore.authorizedResourceAccountList.find((item) => item.id === id)?.name ||
    ''
  );
});

const title = computed(() => {
  if (selectedAccountId.value) return accountName.value;
  if (resourceAccountStore.currentVendor) return VendorMap[resourceAccountStore.currentVendor];
  return '全部账号';
});

const headerExtensionMap = computed(() => {
  const map = { firstLabel: '', firstField: '', secondLabel: '', secondField: '' };
  switch (resourceAccountStore.resourceAccount?.vendor) {
    case VendorEnum.TCLOUD:
      Object.assign(map, {
        firstLabel: '主账号ID',
        firstField: 'cloud_main_account_id',
        secondLabel: '子账号ID',
        secondField: 'cloud_sub_account_id',
      });
      break;
    case VendorEnum.AWS:
      Object.assign(map, {
        firstLabel: '云账号ID',
        firstField: 'cloud_account_id',
        secondLabel: '云iam用户名',
        secondField: 'cloud_iam_username',
      });
      break;
    case VendorEnum.AZURE:
      Object.assign(map, {
        firstLabel: '云租户ID',
        firstField: 'cloud_tenant_id',
        secondLabel: '云订阅名称',
        secondField: 'cloud_subscription_name',
      });
      break;
    case VendorEnum.GCP:
      Object.assign(map, {
        firstLabel: '云项目ID',
        firstField: 'cloud_project_id',
        secondLabel: '云项目名称',
        secondField: 'cloud_project_name',
      });
      break;
    case VendorEnum.HUAWEI:
      Object.assign(map, {
        firstLabel: '子账号ID',
        firstField: 'cloud_sub_account_id',
        secondLabel: '云子账号名称',
        secondField: 'cloud_sub_account_name',
      });
      break;
  }
  return map;
});

// 旧「账号信息 / 资源状态」地址转址时带 showAccountStatus=1：打开侧滑后立刻去掉参数，刷新、后退不再弹出
watch(
  () => route.query.showAccountStatus,
  (flag) => {
    if (flag !== '1') return;
    if (route.query.accountId) isShowAccountStatus.value = true;
    const { showAccountStatus, ...query } = route.query;
    routerAction.redirect({ query }, { replace: true });
  },
  { immediate: true },
);
</script>

<template>
  <div class="account-summary">
    <div class="account-summary-title">
      <span class="main-account-name">{{ title }}</span>
      <div
        v-if="selectedAccountId && resourceAccountStore.resourceAccount?.extension && !isOtherVendor"
        class="account-summary-extension"
      >
        <div class="extension-item">
          {{ headerExtensionMap.firstLabel }}：
          <span class="info-text">
            {{ resourceAccountStore.resourceAccount.extension[headerExtensionMap.firstField] }}
          </span>
        </div>
        <div class="extension-item">
          {{ headerExtensionMap.secondLabel }}：
          <span class="info-text">
            {{ resourceAccountStore.resourceAccount.extension[headerExtensionMap.secondField] }}
          </span>
        </div>
      </div>
    </div>
    <bk-button v-if="selectedAccountId" text theme="primary" @click="isShowAccountStatus = true">
      查看账号状态
    </bk-button>
  </div>

  <AccountStatusSlider v-model:is-show="isShowAccountStatus" :account-id="selectedAccountId" />
</template>

<style lang="scss" scoped>
.account-summary {
  display: flex;
  align-items: center;
  gap: 20px;
  height: 52px;
  padding: 0 24px;
  background: #fff;
  border-bottom: 1px solid #dcdee5;
}

.account-summary-title {
  display: flex;
  align-items: center;
  gap: 20px;
  min-width: 0;
}

.main-account-name {
  font-size: 16px;
  color: #313238;
}

.account-summary-extension {
  display: flex;
  align-items: center;
  gap: 20px;
  font-size: 14px;
  color: #63656e;

  .extension-item {
    display: flex;
    align-items: center;

    .info-text {
      font-size: 14px;
      color: #313238;
    }
  }
}
</style>
