<script setup lang="ts">
import { computed, reactive, ref } from 'vue';
import type { IPermissionPolicyItem } from '../../typings';
import { PolicyApplyStatus } from '../../typings';
import usePage from '@/hooks/use-page';
import ModelInfoDialog from '@/views/cloud-account-manage/permission-policy/children/dialog/info.vue';
import PolicyDiffDialog from '@/views/cloud-account-manage/permission-policy/children/dialog/diff.vue';
import { IPermissionAppliedItem } from '@/store/cloud-account-manage/permission-policy';
import { VendorEnum, SecondaryAccountResourceTypeEnum } from '@/common/constant';
import SecondaryAccountValue from '@/views/cloud-account-manage/components/secondary-account-value.vue';

const props = defineProps<{
  policyData: IPermissionPolicyItem | null;
  list: IPermissionAppliedItem[];
  bizId: number;
  vendor: VendorEnum;
}>();

const bizId = computed(() => props.bizId);
const vendor = computed(() => props.vendor);
const resType = computed(() =>
  bizId.value ? SecondaryAccountResourceTypeEnum.SUB : SecondaryAccountResourceTypeEnum.PERMISSION,
);

const tableData = computed(() => [...props.list]);
const isLoading = ref(false);
const { pagination } = usePage();

// 已选账号（用于批量更新）
const selectedAccounts = ref<IPermissionAppliedItem[]>([]);

// 模板详情数据
const modelInfo = reactive({
  show: false,
  json: '',
  accountId: '',
  id: '',
  name: '',
});

// 策略内容对比数据
const policyDiffInfo = reactive({
  show: false,
  accountId: '',
  cloudContent: {
    version: 1,
    json: '',
  },
});

const policyContent = computed(() => ({
  version: props.policyData?.version,
  json: props.policyData?.policy_document,
}));

// 状态映射
const statusMap: Record<string, { text: string; dotClass: string }> = {
  [PolicyApplyStatus.APPLIED]: { text: '已应用', dotClass: 'status-applied' },
  [PolicyApplyStatus.PENDING]: { text: '待应用', dotClass: 'status-pending' },
  [PolicyApplyStatus.DATA_MISMATCH]: { text: '已应用(数据不一致)', dotClass: 'status-mismatch' },
};

const getStatusInfo = (status: PolicyApplyStatus) => statusMap[status] || { text: status, dotClass: '' };

// 选择变化
const handleSelectionChange = ({ row, checked }: { row: IPermissionAppliedItem; checked: boolean }) => {
  if (checked) {
    if (!selectedAccounts.value.find((item) => item.id === row.id)) {
      selectedAccounts.value.push(row);
    }
  } else {
    selectedAccounts.value = selectedAccounts.value.filter((item) => item.id !== row.id);
  }
};

const handleSelectAll = ({ checked }: { checked: boolean }) => {
  if (checked) {
    selectedAccounts.value = [...tableData.value];
  } else {
    selectedAccounts.value = [];
  }
};

const handleModelInfo = (row: IPermissionAppliedItem) => {
  modelInfo.name = row.name;
  modelInfo.id = row.cloud_id;
  modelInfo.accountId = row.account_id;
  modelInfo.json = row.policy_document;
  modelInfo.show = true;
};

const handlePolicyDiff = (row: IPermissionAppliedItem) => {
  policyDiffInfo.show = true;
  policyDiffInfo.accountId = row.account_id;
  policyDiffInfo.cloudContent = {
    version: row.policy_library_version,
    json: row.policy_document,
  };
};

// 暴露已选数据给父组件（defineExpose 必须是 <script setup> 的最后语句）
defineExpose({
  getSelectedAccounts: () => selectedAccounts.value,
});
</script>

<template>
  <div class="update-applied-table">
    <div class="section-label">选择要同步的账号</div>
    <bk-loading :loading="isLoading">
      <bk-table
        ref="tableRef"
        :data="tableData"
        :pagination="pagination"
        :max-height="400"
        :border="['outer', 'row']"
        row-key="account_id"
        show-overflow-tooltip
        @select="handleSelectionChange"
        @select-all="handleSelectAll"
      >
        <bk-table-column type="selection" align="center" />
        <bk-table-column label="二级账号" min-width="180">
          <template #default="{ row }">
            <SecondaryAccountValue :value="row.account_id" :biz-id="bizId" :vendor="vendor" :res-type="resType" />
          </template>
        </bk-table-column>
        <bk-table-column label="云上模版名称" prop="name" min-width="150" />
        <bk-table-column label="云模版同步时间" prop="policy_library_sync_time" min-width="160">
          <template #default="{ row }">
            <display-value class="info-value" :property="{ type: 'datetime' }" :value="row.policy_library_sync_time" />
          </template>
        </bk-table-column>
        <bk-table-column label="策略库应用版本" prop="policy_library_version">
          <template #default="{ row }">
            <span class="version">v{{ row.policy_library_version }}</span>
          </template>
        </bk-table-column>
        <bk-table-column label="策略库应用状态" min-width="150">
          <template #default="{ row }">
            <span class="status-cell">
              <span :class="['status-dot', getStatusInfo(row.apply_status).dotClass]"></span>
              <span :class="{ 'text-danger': row.apply_status === PolicyApplyStatus.DATA_MISMATCH }">
                {{ getStatusInfo(row.apply_status).text }}
              </span>
            </span>
          </template>
        </bk-table-column>
        <bk-table-column label="策略库应用时间" prop="created_at" min-width="160">
          <template #default="{ row }">
            <display-value class="info-value" :property="{ type: 'datetime' }" :value="row.created_at" />
          </template>
        </bk-table-column>
        <bk-table-column label="操作" width="160" fixed="right">
          <template #default="{ row }">
            <bk-button theme="primary" text class="mr8" @click="handleModelInfo(row)">模版详情</bk-button>
            <bk-button
              v-if="
                row.apply_status === PolicyApplyStatus.PENDING || row.apply_status === PolicyApplyStatus.DATA_MISMATCH
              "
              theme="primary"
              text
              @click="handlePolicyDiff(row)"
            >
              策略对比
            </bk-button>
          </template>
        </bk-table-column>
      </bk-table>
    </bk-loading>
    <!--模版详情弹框-->
    <ModelInfoDialog v-bind="modelInfo" @close="modelInfo.show = false" :res-type="resType" />
    <!--策略对比弹框-->
    <PolicyDiffDialog v-bind="policyDiffInfo" @close="policyDiffInfo.show = false" :policy-content="policyContent" />
  </div>
</template>

<style lang="scss" scoped>
.update-applied-table {
  .section-label {
    font-size: 12px;
    color: #63656e;
    margin-bottom: 8px;
  }

  .version {
    background: #eaebf0;
    padding: 3px 8px;
    border-radius: 5px;
  }

  .status-cell {
    display: inline-flex;
    align-items: center;

    .status-dot {
      display: inline-block;
      width: 8px;
      height: 8px;
      border-radius: 50%;
      margin-right: 6px;
      flex-shrink: 0;

      &.status-applied {
        background-color: #2dcb56;
      }

      &.status-pending {
        background-color: #ff9c01;
      }

      &.status-mismatch {
        background-color: #ea3636;
      }
    }

    .text-danger {
      color: #ea3636;
    }
  }

  .mr8 {
    margin-right: 8px;
  }
}
</style>
