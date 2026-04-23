<script setup lang="ts">
import { ref, inject, computed, h, type Ref, type ComponentPublicInstance, watch, nextTick } from 'vue';
import { Message, Select } from 'bkui-vue';
import { Ediatable, TextPlainColumn, SelectColumn } from '@blueking/ediatable';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import { useAccountStore } from '@/store';
import {
  useTertiaryAccountStore,
  type ISubAccountItem,
  type ISubAccountUpdateParams,
} from '@/store/cloud-account-manage/tertiary-account';
import { useAccountSelectorStore } from '@/store/account-selector';
import { VendorEnum } from '@/common/constant';
import OperationColumn from '@/components/ediatable/operation-column.vue';
import UserSelector from '@/components/user-selector/index.vue';

import BatchUpdatePopConfirm from '@/components/batch-update-popconfirm';

const model = defineModel<boolean>();

const props = defineProps<{
  selectedRows: ISubAccountItem[];
}>();

const emit = defineEmits<{
  (e: 'success'): void;
}>();

const currentVendor = inject<Ref<VendorEnum>>('currentVendor', ref(VendorEnum.TCLOUD));
const accountStore = useAccountStore();
const tertiaryAccountStore = useTertiaryAccountStore();
const accountSelectorStore = useAccountSelectorStore();
const { getBizsId } = useWhereAmI();

interface IBatchRow {
  id: string;
  cloud_id: string;
  name: string;
  account_id: string;
  account_name: string;
  managers: string[];
  bk_biz_ids: string | number | string[];
}

const batchData = ref<IBatchRow[]>([]);
const isSubmitting = ref(false);
const isReady = ref(false);
const bizList = ref<{ value: string; label: string }[]>([]);
const managerRefs = ref<Record<number, ComponentPublicInstance & { getValue: () => Promise<any> }>>({});
const bizRefs = ref<Record<number, ComponentPublicInstance & { getValue: () => Promise<any> }>>({});

watch(
  () => model.value,
  async (val) => {
    if (val) {
      isReady.value = false;
      batchData.value = props.selectedRows.map((row) => ({
        id: row.id,
        cloud_id: row.cloud_id,
        name: row.name,
        account_id: row.account_id,
        account_name: '',
        managers: [...(row.managers || [])],
        bk_biz_ids: (row.bk_biz_ids || []).map((id) => String(id)),
      }));
      // 并行请求业务列表和云账号列表（用于获取二级账号名称）
      const [bizRes, accountList] = await Promise.all([
        accountStore.getBizList(),
        accountSelectorStore.getBusinessAccountList({ bizId: getBizsId(), account_type: 'resource' }),
      ]);
      bizList.value = (bizRes?.data || []).map((item: { id: number; name: string }) => ({
        value: String(item.id),
        label: item.name,
      }));
      // 构建 account_id -> name 的映射，回填二级账号名称
      if (accountList?.length) {
        const nameMap = new Map<string, string>();
        accountList.forEach((item: { id: string; name: string }) => nameMap.set(item.id, item.name));
        batchData.value.forEach((row) => {
          row.account_name = nameMap.get(row.account_id) || '';
        });
      }
      await nextTick();
      isReady.value = true;
    }
  },
);

const handleClose = () => {
  model.value = false;
};

const handleRemoveRow = (index: number) => {
  if (batchData.value.length <= 1) {
    Message({ theme: 'warning', message: '至少保留一行' });
    return;
  }
  batchData.value.splice(index, 1);
};

