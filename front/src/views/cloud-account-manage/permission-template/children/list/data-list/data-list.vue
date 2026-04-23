<script setup lang="ts">
import { inject, type Ref, ref, computed } from 'vue';
import { PaginationType } from '@/typings';
import { ModelPropertyColumn } from '@/model/typings';
import {
  type IPermissionTemplateItem,
  usePermissionTemplateStore,
} from '@/store/cloud-account-manage/permission-template';
import usePage from '@/hooks/use-page';
import useTableSettings from '@/hooks/use-table-settings';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import { SecondaryAccountResourceTypeEnum, VendorEnum } from '@/common/constant';
import { AUTH_UPDATE_PERMISSION_TEMPLATE, AUTH_DELETE_PERMISSION_TEMPLATE } from '@/constants/auth-symbols';
import { getTypeData } from '@/views/cloud-account-manage/permission-template/utils';
import SecondaryAccountValue from '@/views/cloud-account-manage/components/secondary-account-value.vue';
import routeAction from '@/router/utils/action';
import { MENU_BUSINESS_CLOUD_ACCOUNT } from '@/constants/menu-symbol';
import type { LinkPopoverItem } from '@/components/display-value/appearance/link-popover.vue';

export interface IDataListProps {
  columns: ModelPropertyColumn[];
  list: IPermissionTemplateItem[];
  pagination: PaginationType;
}

const props = withDefaults(defineProps<IDataListProps>(), {});

const emit = defineEmits<{
  'view-details': [row: IPermissionTemplateItem];
  delete: [row: IPermissionTemplateItem];
  edit: [row: IPermissionTemplateItem];
}>();

const permissionTemplateStore = usePermissionTemplateStore();
const currentVendor = inject<Ref<VendorEnum>>('currentVendor', ref(VendorEnum.TCLOUD));
const { getBizsId } = useWhereAmI();

const { handlePageChange, handlePageSizeChange, handleSort } = usePage();

const { settings } = useTableSettings(props.columns);

const bizId = computed(() => getBizsId());

const getSubAccountLoadFn = (row: IPermissionTemplateItem) => async (): Promise<LinkPopoverItem[]> => {
  const sub_accounts = await permissionTemplateStore.getPermissionTemplateSubAccountIds(
    bizId.value,
    currentVendor.value,
    row.id,
  );
  return sub_accounts.map(({ id, cloud_id }) => ({ id, label: cloud_id }));
};

const handleGoToTertiaryAccount = (item: LinkPopoverItem) => {
  routeAction.open({
    name: MENU_BUSINESS_CLOUD_ACCOUNT,
    query: { type: 'tertiary-account', id: item.id as string },
  });
};
</script>

<template>
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
      :render="column.render"
    >
      <template #default="{ row }">
        <template v-if="column.id === 'name'">
          <bk-button theme="primary" text @click="emit('view-details', row)">{{ row.name || '--' }}</bk-button>
        </template>
        <template v-else-if="column.id === 'cloud_account_id'">
          <SecondaryAccountValue
            :value="row.cloud_account_id"
            :biz-id="bizId"
            :vendor="currentVendor"
            :res-type="SecondaryAccountResourceTypeEnum.TEMPLATE"
          />
        </template>
        <template v-else-if="column.id === 'associated_sub_account_count'">
          <display-value
            :property="column"
            :value="row.associated_sub_account_count"
            :display="{
              appearance: 'link-popover',
              appearanceProps: {
                loadFn: getSubAccountLoadFn(row),
                onLinkClick: handleGoToTertiaryAccount,
                emptyText: '未查询到关联三级账号',
              },
            }"
          />
        </template>
        <template v-else>
          <display-value :property="column" :value="row[column.id]" :display="column?.meta?.display" />
        </template>
      </template>
    </bk-table-column>
    <bk-table-column :show-overflow-tooltip="false" :label="'操作'">
      <template #default="{ row }">
        <div class="actions">
          <hcm-auth :sign="{ type: AUTH_UPDATE_PERMISSION_TEMPLATE, relation: [bizId] }" v-slot="{ noPerm }">
            <bk-button
              theme="primary"
              text
              :disabled="noPerm || getTypeData(row).isCloudCustom"
              @click="emit('edit', row)"
              v-bk-tooltips="{ content: '仅云自定义模板可编辑', disabled: !getTypeData(row).isCloudCustom }"
            >
              编辑
            </bk-button>
          </hcm-auth>
          <hcm-auth :sign="{ type: AUTH_DELETE_PERMISSION_TEMPLATE, relation: [bizId] }" v-slot="{ noPerm }">
            <bk-button
              theme="primary"
              text
              :disabled="noPerm || getTypeData(row).isCloudCustom || row.associated_sub_account_count > 0"
              @click="emit('delete', row)"
              v-bk-tooltips="{
                content: getTypeData(row).isCloudCustom ? '仅云自定义模板可删除' : '有三级账号关联不可删除',
                disabled: !(getTypeData(row).isCloudCustom || row.associated_sub_account_count > 0),
              }"
            >
              删除
            </bk-button>
          </hcm-auth>
        </div>
      </template>
    </bk-table-column>
  </bk-table>
</template>

<style lang="scss" scoped>
.actions {
  display: flex;
  gap: 12px;
}
</style>
