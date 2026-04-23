<script setup lang="ts">
import { computed, h } from 'vue';
import { PaginationType } from '@/typings';
import { ModelPropertyColumn } from '@/model/typings';
import usePage from '@/hooks/use-page';
import useTableSettings from '@/hooks/use-table-settings';
import { Button } from 'bkui-vue';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import {
  AUTH_UPDATE_PERMISSION_POLICY_LIBRARY,
  AUTH_APPLY_PERMISSION_POLICY_LIBRARY,
  AUTH_BIZ_UPDATE_PERMISSION_POLICY_LIBRARY,
  AUTH_BIZ_APPLY_PERMISSION_POLICY_LIBRARY,
} from '@/constants/auth-symbols';
import type { IPermissionPolicyItem } from '../../typings';
import { getAuthSignByBusinessId } from '@/utils';
import { MENU_BUSINESS_CLOUD_ACCOUNT } from '@/constants/menu-symbol';
import router from '@/router';
import { useRoute } from 'vue-router';

export interface IDataListProps {
  columns: ModelPropertyColumn[];
  list: IPermissionPolicyItem[];
  pagination: PaginationType;
  loading?: boolean;
}

const props = withDefaults(defineProps<IDataListProps>(), {
  loading: false,
});

// 定义事件
const emit = defineEmits<{
  'view-details': [row: IPermissionPolicyItem];
  'apply-to-account': [row: IPermissionPolicyItem];
  'edit-account': [row: IPermissionPolicyItem];
}>();

const { handlePageChange, handlePageSizeChange, handleSort } = usePage();

const { settings } = useTableSettings(props.columns);
const { isBusinessPage, getBizsId } = useWhereAmI();
const route = useRoute();

const bizId = computed(() => (isBusinessPage ? getBizsId() : 0));

// 查看详情
const handleViewDetails = (row: IPermissionPolicyItem) => {
  emit('view-details', row);
};

// 应用到二级账号
const handleApplyToAccount = (row: IPermissionPolicyItem) => {
  emit('apply-to-account', row);
};

// 编辑
const handleEditAccount = (row: IPermissionPolicyItem) => {
  emit('edit-account', row);
};

// 跳转二级账号详情（新开标签页）
const handleGoToAccount = (id: string) => {
  router.push({
    name: MENU_BUSINESS_CLOUD_ACCOUNT,
    query: {
      ...route.query,
      type: 'secondary-account',
      id,
    },
  });
};

// 判断是否为需要自定义渲染的列（排除 related_account_count，它在 template 中单独处理）
const getColumnRender = (column: ModelPropertyColumn) => {
  if (column.id === 'name') {
    return ({ row }: { row: IPermissionPolicyItem }) =>
      h(
        Button,
        {
          text: true,
          theme: 'primary',
          onClick: () => handleViewDetails(row),
        },
        () => row.name || '--',
      );
  }
  return null;
};

// 判断列是否为关联二级账号数
const isRelatedAccountColumn = (column: ModelPropertyColumn) => column.id === 'associated_account_count';
</script>

<template>
  <bk-loading :loading="loading">
    <bk-table
      row-hover="auto"
      :data="list"
      :pagination="pagination"
      :max-height="`calc(100vh - 400px)`"
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
          <!-- 关联二级账号数 - hover 弹出账号列表 -->
          <template v-if="isRelatedAccountColumn(column)">
            <bk-popover
              v-if="row.associated_account_count > 0"
              theme="light"
              trigger="hover"
              placement="right"
              :popover-delay="[200, 150]"
              :max-height="240"
              :arrow="true"
              ext-cls="related-account-popover"
            >
              <span class="related-count-link">{{ row.associated_account_count ?? 0 }}</span>
              <template #content>
                <div class="related-account-list">
                  <div v-for="account in row.related_accounts" :key="account" class="related-account-item">
                    <span class="account-id">{{ account }}</span>
                    <i class="hcm-icon bkhcm-icon-jump-fill account-link-icon" @click="handleGoToAccount(account)" />
                  </div>
                </div>
              </template>
            </bk-popover>
            <span v-else class="related-count-zero">{{ row.associated_account_count ?? 0 }}</span>
          </template>
          <!-- 其他自定义渲染列 -->
          <template v-else-if="getColumnRender(column)">
            <component :is="() => getColumnRender(column)({ row })" />
          </template>
          <!-- 默认渲染 -->
          <template v-else>
            <display-value :property="column" :value="row[column.id]" :display="column?.meta?.display" />
          </template>
        </template>
      </bk-table-column>
      <bk-table-column label="操作" width="180" fixed="right">
        <template #default="{ row }">
          <hcm-auth
            :sign="
              getAuthSignByBusinessId(
                bizId,
                AUTH_UPDATE_PERMISSION_POLICY_LIBRARY,
                AUTH_BIZ_UPDATE_PERMISSION_POLICY_LIBRARY,
              )
            "
            v-slot="{ noPerm }"
          >
            <bk-button theme="primary" text :disabled="noPerm" @click="handleEditAccount(row)" v-if="!isBusinessPage">
              编辑
            </bk-button>
          </hcm-auth>
          <hcm-auth
            :sign="
              getAuthSignByBusinessId(
                bizId,
                AUTH_APPLY_PERMISSION_POLICY_LIBRARY,
                AUTH_BIZ_APPLY_PERMISSION_POLICY_LIBRARY,
              )
            "
            v-slot="{ noPerm }"
          >
            <bk-button theme="primary" :disabled="noPerm" text @click="handleApplyToAccount(row)">
              应用到二级账号
            </bk-button>
          </hcm-auth>
        </template>
      </bk-table-column>
    </bk-table>
  </bk-loading>
</template>

<style lang="scss" scoped>
.related-count-link {
  color: #3a84ff;
  cursor: pointer;
}

.related-count-zero {
  color: #63656e;
}
</style>

<style lang="scss">
.related-account-popover {
  padding: 8px !important;

  .related-account-list {
    max-height: 220px;
    overflow-y: auto;

    // padding: 4px 0;

    .related-account-item {
      display: flex;
      align-items: center;
      padding: 0 12px;
      height: 36px;
      line-height: 20px;
      cursor: pointer;
      transition: background-color 0.15s;

      &:hover {
        background-color: #f0f1f5;
      }

      .account-id {
        font-size: 12px;
        color: #4d4f56;
        margin-right: 8px;
      }

      .account-link-icon {
        font-size: 16px;
        color: #3a84ff;
        flex-shrink: 0;
        font-weight: 400 !important;
      }
    }
  }
}
</style>
