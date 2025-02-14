<script lang="ts" setup>
import { ref } from 'vue';
import UserSelector from '@/components/user-selector/index.vue';

export interface IChargeSelectorProps {
  manager?: string;
  bakManager?: string;
}

defineOptions({ name: 'ChargePersonSelector' });

const props = withDefaults(defineProps<IChargeSelectorProps>(), {
  manager: '',
  bakManager: '',
});

const formRef = ref(null);
const formData = ref({ manager: props.manager, bak_manager: props.bakManager });

const validate = () => {
  return formRef.value.validate();
};

defineExpose({ validate, formData });
</script>

<template>
  <bk-form class="chargePerson" label-width="150" form-type="vertical" :model="formData" ref="formRef">
    <bk-form-item label="主负责人" property="manager" required>
      <user-selector :multiple="false" v-model="formData.manager"></user-selector>
    </bk-form-item>
    <bk-form-item label="备份负责人" property="bak_manager" required>
      <user-selector :multiple="false" v-model="formData.bak_manager"></user-selector>
    </bk-form-item>
  </bk-form>
</template>

<style lang="scss" scoped>
.chargePerson {
  display: flex;
  justify-content: space-around;
  gap: 24px;

  .bk-form-item {
    flex-basis: 50%;
  }
}
</style>
