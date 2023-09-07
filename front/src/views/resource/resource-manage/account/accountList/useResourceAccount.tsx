import { useAccountStore } from '@/store';
import { useResourceAccountStore } from '@/store/useResourceAccountStore';
import { QueryRuleOPEnum } from '@/typings';
import { ref, watch, watchEffect } from 'vue';
import { useRoute, useRouter } from 'vue-router';

export const useResourceAccount = () => {
  // const { resourceAccount, setResourceAccount } = useResourceAccountStore();
  const resourceAccountStore = useResourceAccountStore();
  const accountStore = useAccountStore();
  const router = useRouter();
  const route = useRoute();
  const accountId = ref('');

  const setAccountId = (val: string) => {
    accountId.value = val;
    router.push({
      query: {
        ...route.query,
        accountId: val ? val : undefined,
      },
    });
  };

  watchEffect(async () => {
    if (!accountId.value) {
      resourceAccountStore.setResourceAccount({});
      return;
    }
    const res = await accountStore.getAccountList({
      filter: {
        op: QueryRuleOPEnum.AND,
        rules: [
          { field: 'id', op: QueryRuleOPEnum.EQ, value: accountId.value },
        ],
      },
      page: {
        start: 0,
        limit: 1,
      },
    });
    resourceAccountStore.setResourceAccount(res?.data?.details?.[0] || {});
    console.log(666666, res, resourceAccountStore.resourceAccount);
  });

  watch(
    () => route.query.accountId,
    (id) => {
      accountId.value = id as string;
    },
    {
      immediate: true,
    },
  );

  return {
    resourceAccount: resourceAccountStore.resourceAccount,
    accountId,
    setAccountId,
  };
};
