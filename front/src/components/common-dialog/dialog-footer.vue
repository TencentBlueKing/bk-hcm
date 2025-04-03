<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import { ThemeEnum } from 'bkui-vue/lib/shared';

interface IProps {
  disabled?: boolean;
  loading?: boolean;
  confirmText?: string;
  confirmButtonTheme?: ThemeEnum;
  cancelText?: string;
}

withDefaults(defineProps<IProps>(), {
  disabled: false,
  loading: false,
  confirmText: '确定',
  confirmButtonTheme: ThemeEnum.PRIMARY,
  cancelText: '取消',
});
const emit = defineEmits(['confirm', 'closed']);

const { t } = useI18n();
</script>

<template>
  <bk-button class="mr8" :theme="confirmButtonTheme" :disabled="disabled" :loading="loading" @click="emit('confirm')">
    {{ t(confirmText) }}
  </bk-button>
  <bk-button :disabled="loading" @click="emit('closed')">
    {{ t(cancelText) }}
  </bk-button>
</template>

<style scoped lang="scss"></style>
