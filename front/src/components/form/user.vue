<script setup lang="ts">
import { ref, useAttrs } from 'vue';
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

const userSelectorRef = ref<InstanceType<typeof UserSelector>>();

defineExpose({
  getValue() {
    if (userSelectorRef.value?.getValue) {
      return userSelectorRef.value.getValue();
    }
    return model.value;
  },
});
</script>

<template>
  <user-selector v-model="model" ref="userSelectorRef" :display="display" :rules="rules" v-bind="attrs" />
</template>
