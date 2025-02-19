<script setup lang="ts">
import { reactive, ref } from 'vue';
import { Message } from 'bkui-vue';
import { useSecurityGroupStore, type ISecurityGroupItem, SecurityGroupManageType } from '@/store/security-group';

import HcmFormUser from '@/components/form/user.vue';
import HcmFormBusiness from '@/components/form/business.vue';

const props = defineProps<{ detail: ISecurityGroupItem }>();

const emit = defineEmits<{
  success: [];
}>();

const securityGroupStore = useSecurityGroupStore();

const model = defineModel<boolean>();

const formData = reactive({
  mgmt_type: SecurityGroupManageType.PLATFORM,
  mgmt_biz_id: undefined,
  manager: undefined,
  bak_manager: undefined,
});

const formRef = ref(null);

const closeDialog = () => {
  model.value = false;
};

const handleDialogConfirm = async () => {
  await formRef.value?.validate();

  await securityGroupStore.updateMgmtAttr(props.detail.id, formData);

  Message({ theme: 'success', message: '确认成功' });
  closeDialog();
  emit('success');
};
</script>

<template>
  <bk-dialog
    :title="'管理类型'"
    :width="480"
    :quick-close="false"
    :is-show="model"
    :loading="securityGroupStore.isUpdateMgmtAttrLoading"
    @closed="closeDialog"
    @confirm="handleDialogConfirm"
  >
    <bk-alert
      class="update-alert"
      theme="error"
      title="我是提示我是提示我是提示我是提示我是提示我是提示我是提示我是提示我是提示我是提示"
    />
    <bk-form form-type="vertical" :model="formData" ref="formRef">
      <bk-form-item property="mgmt_type">
        <bk-radio-group v-model="formData.mgmt_type" type="card">
          <bk-radio-button :label="SecurityGroupManageType.PLATFORM">平台管理</bk-radio-button>
          <bk-radio-button :label="SecurityGroupManageType.BIZ">业务管理</bk-radio-button>
        </bk-radio-group>
      </bk-form-item>
      <template v-if="formData.mgmt_type === SecurityGroupManageType.BIZ">
        <bk-form-item label="管理业务" property="mgmt_biz_id">
          <hcm-form-business v-model="formData.mgmt_biz_id" />
        </bk-form-item>
        <bk-form-item label="主负责人" property="manager">
          <hcm-form-user :multiple="false" v-model="formData.manager" />
        </bk-form-item>
        <bk-form-item label="备份负责人" property="bak_manager">
          <hcm-form-user :multiple="false" v-model="formData.bak_manager" />
        </bk-form-item>
      </template>
    </bk-form>
  </bk-dialog>
</template>

<style lang="scss" scoped>
.update-alert {
  margin-bottom: 24px;
}
</style>
