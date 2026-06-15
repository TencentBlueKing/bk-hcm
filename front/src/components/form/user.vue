<script setup lang="ts">
import { useAttrs, useTemplateRef } from 'vue';
import UserSelector from '@/components/user-selector/index.vue';
import { DisplayType } from './typings';
import type { Rules } from '@blueking/ediatable';

defineOptions({ name: 'hcm-form-user' });

const model = defineModel<string | string[]>();

defineProps<{
  display?: DisplayType;
  rules?: Rules;
}>();

const attrs = useAttrs();

const userSelectorRef = useTemplateRef<InstanceType<typeof UserSelector>>('userSelectorRef');

const focus = () => {
  userSelectorRef.value?.focus?.();
};

defineExpose({
  getValue() {
    if (userSelectorRef.value?.getValue) {
      return userSelectorRef.value.getValue();
    }
    return model.value;
  },
  focus,
});
</script>

<template>
  <user-selector
    ref="userSelectorRef"
    v-model="model"
    :display="display"
    :rules="rules"
    :collapse-tags="true"
    :allow-create="false"
    :multiple="true"
    v-bind="attrs"
  />
</template>
