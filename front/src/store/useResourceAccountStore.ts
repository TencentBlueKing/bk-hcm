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
};

export const useResourceAccountStore = defineStore(
  'useResourceAccountStore',
  () => {
    const resourceAccount = ref<IAccount>(null);
    const setResourceAccount = (val: IAccount) => {
      resourceAccount.value = val;
    };

    return {
      resourceAccount,
      setResourceAccount,
    };
  },
);
