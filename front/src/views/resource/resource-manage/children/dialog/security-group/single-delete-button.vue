<script setup lang="ts">
import { useI18n } from 'vue-i18n';
const { t } = useI18n();

defineEmits<(e: 'del') => void>();
</script>
<template>
  <!-- 将透传attrs -->
  <bk-button theme="danger" @click="$emit('del')">
    {{ t('删除') }}
  </bk-button>
</template>
