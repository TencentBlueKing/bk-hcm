import { defineStore } from 'pinia';
import { ref } from 'vue';
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
  }
};

export const useResourceAccountStore = defineStore('useResourceAccountStore', () => {
  const resourceAccount = ref<IAccount>(null);
  const currentVendor = ref<VendorEnum>(null);

  const setResourceAccount = (val: IAccount) => {
    resourceAccount.value = val;
  };
  const setCurrentVendor = (val: VendorEnum) => {
    currentVendor.value = val;
  };

  return {
    resourceAccount,
    setResourceAccount,
    currentVendor,
    setCurrentVendor,
  };
});
