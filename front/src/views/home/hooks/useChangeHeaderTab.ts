import { ref, watch } from 'vue';
import { useRoute, useRouter, type RouteRecordRaw } from 'vue-router';
// import routes
import resource from '@/router/module/resource';
import service from '@/router/module/service';
import { businessViews } from '@/views';

import scheme from '@/router/module/scheme';
import bill from '@/router/module/bill';
// import stores
import { useAccountStore } from '@/store';
import { useResourceAccountStore } from '@/store/useResourceAccountStore';
import { GLOBAL_BIZS_KEY } from '@/common/constant';

// home页切换header-tab相关业务逻辑
export default () => {
  const accountStore = useAccountStore();
  const resourceAccountStore = useResourceAccountStore();
  // use hooks
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
      bizs = accountStore.bizs;
    }
    if (id !== 'resource') {
      resourceAccountStore.clear();
    }
    router.push({ path, query: { [GLOBAL_BIZS_KEY]: bizs } });
  };

  // 更新左侧 menus 菜单, 并更新全局业务id
  const changeMenus = (id: string, ...subPath: string[]) => {
    // 更新当前 active header-tab
    topMenuActiveItem.value = id;
    switch (id) {
      case 'business':
        menus.value = businessViews;
        break;
      case 'resource':
        menus.value = resource;
        accountStore.updateBizsId(0); // 初始化业务ID
        break;
      case 'service':
        menus.value = service;
        break;
      case 'scheme':
        menus.value = scheme;
        accountStore.updateBizsId(0); // 初始化业务ID
        break;
      case 'bill':
        menus.value = bill;
        break;
      default:
        if (subPath[0] === 'biz_access') {
          topMenuActiveItem.value = 'business';
          menus.value = businessViews;
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
