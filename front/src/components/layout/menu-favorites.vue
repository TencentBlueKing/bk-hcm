<script lang="ts" setup>
import { computed, ComputedRef, inject } from 'vue';
import { useRoute } from 'vue-router';
import { toMenuKey, type IMenu } from '@/common/menu-service';
import { useMenuFavorite } from '@/hooks/use-menu-favorite';
import { FAVORITES_SECTION_KEY } from '@/hooks/use-menu-collapse';

// menus 为已按权限过滤的收藏菜单
const props = defineProps<{
  topMenuId: symbol;
  menus: IMenu[];
}>();

const route = useRoute();
const { removeFavorite, setClickSource, lastClickSource } = useMenuFavorite();

const count = computed(() => props.menus.length);

const activeMenuKey = computed(() => {
  const key = route.meta.activeKey;
  return typeof key === 'symbol' ? key : null;
});

// 从常规菜单点击时，收藏区的激活项应弱化
const shouldDim = (menu: IMenu) => {
  return activeMenuKey.value === menu.id && lastClickSource.value === 'menu';
};

const isBusinessNav = inject<ComputedRef<boolean>>('isBusinessNav');
const bizId = inject<ComputedRef<number>>('bizId');

const getMenuLink = (menu: IMenu) =>
  isBusinessNav.value ? { name: menu.route.name, params: { bizId: bizId.value } } : { name: menu.route.name };

const handleRemove = (e: MouseEvent, menuId: symbol) => {
  e.preventDefault();
  e.stopPropagation();
  removeFavorite(props.topMenuId, menuId);
};

const handleItemClick = () => {
  setClickSource('favorites');
};
</script>

<template>
  <bk-submenu :key="FAVORITES_SECTION_KEY">
    <template #icon>
      <i
        :class="[
          'hcm-icon',
          count > 0 ? 'bkhcm-icon-collect' : 'bkhcm-icon-not-favorited',
          { 'fav-star-active': count > 0 },
        ]"
      />
    </template>
    <template #title>
      我的收藏
      <span v-if="count > 0" class="fav-count-badge">{{ count }}</span>
    </template>
    <RouterLink
      v-for="menu in menus"
      :key="toMenuKey(menu.id)"
      :to="getMenuLink(menu)"
      class="favorite-item-link"
      @click="handleItemClick"
    >
      <bk-menu-item :key="toMenuKey(menu.id)" :class="{ 'favorite-dim': shouldDim(menu) }" class="hcm-favorite-item">
        <template #default>
          {{ menu.i18n }}
          <i class="hcm-icon bkhcm-icon-close fav-remove-icon" @click="(e) => handleRemove(e, menu.id as symbol)" />
        </template>
      </bk-menu-item>
    </RouterLink>
  </bk-submenu>
</template>

<style lang="scss" scoped>
// 有收藏项时点亮为金色，无收藏项时跟随组件默认颜色
.fav-star-active {
  color: #f8b64f;
}

.fav-count-badge {
  min-width: 24px;
  height: 18px;
  line-height: 18px;
  text-align: center;
  border-radius: 9px;
  background: #4a5064;
  color: #96a2b9;
  font-size: 12px;
  padding: 0 5px;
  margin-left: 6px;
}

.favorite-item-link {
  text-decoration: none;
}

.hcm-favorite-item {
  // 从常规菜单点击时弱化收藏区激活项
  &.favorite-dim {
    opacity: 0.5;
  }

  .fav-remove-icon {
    display: none;
    font-size: 14px;
    color: #979ba5;
    margin-left: auto;
    margin-right: 16px;

    &:hover {
      color: white;
    }
  }

  &:hover .fav-remove-icon {
    display: block;
  }
}
</style>
