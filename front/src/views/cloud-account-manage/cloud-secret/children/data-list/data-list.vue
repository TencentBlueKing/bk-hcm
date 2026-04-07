<script setup lang="ts">
import { h } from 'vue';
import { Message, Button } from 'bkui-vue';
import { Copy } from 'bkui-vue/lib/icon';
import { PaginationType } from '@/typings';
import { ModelPropertyColumn } from '@/model/typings';
import usePage from '@/hooks/use-page';
import useTableSettings from '@/hooks/use-table-settings';
import useClipboard from 'vue-clipboard3';
import { AUTH_UPDATE_SUB_ACCOUNT_SECRET, AUTH_DELETE_SUB_ACCOUNT_SECRET } from '@/constants/auth-symbols';
import { CONSOLE_LOGIN_MAP } from '../../constants';
import type { ICloudSecretItem } from '../../typings';

export interface IDataListProps {
  columns: ModelPropertyColumn[];
  list: ICloudSecretItem[];
  pagination: PaginationType;
  loading?: boolean;
}

const props = withDefaults(defineProps<IDataListProps>(), {
  loading: false,
});

const emit = defineEmits<{
  'view-details': [row: ICloudSecretItem];
  enable: [row: ICloudSecretItem];
  disable: [row: ICloudSecretItem];
  delete: [row: ICloudSecretItem];
}>();

const { handlePageChange, handlePageSizeChange, handleSort } = usePage();

const { settings } = useTableSettings(props.columns);

const { toClipboard } = useClipboard();

const formatDateTime = (dateStr?: string) => {
  if (!dateStr) return '--';
  return dateStr.replace('T', ' ').replace('Z', '');
};

const maskSecretId = (id: string) => {
  if (!id) return '--';
  if (id.length <= 8) return id;
  return `${id.substring(0, 4)}****${id.substring(id.length - 4)}`;
};

const handleCopySecretId = async (id: string, event: Event) => {
  event.stopPropagation();
  try {
    await toClipboard(id);
    Message({ theme: 'success', message: '复制成功' });
  } catch (e) {
    Message({ theme: 'error', message: '复制失败' });
  }
};

const handleViewDetails = (row: ICloudSecretItem) => {
  emit('view-details', row);
};

const handleEnable = (row: ICloudSecretItem) => {
  emit('enable', row);
};

const handleDisable = (row: ICloudSecretItem) => {
  emit('disable', row);
};

const handleDelete = (row: ICloudSecretItem) => {
  emit('delete', row);
};

const getColumnRender = (column: ModelPropertyColumn) => {
  if (column.id === 'cloud_secret_id') {
    return ({ row }: { row: ICloudSecretItem }) => {
      const secretId = row.cloud_secret_id || row.extension?.cloud_secret_id || '';
      return h('span', { class: 'secret-id-cell' }, [
        h(
          Button,
          {
            text: true,
            theme: 'primary',
            onClick: () => handleViewDetails(row),
          },
          () => maskSecretId(secretId),
        ),
        h(Copy, {
          class: 'copy-icon',
          onClick: (e: Event) => handleCopySecretId(secretId, e),
        }),
      ]);
    };
  }

  if (column.id === 'console_login') {
    return ({ row }: { row: ICloudSecretItem }) => {
      const consoleLogin = row.console_login ?? row.extension?.console_login;
      return CONSOLE_LOGIN_MAP[consoleLogin as number] || '--';
    };
  }

  if (column.id === 'cloud_sub_account_id') {
    return ({ row }: { row: ICloudSecretItem }) =>
      row.cloud_sub_account_id || row.extension?.cloud_sub_account_id || '--';
  }

  if (column.id === 'cloud_main_account_id') {
    return ({ row }: { row: ICloudSecretItem }) =>
      row.cloud_main_account_id || row.extension?.cloud_main_account_id || '--';
  }

  if (['cloud_created_at', 'last_used_time', 'disabled_time'].includes(column.id)) {
    return ({ row }: { row: ICloudSecretItem }) => formatDateTime(row[column.id as keyof ICloudSecretItem] as string);
  }

  return null;
};
</script>

<template>
  <bk-loading :loading="loading">
    <bk-table
      row-hover="auto"
      :data="list"
      :pagination="pagination"
      :max-height="`calc(100vh - 520px)`"
      :settings="settings"
      remote-pagination
      show-overflow-tooltip
      @page-limit-change="handlePageSizeChange"
      @page-value-change="handlePageChange"
      @column-sort="handleSort"
      row-key="id"
    >
      <bk-table-column
        v-for="(column, index) in columns"
        :key="index"
        :prop="column.id"
        :label="column.name"
        :sort="column.sort"
        :width="column.width"
        :min-width="column.minWidth"
        :fixed="column.fixed"
        v-bind="column"
      >
        <template #default="{ row }">
          <template v-if="getColumnRender(column)">
            <component :is="() => getColumnRender(column)({ row })" />
          </template>
          <template v-else>
            <display-value :property="column" :value="row[column.id]" :display="column?.meta?.display" />
          </template>
        </template>
      </bk-table-column>
      <bk-table-column label="操作" width="120" fixed="right">
        <template #default="{ row }">
          <template v-if="row.status === 'enabled'">
            <hcm-auth :sign="{ type: AUTH_UPDATE_SUB_ACCOUNT_SECRET }" v-slot="{ noPerm }">
              <bk-button theme="primary" text :disabled="noPerm" @click="handleDisable(row)">禁用</bk-button>
            </hcm-auth>
          </template>
          <template v-else>
            <hcm-auth :sign="{ type: AUTH_UPDATE_SUB_ACCOUNT_SECRET }" v-slot="{ noPerm }">
              <bk-button theme="primary" text class="mr8" :disabled="noPerm" @click="handleEnable(row)">启用</bk-button>
            </hcm-auth>
            <hcm-auth :sign="{ type: AUTH_DELETE_SUB_ACCOUNT_SECRET }" v-slot="{ noPerm }">
              <bk-button theme="primary" text :disabled="noPerm" @click="handleDelete(row)">删除</bk-button>
            </hcm-auth>
          </template>
        </template>
      </bk-table-column>
    </bk-table>
  </bk-loading>
</template>

<style lang="scss" scoped>
.status-cell {
  display: flex;
  align-items: center;
}

.status-dot {
  display: inline-block;
  width: 8px;
  height: 8px;
  border-radius: 50%;
  margin-right: 6px;
}

:deep(.status-enabled) {
  .status-dot {
    background-color: #2dcb56;
  }
}

:deep(.status-disabled) {
  .status-dot {
    background-color: #979ba5;
  }
}

.dot-enabled {
  background-color: #2dcb56;
}

.dot-disabled {
  background-color: #979ba5;
}

.mr8 {
  margin-right: 8px;
}

:deep(.secret-id-cell) {
  display: inline-flex;
  align-items: center;
  gap: 4px;

  .copy-icon {
    font-size: 14px;
    color: #979ba5;
    cursor: pointer;
    flex-shrink: 0;

    &:hover {
      color: #3a84ff;
    }
  }
}
</style>
