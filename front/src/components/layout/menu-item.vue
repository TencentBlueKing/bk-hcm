<script lang="ts" setup>
import { computed, ComputedRef, inject } from 'vue';
import { type IMenu } from '@/common/menu-service';
import { useMenuFavorite } from '@/hooks/use-menu-favorite';

const props = defineProps<{ menu: IMenu; topMenuId: symbol }>();

const isBusinessNav = inject<ComputedRef<boolean>>('isBusinessNav');
const bizId = inject<ComputedRef<number>>('bizId');

const { isFavorite, toggleFavorite, setClickSource } = useMenuFavorite();

const favorited = computed(() => isFavorite(props.topMenuId, props.menu.id as symbol));

const getMenuLink = (menu: IMenu) => {
  if (isBusinessNav.value) {
    return {
      name: menu.route.name,
      params: {
        bizId: bizId.value,
      },
    };
  }
  return menu.route;
};

const handleToggleFavorite = (e: MouseEvent) => {
  e.preventDefault();
  e.stopPropagation();
  toggleFavorite(props.topMenuId, props.menu.id as symbol);
};

const handleItemClick = () => {
  setClickSource('menu');
};
</script>

<template>
  <RouterLink :to="getMenuLink(menu)" @click="handleItemClick">
    <bk-menu-item :key="menu.id" class="hcm-menu-item">
      <template #icon v-if="menu.icon">
        <i :class="['hcm-icon', menu.icon]" />
      </template>
      <template #default>
        {{ menu.i18n }}
        <span :class="['collect', { collected: favorited }]" @click="handleToggleFavorite">
          <i :class="['hcm-icon', favorited ? 'bkhcm-icon-collect' : 'bkhcm-icon-not-favorited']" />
        </span>
      </template>
    </bk-menu-item>
  </RouterLink>
</template>

<style lang="scss" scoped>
.hcm-menu-item {
  .collect {
    display: none;
    margin-left: auto;
    margin-right: 24px;
    cursor: pointer;

    .hcm-icon {
      font-size: 16px;
    }
    .bkhcm-icon-not-favorited {
      color: #fff;
    }
    .bkhcm-icon-collect {
      color: #f8b64f;
    }

    &.collected {
      display: block;
    }
  }
  &:hover {
    .collect {
      display: block;
    }
  }
}
</style>
