<script setup lang="ts">
import { computed, ref, watchEffect } from 'vue';
import { useI18n } from 'vue-i18n';
import { getPrivateIPs, getPublicIPs } from '@/utils';
import { timeFormatter } from '@/common/util';
import { useHostStore, type CvmsAssignPreviewItem, type IMatchHostsItem } from '@/store';
import { ResourceTypeEnum } from '@/common/resource-constant';

const props = defineProps<{
  action: 'backfill' | 'submit';
  cvm: CvmsAssignPreviewItem;
}>();
const emit = defineEmits<{
  (e: 'backfill', cvm: CvmsAssignPreviewItem, bkBizId: number, bkCloudId: number): void;
  (e: 'submit', cvm: { cvm_id: string; bk_biz_id: number; bk_cloud_id: number }): void;
  (e: 'manual-assign'): void;
}>();
const model = defineModel<boolean>();

const { t } = useI18n();
const hostStore = useHostStore();

const tableData = ref<IMatchHostsItem[]>([]);

watchEffect(async () => {
  if (props.cvm) {
    const { account_id, private_ipv4_addresses } = props.cvm;
    const list = await hostStore.getAssignHostsMatchList(account_id, private_ipv4_addresses);
    tableData.value = list.map((item) => ({
      ...item,
      private_ip_address: getPrivateIPs(item),
      public_ip_address: getPublicIPs(item),
    }));
  }
});

const selected = ref<number>();
const selectedHost = computed(() => tableData.value.find((item) => item.bk_host_id === selected.value));
const handleDelete = () => {
  selected.value = undefined;
};
const handleRowClick = (row: IMatchHostsItem) => {
  // 暂不支持0管控区
  if (row.bk_cloud_id === 0) return;
  selected.value = row.bk_host_id;
};

const columns = [
  { id: 'private_ip_address', name: t('内网IP'), type: 'string' },
  { id: 'public_ip_address', name: t('公网IP'), type: 'string' },
  { id: 'bk_cloud_id', name: t('管控区域'), type: 'cloud-area' },
  { id: 'bk_biz_id', name: t('所属业务'), type: 'business' },
  { id: 'region', name: t('地域'), type: 'region' },
  { id: 'bk_host_name', name: t('主机名称'), type: 'string' },
  { id: 'bk_os_name', name: t('操作系统'), type: 'string', width: 200 },
  {
    id: 'create_time',
    name: t('创建时间'),
    type: 'string',
    width: 180,
    render: ({ cell }: any) => timeFormatter(cell),
  },
];

const handleConfirm = async () => {
  if (props.action === 'backfill') {
    // 批量分配
    emit('backfill', props.cvm, selectedHost.value.bk_biz_id, selectedHost.value.bk_cloud_id);
  } else {
    // 单个分配
    emit('submit', {
      cvm_id: props.cvm.id,
      bk_biz_id: selectedHost.value.bk_biz_id,
      bk_cloud_id: selectedHost.value.bk_cloud_id,
    });
  }
  handleClosed();
};

const handleClosed = () => {
  selected.value = undefined;
  model.value = false;
};

defineExpose({ handleClosed });
</script>

<template>
  <bk-dialog :is-show="model" :title="t('关联配置平台主机')" width="1280" @closed="handleClosed">
    <div class="selected-preview-wrap">
      <span class="label">{{ t('已选') }}</span>
      <span v-if="selected" class="value">
        <span class="mr8">{{ getPrivateIPs(selectedHost) }}</span>
        <bk-button text @click="handleDelete"><i class="hcm-icon bkhcm-icon-close"></i></bk-button>
      </span>
    </div>
    <bk-table
      :data="tableData"
      :thead="{ color: '#F0F1F5' }"
      row-key="id"
      row-hover="auto"
      show-overflow-tooltip
      pagination
      @row-click="(_:any, row: any) => handleRowClick(row)"
      v-bkloading="{ loading: hostStore.isAssignHostsMatchLoading }"
    >
      <bk-table-column prop="radio" width="50" min-width="50">
        <template #default="{ row }">
          <!-- 暂不支持0管控区 -->
          <bk-radio v-model="selected" :label="row.bk_host_id" :disabled="row.bk_cloud_id === 0" class="no-label" />
        </template>
      </bk-table-column>
      <bk-table-column
        v-for="(column, idx) in columns"
        :key="idx"
        :prop="column.id"
        :label="column.name"
        :render="column.render"
        :width="column.width"
      >
        <template #default="{ row }">
          <display-value
            :property="column"
            :value="row[column.id]"
            :vendor="row?.vendor"
            :resource-type="ResourceTypeEnum.CVM"
          />
        </template>
      </bk-table-column>
    </bk-table>
    <div class="help">
      <i class="hcm-icon bkhcm-icon-question-circle-fill"></i>
      {{ t('没有找到想要关联的主机？可尝试') }}
      <bk-button class="button" text @click="emit('manual-assign')">{{ t('手动分配') }}</bk-button>
    </div>

    <template #footer>
      <bk-button
        theme="primary"
        :loading="hostStore.isAssignCvmsToBizsLoading"
        :disabled="!selectedHost"
        v-bk-tooltips="{ content: t('未选择主机'), disabled: selectedHost }"
        @click="handleConfirm"
      >
        {{ t('关联所选主机') }}
      </bk-button>
      <bk-button class="ml8" :disabled="hostStore.isAssignCvmsToBizsLoading" @click="handleClosed">
        {{ t('取消') }}
      </bk-button>
    </template>
  </bk-dialog>
</template>

<style scoped lang="scss">
.selected-preview-wrap {
  margin-bottom: 16px;
  height: 32px;
  line-height: 32px;
  font-size: 12px;

  .label {
    margin-right: 12px;
    font-weight: 700;
  }

  .value {
    display: inline-flex;
    align-items: center;
    height: 32px;
    padding: 0 12px;
    background: #f0f5ff;
    border: 1px solid #a3c5fd;
    border-radius: 2px;
  }
}

.no-label {
  :deep(.bk-radio-label) {
    display: none;
  }
}

.help {
  margin: 0 auto;
  width: 400px;
  height: 32px;
  line-height: 32px;
  background: #f0f1f5;
  border-radius: 21px;
  font-size: 12px;
  text-align: center;

  .hcm-icon,
  .button {
    color: #699df4;
  }

  .hcm-icon {
    font-size: 14px;
  }
}
</style>
