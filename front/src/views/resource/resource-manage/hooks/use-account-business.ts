import { ref, watchEffect } from 'vue';
import { useAccountStore } from '@/store/account';
import { useBusinessGlobalStore, type IBusinessItem } from '@/store/business-global';

export const useAccountBusiness = (accountId: string) => {
  const accountStore = useAccountStore();
  const businessGlobalStore = useBusinessGlobalStore();

  const accountBizList = ref<IBusinessItem[]>([]);

  const isAccountDetailLoading = ref(false);

  watchEffect(async () => {
    isAccountDetailLoading.value = true;

    // 账号业务列表等于-1时，管理业务使用全部业务，否则限定为账号业务列表
    const accountUsageBizRes = await accountStore.getAccountUsageBiz(accountId);
    const accountBizIds = accountUsageBizRes?.data;
    if (accountBizIds?.[0] !== -1) {
      accountBizList.value = businessGlobalStore.businessFullList.filter((item) => accountBizIds.includes(item.id));
    } else {
      // null表示使用全部业务
      accountBizList.value = null;
    }

    isAccountDetailLoading.value = false;
  });

  return {
    accountBizList,
    isAccountDetailLoading,
  };
};
