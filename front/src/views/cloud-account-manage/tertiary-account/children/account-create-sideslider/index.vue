<script setup lang="ts">
import { ref, inject, computed, watch, type Ref } from 'vue';
import { Message } from 'bkui-vue';
import { Ediatable, InputColumn, SelectColumn } from '@blueking/ediatable';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import { useCloudAccountStore, type ISubAccountCreateParams, type ISecondaryAccountItem } from '@/store/cloud-account';
import { VendorEnum } from '@/common/constant';
import { QueryRuleOPEnum, type QueryFilterType } from '@/typings';
import OperationColumn from '@/components/ediatable/operation-column.vue';
import UserSelector from '@/components/user-selector/index.vue';
import SecondaryAccountSelector from './secondary-account-selector.vue';

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

const headList = computed(() => [
  { title: '所属二级账号', minWidth: 140, required: true },
  { title: '三级账号名称', minWidth: 140, required: true },
  { title: '权限模版', minWidth: 140, required: true },
  { title: '手机号', minWidth: 120, required: false },
  { title: '账号邮箱', minWidth: 130, required: false },
  { title: '负责人', minWidth: 180, required: true },
  { title: '账号开通接收邮箱', minWidth: 140, required: true },
  { title: '', width: 112, required: false },
]);

const handleSubmit = async () => {
  const validRows = tableData.value.filter((row) => row.account_id && row.name);
  if (validRows.length === 0) {
    Message({ theme: 'warning', message: '请至少填写一行完整的账号信息' });
    return;
  }

  for (const row of validRows) {
    if (!row.account_id) {
      Message({ theme: 'warning', message: '请选择所属二级账号' });
      return;
    }
    if (!row.name) {
      Message({ theme: 'warning', message: '请输入三级账号名称' });
      return;
    }
    if (!row.managers?.length) {
      Message({ theme: 'warning', message: '请选择负责人' });
      return;
    }
    if (!row.receive_email) {
      Message({ theme: 'warning', message: '请输入账号开通接收邮箱' });
      return;
    }
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
                  <SecondaryAccountSelector
                    v-model="row.account_id"
                    :account-list="secondaryAccountList"
                    :loading="secondaryAccountLoading"
                  />
                </td>
                <td>
                  <InputColumn v-model="row.name" placeholder="请输入" />
                </td>
                <td>
                  <SelectColumn v-model="row.permission_template" :list="[]" multiple placeholder="请选择" />
                </td>
                <td>
                  <InputColumn v-model="row.phone_num" placeholder="请输入" />
                </td>
                <td>
                  <InputColumn v-model="row.email" placeholder="请输入" />
                </td>
                <td>
                  <UserSelector
                    v-model="row.managers"
                    :multiple="true"
                    :collapse-tags="false"
                    :allow-create="true"
                    placeholder="请输入负责人"
                  />
                </td>
                <td>
                  <InputColumn v-model="row.receive_email" placeholder="请输入" />
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
