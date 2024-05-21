import { Ref, ref, watch } from 'vue';
import { useRoute, useRouter, type RouteRecordRaw } from 'vue-router';
// import routes
import workbench from '@/router/module/workbench';
import resource from '@/router/module/resource';
import service from '@/router/module/service';
import business from '@/router/module/business';
import scheme from '@/router/module/scheme';
// import stores
import { useAccountStore } from '@/store';
// import hooks
import { useWhereAmI, Senarios } from '@/hooks/useWhereAmI';
import { localStorageActions } from '@/common/util';
// 点击跳转header-tab时清除一下pinia
import { useResourceAccountStore } from '@/store/useResourceAccountStore';
// home页切换header-tab相关业务逻辑
export default (businessId: Ref<number>, businessList: Ref<any[]>) => {
  const accountStore = useAccountStore();
  const resourceAccountStore = useResourceAccountStore();
  // use hooks
  const { whereAmI } = useWhereAmI();
  const route = useRoute();
  const router = useRouter();

  // define data
  const topMenuActiveItem = ref(''); // 当前 active header-tab
  const menus = ref<RouteRecordRaw[]>([]); // 左侧 menus 菜单
  const curPath = ref(''); // 控制 router-view 是否有 padding

  // 点击 header-tab handler
  const handleHeaderMenuClick = (id: string, path: string) => {
    let bizs;
    if (id === 'business') {
      bizs = localStorageActions.get('bizs');
    }
    resourceAccountStore.setResourceAccount({});
    router.push({ path, query: { bizs } });
  };

  // 获取业务列表
  const getBusinessList = async () => {
    const res = await accountStore.getBizListWithAuth();
    // 更新业务列表
    businessList.value = res.data;
    // 没有权限
    if (!businessList.value.length && whereAmI.value === Senarios.business) {
      router.push({ name: '403', params: { id: 'biz_access' } });
      return;
    }
    // 先从 url 中获取 bizs 参数, 如果没有, 则从 localStorage 中获取, 如果还是没有, 则取第一个
    let { bizs } = route.query;
    if (!bizs) {
      bizs = localStorageActions.get('bizs');
    }
    businessId.value = Number(bizs) || res.data[0].id;
    // 设置全局业务id
    accountStore.updateBizsId(businessId.value);
    // 持久化存储全局业务id
    localStorageActions.set('bizs', businessId.value);
  };

  // 更新左侧 menus 菜单, 并更新全局业务id
  const changeMenus = (id: string, ...subPath: string[]) => {
    // openedKeys.push(`/${id}`); // 这个其实没用，因为 openedKeys 是打开的submenu key值, 但咱们的左侧 menu 菜单都只有一级

    // 更新当前 active header-tab
    topMenuActiveItem.value = id;
    switch (id) {
      case 'business':
        menus.value = business;
        // 业务下需要获取业务列表
        getBusinessList();
        break;
      case 'resource':
        menus.value = resource;
        accountStore.updateBizsId(0); // 初始化业务ID
        break;
      case 'service':
        menus.value = service;
        break;
      case 'workbench':
        menus.value = workbench;
        accountStore.updateBizsId(0); // 初始化业务ID
        break;
      case 'scheme':
        menus.value = scheme;
        accountStore.updateBizsId(0); // 初始化业务ID
        break;
      default:
        if (subPath[0] === 'biz_access') {
          topMenuActiveItem.value = 'business';
          menus.value = business;
        } else {
          topMenuActiveItem.value = 'resource';
          menus.value = resource;
        }
        accountStore.updateBizsId(''); // 初始化业务ID
        break;
    }
  };

  // 一进来 home 页面, 就调用一次 changeMenus 函数
  watch(
    () => route.path,
    (val) => {
      curPath.value = val;
      const pathArr = val.slice(1, val.length).split('/');
      changeMenus(pathArr.shift(), ...pathArr);
    },
    {
      immediate: true,
    },
  );

  return {
    topMenuActiveItem,
    menus,
    curPath,
    handleHeaderMenuClick,
  };
};
