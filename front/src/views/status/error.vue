<script setup lang="ts">
import { computed } from 'vue';
import { useRoute } from 'vue-router';
import { useI18n } from 'vue-i18n';

const route = useRoute();
const { t } = useI18n();

const description = computed(() => (route.meta.errorMessage as string) || t('页面加载出错，请稍后重试'));

const handleRetry = () => {
  window.location.reload();
};
</script>

<template>
  <bk-exception type="500" :description="description">
    <bk-button theme="primary" @click="handleRetry">{{ t('刷新页面') }}</bk-button>
  </bk-exception>
</template>
