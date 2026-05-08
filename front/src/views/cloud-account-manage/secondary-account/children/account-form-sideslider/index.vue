<script setup lang="ts">
import { ref, computed, watch, nextTick } from 'vue';
import { Message } from 'bkui-vue';
import { BkRadioButton, BkRadioGroup } from 'bkui-vue/lib/radio';
import UserSelector from '@/components/user-selector/index.vue';
import BusinessSelector from '@/components/business-selector/business.vue';
import { useSecondaryAccountStore, type ISecondaryAccountItem } from '@/store/cloud-account-manage/secondary-account';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import { useUserStore } from '@/store/user';
import { SITE_TYPE } from '@/constants/account';
import { MENU_SERVICE_TICKET_DETAILS } from '@/constants/menu-symbol';
import routerAction from '@/router/utils/action';

// 双向绑定控制显示状态
const model = defineModel<boolean>();

// Props 定义
const props = defineProps<{
  isEdit?: boolean;
  accountData?: ISecondaryAccountItem | null;
}>();

// Emits 定义
const emit = defineEmits<{
  success: [updatedData?: ISecondaryAccountItem];
}>();

// Store 和 Hooks
const secondaryAccountStore = useSecondaryAccountStore();
const userStore = useUserStore();
const { getBizsId } = useWhereAmI();

// 表单引用
const formRef = ref();

// 提交加载状态
const submitLoading = ref(false);

// 表单数据
const formData = ref({
  site: 'china',
  name: '',
  cloud_main_account_id: '',
  managers: [] as string[],
  security_managers: [] as string[],
  bk_biz_id: getBizsId() as number | undefined,
  usage_biz_ids: getBizsId() ? [getBizsId()] : ([] as number[]),
  memo: '',
});

// 侧栏标题
const sidesliderTitle = computed(() => (props.isEdit ? '编辑二级账号' : '录入二级账号'));

// 使用业务校验器：必须包含管理业务
const usageBizValidator = (value: number[]) => {
  const manageBizId = formData.value.bk_biz_id;
  if (manageBizId !== undefined && manageBizId !== null) {
    if (!value || !value.includes(manageBizId)) {
      return false;
    }
  }
  return true;
};

// 表单校验规则
const formRules = {
  site: [{ required: true, message: '请选择站点类型', trigger: 'change' }],
  name: [{ required: true, message: '请输入二级账号名称', trigger: 'blur' }],
  cloud_main_account_id: [{ required: true, message: '请输入二级账号ID', trigger: 'blur' }],
  managers: [{ required: true, message: '请选择负责人', trigger: 'change', type: 'array' }],
  security_managers: [{ required: true, message: '请选择安全负责人', trigger: 'change', type: 'array' }],
  usage_biz_ids: [{ validator: usageBizValidator, trigger: 'change', message: '使用业务必须包含当前选择的管理业务' }],
};

// 重置表单
const resetForm = () => {
  formData.value = {
    site: 'china',
    name: '',
    cloud_main_account_id: '',
    managers: [userStore.username],
    security_managers: [userStore.username],
    bk_biz_id: getBizsId(),
    usage_biz_ids: getBizsId() ? [getBizsId()] : [],
    memo: '',
  };
  nextTick(() => {
    formRef.value?.clearValidate();
  });
};

// 填充编辑数据
const fillEditData = () => {
  if (props.isEdit && props.accountData) {
    const data = props.accountData;
    formData.value = {
      site: data.site || 'china',
      name: data.name || '',
      cloud_main_account_id: data.extension?.cloud_main_account_id || '',
      managers: data.managers || [],
      security_managers: data.security_managers || [],
      bk_biz_id: data.bk_biz_id,
      usage_biz_ids: data.usage_biz_ids || [],
      memo: data.memo || '',
    };
  }
};

// 监听显示状态变化
watch(
  () => model.value,
  (newVal) => {
    if (newVal) {
      if (props.isEdit) {
        fillEditData();
      } else {
        resetForm();
      }
    }
  },
);

// 监听 accountData 变化（编辑模式）
watch(
  () => props.accountData,
  () => {
    if (model.value && props.isEdit) {
      fillEditData();
    }
  },
);

// 监听管理业务变化，自动添加到使用业务中
watch(
  () => formData.value.bk_biz_id,
  (newBizId) => {
    if (newBizId !== undefined && newBizId !== null) {
      // 如果使用业务中不包含当前管理业务，则自动添加
      if (!formData.value.usage_biz_ids.includes(newBizId)) {
        formData.value.usage_biz_ids = [...formData.value.usage_biz_ids, newBizId];
      }
    }
  },
);

