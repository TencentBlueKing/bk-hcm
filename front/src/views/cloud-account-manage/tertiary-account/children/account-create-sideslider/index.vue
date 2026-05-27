<script setup lang="ts">
import { ref, inject, computed, watch, nextTick, h, type Ref } from 'vue';
import { Message, Select } from 'bkui-vue';
import { Ediatable, InputColumn, SelectColumn } from '@blueking/ediatable';
import { parsePhoneNumberFromString } from 'libphonenumber-js';
import isEmail from 'validator/lib/isEmail';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import { useSecondaryAccountStore } from '@/store/cloud-account-manage/secondary-account';
import { useTertiaryAccountStore, type ISubAccountCreateParams } from '@/store/cloud-account-manage/tertiary-account';
import type { ISecondaryAccountItem } from '@/store/cloud-account-manage/secondary-account';
import { VendorEnum } from '@/common/constant';
import { QueryRuleOPEnum, type QueryFilterType } from '@/typings';
import OperationColumn from '@/components/ediatable/operation-column.vue';
import UserSelector from '@/components/user-selector/index.vue';
import BatchUpdatePopConfirm from '@/components/batch-update-popconfirm';

import { usePermissionTemplateStore } from '@/store/cloud-account-manage/permission-template';
import routerAction from '@/router/utils/action';
import { MENU_SERVICE_TICKET_DETAILS, MENU_SERVICE_TICKET_MANAGEMENT } from '@/constants/menu-symbol';
import { PERMISSION_TEMPLATE_TYPES } from '@/views/cloud-account-manage/permission-template/constants';

interface IRowData {
  account_id: string;
  account_name: string;
  name: string;
  permission_template_ids: string[];
  phone_num: string;
  country_code: string;
  email: string;
  managers: string[];
  receive_email: string;
}

const model = defineModel<boolean>();

const props = defineProps<{
  defaultAccountId?: string;
}>();

const emit = defineEmits<{
  (e: 'success'): void;
}>();

const currentVendor = inject<Ref<VendorEnum>>('currentVendor', ref(VendorEnum.TCLOUD));
const secondaryAccountStore = useSecondaryAccountStore();
const tertiaryAccountStore = useTertiaryAccountStore();
const permissionTemplateStore = usePermissionTemplateStore();

const { getBizsId } = useWhereAmI();

const secondaryAccountList = ref<ISecondaryAccountItem[]>([]);
const secondaryAccountLoading = ref(false);

const loadSecondaryAccountList = async () => {
  secondaryAccountLoading.value = true;
  try {
    const filter: QueryFilterType = {
      op: 'and',
      rules: [
        { field: 'vendor', op: QueryRuleOPEnum.EQ, value: currentVendor.value },
        { field: 'type', op: QueryRuleOPEnum.EQ, value: 'resource' },
      ],
    };
    const list = await secondaryAccountStore.getSecondaryAccountFullList(getBizsId(), filter);
    secondaryAccountList.value = list;
  } catch (error) {
    console.error('加载二级账号列表失败:', error);
  } finally {
    secondaryAccountLoading.value = false;
  }
};

watch(
  () => model.value,
  async (val) => {
    if (val) {
      await loadSecondaryAccountList();
      const autoAccountId =
        props.defaultAccountId || (secondaryAccountList.value.length === 1 ? secondaryAccountList.value[0].id : '');
      if (autoAccountId) {
        tableData.value = [{ ...defaultRow(), account_id: autoAccountId }];
      }
    }
  },
);

const accountType = ref<number>(1);

