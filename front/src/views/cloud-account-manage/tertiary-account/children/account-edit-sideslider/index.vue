<script setup lang="ts">
import { ref, inject, computed, type Ref, watch } from 'vue';
import { Message } from 'bkui-vue';
import { parsePhoneNumberFromString } from 'libphonenumber-js';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import {
  useTertiaryAccountStore,
  type ISubAccountItem,
  type ISubAccountUpdateParams,
} from '@/store/cloud-account-manage/tertiary-account';
import { VendorEnum } from '@/common/constant';
import UserSelector from '@/components/user-selector/index.vue';
import BusinessSelector from '@/components/business-selector/business.vue';
import { usePermissionTemplateStore } from '@/store/cloud-account-manage/permission-template';

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

const formData = ref({
  name: '',
  managers: [] as string[],
  bk_biz_id: undefined as number | undefined,
  permission_template_ids: [] as string[],
  phone_num: '',
  country_code: '',
  email: '',
  memo: '',
});

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

// 组合 country_code + phone_num 用于显示
const phoneDisplay = computed({
  get: () => {
    const code = formData.value.country_code;
    const num = formData.value.phone_num;
    if (!num) return '';
    return code ? `+${code}${num}` : num;
  },
  set: (val: string) => {
    const { countryCode, phoneNum } = parsePhoneInput(val);
    formData.value.country_code = countryCode;
    formData.value.phone_num = phoneNum;
  },
});

const isSubmitting = ref(false);

const isTcloud = computed(() => currentVendor.value === VendorEnum.TCLOUD);

const formRules = {
  name: [{ required: true, message: '请输入三级账号名称', trigger: 'blur' }],
  managers: [{ required: true, message: '请选择负责人', trigger: 'change', type: 'array' }],
  bk_biz_id: [{ required: true, message: '请选择所属业务', trigger: 'change' }],
  permission_template_ids: [{ required: true, message: '请选择权限模板', trigger: 'change', type: 'array' }],
};

watch(
  () => model.value,
  (val) => {
    if (val && props.accountData) {
      formData.value = {
        name: props.accountData.name || '',
        managers: [...(props.accountData.managers || [])],
        bk_biz_id: props.accountData.bk_biz_ids?.[0] ?? undefined,
        permission_template_ids: props.accountData.permission_template_ids || [],
        phone_num: props.accountData.phone_num || '',
        country_code: props.accountData.country_code || '',
        email: props.accountData.email || '',
        memo: props.accountData.memo || '',
      };
    }
  },
);

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

  const subAccounts: ISubAccountUpdateParams[] = [
    {
      id: props.accountData.id,
      email: formData.value.email || undefined,
      phone_num: formData.value.phone_num || undefined,
      permission_template_ids: formData.value.permission_template_ids,
      bk_biz_id: formData.value.bk_biz_id,
      country_code: formData.value.country_code,
      managers: formData.value.managers,
      memo: formData.value.memo || undefined,
    },
  ];

  isSubmitting.value = true;
  try {
    await tertiaryAccountStore.updateSubAccount(getBizsId(), currentVendor.value, subAccounts);
    Message({ theme: 'success', message: '更新申请提交成功' });
    handleClose();
    emit('success');
  } catch (error) {
    console.error('更新三级账号失败:', error);
  } finally {
    isSubmitting.value = false;
  }
};

const parentAccountDisplay = () => {
  if (!props.accountData) return '--';
  return `${props.accountData.account_id || '--'}`;
};

const permissionTemplateStore = usePermissionTemplateStore();
const listGenerator = computed(() =>
  permissionTemplateStore.createPermissionTemplateListGenerator(getBizsId(), currentVendor.value),
);
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
          <bk-form-item label="所属二级账号" required>
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
            description='填写"+地域代码号码"，如+8613212345678。+86是地区代码 13212345678是电话号码'
          >
            <bk-input v-model="phoneDisplay" placeholder="请输入手机号" />
          </bk-form-item>

          <bk-form-item label="账号邮箱">
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
        <bk-button theme="primary" :loading="isSubmitting" @click="handleSubmit">提交</bk-button>
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
