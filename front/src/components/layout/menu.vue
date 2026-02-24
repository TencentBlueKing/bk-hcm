<script lang="ts" setup>
import { computed, provide, ref, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { getMenus, type IMenu } from '@/common/menu-service';
import { MENU_BUSINESS } from '@/constants/menu-symbol';
import { useAuthStore } from '@/store/auth';
import { useMenuFavorite } from '@/hooks/use-menu-favorite';
import { useMenuCollapse, FAVORITES_SECTION_KEY } from '@/hooks/use-menu-collapse';
import GlobalBusinessSelector from '@/components/business-selector/global.vue';
import MenuFavorites from './menu-favorites.vue';
import MenuItem from './menu-item.vue';

const route = useRoute();
const router = useRouter();
const authStore = useAuthStore();
const { lastClickSource, isFavorite, getFavoriteCount } = useMenuFavorite();
const { setCollapsed, getOpenedKeys } = useMenuCollapse();

const topMenus = getMenus();
const ungrouped = Symbol('ungrouped');

const filterMenusByPermission = (menus: IMenu[]): IMenu[] => {
  return menus.filter((menu) => {
    if (typeof menu.id === 'symbol') {
      return authStore.hasViewPermission(menu.id);
    }
    return true;
  });
};

const currentTopMenuId = computed(() => {
  return route.matched[0]?.name as symbol | undefined;
});

const topMenuKey = computed(() => currentTopMenuId.value?.description ?? '');

const filteredMenus = computed(() => {
  const topMenu = topMenus.find((menu) => menu.id === currentTopMenuId.value);
  return filterMenusByPermission(topMenu?.menu ?? []);
});

const currentMenus = computed(() => {
  const menuGroup = new Map<string | symbol, IMenu[]>();
  for (const menu of filteredMenus.value) {
    const group = menu.group ?? ungrouped;
    if (menuGroup.has(group)) {
      menuGroup.set(group, [...menuGroup.get(group), menu]);
    } else {
      menuGroup.set(group, [menu]);
    }
  }
  return menuGroup;
});

const activeMenuKey = computed(() => {
  const key = route.meta.activeKey;
  return typeof key === 'symbol' ? key.toString() : (key as string) ?? '';
});

const hasFavorites = computed(() => {
  return currentTopMenuId.value && getFavoriteCount(currentTopMenuId.value) > 0;
});

// 收藏区 openedKeys（独立 bk-menu）：无收藏时强制收起
const favOpenedKeys = ref<string[]>([]);
const computeFavOpenedKeys = () => {
  if (!topMenuKey.value || !hasFavorites.value) {
    favOpenedKeys.value = [];
    return;
  }
  favOpenedKeys.value = getOpenedKeys(topMenuKey.value, [FAVORITES_SECTION_KEY]);
};

// 常规菜单 openedKeys
const allGroupKeys = computed(() => {
  return Array.from(currentMenus.value.keys()).filter((k): k is string => typeof k === 'string');
});
const menuOpenedKeys = ref<string[]>([]);
const computeMenuOpenedKeys = () => {
  if (!topMenuKey.value) return;
  menuOpenedKeys.value = getOpenedKeys(topMenuKey.value, allGroupKeys.value);
};

watch([topMenuKey, allGroupKeys], computeMenuOpenedKeys, { immediate: true });
watch([topMenuKey, hasFavorites], computeFavOpenedKeys, { immediate: true });

// 收藏区折叠事件：无收藏时忽略展开操作
const handleFavOpenChange = (opened: boolean, info: { key: string }) => {
  if (!topMenuKey.value || !info?.key) return;
  if (!hasFavorites.value && opened) return;
  setCollapsed(topMenuKey.value, info.key, !opened);
  computeFavOpenedKeys();
};

// 常规菜单折叠事件
const handleMenuOpenChange = (opened: boolean, info: { key: string }) => {
  if (!topMenuKey.value || !info?.key) return;
  setCollapsed(topMenuKey.value, info.key, !opened);
  computeMenuOpenedKeys();
};

const isBusinessNav = computed(() => {
  const {
    matched: [topRoute],
  } = route;
  return topRoute?.name === MENU_BUSINESS;
});

const bizId = computed(() => Number(route.params.bizId));

const handleChangeBusiness = (id: number) => {
  router.push({
    name: route.name,
    params: { ...route.params, bizId: id },
    query: route.query,
  });
};

const isActiveItemFavorited = computed(() => {
  const { activeKey } = route.meta;
  if (!currentTopMenuId.value || typeof activeKey !== 'symbol') return false;
  return isFavorite(currentTopMenuId.value, activeKey);
});

const dimRegularMenu = computed(() => {
  return lastClickSource.value === 'favorites' && isActiveItemFavorited.value;
});

provide('isBusinessNav', isBusinessNav);
provide('bizId', bizId);
</script>

<template>
  <div class="hcm-sidebar">
    <!-- 固定区域：业务选择器 + 我的收藏 -->
    <div class="sidebar-fixed">
      <GlobalBusinessSelector v-if="isBusinessNav" :value="bizId" @change="handleChangeBusiness" />
      <bk-menu
        v-if="currentTopMenuId"
        :unique-open="false"
        :active-key="activeMenuKey"
        :opened-keys="favOpenedKeys"
        :class="{ 'fav-menu-empty': !hasFavorites }"
        @open-change="handleFavOpenChange"
      >
        <MenuFavorites :top-menu-id="currentTopMenuId" :menus="filteredMenus" />
      </bk-menu>
    </div>

    <!-- 可滚动区域：常规菜单 -->
    <div class="sidebar-scroll">
      <bk-menu
        :unique-open="false"
        :active-key="activeMenuKey"
        :opened-keys="menuOpenedKeys"
        @open-change="handleMenuOpenChange"
      >
        <div :class="{ 'menu-regular-dim': dimRegularMenu }">
          <template v-for="[group, menus] of currentMenus">
            <template v-if="group === ungrouped">
              <menu-item v-for="menu in menus" :key="menu.id" :menu="menu" :top-menu-id="currentTopMenuId" />
            </template>
            <bk-submenu v-else :title="group" :key="group">
              <template #icon>
                <i :class="['hcm-icon', menus?.[0]?.groupIcon]" />
              </template>
              <menu-item v-for="menu in menus" :key="menu.id" :menu="menu" :top-menu-id="currentTopMenuId" />
            </bk-submenu>
          </template>
        </div>
      </bk-menu>
    </div>
  </div>
</template>

<style lang="scss" scoped>
.hcm-sidebar {
  display: flex;
  flex-direction: column;
  height: 100%;
  overflow: hidden;

  .sidebar-fixed {
    flex-shrink: 0;
  }

  .sidebar-scroll {
    flex: 1;
    overflow-y: auto;
    overflow-x: hidden;
  }
}

// 无收藏时隐藏折叠箭头、禁止展开
.fav-menu-empty {
  :deep(.submenu-header) {
    cursor: default;
    pointer-events: none;
  }

  :deep(.submenu-header-collapse) {
    display: none;
  }
}

.menu-regular-dim {
  :deep(.bk-menu-item.is-active) {
    opacity: 0.5;
  }
}
</style>