const handleSubmit = async () => {
  try {
    const allRefs = batchData.value
      .flatMap((_, index) => [managerRefs.value[index], bizRefs.value[index]])
      .filter(Boolean);
    await Promise.all(allRefs.map((r) => r.getValue()));
  } catch {
    return;
  }

  const rows = batchData.value;

  const subAccounts: ISubAccountUpdateParams[] = rows.map((row) => {
    const bizId = Array.isArray(row.bk_biz_ids) ? row.bk_biz_ids[0] : row.bk_biz_ids;
    return {
      id: row.id,
      managers: row.managers,
      bk_biz_id: bizId !== undefined && bizId !== '' ? Number(bizId) : undefined,
    };
  });

  isSubmitting.value = true;
  try {
    await tertiaryAccountStore.updateSubAccount(getBizsId(), currentVendor.value, subAccounts);
    Message({ theme: 'success', message: '批量更新申请提交成功' });
    handleClose();
    emit('success');
  } catch (error) {
    console.error('批量更新失败:', error);
  } finally {
    isSubmitting.value = false;
  }
};

const handleBatchUpdateManagers = async (val: string | string[]) => {
  const managers = Array.isArray(val) ? val : [val];
  if (!managers.length) return;
  batchData.value.forEach((row) => {
    row.managers = [...managers];
  });
  await nextTick();
  Object.values(managerRefs.value).forEach((r) => r?.getValue?.());
};

const handleBatchUpdateBiz = async (val: string) => {
  if (!val) return;
  batchData.value.forEach((row) => {
    row.bk_biz_ids = [String(val)];
  });
  await nextTick();
  Object.values(bizRefs.value).forEach((r) => r?.getValue?.());
};

const headList = computed(() => [
  { title: '三级账号ID', minWidth: 120, required: false },
  { title: '三级账号名称', minWidth: 140, required: false },
  { title: '所属二级账号ID', minWidth: 130, required: false },
  { title: '所属二级账号名称', minWidth: 140, required: false },
  {
    title: '三级账号负责人',
    minWidth: 240,
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
  {
    title: '三级账号业务',
    minWidth: 160,
    renderAppend: () =>
      h(
        BatchUpdatePopConfirm,
        { title: '业务', onUpdateValue: handleBatchUpdateBiz },
        {
          content: ({ value, updateValue }: { value: string; updateValue: (v: string) => void }) =>
            h(
              Select,
              {
                modelValue: value || '',
                'onUpdate:modelValue': updateValue,
                filterable: true,
                placeholder: '请选择业务',
                popoverOptions: { boundary: 'parent' },
              },
              () =>
                bizList.value.map((item) =>
                  h(Select.Option, { key: item.value, value: item.value, label: item.label }),
                ),
            ),
        },
      ),
  },
  { title: '', width: 48, required: false },
]);
</script>

<template>
  <bk-sideslider
    :is-show="model"
    :width="1200"
    title="批量更新三级账号信息"
    :before-close="handleClose"
    @closed="handleClose"
  >
    <template #default>
      <div class="batch-update-form">
        <div class="selected-count">
          共选择
          <span class="highlight">{{ batchData.length }}</span>
          个三级账号
        </div>

        <bk-loading :loading="!isReady">
          <Ediatable v-if="isReady" :thead-list="headList">
            <template #data>
              <tr v-for="(row, index) in batchData" :key="row.id">
                <td>
                  <TextPlainColumn :data="row.cloud_id" />
                </td>
                <td>
                  <TextPlainColumn :data="row.name" />
                </td>
                <td>
                  <TextPlainColumn :data="row.account_id" />
                </td>
                <td>
                  <TextPlainColumn :data="row.account_name" />
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
                  <SelectColumn
                    v-model="row.bk_biz_ids"
                    :ref="(el: any) => (bizRefs[index] = el)"
                    :list="bizList"
                    filterable
                    :rules="[{ validator: (v: any) => Boolean(v?.length), message: '请选择业务' }]"
                  />
                </td>
                <OperationColumn
                  :show-add="false"
                  :show-copy="false"
                  :removable="batchData.length > 1"
                  remove-text="移除此行"
                  @remove="handleRemoveRow(index)"
                />
              </tr>
            </template>
          </Ediatable>
        </bk-loading>
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
.batch-update-form {
  padding: 28px 40px;

  .selected-count {
    font-size: 12px;
    color: #63656e;
    margin-bottom: 12px;

    .highlight {
      color: #3a84ff;
      font-weight: 600;
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
