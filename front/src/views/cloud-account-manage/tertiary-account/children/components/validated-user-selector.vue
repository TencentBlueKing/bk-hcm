<script setup lang="ts">
import { ref } from 'vue';
import UserSelector from '@/components/user-selector/index.vue';
import { ExclamationCircleShape } from 'bkui-vue/lib/icon';

defineOptions({ name: 'ValidatedUserSelector' });

const model = defineModel<string[]>({ default: () => [] });

withDefaults(
  defineProps<{
    multiple?: boolean;
    collapseTags?: boolean;
    allowCreate?: boolean;
    placeholder?: string;
  }>(),
  {
    multiple: true,
    collapseTags: false,
    allowCreate: true,
    placeholder: '请输入负责人',
  },
);

const error = ref(false);
const errorMessage = ref('');

defineExpose({
  async getValue() {
    if (!model.value?.length) {
      error.value = true;
      errorMessage.value = '负责人不能为空';
      return Promise.reject(new Error(errorMessage.value));
    }
    error.value = false;
    errorMessage.value = '';
    return model.value;
  },
});
</script>

<template>
  <div class="validated-user-selector" :class="{ 'is-error': error }">
    <UserSelector
      v-model="model"
      :multiple="multiple"
      :collapse-tags="collapseTags"
      :allow-create="allowCreate"
      :placeholder="placeholder"
      @update:model-value="error = false"
    />
    <exclamation-circle-shape
      v-if="error"
      fill="#ea3636"
      class="icon"
      width="14"
      height="14"
      v-bk-tooltips="{ content: errorMessage, disabled: !error }"
    />
  </div>
</template>

<style lang="scss" scoped>
.validated-user-selector {
  position: relative;

  .icon {
    position: absolute;
    right: 20px;
    top: 50%;
    transform: translateY(-50%);
    z-index: 99;
  }

  &.is-error {
    :deep(.bk-tag-input-trigger) {
      background-color: #fff0f1;
    }
  }
}
</style>
