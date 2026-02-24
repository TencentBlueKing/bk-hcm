<script lang="ts" setup>
import { computed, ComputedRef, inject } from 'vue';
import { useRoute } from 'vue-router';
import { type IMenu } from '@/common/menu-service';
import { useMenuFavorite } from '@/hooks/use-menu-favorite';
import { FAVORITES_SECTION_KEY } from '@/hooks/use-menu-collapse';

const props = defineProps<{
  topMenuId: symbol;
  menus: IMenu[];
}>();

const route = useRoute();
const { getFavoriteMenus, getFavoriteCount, removeFavorite, setClickSource, lastClickSource } = useMenuFavorite();

const isBusinessNav = inject<ComputedRef<boolean>>('isBusinessNav');
const bizId = inject<ComputedRef<number>>('bizId');

const count = computed(() => getFavoriteCount(props.topMenuId));
const favoriteMenus = computed(() => getFavoriteMenus(props.topMenuId, props.menus));

const activeMenuKey = computed(() => {
  const key = route.meta.activeKey;
  return typeof key === 'symbol' ? key : null;
});

// 从常规菜单点击时，收藏区的激活项应弱化
const shouldDim = (menu: IMenu) => {
  return activeMenuKey.value === menu.id && lastClickSource.value === 'menu';
};

const getMenuLink = (menu: IMenu) => {
  if (isBusinessNav?.value) {
    return { name: menu.route.name, params: { bizId: bizId.value } };
  }
  return menu.route;
};

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
      <i :class="['hcm-icon bkhcm-icon-collect', { 'fav-star-active': count > 0 }]" />
    </template>
    <template #title>
      我的收藏
      <span v-if="count > 0" class="fav-count-badge">{{ count }}</span>
    </template>
    <RouterLink
      v-for="menu in favoriteMenus"
      :key="menu.id"
      :to="getMenuLink(menu)"
      class="favorite-item-link"
      @click="handleItemClick"
    >
      <bk-menu-item :key="menu.id" :class="{ 'favorite-dim': shouldDim(menu) }" class="hcm-favorite-item">
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
