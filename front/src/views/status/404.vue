<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';

const { t } = useI18n();
const router = useRouter();

const handleBack = () => {
  router.push('/');
};
</script>

<template>
  <bk-exception type="404" :description="t('当前页面不存在或功能暂未开放')">
    <bk-button theme="primary" @click="handleBack">{{ t('返回首页') }}</bk-button>
  </bk-exception>
</template>
