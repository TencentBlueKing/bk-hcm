<script setup lang="ts">
import { ref, watch, defineEmits, defineExpose } from 'vue';

const emit = defineEmits(['change']);

const formData = ref({
  eip_name: '',
  eip_count: 1,
  service_provider: 'BGP',
  address_type: 'EIP',
});
const formRef = ref(null);
const rules = {
  eip_count: [
    {
      validator: (value: number) => value > 0,
      message: '请输入大于0的整数',
      trigger: 'blur',
    },
  ],
  eip_name: [
    {
      validator: (value: string) => {
        return value.trim().length > 0;
      },
      message: '名称必填',
      trigger: 'blur',
    },
  ],
};

const handleChange = () => {
  emit('change', formData.value);
};

const validate = () => {
  return formRef.value.validate();
};

watch(() => formData, handleChange, {
  immediate: true,
  deep: true,
});

defineExpose([validate]);
</script>

<template>
  <bk-form label-width="150" ref="formRef" :model="formData" :rules="rules">
    <bk-form-item label="名称" property="eip_name" required>
      <bk-input v-model="formData.eip_name" placeholder="请输入名称" />
    </bk-form-item>
    <bk-form-item label="IP地址类型" required>
      <bk-radio model-value="常规BGP IP" label="常规BGP IP" />
    </bk-form-item>
    <bk-form-item label="数量" property="eip_count" required>
      <bk-input v-model="formData.eip_count" placeholder="请输入数量" />
    </bk-form-item>
  </bk-form>
</template>
