<script setup lang="ts">
import { useTemplateRef } from 'vue';
import type { FieldTcloud } from './field-tcloud';
import type { IPermissionTemplateItem } from '@/store/cloud-account-manage/permission-template';
import Form from './form.vue';

defineProps<{
  data: IPermissionTemplateItem & FieldTcloud;
}>();

const formRef = useTemplateRef<typeof Form>('formRef');

defineExpose({
  validate: () => formRef.value.validate(),
  getFormData: () => formRef.value.getFormData(),
});
</script>

<template>
  <div class="permission-template-create">
    <Form ref="formRef" :data="data" />
  </div>
</template>

<style lang="scss" scoped>
.permission-template-create {
  padding: 24px 24px 0;
}
</style>
