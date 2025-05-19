<script setup lang="ts">
import { computed, provide, reactive, ref } from 'vue';
import { isEqual } from 'lodash';
import { Message } from 'bkui-vue';
import { isEmpty } from '@/common/util';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import {
  useSecurityGroupStore,
  type ISecurityGroupItem,
  type SecurityGroupMgmtAttrSingleType,
} from '@/store/security-group';
import UsageBizFormItem from './usage-biz-form-item.vue';
import { useAccountBusiness } from '@/views/resource/resource-manage/hooks/use-account-business';

const props = defineProps<{
  detail: ISecurityGroupItem;
  field: SecurityGroupMgmtAttrSingleType;
}>();

const emit = defineEmits<{
  success: [];
}>();

const whereAmI = useWhereAmI();
const securityGroupStore = useSecurityGroupStore();

const { accountBizList } = useAccountBusiness(props.detail.account_id);

const model = defineModel<boolean>();

const formData = reactive<Record<SecurityGroupMgmtAttrSingleType, any>>({
  mgmt_biz_id: props.detail.mgmt_biz_id === -1 ? undefined : props.detail.mgmt_biz_id,
  manager: props.detail.manager,
  bak_manager: props.detail.bak_manager,
  usage_biz_ids: props.detail.usage_biz_ids,
});

const changeConfirmed = ref(false);

const closeDialog = () => {
  model.value = false;
};

const isEditUsageBiz = computed(() => props.field === 'usage_biz_ids');
const isBusinessEditUsageBiz = computed(() => whereAmI.isBusinessPage && isEditUsageBiz.value);

const isValueChanged = computed(() => !isEqual(formData[props.field], props.detail[props.field]));

const confirmButtonDisabled = computed(() => {
  const isValueEmpty = isEmpty(formData[props.field]);

  if (isBusinessEditUsageBiz.value) {
    return isValueEmpty || !isValueChanged.value || !changeConfirmed.value;
  }

  return isValueEmpty || !isValueChanged.value;
});

const handleDialogConfirm = async () => {
  if (confirmButtonDisabled.value) {
    return;
  }

  await securityGroupStore.updateMgmtAttr(props.detail.id, { [props.field]: formData[props.field] });

  Message({ theme: 'success', message: '编辑成功' });
  closeDialog();
  emit('success');
};

provide('isBusinessEditUsageBiz', isBusinessEditUsageBiz);
</script>

<template>
  <bk-dialog
    :title="isBusinessEditUsageBiz ? '使用范围' : '编辑资产属性'"
    :width="isEditUsageBiz ? 560 : 480"
    :quick-close="false"
    :is-show="model"
    @closed="closeDialog"
  >
    <bk-form form-type="vertical" :model="formData">
      <bk-form-item label="管理业务" property="mgmt_biz_id" v-if="field === 'mgmt_biz_id'">
        <hcm-form-business :data="accountBizList" v-model="formData.mgmt_biz_id" />
      </bk-form-item>
      <bk-form-item label="主负责人" property="manager" v-if="field === 'manager'">
        <hcm-form-user :multiple="false" v-model="formData.manager" />
      </bk-form-item>
      <bk-form-item label="备份负责人" property="bak_manager" v-if="field === 'bak_manager'">
        <hcm-form-user :multiple="false" v-model="formData.bak_manager" />
      </bk-form-item>
      <usage-biz-form-item
        :detail="detail"
        :account-biz-list="accountBizList"
        v-model="formData.usage_biz_ids"
        v-if="isEditUsageBiz"
      />
    </bk-form>
    <template #footer>
      <div class="dialog-custom-footer">
        <bk-checkbox class="confirm-checkbox" v-model="changeConfirmed" v-if="isBusinessEditUsageBiz">
          已知晓变更影响，仍需变更
        </bk-checkbox>
        <bk-button
          theme="primary"
          :disabled="confirmButtonDisabled"
          :loading="securityGroupStore.isUpdateMgmtAttrLoading"
          @click="handleDialogConfirm"
        >
          {{ isBusinessEditUsageBiz ? '提交配置' : '确定' }}
        </bk-button>
        <bk-button @click="closeDialog">取消</bk-button>
      </div>
    </template>
  </bk-dialog>
</template>

<style lang="scss" scoped>
.dialog-custom-footer {
  display: flex;
  justify-content: flex-end;
  gap: 8px;

  .confirm-checkbox {
    margin-right: auto;
  }
}
</style>
