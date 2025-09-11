<script setup lang="ts">
import { useI18n } from 'vue-i18n';

defineOptions({ name: 'device-container-large-data' });

const { t } = useI18n();
</script>

<template>
  <div class="device-container-empty">
    <bk-exception
      class="exception-wrap-item exception-part"
      :title="t('查询结果超出数量限制')"
      :description="t('当前检索条件展示上限为10,000条，请添加检索条件后查询')"
      scene="part"
      type="500"
    />
  </div>
</template>

<style scoped lang="scss"></style>
