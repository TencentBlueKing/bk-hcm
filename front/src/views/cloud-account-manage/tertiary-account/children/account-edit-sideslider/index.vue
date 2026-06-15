<script setup lang="ts">
import { ref, inject, computed, type Ref } from 'vue';
import { Message } from 'bkui-vue';
import { parsePhoneNumberFromString } from 'libphonenumber-js';
import isEmail from 'validator/lib/isEmail';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import { useFormChange } from '@/hooks/use-form-change';
import {
  useTertiaryAccountStore,
  type ISubAccountItem,
  type ISubAccountUpdateParams,
} from '@/store/cloud-account-manage/tertiary-account';
import { VendorEnum } from '@/common/constant';
import UserSelector from '@/components/user-selector/index.vue';
import BusinessSelector from '@/components/business-selector/business.vue';
import { usePermissionTemplateStore } from '@/store/cloud-account-manage/permission-template';
import routerAction from '@/router/utils/action';
import { MENU_SERVICE_TICKET_DETAILS, MENU_SERVICE_TICKET_MANAGEMENT } from '@/constants/menu-symbol';
import { PERMISSION_TEMPLATE_TYPES } from '@/views/cloud-account-manage/permission-template/constants';

const model = defineModel<boolean>();

const props = defineProps<{
  accountData: ISubAccountItem | null;
}>();

const emit = defineEmits<{
  (e: 'success'): void;
}>();

const formRef = ref();
const currentVendor = inject<Ref<VendorEnum>>('currentVendor', ref(VendorEnum.TCLOUD));
const tertiaryAccountStore = useTertiaryAccountStore();
const { getBizsId } = useWhereAmI();

const initialFormData = computed(() => {
  if (!model.value || !props.accountData) return null;

  const { phone_num, country_code } = props.accountData;
  return {
    name: props.accountData.name || '',
    managers: [...(props.accountData.managers || [])],
    bk_biz_id: props.accountData.bk_biz_ids?.[0] ?? undefined,
    permission_template_ids: props.accountData.permission_template_ids || [],
    phone_num: country_code ? `+${country_code}${phone_num}` : phone_num,
    country_code: props.accountData.country_code || '',
    email: props.accountData.email || '',
    memo: props.accountData.memo || '',
  };
});

const { formData, isChanged, changedFormData } = useFormChange<{
  name: string;
  managers: string[];
  bk_biz_id: number | undefined;
  permission_template_ids: string[];
  phone_num: string;
  country_code: string;
  email: string;
  memo: string;
}>(initialFormData);

// 使用 libphonenumber-js 解析手机号，自动识别国家区号
const parsePhoneInput = (input: string): { countryCode: string; phoneNum: string } => {
  const trimmed = input.trim();
  if (!trimmed) return { countryCode: '', phoneNum: '' };
  const phoneNumber = parsePhoneNumberFromString(trimmed);
  if (phoneNumber) {
    return {
      countryCode: String(phoneNumber.countryCallingCode),
      phoneNum: phoneNumber.nationalNumber,
    };
  }
  return { countryCode: '', phoneNum: trimmed };
};

const isSubmitting = ref(false);

const isTcloud = computed(() => currentVendor.value === VendorEnum.TCLOUD);

const formRules = {
  name: [{ required: true, message: '请输入三级账号名称', trigger: 'blur' }],
  managers: [{ required: true, message: '请选择负责人', trigger: 'change', type: 'array' }],
  bk_biz_id: [{ required: true, message: '请选择所属业务', trigger: 'change' }],
  permission_template_ids: [{ required: true, message: '请选择权限模板', trigger: 'change', type: 'array' }],
  phone_num: [
    {
      validator: (value: string) => {
        if (!changedFormData.value?.phone_num) return true;
        if (!value || !value.trim()) return true;
        const phoneNumber = parsePhoneNumberFromString(value.trim());
        return phoneNumber?.isValid() || '手机号格式不正确';
      },
      trigger: 'change',
    },
  ],
  email: [
    {
      validator: (value: string) => {
        if (!changedFormData.value?.email) return true;
        if (!value || !value.trim()) return true;
        return isEmail(value.trim()) || '邮箱格式不正确';
      },
      trigger: 'change',
    },
  ],
};

const handleClose = () => {
  model.value = false;
};

