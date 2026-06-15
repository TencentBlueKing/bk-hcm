<script setup lang="ts">
import { computed, h, inject, ref, type Ref } from 'vue';
import { PaginationType } from '@/typings';
import { ModelPropertyColumn } from '@/model/typings';
import usePage from '@/hooks/use-page';
import useTableSettings from '@/hooks/use-table-settings';
import { Button } from 'bkui-vue';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import { VendorEnum, SecondaryAccountResourceTypeEnum } from '@/common/constant';
import SecondaryAccountValue from '@/views/cloud-account-manage/components/secondary-account-value.vue';
import {
  AUTH_UPDATE_PERMISSION_POLICY_LIBRARY,
  AUTH_APPLY_PERMISSION_POLICY_LIBRARY,
  AUTH_BIZ_UPDATE_PERMISSION_POLICY_LIBRARY,
  AUTH_BIZ_APPLY_PERMISSION_POLICY_LIBRARY,
} from '@/constants/auth-symbols';
import type { IPermissionPolicyItem } from '../../typings';
import { getAuthSignByBusinessId } from '@/utils';
import { MENU_BUSINESS_CLOUD_ACCOUNT } from '@/constants/menu-symbol';
import type { LinkPopoverItem } from '@/components/display-value/appearance/link-popover.vue';
import routeAction from '@/router/utils/action';

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

const currentVendor = inject<Ref<VendorEnum>>('currentVendor', ref(VendorEnum.TCLOUD));

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
const handleGoToAccount = (item: LinkPopoverItem) => {
  routeAction.open({
    name: MENU_BUSINESS_CLOUD_ACCOUNT,
    query: { type: 'secondary-account', id: item.id },
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
          <template v-if="column.id === 'associated_account_count'">
            <display-value
              :property="column"
              :value="row.associated_account_count"
              :display="{
                appearance: 'link-popover',
                appearanceProps: {
                  onLinkClick: handleGoToAccount,
                  emptyText: '未查询到关联二级账号',
                  list: row?.related_accounts?.map((id: string) => ({ id, label: id })),
                },
              }"
            >
              <template #item-label="{ item }">
                <SecondaryAccountValue
                  :value="item.id"
                  :vendor="currentVendor"
                  :res-type="SecondaryAccountResourceTypeEnum.PERMISSION"
                  :biz-id="bizId"
                  :label-formatter="(item) => item?.extension?.cloud_main_account_id"
                />
              </template>
            </display-value>
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
          <div class="actions">
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
          </div>
        </template>
      </bk-table-column>
    </bk-table>
  </bk-loading>
</template>

<style lang="scss" scoped>
.actions {
  display: flex;
  gap: 12px;
}
</style>
