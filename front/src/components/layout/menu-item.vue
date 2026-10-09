<script lang="ts" setup>
import { computed } from 'vue';
import type { RouteLocationRaw } from 'vue-router';
import { GLOBAL_BIZS_KEY } from '@/common/constant';
import { toMenuKey, type IMenu } from '@/common/menu-service';
import { useAccountStore } from '@/store';

const props = defineProps<{ menu: IMenu; isBusiness: boolean }>();

const accountStore = useAccountStore();

const link = computed<RouteLocationRaw>(() => ({
  name: props.menu.route.name,
  query: props.isBusiness ? { [GLOBAL_BIZS_KEY]: accountStore.bizs } : undefined,
}));
</script>

<template>
  <RouterLink :to="link">
    <bk-menu-item :key="toMenuKey(menu.id)">
      <template #icon v-if="menu.icon">
        <i :class="['hcm-icon', menu.icon]" />
      </template>
      <p class="menu-item-content">
        <span class="text-ov">{{ menu.i18n }}</span>
      </p>
    </bk-menu-item>
  </RouterLink>
</template>

<style lang="scss" scoped>
.menu-item-content {
  display: flex;
  flex: 1;
  align-items: center;
  min-width: 0;
  padding-right: 16px;

  .text-ov {
    flex: 1;
  }
}
</style>