const defaultRow = (): IRowData => ({
  account_id: undefined,
  account_name: '',
  name: '',
  permission_template_ids: [],
  phone_num: '',
  country_code: '',
  email: '',
  managers: [],
  receive_email: '',
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

const tableData = ref<IRowData[]>([defaultRow()]);

const isSubmitting = ref(false);

interface InputColumnExpose {
  getValue: () => Promise<string | number>;
  focus: () => void;
}

const accountRefs = ref<Record<number, InstanceType<typeof SelectColumn>>>({});
const nameRefs = ref<Record<number, InputColumnExpose>>({});
const permissionTemplateRefs = ref<Record<number, { getValue: () => Promise<any> }>>({});
const managerRefs = ref<Record<number, { getValue: () => Promise<any> }>>({});
const receiveEmailRefs = ref<Record<number, InputColumnExpose>>({});
const phoneRefs = ref<Record<number, InputColumnExpose>>({});
const emailRefs = ref<Record<number, InputColumnExpose>>({});

const secondaryAccountSelectList = computed(() =>
  secondaryAccountList.value.map((item) => ({
    value: item.id,
    label: `${item.name}(${item?.extension?.cloud_main_account_id})`,
  })),
);

const getRowPermTemplateListGenerator = (row: IRowData) => {
  const secondaryAccount = secondaryAccountList.value.find((item) => item.id === row.account_id);
  if (!secondaryAccount?.extension?.cloud_main_account_id) return null;
  const defaultParams = {
    extension: {
      cloud_main_account_ids: [secondaryAccount.extension.cloud_main_account_id],
    },
    permission_template_type: PERMISSION_TEMPLATE_TYPES.SYNC_WITH_LIBRARY,
  };
  return permissionTemplateStore.createPermissionTemplateListGenerator(getBizsId(), currentVendor.value, defaultParams);
};

const handleClose = () => {
  model.value = false;
  accountType.value = 1;
  tableData.value = [defaultRow()];
};

const handleAddRow = (index: number) => {
  tableData.value.splice(index + 1, 0, defaultRow());
};

const handleCopyRow = (index: number) => {
  const copiedRow = { ...tableData.value[index] };
  copiedRow.managers = [...tableData.value[index].managers];
  copiedRow.permission_template_ids = [...tableData.value[index].permission_template_ids];
  tableData.value.splice(index + 1, 0, copiedRow);
};

const handleRemoveRow = (index: number) => {
  if (tableData.value.length <= 1) {
    Message({ theme: 'warning', message: '至少保留一行' });
    return;
  }
  tableData.value.splice(index, 1);
};

const handleBatchUpdateAccount = async (val: string) => {
  if (!val) return;
  tableData.value.forEach((row) => {
    row.account_id = val;
  });
  await nextTick();
  Object.values(accountRefs.value).forEach((r) => r?.getValue?.());
};

const handleBatchUpdateName = async (val: string) => {
  if (!val) return;
  tableData.value.forEach((row) => {
    row.name = val;
  });
  await nextTick();
  Object.values(nameRefs.value).forEach((r) => r?.getValue?.());
};

const handleBatchUpdateManagers = async (val: string | string[]) => {
  const managers = Array.isArray(val) ? val : [val];
  if (!managers.length) return;
  tableData.value.forEach((row) => {
    row.managers = [...managers];
  });
  await nextTick();
  Object.values(managerRefs.value).forEach((r) => r?.getValue?.());
};

const headList = computed(() => [
  {
    title: '所属二级账号',
    minWidth: 140,
    required: true,
    renderAppend: () =>
      h(
        BatchUpdatePopConfirm,
        { title: '所属二级账号', onUpdateValue: handleBatchUpdateAccount },
        {
          content: ({ value, updateValue }: { value: string; updateValue: (v: string) => void }) =>
            h(
              Select,
              {
                modelValue: value || '',
                'onUpdate:modelValue': updateValue,
                filterable: true,
                placeholder: '请选择二级账号',
                popoverOptions: { boundary: 'parent' },
              },
              () =>
                secondaryAccountList.value.map((item) =>
                  h(Select.Option, {
                    key: item.id,
                    value: item.id,
                    label: `${item.name}(${item.id})`,
                  }),
                ),
            ),
        },
      ),
  },
  {
    title: '三级账号名称',
    minWidth: 140,
    required: true,
    renderAppend: () =>
      h(BatchUpdatePopConfirm, {
        title: '三级账号名称',
        valueType: 'string',
        onUpdateValue: handleBatchUpdateName,
      }),
  },
  { title: '权限模板', minWidth: 220, required: true },
  {
    title: '手机号',
    minWidth: 120,
    required: false,
    memo: '填写"+地域代码号码"，如+8613212345678。+86是地区代码 13212345678是电话号码',
  },
  { title: '账号邮箱', minWidth: 130, required: false },
  {
    title: '负责人',
    minWidth: 240,
    required: true,
    renderAppend: () =>
      h(
        BatchUpdatePopConfirm,
        { title: '负责人', onUpdateValue: handleBatchUpdateManagers },
        {
          content: ({ value, updateValue }: { value: any; updateValue: (v: any) => void }) =>
            h(UserSelector, {
              modelValue: value || [],
              'onUpdate:modelValue': updateValue,
              multiple: true,
              collapseTags: false,
              allowCreate: true,
              placeholder: '请输入负责人',
            }),
        },
      ),
  },
  { title: '账号开通接收邮箱', minWidth: 140, required: true },
  { title: '', width: 112, required: false },
]);

const handleSubmit = async () => {
  try {
    const allRefs = tableData.value
      .flatMap((_, index) => [
        accountRefs.value[index],
        nameRefs.value[index],
        permissionTemplateRefs.value[index],
        managerRefs.value[index],
        receiveEmailRefs.value[index],
        phoneRefs.value[index],
        emailRefs.value[index],
      ])
      .filter(Boolean);
    await Promise.all(allRefs.map((r) => r.getValue()));
  } catch {
    return;
  }

  const validRows = tableData.value.filter((row) => row.account_id && row.name);
  if (validRows.length === 0) {
    Message({ theme: 'warning', message: '请至少填写一行完整的账号信息' });
    return;
  }

  const subAccounts: ISubAccountCreateParams[] = validRows.map((row) => {
    const { countryCode, phoneNum } = parsePhoneInput(row.phone_num);
    return {
      account_id: row.account_id,
      name: row.name,
      receive_email: row.receive_email,
      email: row.email || undefined,
      phone_num: phoneNum || undefined,
      country_code: countryCode,
      managers: row.managers,
      memo: '',
      permission_template_ids: row.permission_template_ids,
      extension: {
        console_login: accountType.value,
      },
    };
  });

  isSubmitting.value = true;
  try {
    const result = await tertiaryAccountStore.createSubAccount(getBizsId(), currentVendor.value, subAccounts);
    Message({ theme: 'success', message: '申请提交成功' });
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
    console.error('创建三级账号失败:', error);
  } finally {
    isSubmitting.value = false;
  }
};

const handleChangeSecondaryAccount = (index: number) => {
  // 变更二级账号时，清空权限模板选择数据
  tableData.value[index].permission_template_ids = [];
};

const receiveEmailRules = computed(() => [
  { validator: (v: any) => Boolean(v), message: '请输入账号开通接收邮箱' },
  {
    validator: (v: any) => {
      if (!v || !v.trim()) return true;
      return isEmail(v.trim());
    },
    message: '邮箱格式不正确',
  },
]);
</script>

<template>
  <bk-sideslider
    :is-show="model"
    :width="1280"
    title="创建三级账号"
    :before-close="handleClose"
    render-directive="if"
    @closed="handleClose"
  >
    <template #default>
      <div class="create-form">
        <div class="form-item">
          <label class="form-label required">账号类型</label>
          <bk-radio-group v-model="accountType">
            <bk-radio-button :label="1">控制台账号</bk-radio-button>
            <bk-radio-button :label="0">编程账号</bk-radio-button>
          </bk-radio-group>
        </div>

        <div class="form-item">
          <label class="form-label">账号信息录入</label>
          <Ediatable :thead-list="headList">
            <template #data>
              <tr v-for="(row, index) in tableData" :key="index">
                <td>
                  <SelectColumn
                    v-model="row.account_id"
                    :ref="(el: any) => (accountRefs[index] = el)"
                    :list="secondaryAccountSelectList"
                    :loading="secondaryAccountLoading"
                    :rules="[{ validator: (v: any) => Boolean(v), message: '请选择所属二级账号' }]"
                    filterable
                    placeholder="请选择"
                    @change="handleChangeSecondaryAccount(index)"
                  />
                </td>
                <td>
                  <InputColumn
                    v-model="row.name"
                    :ref="(el: any) => (nameRefs[index] = el)"
                    :rules="[{ validator: (v: any) => Boolean(v), message: '请输入三级账号名称' }]"
                    placeholder="请输入"
                  />
                </td>
                <td>
                  <hcm-form-list
                    v-model="row.permission_template_ids"
                    :list-generator="getRowPermTemplateListGenerator(row)"
                    :ref="(el: any) => (permissionTemplateRefs[index] = el)"
                    placeholder="请选择"
                    display-key="name"
                    id-key="id"
                    :display="{ on: 'cell' }"
                    collapse-tags
                    multiple
                    :rules="[{ validator: (v: any) => (Boolean(v?.length)), message: '请选择权限模板' }]"
                  ></hcm-form-list>
                </td>
                <td>
                  <InputColumn
                    v-model="row.phone_num"
                    :ref="(el: any) => (phoneRefs[index] = el)"
                    :rules="[{
                      validator: (v: any) => {
                        if (!v || !v.trim()) return true;
                        const phoneNumber = parsePhoneNumberFromString(v.trim());
                        return phoneNumber?.isValid() === true;
                      },
                      message: '手机号格式不正确'
                    }]"
                    placeholder="请输入"
                  />
                </td>
                <td>
                  <InputColumn
                    v-model="row.email"
                    :ref="(el: any) => (emailRefs[index] = el)"
                    :rules="[{
                      validator: (v: any) => {
                        if (!v || !v.trim()) return true;
                        return isEmail(v.trim());
                      },
                      message: '邮箱格式不正确'
                    }]"
                    placeholder="请输入"
                  />
                </td>
                <td>
                  <hcm-form-user
                    v-model="row.managers"
                    :ref="(el: any) => (managerRefs[index] = el)"
                    :display="{ on: 'cell' }"
                    :clearable="false"
                    :rules="[{ validator: (v: any) => Boolean(v?.length), message: '负责人不能为空' }]"
                    placeholder="请输入"
                  />
                </td>
                <td>
                  <InputColumn
                    v-model="row.receive_email"
                    :ref="(el: any) => (receiveEmailRefs[index] = el)"
                    :rules="receiveEmailRules"
                    placeholder="请输入"
                  />
                </td>
                <OperationColumn
                  :show-copy="true"
                  :show-add="true"
                  :show-remove="true"
                  :removable="tableData.length > 1"
                  @copy="handleCopyRow(index)"
                  @add="handleAddRow(index)"
                  @remove="handleRemoveRow(index)"
                />
              </tr>
            </template>
          </Ediatable>
        </div>
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
.create-form {
  padding: 26px 40px;

  .form-item {
    margin-bottom: 24px;

    .form-label {
      display: block;
      margin-bottom: 8px;
      font-size: 14px;
      color: #313238;

      &.required::after {
        content: '*';
        color: #ea3636;
        margin-left: 4px;
      }
    }

    :deep(.bk-radio-button) {
      .bk-radio-button-label {
        width: 150px;
        text-align: center;
      }
    }
  }
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
