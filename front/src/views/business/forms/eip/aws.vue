<script setup lang="ts">
import {
  ref,
  watch,
  defineEmits,
  defineExpose,
} from 'vue';

const props = defineProps({
  region: {
    type: String
  }
});

const emit = defineEmits(['change']);

const formData = ref({
  public_ipv4_pool: '',
  network_border_group: props.region,
  eip_count: 1,
})
const formRef = ref(null);


const handleChange = () => {
  formData.value.network_border_group = props.region
  emit('change', formData.value)
}

const validate = () => {
  return formRef.value.validate()
}

watch(
  () => formData,
  handleChange,
  {
    immediate: true,
    deep: true
  }
)

defineExpose([validate]);
</script>

<template>
  <bk-form
    ref="formRef"
    :model="formData"
  >
    <bk-form-item
      label="网络边界组"
    >
      {{ region }}
    </bk-form-item>
    <bk-form-item
      label="公有IPv4地址池"
    >
      <bk-radio
        model-value="Amazon的IPv4地址池"
        label="Amazon的IPv4地址池"
      />
    </bk-form-item>
    <bk-form-item
      label="数量"
      required
    >
      1
    </bk-form-item>
  </bk-form>
</template>
