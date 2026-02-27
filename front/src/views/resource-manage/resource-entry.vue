<script setup lang="ts">
/**
 * 资源纳管入口组件
 * - 左侧：AccountVendorGroup 云厂商&账号选择器
 * - 右侧：列表页显示账号头部信息，详情/申请页显示面包屑 + RouterView
 */
import { computed, ref, watch } from 'vue';
import { RouterView, useRoute } from 'vue-router';
import AccountVendorGroup from './account/vendor-group/index.vue';
import HcmBreadcrumb from '@/components/layout/breadcrumb.vue';
import { useAccountSelectorStore } from '@/store/account-selector';
import { useAccountStore } from '@/store';
import { VendorEnum } from '@/common/constant';
import { MENU_RESOURCE_RESOURCE_LIST } from '@/constants/menu-symbol';

const route = useRoute();
const accountSelectorStore = useAccountSelectorStore();
const accountStore = useAccountStore();

const isListPage = computed(() => route.name === MENU_RESOURCE_RESOURCE_LIST);

// 从 route.query 获取 accountId、vendor
const accountId = computed(() => (route.query.accountId as string) || '');
const queryVendor = computed(() => (route.query.vendor as VendorEnum) || null);

// 列表中的账号（含 name、vendor）
const accountFromList = computed(() => {
  if (!accountId.value) return null;
  return (
    accountSelectorStore.authorizedResourceAccountList.find((a: { id: string }) => a.id === accountId.value) || null
  );
});

// 账号详情（含 extension，用于头部展示）
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

// 账号 extension 信息
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
  <div class="resource-entry">
    <div class="resource-entry-sidebar">
      <AccountVendorGroup />
    </div>
    <div class="resource-entry-main">
      <template v-if="isListPage">
        <div class="resource-entry-header">
          <p class="resource-title">
            <span class="main-account-name">
              {{ currentAccount?.name || '全部账号' }}
            </span>
            <template v-if="(currentAccount as any)?.extension && !isOtherVendor">
              <div class="extension">
                <span>
                  {{ headerExtensionMap.firstLabel }}：
                  <span class="info-text">
                    {{ (currentAccount as any).extension?.[headerExtensionMap.firstField] }}
                  </span>
                </span>
                <span>
                  {{ headerExtensionMap.secondLabel }}：
                  <span class="info-text">
                    {{ (currentAccount as any).extension?.[headerExtensionMap.secondField] }}
                  </span>
                </span>
              </div>
            </template>
          </p>
        </div>
      </template>
      <HcmBreadcrumb v-else />
      <div class="resource-entry-content">
        <RouterView />
      </div>
    </div>
  </div>
</template>

<style lang="scss" scoped>
.resource-entry {
  display: flex;
  height: 100%;
}

.resource-entry-sidebar {
  flex-shrink: 0;
  width: 240px;
  height: 100%;
  border-right: 1px solid #eaebf0;
  overflow-y: auto;
}

.resource-entry-main {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.resource-entry-header {
  flex-shrink: 0;
  background: #fff;
  border-bottom: 1px solid #dcdee5;
}

.resource-entry-content {
  flex: 1;
  overflow-y: auto;
}

.resource-title {
  font-size: 16px;
  color: #313238;
  line-height: 24px;
  padding: 14px 0 9px 24px;
  display: flex;
  align-items: center;

  .extension {
    font-size: 14px;
    color: #63656e;

    & > span {
      margin-left: 20px;

      .info-text {
        color: #313238;
      }
    }
  }
}
</style>
