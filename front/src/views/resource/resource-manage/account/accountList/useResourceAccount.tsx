import { useAccountStore } from '@/store';
import { useResourceAccountStore } from '@/store/useResourceAccountStore';
import { onBeforeMount, ref, watch, watchEffect } from 'vue';
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
    const WHITE_LIST = [
      '/resource/service-apply/cvm',
      '/resource/service-apply/subnet',
      '/resource/service-apply/disk',
      '/resource/service-apply/vpc',
      '/resource/resource/recycle',
    ];
    const WHITE_LIST_2 = [
      '/resource/resource/account/detail',
      '/resource/resource/account/resource',
      '/resource/resource/account/manage',
    ];
    let path = '/resource/resource/';
    if (accountId.value && WHITE_LIST_2.includes(route.path)) path = '';
    if (WHITE_LIST.includes(route.path)) path = '';
    router.push({
      path,
      query: {
        ...route.query,
        accountId: val ? val : undefined,
        id: undefined,
      },
    });
  };

  watchEffect(async () => {
    if (!accountId.value) {
      resourceAccountStore.setResourceAccount({});
      resourceAccountStore.setCurrentAccountSimpleInfo(null);
      return;
    }
    const res = await accountStore.getAccountDetail(accountId.value);
    // !这里由于有网络请求，所以一些路由跳转使用到账号id的地方不能用resourceAccount，会出问题（比如两个账号之间切换 & 有连续的路由跳转操作）
    resourceAccountStore.setResourceAccount(res?.data || {});
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

  onBeforeMount(() => {
    // 设置当前账号id，用于资源下页面有关账号id的初始化操作
    const id = route.query?.accountId as string;
    id && resourceAccountStore.setCurrentAccountSimpleInfo({ id: id as string });
  });

  return {
    resourceAccount: resourceAccountStore.resourceAccount,
    accountId,
    setAccountId,
  };
};
