import { defineStore } from 'pinia';
import { computed, ref } from 'vue';
import { VendorEnum } from '@/common/constant';

export type IAccount = {
  bk_biz_ids: number[];
  created_at: string;
  creator: string;
  id: string;
  managers: string[];
  memo: string;
  name: string;
  price: string;
  price_unit: string;
  reviser: string;
  site: string;
  sync_failed_reason: string;
  sync_status: string;
  type: string;
  updated_at: string;
  vendor: VendorEnum;
  recycle_reserve_time: number;
  extension?: {
    cloud_main_account_id?: string;
    cloud_account_id?: string;
    cloud_iam_username?: string;
    cloud_project_id?: string;
    cloud_project_name?: string;
    cloud_tenant_id?: string;
    cloud_subscription_name?: string;
    cloud_sub_account_id?: string;
    cloud_sub_account_name?: string;
    [key: string]: any;
  };
};

export const useResourceAccountStore = defineStore('useResourceAccountStore', () => {
  const resourceAccount = ref<IAccount>(null);
  const currentAccountSimpleInfo = ref<Partial<IAccount>>(null); // 存储当前账号的简单信息（list接口返回值）
  const currentVendor = ref<VendorEnum>(null); // 当前选中的云厂商
  const selectedAccountId = computed(() => currentAccountSimpleInfo.value?.id || resourceAccount.value?.id || '');
  const vendorInResourcePage = computed(() => {
    return currentVendor.value || currentAccountSimpleInfo.value?.vendor || resourceAccount.value?.vendor;
  });

  const setResourceAccount = (val: IAccount) => {
    resourceAccount.value = val;
  };
  const setCurrentAccountSimpleInfo = (val: Partial<IAccount>) => {
    currentAccountSimpleInfo.value = val;
  };
  const setCurrentVendor = (val: VendorEnum) => {
    currentVendor.value = val;
  };

  // 当页面离开resource-page时，需要清空数据
  const clear = () => {
    setResourceAccount(null);
    setCurrentVendor(null);
    setCurrentAccountSimpleInfo(null);
  };

  return {
    resourceAccount,
    setResourceAccount,
    currentAccountSimpleInfo,
    setCurrentAccountSimpleInfo,
    currentVendor,
    setCurrentVendor,
    selectedAccountId,
    vendorInResourcePage,
    clear,
  };
});
