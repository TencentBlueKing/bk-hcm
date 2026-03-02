<script setup lang="ts">
import { h } from 'vue';
import { PaginationType } from '@/typings';
import { ModelPropertyColumn } from '@/model/typings';
import usePage from '@/hooks/use-page';
import useTableSettings from '@/hooks/use-table-settings';
import { Button } from 'bkui-vue';
import { SECRET_STATUS_MAP, CONSOLE_LOGIN_MAP } from '../../constants';
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

// 定义事件
const emit = defineEmits<{
  'view-details': [row: ICloudSecretItem];
  enable: [row: ICloudSecretItem];
  disable: [row: ICloudSecretItem];
  delete: [row: ICloudSecretItem];
}>();

const { handlePageChange, handlePageSizeChange, handleSort } = usePage();

const { settings } = useTableSettings(props.columns);

// 格式化时间展示
const formatDateTime = (dateStr?: string) => {
  if (!dateStr) return '--';
  return dateStr.replace('T', ' ').replace('Z', '');
};

// 查看详情 - 触发事件
const handleViewDetails = (row: ICloudSecretItem) => {
  emit('view-details', row);
};

// 启用
const handleEnable = (row: ICloudSecretItem) => {
  emit('enable', row);
};

// 禁用
const handleDisable = (row: ICloudSecretItem) => {
  emit('disable', row);
};

// 删除
const handleDelete = (row: ICloudSecretItem) => {
  emit('delete', row);
};

// 自定义渲染列
const getColumnRender = (column: ModelPropertyColumn) => {
  // 密钥ID列 - 点击打开详情侧栏
  if (column.id === 'cloud_secret_id') {
    return ({ row }: { row: ICloudSecretItem }) =>
      h(
        Button,
        {
          text: true,
          theme: 'primary',
          onClick: () => handleViewDetails(row),
        },
        () => row.cloud_secret_id || row.extension?.cloud_secret_id || '--',
      );
  }

  // 密钥状态列
  if (column.id === 'status') {
    return ({ row }: { row: ICloudSecretItem }) => {
      const status = SECRET_STATUS_MAP[row.status] || { class: '', text: row.status || '--', dotClass: '' };
      return h('span', { class: ['status-cell', status.class] }, [
        h('span', { class: ['status-dot', status.dotClass] }),
        status.text,
      ]);
    };
  }

  // 三级账号类型列
  if (column.id === 'console_login') {
    return ({ row }: { row: ICloudSecretItem }) => {
      const consoleLogin = row.console_login ?? row.extension?.console_login;
      return CONSOLE_LOGIN_MAP[consoleLogin as number] || '--';
    };
  }

  // 所属三级账号ID列
  if (column.id === 'cloud_sub_account_id') {
    return ({ row }: { row: ICloudSecretItem }) =>
      row.cloud_sub_account_id || row.extension?.cloud_sub_account_id || '--';
  }

  // 所属二级账号ID列
  if (column.id === 'cloud_main_account_id') {
    return ({ row }: { row: ICloudSecretItem }) =>
      row.cloud_main_account_id || row.extension?.cloud_main_account_id || '--';
  }

  // 时间类型列
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
      :max-height="`calc(100vh - 450px)`"
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
          <!-- 已启用状态：显示禁用按钮 -->
          <template v-if="row.status === 'enabled'">
            <bk-button theme="primary" text @click="handleDisable(row)">禁用</bk-button>
          </template>
          <!-- 已禁用状态：显示启用和删除按钮 -->
          <template v-else>
            <bk-button theme="primary" text class="mr8" @click="handleEnable(row)">启用</bk-button>
            <bk-button theme="danger" text @click="handleDelete(row)">删除</bk-button>
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
</style>
