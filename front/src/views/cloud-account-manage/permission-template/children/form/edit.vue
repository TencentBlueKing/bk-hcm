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
  get isChanged() {
    return formRef.value?.isChanged;
  },
  get changedFormData() {
    return formRef.value?.changedFormData;
  },
});
</script>

<template>
  <div class="permission-template-edit">
    <bk-alert theme="info" closable class="alert-info">
      <template #title>
        <div>影响说明：修改权限模板，将直接影响所关联三级账号的权限。</div>
        <div>生效流程：修改需求提交后，系统将生成审批单。此变更需审批通过后，方可正式生效。</div>
      </template>
    </bk-alert>
    <Form ref="formRef" :data="data" :is-edit="true" />
  </div>
</template>

<style lang="scss" scoped>
.permission-template-edit {
  padding: 24px 24px 0;
}

.alert-info {
  margin-bottom: 16px;
}
</style>
