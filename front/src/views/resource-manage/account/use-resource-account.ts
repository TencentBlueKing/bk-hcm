import { onBeforeMount, ref, watch, watchEffect } from 'vue';
import { useRoute } from 'vue-router';
import { useAccountStore } from '@/store';
import { type IAccount, useResourceAccountStore } from '@/store/useResourceAccountStore';
import routerAction from '@/router/utils/action';
import {
  MENU_RESOURCE_DISK_APPLY,
  MENU_RESOURCE_HOST_APPLY,
  MENU_RESOURCE_LOAD_BALANCER_APPLY,
  MENU_RESOURCE_RESOURCE_LIST,
  MENU_RESOURCE_SUBNET_APPLY,
  MENU_RESOURCE_VPC_APPLY,
} from '@/constants/menu-symbol';

// 切换账号时留在原页面、只改 query 的路由；其余（资源详情）回到对应资源类型的列表
const STAY_ROUTES: symbol[] = [
  MENU_RESOURCE_RESOURCE_LIST,
  MENU_RESOURCE_HOST_APPLY,
  MENU_RESOURCE_VPC_APPLY,
  MENU_RESOURCE_DISK_APPLY,
  MENU_RESOURCE_SUBNET_APPLY,
  MENU_RESOURCE_LOAD_BALANCER_APPLY,
];

export const useResourceAccount = () => {
  const resourceAccountStore = useResourceAccountStore();
  const accountStore = useAccountStore();
  const route = useRoute();
  const accountId = ref('');

  const setAccountId = (val: string) => {
    accountId.value = val;
    if (STAY_ROUTES.includes(route.name as symbol)) {
      routerAction.redirect({ query: { ...route.query, accountId: val || undefined, id: undefined } });
      return;
    }
    routerAction.redirect(
      {
        name: MENU_RESOURCE_RESOURCE_LIST,
        params: { resourceType: route.params.resourceType },
        query: { accountId: val || undefined },
      },
      { replace: true },
    );
  };

  watchEffect(async () => {
    if (!accountId.value) {
      resourceAccountStore.setResourceAccount({} as IAccount);
      resourceAccountStore.setCurrentAccountSimpleInfo(null);
      return;
    }
    const res = await accountStore.getAccountDetail(accountId.value);
    // 详情是异步的：依赖账号 ID 的路由跳转要读 accountId，不能读 resourceAccount（连续切换账号时会读到旧值）
    resourceAccountStore.setResourceAccount(res?.data || {});
  });

  watch(
    () => route.query.accountId,
    (id) => {
      accountId.value = id as string;
    },
    { immediate: true },
  );

  onBeforeMount(() => {
    // 资源页面初始化时依赖账号 ID
    const id = route.query?.accountId as string;
    id && resourceAccountStore.setCurrentAccountSimpleInfo({ id });
  });

  return {
    accountId,
    setAccountId,
  };
};
