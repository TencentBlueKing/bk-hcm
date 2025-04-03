<script setup lang="ts">
import { ISecurityGroupOperateItem } from '@/store/security-group';
import { TagThemeEnum } from 'bkui-vue/lib/shared';

defineProps<{ resources: ISecurityGroupOperateItem['resources']; tagTheme?: TagThemeEnum }>();
</script>

<template>
  <div class="rel-res-wrap">
    <div class="rel-res-item" v-for="resource in resources" :key="resource.res_name">
      {{ resource.res_name }}
      <bk-tag :theme="tagTheme" size="small" class="ml2">
        <span class="number">{{ resource.count }}</span>
      </bk-tag>
    </div>
  </div>
</template>

<style scoped lang="scss">
.rel-res-wrap {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  row-gap: 4px;

  .rel-res-item {
    display: flex;
    align-items: center;

    .number {
      padding: 0 4px;
      font-size: 12px;
    }

    &:not(:last-of-type)::after {
      content: '|';
      margin: 0 4px;
    }
  }
}
</style>
