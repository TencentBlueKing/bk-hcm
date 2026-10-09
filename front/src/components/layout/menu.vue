<script lang="ts" setup>
import { computed } from 'vue';
import { useRoute } from 'vue-router';
import { getCurrentTopMenu, toMenuKey, type IMenu } from '@/common/menu-service';
import { MENU_BUSINESS } from '@/constants/menu-symbol';
import { viewAuthConfig } from '@/constants/view-auth';
import { useAuthStore } from '@/store/auth';
import BusinessSelector from './business-selector';
import MenuItem from './menu-item.vue';

defineProps<{ isOpen: boolean; reload: () => void }>();

const route = useRoute();
const authStore = useAuthStore();

const ungrouped = Symbol('ungrouped');

const isBusiness = computed(() => route.matched[0]?.name === MENU_BUSINESS);

// 菜单 id 与 viewId 是同一个 Symbol 常量，按引用比较
const hideMenuIds = new Set(viewAuthConfig.filter((config) => config.hideMenu).map((config) => config.viewId));

// 受控菜单需权限数据已返回且有权限才显示，接口失败时不显示
const isMenuAuthorized = (menu: IMenu) =>
  !hideMenuIds.has(menu.id) || (authStore.isViewPermissionCached(menu.id) && authStore.hasViewPermission(menu.id));

const menuGroups = computed(() => {
  const groups = new Map<string | symbol, IMenu[]>();
  (getCurrentTopMenu(route)?.menu ?? [])
    .filter((menu) => menu.route?.name && isMenuAuthorized(menu))
    .forEach((menu) => {
      const group = menu.group ?? ungrouped;
      groups.set(group, [...(groups.get(group) ?? []), menu]);
    });
  return groups;
});

const activeKey = computed(() => toMenuKey(route.meta.activeKey));
</script>

<template>
  <div class="home-menu">
    <BusinessSelector v-if="isBusiness && isOpen" :reload="reload" />
    <bk-menu class="menu-warp" :unique-open="false" :active-key="activeKey">
      <template v-for="[group, menus] of menuGroups" :key="group">
        <template v-if="group === ungrouped">
          <MenuItem v-for="menu in menus" :key="toMenuKey(menu.id)" :menu="menu" :is-business="isBusiness" />
        </template>
        <bk-menu-group v-else :name="group">
          <MenuItem v-for="menu in menus" :key="toMenuKey(menu.id)" :menu="menu" :is-business="isBusiness" />
        </bk-menu-group>
      </template>
    </bk-menu>
  </div>
</template>

<style lang="scss" scoped>
.menu-warp {
  width: 240px;
}
</style>
