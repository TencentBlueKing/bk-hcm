<script lang="ts" setup>
import { ref, nextTick, watch } from 'vue';

export interface IChargeSelectorProps {
  manager?: string;
  bakManager?: string;
}

defineOptions({ name: 'SecurityGroupManagerSelector' });

const props = withDefaults(defineProps<IChargeSelectorProps>(), {
  manager: '',
  bakManager: '',
});

const formRef = ref(null);
const formData = ref({ manager: props.manager, bak_manager: props.bakManager });

const validate = () => {
  return formRef.value.validate();
};

const reset = () => {
  formData.value.manager = '';
  formData.value.bak_manager = '';
  nextTick(() => formRef.value?.clearValidate());
};

watch(
  () => [props.manager, props.bakManager],
  (val) => {
    const [manager, bakManager] = val;
    formData.value.manager = manager;
    formData.value.bak_manager = bakManager;
  },
);

defineExpose({ validate, formData, reset });
</script>

<template>
  <bk-form class="manager-selector" label-width="150" form-type="vertical" :model="formData" ref="formRef">
    <bk-form-item label="主负责人" property="manager" required>
      <hcm-form-user :multiple="false" v-model="formData.manager"></hcm-form-user>
    </bk-form-item>
    <bk-form-item label="备份负责人" property="bak_manager" required>
      <hcm-form-user :multiple="false" v-model="formData.bak_manager"></hcm-form-user>
    </bk-form-item>
  </bk-form>
</template>

<style lang="scss" scoped>
.manager-selector {
  display: flex;
  justify-content: space-around;
  gap: 24px;

  .bk-form-item {
    flex-basis: 50%;
  }
}
</style>
