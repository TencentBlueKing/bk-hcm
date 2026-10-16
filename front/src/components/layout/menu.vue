<script lang="ts" setup>
import { computed, provide, ref, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { getCurrentTopMenu, toMenuKey, type IMenu } from '@/common/menu-service';
import { MENU_BUSINESS } from '@/constants/menu-symbol';
import { viewAuthConfig } from '@/constants/view-auth';
import { useAuthStore } from '@/store/auth';
import { useMenuFavorite } from '@/hooks/use-menu-favorite';
import { useMenuCollapse, FAVORITES_SECTION_KEY } from '@/hooks/use-menu-collapse';
import GlobalBusinessSelector from '@/components/business-selector/global.vue';
import MenuFavorites from './menu-favorites.vue';
import MenuItem from './menu-item.vue';

const route = useRoute();
const router = useRouter();
const authStore = useAuthStore();
const { lastClickSource, isFavorite, getFavoriteMenus } = useMenuFavorite();
const { setCollapsed, getOpenedKeys } = useMenuCollapse();

const ungrouped = Symbol('ungrouped');

const isBusiness = computed(() => route.matched[0]?.name === MENU_BUSINESS);

const currentTopMenuId = computed(() => route.matched[0]?.name as symbol | undefined);
const topMenuKey = computed(() => currentTopMenuId.value?.description ?? '');

// 菜单 id 与 viewId 是同一个 Symbol 常量，按引用比较
const hideMenuIds = new Set(viewAuthConfig.filter((config) => config.hideMenu).map((config) => config.viewId));

// 受控菜单需权限数据已返回且有权限才显示，接口失败时不显示
const isMenuAuthorized = (menu: IMenu) =>
  !hideMenuIds.has(menu.id) || (authStore.isViewPermissionCached(menu.id) && authStore.hasViewPermission(menu.id));

const allMenus = computed(() => (getCurrentTopMenu(route)?.menu ?? []).filter((menu) => menu.route?.name));
const authorizedMenus = computed(() => allMenus.value.filter(isMenuAuthorized));

// 失效收藏按菜单是否存在清理；无权限的只是不显示，权限恢复后收藏仍在
const favoriteMenus = computed(() => {
  if (!currentTopMenuId.value) return [];
  return getFavoriteMenus(currentTopMenuId.value, allMenus.value).filter(isMenuAuthorized);
});
const hasFavorites = computed(() => favoriteMenus.value.length > 0);

const menuGroups = computed(() => {
  const groups = new Map<string | symbol, IMenu[]>();
  authorizedMenus.value.forEach((menu) => {
    const group = menu.group ?? ungrouped;
    groups.set(group, [...(groups.get(group) ?? []), menu]);
  });
  return groups;
});

const activeKey = computed(() => toMenuKey(route.meta.activeKey));

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
const allGroupKeys = computed(() =>
  Array.from(menuGroups.value.keys()).filter((key): key is string => typeof key === 'string'),
);
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

const isActiveItemFavorited = computed(() => {
  const { activeKey } = route.meta;
  if (!currentTopMenuId.value || typeof activeKey !== 'symbol') return false;
  return isFavorite(currentTopMenuId.value, activeKey);
});

// 从收藏区点击时，常规菜单的激活项弱化
const dimRegularMenu = computed(() => lastClickSource.value === 'favorites' && isActiveItemFavorited.value);

const bizId = computed(() => Number(route.params.bizId));

provide('isBusinessNav', isBusiness);
provide('bizId', bizId);

// 切换业务后整页刷新，详情等页面回到所属列表
const handleChangeBusiness = (id: number) => {
  const targetRoute = route.meta.menu?.relative ?? route.name;
  const resolved = router.resolve({ name: targetRoute, params: { bizId: id } });
  window.location.hash = resolved.href;
  window.location.reload();
};
</script>

<template>
  <div class="hcm-sidebar">
    <!-- 固定区域：业务选择器 + 我的收藏 -->
    <div class="sidebar-fixed">
      <GlobalBusinessSelector v-if="isBusiness" :value="bizId" @change="handleChangeBusiness" />
      <bk-menu
        v-if="currentTopMenuId"
        :unique-open="false"
        :active-key="activeKey"
        :opened-keys="favOpenedKeys"
        :class="{ 'fav-menu-empty': !hasFavorites }"
        @open-change="handleFavOpenChange"
      >
        <MenuFavorites :top-menu-id="currentTopMenuId" :menus="favoriteMenus" />
      </bk-menu>
    </div>

    <!-- 可滚动区域：常规菜单 -->
    <div class="sidebar-scroll">
      <bk-menu
        :unique-open="false"
        :active-key="activeKey"
        :opened-keys="menuOpenedKeys"
        @open-change="handleMenuOpenChange"
      >
        <div :class="{ 'menu-regular-dim': dimRegularMenu }">
          <template v-for="[group, menus] of menuGroups">
            <template v-if="group === ungrouped">
              <MenuItem v-for="menu in menus" :key="toMenuKey(menu.id)" :menu="menu" :top-menu-id="currentTopMenuId" />
            </template>
            <bk-submenu v-else :title="group" :key="group">
              <template #icon>
                <i :class="['hcm-icon', menus?.[0]?.groupIcon]" />
              </template>
              <MenuItem v-for="menu in menus" :key="toMenuKey(menu.id)" :menu="menu" :top-menu-id="currentTopMenuId" />
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
    overflow: hidden auto;
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
