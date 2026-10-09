<script lang="ts" setup>
import { computed } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { getCurrentTopMenu, toMenuKey, type IMenu } from '@/common/menu-service';
import { MENU_BUSINESS } from '@/constants/menu-symbol';
import { useVerify } from '@/hooks';
import BusinessSelector from './business-selector';
import MenuItem from './menu-item.vue';

defineProps<{ isOpen: boolean; reload: () => void }>();

const route = useRoute();
const router = useRouter();
const { authVerifyData } = useVerify();

const ungrouped = Symbol('ungrouped');

const isBusiness = computed(() => route.matched[0]?.name === MENU_BUSINESS);

const isMenuAuthorized = (menu: IMenu) => {
  const { checkAuth } = router.resolve({ name: menu.route.name }).meta;
  return !checkAuth || !!authVerifyData.value?.permissionAction?.[checkAuth];
};

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
