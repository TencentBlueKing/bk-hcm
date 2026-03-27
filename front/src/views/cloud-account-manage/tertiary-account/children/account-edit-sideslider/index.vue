<script setup lang="ts">
import { ref, inject, computed, type Ref, watch } from 'vue';
import { Message } from 'bkui-vue';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import { useCloudAccountStore, type ISubAccountItem, type ISubAccountUpdateParams } from '@/store/cloud-account';
import { VendorEnum } from '@/common/constant';
import UserSelector from '@/components/user-selector/index.vue';
import BusinessSelector from '@/components/business-selector/business.vue';

const props = defineProps<{
  modelValue: boolean;
  accountData: ISubAccountItem | null;
}>();

const emit = defineEmits<{
  (e: 'update:modelValue', val: boolean): void;
  (e: 'success'): void;
}>();

const currentVendor = inject<Ref<VendorEnum>>('currentVendor', ref(VendorEnum.TCLOUD));
const cloudAccountStore = useCloudAccountStore();
const { getBizsId } = useWhereAmI();

const formData = ref({
  name: '',
  managers: [] as string[],
  bk_biz_id: undefined as number | undefined,
  permission_template: [] as string[],
  phone_num: '',
  email: '',
  memo: '',
});

const isSubmitting = ref(false);

// 腾讯云账号不允许修改名称
const isTcloud = computed(() => currentVendor.value === VendorEnum.TCLOUD);

watch(
  () => props.modelValue,
  (val) => {
    if (val && props.accountData) {
      formData.value = {
        name: props.accountData.name || '',
        managers: [...(props.accountData.managers || [])],
        bk_biz_id: props.accountData.bk_biz_ids?.[0] ?? undefined,
        permission_template: [],
        phone_num: props.accountData.phone_num || '',
        email: props.accountData.email || '',
        memo: props.accountData.memo || '',
      };
    }
  },
);

const handleClose = () => {
  emit('update:modelValue', false);
};

const handleSubmit = async () => {
  if (!props.accountData?.id) return;

  if (!isTcloud.value && !formData.value.name) {
    Message({ theme: 'warning', message: '请输入三级账号名称' });
    return;
  }
  if (!formData.value.managers?.length) {
    Message({ theme: 'warning', message: '请选择负责人' });
    return;
  }

  const subAccounts: ISubAccountUpdateParams[] = [
    {
      id: props.accountData.id,
      ...(!isTcloud.value ? { name: formData.value.name } : {}),
      email: formData.value.email || undefined,
      phone_num: formData.value.phone_num || undefined,
      country_code: '86',
      managers: formData.value.managers,
      bk_biz_id: formData.value.bk_biz_id,
      memo: formData.value.memo || undefined,
    },
  ];

  isSubmitting.value = true;
  try {
    await cloudAccountStore.updateSubAccount(getBizsId(), currentVendor.value, subAccounts);
    Message({ theme: 'success', message: '更新申请提交成功' });
    handleClose();
    emit('success');
  } catch (error) {
    console.error('更新三级账号失败:', error);
  } finally {
    isSubmitting.value = false;
  }
};

// 所属二级账号展示
const parentAccountDisplay = () => {
  if (!props.accountData) return '--';
  return `${props.accountData.account_id || '--'}`;
};
</script>

<template>
  <bk-sideslider
    :is-show="modelValue"
    :width="640"
    title="编辑三级账号"
    :before-close="handleClose"
    @closed="handleClose"
  >
    <template #default>
      <div v-if="accountData" class="edit-form">
        <bk-form form-type="vertical" :model="formData">
          <!-- 所属二级账号（只读） -->
          <bk-form-item label="所属二级账号" required>
            <bk-input :model-value="parentAccountDisplay()" disabled />
          </bk-form-item>

          <!-- 三级账号ID（只读） -->
          <bk-form-item label="三级账号ID" required>
            <bk-input :model-value="accountData.cloud_id" disabled />
          </bk-form-item>

          <!-- 三级账号名称 -->
          <bk-form-item label="三级账号名称" required>
            <bk-input v-model="formData.name" placeholder="请输入三级账号名称" :disabled="isTcloud" />
          </bk-form-item>

          <!-- 负责人 -->
          <bk-form-item label="负责人" required>
            <UserSelector v-model="formData.managers" placeholder="请输入用户名" />
          </bk-form-item>

          <!-- 所属业务 -->
          <bk-form-item label="所属业务" required>
            <BusinessSelector v-model="formData.bk_biz_id" placeholder="请选择业务" clearable />
          </bk-form-item>

          <!-- 权限模版 -->
          <bk-form-item label="权限模版" required>
            <bk-select v-model="formData.permission_template" placeholder="请选择" multiple>
              <!-- 权限模版选项后续对接 -->
            </bk-select>
          </bk-form-item>

          <!-- 手机号 -->
          <bk-form-item label="手机号">
            <bk-input v-model="formData.phone_num" placeholder="请输入手机号" />
          </bk-form-item>

          <!-- 账号邮箱 -->
          <bk-form-item label="账号邮箱">
            <bk-input v-model="formData.email" placeholder="请输入邮箱" />
          </bk-form-item>

          <!-- 备注 -->
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
