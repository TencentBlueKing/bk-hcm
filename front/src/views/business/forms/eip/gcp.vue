<script setup lang="ts">
import { ref, watch, defineEmits, defineExpose, defineProps } from 'vue';

const emit = defineEmits(['change']);
const props = defineProps({
  region: {
    type: String,
  },
});

const formData = ref({
  eip_name: '',
  eip_count: 1,
  network_tier: 'PREMIUM',
  ip_version: 'IPV4',
  region: '',
});
const formRef = ref(null);
const type = ref('area');
const rules = {
  eip_count: [
    {
      validator: (value: number) => value > 0,
      message: '数量必填',
      trigger: 'blur',
    },
  ],
  eip_name: [
    {
      validator: (value: string) => /^(?:[a-z](?:[-a-z0-9]{0,61}[a-z0-9])?)$/.test(value),
      message: '名称需要符合如下正则表达式: /(?:[a-z](?:[-a-z0-9]{0,61}[a-z0-9])?)/',
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

watch(
  () => formData,
  () => {
    if (formData.value.network_tier === 'STANDARD') {
      formData.value.ip_version = 'IPV4';
      type.value = 'area';
    }
    formData.value.region = type.value === 'area' ? props.region : 'global';
    handleChange();
  },
  {
    immediate: true,
    deep: true,
  },
);

defineExpose([validate]);
</script>

<template>
  <bk-form label-width="150" ref="formRef" :model="formData" :rules="rules">
    <bk-form-item label="名称" property="eip_name" required>
      <bk-input v-model="formData.eip_name" placeholder="请输入名称" />
    </bk-form-item>
    <bk-form-item label="网络服务层级" required>
      <bk-radio v-model="formData.network_tier" label="PREMIUM">高级</bk-radio>
      <bk-radio v-model="formData.network_tier" label="STANDARD">标准</bk-radio>
    </bk-form-item>
    <bk-form-item label="IP版本" required>
      <bk-radio v-model="formData.ip_version" label="IPV4">IPv4</bk-radio>
      <bk-radio v-model="formData.ip_version" label="IPV6" :disabled="formData.network_tier === 'STANDARD'">
        IPv6
      </bk-radio>
    </bk-form-item>
    <bk-form-item label="类型">
      <bk-radio v-model="type" label="area">区域级</bk-radio>
      <bk-radio v-model="type" label="global" :disabled="formData.network_tier === 'STANDARD'">全球</bk-radio>
    </bk-form-item>
    <bk-form-item v-if="type === 'area'" label="区域">
      {{ region }}
    </bk-form-item>
    <bk-form-item label="数量" property="eip_count" required>1</bk-form-item>
  </bk-form>
</template>
