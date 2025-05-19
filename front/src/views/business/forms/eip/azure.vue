<script setup lang="ts">
import { ref, watch, defineEmits, defineExpose, defineProps } from 'vue';

import ZoneSelector from '@/components/zone-selector/index.vue';
import ResourceGroup from '@/components/resource-group/index.vue';

const emit = defineEmits(['change']);
defineProps({
  region: {
    type: String,
  },
  vendor: {
    type: String,
  },
});

const formData = ref({
  eip_name: '',
  eip_count: 1,
  ip_version: 'ipv4',
  sku_name: 'Standard', // Standard|Basic
  sku_tier: 'Regional', // Regional|Global
  allocation_method: 'Dynamic', // Dynamic|Static
  zone: '',
  resource_group_name: '',
  idle_timeout_in_minutes: 4,
});
const formRef = ref(null);
const rules = {
  resource_group_name: [
    {
      validator: (value: string) => value.length > 0,
      message: '资源组必填',
      trigger: 'blur',
    },
  ],
  zone: [
    {
      validator: (value: string) => value.length > 0,
      message: '可用性区域必填',
      trigger: 'blur',
    },
  ],
  eip_name: [
    {
      validator: (value: string) => value.length > 0,
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

const getIPAddressName = (val: string) => {
  return val === 'Dynamic' ? '动态' : '静态';
};

watch(
  () => formData,
  () => {
    formData.value.allocation_method = formData.value.sku_name === 'Basic' ? 'Dynamic' : 'Static';
    if (formData.value.ip_version === 'IPV6') {
      formData.value.idle_timeout_in_minutes = 4;
    }
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
    <bk-form-item label="资源组" property="resource_group_name" required>
      <resource-group :vendor="vendor" :region="region" v-model="formData.resource_group_name" />
    </bk-form-item>
    <bk-form-item label="名称" property="eip_name" required>
      <bk-input v-model="formData.eip_name" placeholder="请输入名称" />
    </bk-form-item>
    <bk-form-item label="IP版本" required>
      <bk-radio v-model="formData.ip_version" label="ipv4">IPv4</bk-radio>
      <bk-radio v-model="formData.ip_version" label="ipv6">IPv6</bk-radio>
    </bk-form-item>
    <bk-form-item
      label="SKU"
      description="公共 IP 的 SKU 必须与搭配使用的负载均衡器的 SKU 一致。了解详细信息（https://learn.microsoft.com/zh-cn/azure/load-balancer/skus#skus）"
      required
    >
      <bk-radio v-model="formData.sku_name" label="Standard">标准</bk-radio>
      <bk-radio v-model="formData.sku_name" label="Basic">基本</bk-radio>
    </bk-form-item>
    <bk-form-item
      v-if="formData.ip_version === 'ipv4' && formData.sku_name === 'Standard'"
      label="网络服务层级"
      property="sku_tier"
      required
    >
      <bk-radio v-model="formData.sku_tier" label="Regional">区域级</bk-radio>
      <bk-radio v-model="formData.sku_tier" label="Global">全局</bk-radio>
    </bk-form-item>
    <bk-form-item label="IP地址分配">
      {{ getIPAddressName(formData.allocation_method) }}
    </bk-form-item>
    <bk-form-item v-if="formData.sku_name === 'Standard'" label="可用性区域" property="zone" required>
      <zone-selector :vendor="vendor" :region="region" v-model="formData.zone" />
    </bk-form-item>
    <bk-form-item label="空闲超时(分钟)" property="idle_timeout_in_minutes" required min="4" max="30">
      <bk-input
        v-if="formData.ip_version === 'ipv4'"
        v-model.number="formData.idle_timeout_in_minutes"
        type="number"
        placeholder="请输入空闲超时，最小值4，最大值30"
      />
      <span v-else>{{ formData.idle_timeout_in_minutes }}</span>
    </bk-form-item>
    <bk-form-item label="数量">
      {{ formData.eip_count }}
    </bk-form-item>
  </bk-form>
</template>