const handleSubmit = async () => {
  if (!props.accountData?.id) return;

  try {
    await formRef.value?.validate();
  } catch {
    return;
  }

  const { countryCode, phoneNum } = parsePhoneInput(formData.value.phone_num);
  const changed = changedFormData.value;
  const options: Partial<ISubAccountUpdateParams> = {};
  if (Object.prototype.hasOwnProperty.call(changed, 'phone_num')) {
    options.phone_num = phoneNum || undefined;
    options.country_code = countryCode || undefined;
  }
  const subAccounts: ISubAccountUpdateParams[] = [
    {
      id: props.accountData!.id,
      ...changed,
      ...options,
    } as ISubAccountUpdateParams,
  ];

  isSubmitting.value = true;
  try {
    const result = await tertiaryAccountStore.updateSubAccount(getBizsId(), currentVendor.value, subAccounts);
    Message({ theme: 'success', message: '更新申请提交成功' });
    handleClose();
    emit('success');
    // 跳转到审批单页面
    if (result?.ids?.length) {
      if (result.ids.length === 1) {
        routerAction.redirect({
          name: MENU_SERVICE_TICKET_DETAILS,
          query: { id: result.ids[0], type: 'account' },
        });
      } else {
        routerAction.redirect({ name: MENU_SERVICE_TICKET_MANAGEMENT, query: { type: 'account' } });
      }
    }
  } catch (error) {
    console.error('更新三级账号失败:', error);
  } finally {
    isSubmitting.value = false;
  }
};

const parentAccountDisplay = () => {
  if (!props.accountData) return '--';
  return `${props.accountData?.extension?.cloud_main_account_id || '--'}`;
};

const permissionTemplateStore = usePermissionTemplateStore();
const listGenerator = computed(() => {
  if (!props.accountData?.extension?.cloud_main_account_id) return null;
  return permissionTemplateStore.createPermissionTemplateListGenerator(getBizsId(), currentVendor.value, {
    extension: {
      cloud_main_account_ids: [props.accountData.extension.cloud_main_account_id],
    },
    permission_template_type: PERMISSION_TEMPLATE_TYPES.SYNC_WITH_LIBRARY,
  });
});
</script>

<template>
  <bk-sideslider
    :is-show="model"
    :width="640"
    title="编辑三级账号"
    :before-close="handleClose"
    render-directive="if"
    @closed="handleClose"
  >
    <template #default>
      <div v-if="accountData" class="edit-form">
        <bk-form ref="formRef" form-type="vertical" :model="formData" :rules="formRules">
          <bk-form-item label="所属二级账号ID" required>
            <bk-input :model-value="parentAccountDisplay()" disabled />
          </bk-form-item>

          <bk-form-item label="三级账号ID" required>
            <bk-input :model-value="accountData.cloud_id" disabled />
          </bk-form-item>

          <bk-form-item label="三级账号名称" property="name" required>
            <bk-input v-model="formData.name" placeholder="请输入三级账号名称" :disabled="isTcloud" />
          </bk-form-item>

          <bk-form-item label="负责人" property="managers" required>
            <UserSelector v-model="formData.managers" placeholder="请输入用户名" />
          </bk-form-item>

          <bk-form-item label="所属业务" property="bk_biz_id" required>
            <BusinessSelector v-model="formData.bk_biz_id" placeholder="请选择业务" clearable />
          </bk-form-item>

          <bk-form-item label="权限模板" property="permission_template_ids" required>
            <hcm-form-list
              v-model="formData.permission_template_ids"
              :list-generator="listGenerator"
              placeholder="请选择"
              display-key="name"
              id-key="id"
              collapse-tags
              multiple
            />
          </bk-form-item>

          <bk-form-item
            label="手机号"
            property="phone_num"
            description='填写"+地域代码号码"，如+8613212345678。+86是地区代码 13212345678是电话号码'
          >
            <bk-input v-model="formData.phone_num" placeholder="请输入手机号" />
          </bk-form-item>

          <bk-form-item label="账号邮箱" property="email">
            <bk-input v-model="formData.email" placeholder="请输入邮箱" />
          </bk-form-item>

          <bk-form-item label="备注">
            <bk-input v-model="formData.memo" type="textarea" :maxlength="100" :rows="3" placeholder="请输入" />
          </bk-form-item>
        </bk-form>
      </div>
    </template>
    <template #footer>
      <div class="sideslider-footer">
        <bk-button theme="primary" :loading="isSubmitting" :disabled="!isChanged" @click="handleSubmit">提交</bk-button>
        <bk-button @click="handleClose">取消</bk-button>
      </div>
    </template>
  </bk-sideslider>
</template>

<style lang="scss" scoped>
.edit-form {
  padding: 28px 40px;
}

.sideslider-footer {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 0 24px;

  .bk-button {
    min-width: 88px;
  }
}
</style>