// 提交表单
const handleSubmit = async () => {
  try {
    await formRef.value?.validate();
  } catch {
    return;
  }

  submitLoading.value = true;

  try {
    // 真实接口调用
    if (props.isEdit) {
      // 编辑接口
      await secondaryAccountStore.updateSecondaryAccount(getBizsId(), props.accountData!.id, {
        name: formData.value.name,
        managers: formData.value.managers,
        security_managers: formData.value.security_managers,
        bk_biz_id: formData.value.bk_biz_id,
        usage_biz_ids: formData.value.usage_biz_ids,
        memo: formData.value.memo,
      });
      Message({ theme: 'success', message: '编辑成功' });
      // 编辑模式时，返回更新后的数据
      const updatedData: ISecondaryAccountItem = {
        ...props.accountData!,
        name: formData.value.name,
        managers: formData.value.managers,
        security_managers: formData.value.security_managers,
        bk_biz_id: formData.value.bk_biz_id || 0,
        usage_biz_ids: formData.value.usage_biz_ids,
        memo: formData.value.memo,
        updated_at: new Date().toISOString(),
      };
      model.value = false;
      emit('success', updatedData);
    } else {
      // 录入接口
      const result = await secondaryAccountStore.createSecondaryAccount(getBizsId(), {
        vendor: 'tcloud',
        name: formData.value.name,
        managers: formData.value.managers,
        security_managers: formData.value.security_managers,
        type: 'resource',
        site: formData.value.site,
        bk_biz_id: formData.value.bk_biz_id,
        usage_biz_ids: formData.value.usage_biz_ids,
        memo: formData.value.memo,
        extension: {
          cloud_main_account_id: formData.value.cloud_main_account_id,
        },
      });
      Message({ theme: 'success', message: '录入申请已提交' });
      model.value = false;
      emit('success');
      if (result?.id) {
        routerAction.redirect({
          name: MENU_SERVICE_TICKET_DETAILS,
          query: {
            id: result.id,
            type: 'account',
          },
        });
      }
    }
  } catch (error) {
    console.error('提交失败:', error);
  } finally {
    submitLoading.value = false;
  }
};

// 取消
const handleCancel = () => {
  model.value = false;
};
</script>

<template>
  <bk-sideslider v-model:is-show="model" :title="sidesliderTitle" :width="640" quick-close render-directive="if">
    <template #default>
      <div class="account-form-container">
        <bk-form ref="formRef" :model="formData" :rules="formRules" form-type="vertical">
          <!-- 站点类型 -->
          <bk-form-item label="站点类型" property="site" required>
            <BkRadioGroup v-model="formData.site" :disabled="isEdit" type="card">
              <BkRadioButton v-for="item in SITE_TYPE" :key="item.value" :label="item.value" class="site-radio-btn">
                {{ item.label }}
              </BkRadioButton>
            </BkRadioGroup>
          </bk-form-item>

          <!-- 二级账号名称 -->
          <bk-form-item label="二级账号名称" property="name" required>
            <bk-input v-model="formData.name" placeholder="请输入二级名称" :maxlength="64" show-word-limit />
          </bk-form-item>

          <!-- 二级账号ID -->
          <bk-form-item label="二级账号ID" property="cloud_main_account_id" required>
            <bk-input v-model="formData.cloud_main_account_id" placeholder="请输入二级账号ID" :disabled="isEdit" />
          </bk-form-item>

          <!-- 负责人 -->
          <bk-form-item label="负责人" property="managers" required>
            <UserSelector v-model="formData.managers" placeholder="请输入用户名" />
          </bk-form-item>

          <!-- 安全负责人 -->
          <bk-form-item label="安全负责人" property="security_managers" required>
            <UserSelector v-model="formData.security_managers" placeholder="请输入用户名" />
          </bk-form-item>

          <!-- 管理业务（只读，录入时默认当前业务，编辑时使用已有数据） -->
          <bk-form-item label="管理业务" property="bk_biz_id">
            <BusinessSelector v-model="formData.bk_biz_id" placeholder="请选择业务" disabled />
          </bk-form-item>

          <!-- 使用业务 -->
          <bk-form-item label="使用业务" property="usage_biz_ids">
            <BusinessSelector
              v-model="formData.usage_biz_ids"
              placeholder="请选择业务"
              multiple
              clearable
              collapse-tags
            />
          </bk-form-item>

          <!-- 备注 -->
          <bk-form-item label="备注" property="memo">
            <bk-input
              v-model="formData.memo"
              type="textarea"
              placeholder="请输入"
              :rows="4"
              :maxlength="100"
              show-word-limit
            />
          </bk-form-item>
        </bk-form>
      </div>
    </template>

    <template #footer>
      <div class="sideslider-footer">
        <bk-button theme="primary" :loading="submitLoading" @click="handleSubmit">提交</bk-button>
        <bk-button :disabled="submitLoading" @click="handleCancel">取消</bk-button>
      </div>
    </template>
  </bk-sideslider>
</template>

<style lang="scss" scoped>
.account-form-container {
  padding: 24px 40px 0;

  :deep(.bk-form-item) {
    margin-bottom: 24px;
  }

  :deep(.bk-radio-button-input:checked + .bk-radio-button-label) {
    background-color: #e1ecff;
    border-color: #3a84ff;
    color: #3a84ff;
  }
}

.sideslider-footer {
  display: flex;
  align-items: center;
  padding: 0 18px;

  .bk-button {
    min-width: 88px;
    margin-right: 8px;
  }
}
</style>
