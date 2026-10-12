<script setup lang="ts">
import { h, reactive, ref, watchEffect } from 'vue';
import { useI18n } from 'vue-i18n';
import { useAccountStore } from '@/store';
import { useBusinessMapStore } from '@/store/useBusinessMap';
import useTableSettings from '@/hooks/use-table-settings';
import usePage from '@/hooks/use-page';
import routerAction from '@/router/utils/action';
import { MENU_RESOURCE_ACCOUNT_DETAIL } from '@/constants/menu-symbol';
import { AUTH_DELETE_ACCOUNT, AUTH_UPDATE_ACCOUNT } from '@/constants/auth-symbols';

import { ACCOUNT_TYPES, SITE_TYPE_MAP, SITE_TYPES, VendorMap, VENDORS } from '@/common/constant';
import type { ModelPropertyColumn } from '@/model/typings';
import type { FilterType, IAccountItem, IListResData } from '@/typings';
import { timeFormatter } from '@/common/util';

import { Button, Message } from 'bkui-vue';

interface IState {
  loading: boolean;
  dataList: IAccountItem[];
  isAccurate: boolean;
  filter: FilterType;
}

const { t } = useI18n();
const accountStore = useAccountStore();
const { getNameFromBusinessMap } = useBusinessMapStore();

// TODO(菜单整合三期)：迁至工作台后按 PR 不展示资源账号（账号类型去掉资源账号、请求追加 type neq resource）
const searchData = [
  { name: t('名称'), id: 'name' },
  { name: t('账号类型'), id: 'type', children: ACCOUNT_TYPES },
  { name: t('云厂商'), id: 'vendor', children: VENDORS },
  { name: t('站点类型'), id: 'site', children: SITE_TYPES },
  { name: t('负责人'), id: 'managers' },
  { name: t('创建人'), id: 'creator' },
  { name: t('修改人'), id: 'reviser' },
];
const searchVal = ref([]);

const columns = reactive([
  { id: 'id', name: 'ID', type: 'number', width: 80 },
  {
    id: 'name',
    name: t('名称'),
    type: 'string',
    render: ({ row, cell }: any) =>
      h(
        Button,
        {
          text: true,
          theme: 'primary',
          onClick() {
            handleJump(row.id);
          },
        },
        cell,
      ),
    width: 120,
  },
  {
    id: 'type',
    name: t('账号类型'),
    type: 'enum',
    option: ACCOUNT_TYPES.reduce<{ [key: string]: string }>((acc, cur) => ({ ...acc, [cur.id]: cur.name }), {}),
    width: 90,
  },
  { id: 'vendor', name: t('云厂商'), type: 'enum', option: VendorMap, width: 90 },
  { id: 'site', name: t('站点类型'), type: 'enum', option: SITE_TYPE_MAP, width: 90 },
  {
    id: 'usage_biz_ids',
    name: t('所属业务'),
    type: 'array',
    render: ({ cell }: { cell: number[] }) =>
      cell?.map((v: number) => (v === -1 ? '全部业务' : getNameFromBusinessMap(v)))?.join(',') ?? '--',
    width: 120,
    filter: { list: [] },
  },
  {
    id: 'bk_biz_id',
    name: t('管理业务'),
    type: 'string',
    width: 120,
    render: ({ cell }: { cell: number }) => (!cell ? '--' : getNameFromBusinessMap(cell)),
  },
  { id: 'managers', name: t('负责人'), type: 'user', width: 120 },
  { id: 'creator', name: t('创建人'), type: 'user', width: 120, defaultHidden: true },
  { id: 'reviser', name: t('修改人'), type: 'user', width: 120 },
  {
    id: 'created_at',
    name: t('创建时间'),
    type: 'string',
    render: ({ cell }: { cell: string }) => timeFormatter(cell),
    width: 150,
  },
  {
    id: 'updated_at',
    name: t('修改时间'),
    type: 'string',
    render: ({ cell }: { cell: string }) => timeFormatter(cell),
    width: 150,
  },
  { id: 'memo', name: t('备注'), type: 'string', width: 100 },
]);
const { settings } = useTableSettings(columns as ModelPropertyColumn[]);
const { pagination, getDefaultPagination } = usePage();

const state = reactive<IState>({ loading: false, dataList: [], isAccurate: false, filter: { op: 'and', rules: [] } });
const dataList = ref([]);
const getAccountList = async () => {
  state.loading = true;
  try {
    const res: IListResData<IAccountItem[]> = await accountStore.getAccountList({ filter: state.filter });
    const datalist =
      res.data?.details?.sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime()) ?? [];
    state.dataList = datalist;

    // 给业务columns添加filter
    updateBkBizIdsFilter(datalist);
  } catch (error) {
    console.error(error);
    dataList.value = [];
    Object.assign(pagination, getDefaultPagination());
  } finally {
    state.loading = false;
  }
};

