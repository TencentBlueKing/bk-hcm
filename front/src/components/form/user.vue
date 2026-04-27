<script setup lang="ts">
import { useAttrs, useTemplateRef } from 'vue';
import UserSelector from '@/components/user-selector/index.vue';

defineOptions({ name: 'hcm-form-user' });

const model = defineModel<string | string[]>();

const attrs = useAttrs();

const userSelectorRef = useTemplateRef<InstanceType<typeof UserSelector>>('userSelectorRef');

const focus = () => {
  userSelectorRef.value?.focus?.();
};

defineExpose({ focus });
</script>

<template>
  <user-selector
    ref="userSelectorRef"
    v-model="model"
    :collapse-tags="true"
    :allow-create="false"
    :multiple="true"
    v-bind="attrs"
  />
</template>
