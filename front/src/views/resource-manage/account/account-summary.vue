<script setup lang="ts">
import { ref, computed, watch } from 'vue';
import { useRoute } from 'vue-router';
import { useAccountSelectorStore } from '@/store/account-selector';
import { useAccountStore } from '@/store';
import { VendorEnum } from '@/common/constant';
import AccountStatusSlider from './account-status/index.vue';

const route = useRoute();
const accountSelectorStore = useAccountSelectorStore();
const accountStore = useAccountStore();

const isShowAccountStatus = ref(false);

const accountId = computed(() => (route.query.accountId as string) || '');
const queryVendor = computed(() => (route.query.vendor as VendorEnum) || null);

const accountFromList = computed(() => {
  if (!accountId.value) return null;
  return (
    accountSelectorStore.authorizedResourceAccountList.find((a: { id: string }) => a.id === accountId.value) || null
  );
});

const accountDetail = ref<{ name?: string; vendor?: VendorEnum; extension?: Record<string, any> } | null>(null);
watch(
  accountId,
  async (id) => {
    if (!id) {
      accountDetail.value = null;
      return;
    }
    try {
      const res = await accountStore.getAccountDetail(id);
      accountDetail.value = res?.data || res;
    } catch {
      accountDetail.value = null;
    }
  },
  { immediate: true },
);

const currentAccount = computed(() => accountDetail.value || accountFromList.value);

const isOtherVendor = computed(() => {
  const vendor = queryVendor.value || currentAccount.value?.vendor;
  return vendor === VendorEnum.OTHER;
});

const headerExtensionMap = computed(() => {
  const map = { firstLabel: '', firstField: '', secondLabel: '', secondField: '' };
  const vendor = queryVendor.value || currentAccount.value?.vendor;
  switch (vendor) {
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
</script>

<template>
  <div class="account-summary">
    <div class="account-summary-title">
      <span class="main-account-name">
        {{ currentAccount?.name || '全部账号' }}
      </span>
      <template v-if="(currentAccount as any)?.extension && !isOtherVendor">
        <div class="account-summary-extension">
          <div class="extension-item">
            {{ headerExtensionMap.firstLabel }}：
            <span class="info-text">
              {{ (currentAccount as any).extension?.[headerExtensionMap.firstField] }}
            </span>
          </div>
          <div class="extension-item">
            {{ headerExtensionMap.secondLabel }}：
            <span class="info-text">
              {{ (currentAccount as any).extension?.[headerExtensionMap.secondField] }}
            </span>
          </div>
        </div>
      </template>
    </div>
    <bk-button v-if="accountId" text theme="primary" @click="isShowAccountStatus = true">查看账号状态</bk-button>
  </div>

  <AccountStatusSlider v-model:is-show="isShowAccountStatus" :account-id="accountId" />
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
