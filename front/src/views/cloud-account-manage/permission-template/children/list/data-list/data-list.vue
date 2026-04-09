<script setup lang="ts">
import { inject, reactive, type Ref, ref } from 'vue';
import { Share } from 'bkui-vue/lib/icon';
import { PaginationType } from '@/typings';
import { ModelPropertyColumn } from '@/model/typings';
import {
  type IPermissionTemplateItem,
  usePermissionTemplateStore,
} from '@/store/cloud-account-manage/permission-template';
import usePage from '@/hooks/use-page';
import useTableSettings from '@/hooks/use-table-settings';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import { VendorEnum } from '@/common/constant';

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

const subAccountCache = reactive<Record<string, string[]>>({});
const subAccountLoading = reactive<Record<string, boolean>>({});

const afterAssociatedSubAccountPopoverShow = async (_$event: any, row: IPermissionTemplateItem) => {
  if (subAccountCache[row.id]) return;

  subAccountLoading[row.id] = true;
  try {
    const ids = await permissionTemplateStore.getPermissionTemplateSubAccountIds(
      getBizsId(),
      currentVendor.value,
      row.id,
    );
    subAccountCache[row.id] = ids;
  } finally {
    subAccountLoading[row.id] = false;
  }
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
        <template v-else-if="column.id === 'associated_sub_account_count'">
          <bk-popover
            theme="light"
            component-event-delay="300"
            trigger="click"
            render-type="shown"
            placement="right"
            :popover-delay="[300, 0]"
            @after-show="($event: any) => afterAssociatedSubAccountPopoverShow($event, row)"
          >
            <bk-button theme="primary" text>{{ row.associated_sub_account_count || '--' }}</bk-button>
            <template #content>
              <bk-loading theme="primary" mode="spin" size="mini" :opacity="1" v-if="subAccountLoading[row.id]" />
              <ul v-else-if="subAccountCache[row.id]?.length" class="sub-account-list">
                <li v-for="id in subAccountCache[row.id]" :key="id" class="sub-account-item">
                  <span class="sub-account-id">{{ id }}</span>
                  <Share class="sub-account-link-icon" />
                </li>
              </ul>
              <div v-else class="sub-account-empty">未查询到关联三级账号</div>
            </template>
          </bk-popover>
        </template>
        <template v-else>
          <display-value :property="column" :value="row[column.id]" :display="column?.meta?.display" />
        </template>
      </template>
    </bk-table-column>
    <bk-table-column :show-overflow-tooltip="false" :label="'操作'">
      <template #default="{ row }">
        <div class="actions">
          <bk-button theme="primary" text @click="emit('edit', row)">编辑</bk-button>
          <bk-button theme="primary" text @click="emit('delete', row)">删除</bk-button>
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

.sub-account-list {
  margin: 0;
  padding: 0;
  list-style: none;
  width: 162px;
  height: 140px;
  overflow-y: auto;

  .sub-account-item {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 4px 8px;
    line-height: 28px;

    &:nth-child(even) {
      background: #fafbfd;
    }

    .sub-account-id {
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
    }

    .sub-account-link-icon {
      font-size: 12px;
      color: #3a84ff;
      cursor: pointer;
    }
  }
}

.sub-account-empty {
  padding: 8px 0;
  color: #c4c6cc;
  text-align: center;
}
</style>
