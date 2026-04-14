<script setup lang="ts">
import { ref, inject, computed, watch, nextTick, h, type Ref } from 'vue';
import { Message, Select } from 'bkui-vue';
import { Ediatable, InputColumn, SelectColumn } from '@blueking/ediatable';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import { useCloudAccountStore, type ISubAccountCreateParams, type ISecondaryAccountItem } from '@/store/cloud-account';
import { VendorEnum } from '@/common/constant';
import { QueryRuleOPEnum, type QueryFilterType } from '@/typings';
import OperationColumn from '@/components/ediatable/operation-column.vue';
import UserSelector from '@/components/user-selector/index.vue';
import BatchUpdatePopConfirm from '@/components/batch-update-popconfirm';
import ValidatedUserSelector from '../components/validated-user-selector.vue';
import ValidatedPermissionTemplateSelector from '../components/validated-permission-template-selector.vue';

const props = defineProps<{
  modelValue: boolean;
  defaultAccountId?: string;
}>();

const emit = defineEmits<{
  (e: 'update:modelValue', val: boolean): void;
  (e: 'success'): void;
}>();

const currentVendor = inject<Ref<VendorEnum>>('currentVendor', ref(VendorEnum.TCLOUD));
const cloudAccountStore = useCloudAccountStore();
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
    const list = await cloudAccountStore.getSecondaryAccountFullList(getBizsId(), filter);
    secondaryAccountList.value = list;
  } catch (error) {
    console.error('加载二级账号列表失败:', error);
  } finally {
    secondaryAccountLoading.value = false;
  }
};

watch(
  () => props.modelValue,
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

interface IRowData {
  account_id: string;
  account_name: string;
  name: string;
  permission_template: string[];
  phone_num: string;
  email: string;
  managers: string[];
  receive_email: string;
}

const defaultRow = (): IRowData => ({
  account_id: '',
  account_name: '',
  name: '',
  permission_template: [],
  phone_num: '',
  email: '',
  managers: [],
  receive_email: '',
});

const tableData = ref<IRowData[]>([defaultRow()]);

const isSubmitting = ref(false);

const accountRefs = ref<Record<number, InstanceType<typeof SelectColumn>>>({});
const nameRefs = ref<Record<number, InstanceType<typeof InputColumn>>>({});
const permissionTemplateRefs = ref<Record<number, InstanceType<typeof ValidatedPermissionTemplateSelector>>>({});
const managerRefs = ref<Record<number, InstanceType<typeof ValidatedUserSelector>>>({});
const receiveEmailRefs = ref<Record<number, InstanceType<typeof InputColumn>>>({});

const secondaryAccountSelectList = computed(() =>
  secondaryAccountList.value.map((item) => ({
    value: item.id,
    label: `${item.name}(${item.id})`,
  })),
);

const handleClose = () => {
  emit('update:modelValue', false);
  accountType.value = 1;
  tableData.value = [defaultRow()];
};

const handleAddRow = (index: number) => {
  tableData.value.splice(index + 1, 0, defaultRow());
};

const handleCopyRow = (index: number) => {
  const copiedRow = { ...tableData.value[index] };
  copiedRow.managers = [...tableData.value[index].managers];
  copiedRow.permission_template = [...tableData.value[index].permission_template];
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
  { title: '权限模版', minWidth: 140, required: true },
  { title: '手机号', minWidth: 120, required: false },
  { title: '账号邮箱', minWidth: 130, required: false },
  {
    title: '负责人',
    minWidth: 180,
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

  const subAccounts: ISubAccountCreateParams[] = validRows.map((row) => ({
    account_id: row.account_id,
    name: row.name,
    receive_email: row.receive_email,
    email: row.email || undefined,
    phone_num: row.phone_num || undefined,
    country_code: '86',
    managers: row.managers,
    memo: '',
    extension: {
      console_login: accountType.value,
    },
  }));

  isSubmitting.value = true;
  try {
    await cloudAccountStore.createSubAccount(getBizsId(), currentVendor.value, subAccounts);
    Message({ theme: 'success', message: '申请提交成功' });
    handleClose();
    emit('success');
  } catch (error) {
    console.error('创建三级账号失败:', error);
  } finally {
    isSubmitting.value = false;
  }
};
</script>

<template>
  <bk-sideslider
    :is-show="modelValue"
    :width="1280"
    title="创建三级账号"
    :before-close="handleClose"
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
                  <ValidatedPermissionTemplateSelector
                    v-model="row.permission_template"
                    :ref="(el: any) => (permissionTemplateRefs[index] = el)"
                    :display="{ on: 'cell' }"
                    placeholder="请选择"
                  />
                </td>
                <td>
                  <InputColumn v-model="row.phone_num" placeholder="请输入" />
                </td>
                <td>
                  <InputColumn v-model="row.email" placeholder="请输入" />
                </td>
                <td>
                  <ValidatedUserSelector
                    v-model="row.managers"
                    :ref="(el: any) => (managerRefs[index] = el)"
                    :multiple="true"
                    :collapse-tags="false"
                    :allow-create="true"
                    placeholder="请输入负责人"
                  />
                </td>
                <td>
                  <InputColumn
                    v-model="row.receive_email"
                    :ref="(el: any) => (receiveEmailRefs[index] = el)"
                    :rules="[{ validator: (v: any) => Boolean(v), message: '请输入账号开通接收邮箱' }]"
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

/* stylelint-disable selector-class-pattern */
:deep(.user-selector .bk-tag-input-trigger) {
  min-height: 42px;
  border-color: transparent;
  border-radius: 0;

  .placeholder {
    line-height: 42px;
    top: 0;
  }
}

:deep(.user-selector .bk-tag-input-trigger:hover) {
  background-color: #fafbfd;
  border-color: #a3c5fd !important;
}

:deep(.user-selector .bk-tag-input-trigger.active) {
  border-color: #3a84ff !important;
}
/* stylelint-enable selector-class-pattern */
</style>