const updateBkBizIdsFilter = (datalist: IAccountItem[]) => {
  if (!datalist.length) return;
  const bkBizIdsColumn = columns.find((col) => col.id === 'bk_biz_ids');
  if (bkBizIdsColumn) {
    const uniqueBkBizIds = [...new Set(datalist.flatMap((item) => item.bk_biz_ids))];
    Object.assign(bkBizIdsColumn, {
      filter: {
        list: uniqueBkBizIds.map((v) => {
          const name = getNameFromBusinessMap(v);
          // TODO：组件库2.0.1-beta.34只匹配了label、value，没有匹配text
          return { text: name, value: v, label: name };
        }),
        filterFn: (checked: number[], row: IAccountItem) => checked.some((v) => row.bk_biz_ids?.includes(v)),
      },
    });
  }
};

// 名称与「编辑」都进账号详情（基本信息可编辑）
const handleJump = (id: string) => {
  routerAction.redirect({ name: MENU_RESOURCE_ACCOUNT_DETAIL, params: { accountId: id } }, { history: true });
};

// 删除
const deleteOptions = reactive({
  isDialogShow: false,
  loading: false,
  id: '',
});
const handleDelete = async (id: string) => {
  await accountStore.accountDeleteValidate(id);
  deleteOptions.isDialogShow = true;
  deleteOptions.id = id;
};
const handleDeleteConfirm = async () => {
  deleteOptions.loading = true;
  try {
    await accountStore.accountDelete(deleteOptions.id);
    Message({ message: t('删除成功'), theme: 'success' });
    getAccountList();
  } catch (error) {
    console.error(error);
  } finally {
    deleteOptions.loading = false;
  }
};

watchEffect(() => {
  state.filter.rules = searchVal.value.reduce((p, v) => {
    if (v.type === 'condition') {
      state.filter.op = v.id || 'and';
    } else {
      if (v.id === 'managers') {
        p.push({ field: v.id, op: 'json_contains', value: v.values[0].id });
      } else {
        p.push({ field: v.id, op: state.isAccurate ? 'eq' : 'cs', value: v.values[0].id });
      }
    }
    return p;
  }, []);

  getAccountList();
});
</script>

<template>
  <div class="account-manage-page">
    <div class="account-manage-content">
      <!-- search -->
      <div class="tools">
        <bk-checkbox v-model="state.isAccurate">{{ t('精确') }}</bk-checkbox>
        <bk-search-select v-model="searchVal" :data="searchData" />
      </div>
      <!-- table -->
      <div class="table-wrap">
        <bk-table
          max-height="100%"
          :data="state.dataList"
          row-key="id"
          row-hover="auto"
          show-overflow-tooltip
          :pagination="pagination"
          :settings="settings"
          v-bkloading="{ loading: state.loading }"
        >
          <bk-table-column
            v-for="column in columns"
            :key="column.id"
            :label="column.name"
            :prop="column.id"
            :render="column.render"
            :width="column.width"
            :filter="column.filter"
          >
            <template #default="{ row }">
              <display-value :property="column" :value="row[column.id]" :vendor="row?.vendor" />
            </template>
          </bk-table-column>
          <bk-table-column :label="t('操作')" fixed="right" width="120">
            <template #default="{ row }">
              <hcm-auth :sign="{ type: AUTH_UPDATE_ACCOUNT, relation: [row.id] }" tag="span" v-slot="{ noPerm }">
                <bk-button text theme="primary" :disabled="noPerm" @click="handleJump(row.id)">
                  {{ t('编辑') }}
                </bk-button>
              </hcm-auth>
              <hcm-auth :sign="{ type: AUTH_DELETE_ACCOUNT, relation: [row.id] }" tag="span" v-slot="{ noPerm }">
                <bk-button class="ml8" theme="primary" text :disabled="noPerm" @click="handleDelete(row.id)">
                  {{ t('删除') }}
                </bk-button>
              </hcm-auth>
            </template>
          </bk-table-column>
        </bk-table>
      </div>
    </div>
    <!-- dialog -->
    <bk-dialog
      v-model:is-show="deleteOptions.isDialogShow"
      :title="t('确认删除')"
      :is-loading="deleteOptions.loading"
      @confirm="handleDeleteConfirm"
    >
      {{ t('删除之后无法恢复账户信息') }}
    </bk-dialog>
  </div>
</template>

<style scoped lang="scss">
.account-manage-page {
  height: 100%;
  padding: 24px;

  .account-manage-content {
    height: 100%;
    padding: 16px 24px;
    background: #fff;
  }

  .tools {
    margin-bottom: 16px;
    display: flex;
    justify-content: flex-end;
    align-items: center;

    :deep(.bk-search-select) {
      margin-left: 24px;
      width: 300px;
    }
  }

  .table-wrap {
    height: calc(100% - 48px);
  }
}
</style>
