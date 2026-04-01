<script setup lang="ts">
import { ref, inject, computed, type Ref, watch, nextTick } from 'vue';
import { Message } from 'bkui-vue';
import { Ediatable, TextPlainColumn, SelectColumn } from '@blueking/ediatable';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import { useAccountStore } from '@/store';
import { useCloudAccountStore, type ISubAccountItem, type ISubAccountUpdateParams } from '@/store/cloud-account';
import { VendorEnum } from '@/common/constant';
import OperationColumn from '@/components/ediatable/operation-column.vue';
import UserSelector from '@/components/user-selector/index.vue';

const props = defineProps<{
  modelValue: boolean;
  selectedRows: ISubAccountItem[];
}>();

const emit = defineEmits<{
  (e: 'update:modelValue', val: boolean): void;
  (e: 'success'): void;
}>();

const currentVendor = inject<Ref<VendorEnum>>('currentVendor', ref(VendorEnum.TCLOUD));
const accountStore = useAccountStore();
const cloudAccountStore = useCloudAccountStore();
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

watch(
  () => props.modelValue,
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
      const res = await accountStore.getBizList();
      bizList.value = (res?.data || []).map((item: { id: number; name: string }) => ({
        value: String(item.id),
        label: item.name,
      }));
      await nextTick();
      isReady.value = true;
    }
  },
);

const handleClose = () => {
  emit('update:modelValue', false);
};

const handleRemoveRow = (index: number) => {
  if (batchData.value.length <= 1) {
    Message({ theme: 'warning', message: '至少保留一行' });
    return;
  }
  batchData.value.splice(index, 1);
};

const handleSubmit = async () => {
  const rows = batchData.value;
  for (const row of rows) {
    if (!row.managers?.length) {
      Message({ theme: 'warning', message: '请填写所有三级账号的负责人' });
      return;
    }
  }

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
    await cloudAccountStore.updateSubAccount(getBizsId(), currentVendor.value, subAccounts);
    Message({ theme: 'success', message: '批量更新申请提交成功' });
    handleClose();
    emit('success');
  } catch (error) {
    console.error('批量更新失败:', error);
  } finally {
    isSubmitting.value = false;
  }
};

const headList = computed(() => [
  { title: '三级账号ID', minWidth: 120, required: false },
  { title: '三级账号名称', minWidth: 140, required: false },
  { title: '所属二级账号ID', minWidth: 130, required: false },
  { title: '所属二级账号名称', minWidth: 140, required: false },
  { title: '三级账号负责人', minWidth: 180, required: true },
  { title: '三级账号业务', minWidth: 160, required: true },
  { title: '', width: 50, required: false },
]);
</script>

<template>
  <bk-sideslider
    :is-show="modelValue"
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
                  <TextPlainColumn>{{ row.cloud_id }}</TextPlainColumn>
                </td>
                <td>
                  <TextPlainColumn>{{ row.name || '--' }}</TextPlainColumn>
                </td>
                <td>
                  <TextPlainColumn>{{ row.account_id }}</TextPlainColumn>
                </td>
                <td>
                  <TextPlainColumn>{{ row.account_name || '--' }}</TextPlainColumn>
                </td>
                <td>
                  <UserSelector
                    v-model="row.managers"
                    :multiple="true"
                    :collapse-tags="false"
                    :allow-create="true"
                    placeholder="请输入负责人"
                    :class="{ 'is-error': !row.managers?.length }"
                  />
                </td>
                <td>
                  <SelectColumn v-model="row.bk_biz_ids" :list="bizList" filterable />
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

/* stylelint-disable selector-class-pattern */
:deep(.user-selector .bk-tag-input-trigger) {
  min-height: 42px;
  border-color: transparent;
  border-radius: 0;
}

:deep(.user-selector .bk-tag-input-trigger:hover) {
  background-color: #fafbfd;
  border-color: #a3c5fd !important;
}

:deep(.user-selector .bk-tag-input-trigger.active) {
  border-color: #3a84ff !important;
}

:deep(.is-error .user-selector .bk-tag-input-trigger) {
  background-color: #fff0f1;
}
/* stylelint-enable selector-class-pattern */
</style>
