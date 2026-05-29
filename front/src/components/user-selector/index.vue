<script setup lang="ts">
import { computed, ref, useAttrs, watch } from 'vue';
import BkUserSelector from '@blueking/bk-user-selector';
import '@blueking/bk-user-selector/vue3/vue3.css';
import { useUserStore } from '@/store/user';
import type { DisplayType } from '@/components/form/typings';

interface RuleItem {
  validator: (val: any) => boolean;
  message: string;
}

export interface IUserSelectorProps {
  multiple?: boolean;
  disabled?: boolean;
  clearable?: boolean;
  placeholder?: string;
  fastSelect?: boolean;
  allowCreate?: boolean;
  display?: DisplayType;
  rules?: RuleItem[];
}

defineOptions({ name: 'user-selector' });

const model = defineModel<string | string[]>();

const props = withDefaults(defineProps<IUserSelectorProps>(), {
  multiple: true,
  allowCreate: true,
  clearable: true,
  placeholder: '请输入',
  fastSelect: true,
});

const emit = defineEmits<{
  change: [val: string | string[]];
}>();

const attrs = useAttrs();
const userStore = useUserStore();

// 校验状态
const isError = ref(false);
const errorMessage = ref('');

const localModel = computed<string[]>({
  get() {
    if (!model.value) {
      return [];
    }
    if (!Array.isArray(model.value)) {
      return [model.value];
    }
    return model.value;
  },
  set(val) {
    if (!props.multiple) {
      [model.value] = val;
    } else {
      model.value = val;
    }
  },
});

const tenantId = computed(() => userStore.tenantId);
const currentUserId = computed(() => props.fastSelect && userStore.username);
const apiBaseUrl = window.PROJECT_CONFIG.USER_MANAGE_URL;

const tagInputRef = ref();

const validate = (): boolean => {
  if (!props.rules?.length) {
    isError.value = false;
    errorMessage.value = '';
    return true;
  }
  for (const rule of props.rules) {
    if (!rule.validator(model.value)) {
      isError.value = true;
      errorMessage.value = rule.message;
      return false;
    }
  }
  isError.value = false;
  errorMessage.value = '';
  return true;
};

const handleChange = (val: string | string[]) => {
  emit('change', val);
};

// v-model 绑定值变化时自动校验
watch(model, () => {
  validate();
});

const focus = () => {
  tagInputRef.value?.focusInputTrigger?.();
};

defineExpose({
  getValue() {
    const valid = validate();
    if (!valid) {
      if (props.display?.on === 'cell') {
        return Promise.reject(new Error(errorMessage.value));
      }
      return undefined;
    }
    if (props.display?.on === 'cell') {
      return Promise.resolve(model.value);
    }
    return model.value;
  },
  focus,
  validate,
});
</script>

<template>
  <div class="user-selector-wrapper" :class="{ 'is-error': isError }">
    <bk-user-selector
      :class="{ 'bk-user-selector-cell': display?.on === 'cell' }"
      ref="tagInputRef"
      v-model="localModel"
      :multiple="multiple"
      :placeholder="placeholder"
      :tenant-id="tenantId"
      :current-user-id="currentUserId"
      :api-base-url="apiBaseUrl"
      :disabled="disabled"
      :clearable="clearable"
      v-bind="attrs"
      @change="handleChange"
    />
    <div v-if="isError" v-bk-tooltips="{ content: errorMessage }" class="select-error">
      <i class="ediatable-icon icon-exclamation-fill"></i>
    </div>
  </div>
</template>

<style lang="scss" scoped>
.user-selector-wrapper {
  position: relative;
  width: 100%;
  height: 100%;

  &.is-error {
    :deep(.tags-container) {
      background-color: #fff0f1;
    }
  }

  .select-error {
    position: absolute;
    top: 0;
    right: 30px;
    bottom: 0;
    display: flex;
    padding-right: 6px;
    font-size: 14px;
    color: #ea3636;
    align-items: center;
    cursor: pointer;
  }
}

.bk-user-selector-cell {
  width: 100%;
  height: 100%;
  border: 1px solid transparent;
  border-radius: 0;
  transition: all 0.3s;
  cursor: pointer;

  &:hover {
    border: 1px solid #a3c5fd;
  }

  :deep(.tags-container) {
    display: flex;
    border-color: transparent;
    min-height: 42px;
    padding-left: 14px;

    &:hover {
      cursor: pointer;
      background-color: #fafbfd !important;
    }

    // 无 tag 时始终显示输入框
    &:not(:has(.bk-tag)) {
      .search-input {
        display: block;
      }
    }

    // 有 tag 时默认隐藏输入框
    &:has(.bk-tag) {
      .search-input {
        display: none;
      }
    }

    // 聚焦时显示输入框（特异性高于上方，放在后面覆盖）
    &.focused {
      box-shadow: none;

      .search-input {
        display: block;
      }
    }
  }
}
</style>
