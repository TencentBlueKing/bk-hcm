<script setup lang="ts">
import { useI18n } from 'vue-i18n';

defineOptions({ name: 'device-container-empty' });

const { t } = useI18n();
</script>

<template>
  <div class="device-container-empty">
    <bk-exception
      class="exception-wrap-item exception-part"
      :title="t('暂无搜索结果')"
      :description="t('请在左侧选择检索条件后查询')"
      scene="part"
      type="search-empty"
    />
  </div>
</template>

<style scoped lang="scss"></style>
