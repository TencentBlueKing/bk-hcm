<script setup lang="ts">
import { useI18n } from 'vue-i18n';
const { t } = useI18n();

withDefaults(defineProps<{ disabled: boolean; loading: boolean }>(), {
  disabled: true,
  loading: false,
});

const emit = defineEmits<(e: 'del') => void>();

const handleClick = () => {
  emit('del');
};
</script>
<template>
  <bk-button theme="danger" :disabled="disabled" :loading="loading" @click="handleClick">
    {{ t('删除') }}
  </bk-button>
</template>
