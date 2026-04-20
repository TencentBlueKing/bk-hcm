<script setup lang="ts">
import { ref, inject, computed, type Ref } from 'vue';
import FormList from '@/components/form/list.vue';
import { ExclamationCircleShape } from 'bkui-vue/lib/icon';
import { usePermissionTemplateStore } from '@/store/cloud-account-manage/permission-template';
import { VendorEnum } from '@/common/constant';
import { DisplayType } from '@/components/form/typings';
import { useWhereAmI } from '@/hooks/useWhereAmI';

defineOptions({ name: 'ValidatedPermissionTemplateSelector' });

const model = defineModel<string[]>({ default: () => [] });

const props = withDefaults(
  defineProps<{
    placeholder?: string;
    display?: DisplayType;
    multiple?: boolean;
  }>(),
  {
    placeholder: '请选择',
    multiple: true,
  },
);

const currentVendor = inject<Ref<VendorEnum>>('currentVendor', ref(VendorEnum.TCLOUD));
const { getBizsId } = useWhereAmI();
const permissionTemplateStore = usePermissionTemplateStore();

const error = ref(false);
const errorMessage = ref('');
const isSelectColumn = computed(() => props.display?.on === 'cell'); // 是否可编辑表格

const listGenerator = computed(() =>
  permissionTemplateStore.createPermissionTemplateListGenerator(getBizsId(), currentVendor.value),
);

defineExpose({
  async getValue() {
    if (!model.value?.length) {
      error.value = true;
      errorMessage.value = '请选择权限模板';
      return Promise.reject(new Error(errorMessage.value));
    }
    error.value = false;
    errorMessage.value = '';
    return model.value;
  },
});
</script>

<template>
  <div
    class="validated-permission-template-selector"
    :class="{ 'is-error': isSelectColumn && error, 'select-column': isSelectColumn }"
  >
    <FormList
      v-model="model"
      :list-generator="listGenerator"
      :multiple="multiple"
      :placeholder="placeholder"
      display-key="name"
      id-key="id"
      collapse-tags
      v-bind="$attrs"
      @change="error = false"
    >
      <template #suffix v-if="isSelectColumn"></template>
    </FormList>
    <exclamation-circle-shape
      v-if="isSelectColumn && error"
      fill="#ea3636"
      class="icon"
      width="14"
      height="14"
      v-bk-tooltips="{ content: errorMessage, disabled: !error }"
    />
  </div>
</template>

<style lang="scss" scoped>
.validated-permission-template-selector {
  position: relative;
  height: 100%;

  .icon {
    position: absolute;
    right: 10px;
    top: 50%;
    transform: translateY(-50%);
    z-index: 99;
  }

  // 普通表单模式（编辑页面使用）
  &:not(.select-column) {
    :deep(.bk-select) {
      .bk-select-tag-input {
        font-size: 12px;
      }
    }
  }

  &.select-column {
    :deep(.bk-select) {
      height: 100%;

      &.is-focus .bk-select-tag {
        border-color: #3a84ff !important;
      }

      .bk-select-trigger {
        height: 100% !important;

        .bk-select-placeholder {
          line-height: 42px;
        }

        .bk-select-tag {
          min-height: 42px;
          border-color: transparent;
          border-radius: 0;
          padding: 0 16px;

          &:hover {
            background-color: #fafbfd;
            border-color: #a3c5fd !important;
          }
        }
      }
    }
  }

  &.is-error {
    :deep(.bk-select-tag) {
      background-color: #fff0f1;
    }
  }
}
</style>
