<script setup lang="ts">
import {
  ref,
  watch,
  defineEmits,
  defineExpose,
} from 'vue';

const emit = defineEmits(['change']);

const formData = ref({
  public_ipv4_pool: '',
  network_border_group: '',
  eip_count: 1,
})
const formRef = ref(null);


const handleChange = () => {
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
    form-type="vertical"
    ref="formRef"
    :model="formData"
  >
    <bk-form-item
      label="数量"
      required
    >
      1
    </bk-form-item>
  </bk-form>
</template>
