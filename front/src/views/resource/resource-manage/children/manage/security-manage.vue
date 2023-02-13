<script setup lang="ts">
import type {
  // PlainObject,
  DoublePlainObject,
  FilterType,
} from '@/typings/resource';
import { GcpTypeEnum, CloudType } from '@/typings';
import {
  Button,
  Message } from 'bkui-vue';

import {
  ref,
  h,
  PropType,
  watch,
  reactive,
} from 'vue';

import {
  useI18n,
} from 'vue-i18n';
import {
  useRouter,
} from 'vue-router';
import {
  useResourceStore,
} from '@/store/resource';
import useBusiness from '../../hooks/use-business';
import useQueryList from '../../hooks/use-query-list';
import useColumns from '../../hooks/use-columns';
import useDelete from '../../hooks/use-delete';
import useSelection from '../../hooks/use-selection';

const props = defineProps({
  filter: {
    type: Object as PropType<FilterType>,
  },
});

// use hooks
const {
  t,
} = useI18n();

const router = useRouter();

const resourceStore = useResourceStore();

const activeType = ref('group');

const state = reactive<any>({
  datas: [],
  pagination: {
    current: 1,
    limit: 10,
    count: 0,
  },
  isLoading: true,
  handlePageChange: () => {},
  handlePageSizeChange: () => {},
  handleSort: () => {},
  columns: useColumns('group'),
});

let securityHandleShowDelete: any;
let SecurityDeleteDialog: any;

const {
  isShowDistribution,
  handleDistribution,
  ResourceBusiness,
} = useBusiness();

const {
  selections,
  handleSelectionChange,
} = useSelection();


const fetchList = async (fetchType: string) => {
  console.log('fetchType', fetchType, props);
  const {
    datas,
    pagination,
    isLoading,
    handlePageChange,
    handlePageSizeChange,
    handleSort,
  } = await useQueryList(props, fetchType);
  return {
    datas,
    pagination,
    isLoading,
    handlePageChange,
    handlePageSizeChange,
    handleSort,
  };
};

const showDeleteDialog = (fetchType: string, title: string) => {
  const {
    handleShowDelete,
    DeleteDialog,
  } = useDelete(
    state.columns,
    selections,
    fetchType,
    t(title),
    true,
  );
  return {
    handleShowDelete,
    DeleteDialog,
  };
};

// 状态保持
watch(
  () => activeType.value,
  (v) => {
    selections.value = [];
    state.isLoading = true;
    handleSwtichType(v);
  },
);

const handleSwtichType = async (type: string) => {
  const params = {
    fetchUrl: 'security_groups',
    columns: 'group',
    dialogName: t('删除安全组'),
  };
  if (type === 'gcp') {
    params.fetchUrl = 'vendors/gcp/firewalls/rules';
    params.columns = 'gcp';
    params.dialogName = t('删除防火墙规则');
  }
  // eslint-disable-next-line max-len
  const { datas, pagination, isLoading, handlePageChange, handlePageSizeChange, handleSort } = await fetchList(params.fetchUrl);
  state.datas = datas;
  state.isLoading = isLoading;
  state.pagination = pagination;
  state.handlePageChange = handlePageChange;
  state.handlePageSizeChange = handlePageSizeChange;
  state.handleSort = handleSort;
  state.columns = useColumns(params.columns);
  const { handleShowDelete, DeleteDialog } = showDeleteDialog(params.fetchUrl, params.dialogName);
  securityHandleShowDelete = handleShowDelete;
  SecurityDeleteDialog = DeleteDialog;
};

handleSwtichType(activeType.value);

const groupColumns = [
  {
    type: 'selection',
    width: 100,
  },
  {
    label: 'ID',
    field: 'id',
    sort: true,
    render({ data }: any) {
      return h(
        Button,
        {
          text: true,
          theme: 'primary',
          disabled: data.bk_biz_id !== -1,
          onClick() {
            router.push({
              name: 'resourceDetail',
              params: {
                type: 'security',
              },
              query: {
                id: data.id,
                vendor: data.vendor,
              },
            });
          },
        },
        [
          data.id || '--',
        ],
      );
    },
  },
  {
    label: t('账号 ID'),
    field: 'account_id',
    sort: true,
  },
  {
    label: t('云账号 ID'),
    field: 'cloud_id',
    sort: true,
  },
  {
    label: t('名称'),
    field: 'name',
    sort: true,
  },
  {
    label: t('云厂商'),
    render({ data }: any) {
      return h(
        'span',
        {},
        [
          CloudType[data.vendor],
        ],
      );
    },
  },
  {
    label: t('地域'),
    field: 'region',
  },
  {
    label: t('描述'),
    field: 'memo',
  },
  // {
  //   label: t('关联模板'),
  //   field: '',
  // },
  {
    label: t('修改时间'),
    field: 'updated_at',
    sort: true,
  },
  {
    label: t('创建时间'),
    field: 'created_at',
    sort: true,
  },
  {
    label: t('操作'),
    field: '',
    render({ data }: any) {
      return h(
        'span',
        {},
        [
          h(
            Button,
            {
              text: true,
              disabled: data.bk_biz_id !== -1,
              theme: 'primary',
              onClick() {
                router.push({
                  name: 'resourceDetail',
                  params: {
                    type: 'security',
                  },
                  query: {
                    activeTab: 'rule',
                    id: data.id,
                    vendor: data.vendor,
                  },
                });
              },
            },
            [
              t('配置规则'),
            ],
          ),
          h(
            Button,
            {
              class: 'ml10',
              disabled: data.bk_biz_id !== -1,
              text: true,
              theme: 'primary',
              onClick() {
                securityHandleShowDelete([data.id]);
              },
            },
            [
              t('删除'),
            ],
          ),
        ],
      );
    },
  },
];
const gcpColumns = [
  {
    type: 'selection',
  },
  {
    label: 'ID',
    field: 'id',
    sort: true,
    render({ data }: any) {
      return h(
        Button,
        {
          text: true,
          theme: 'primary',
          disabled: data.bk_biz_id !== -1,
          onClick() {
            router.push({
              name: 'resourceDetail',
              params: {
                type: 'gcp',
              },
              query: {
                id: data.id,
              },
            });
          },
        },
        [
          data.id || '--',
        ],
      );
    },
  },
  {
    label: t('资源 ID'),
    field: 'account_id',
    sort: true,
  },
  {
    label: t('名称'),
    field: 'name',
    sort: true,
  },
  {
    label: t('云厂商'),
    render() {
      return h(
        'span',
        {},
        [
          t('谷歌云'),
        ],
      );
    },
  },
  {
    label: 'VPC',
    field: 'vpc_id',
  },
  {
    label: t('类型'),
    render({ data }: any) {
      return h(
        'span',
        {},
        [
          GcpTypeEnum[data.type],
        ],
      );
    },
  },
  {
    label: t('目标'),
    render({ data }: any) {
      return h(
        'span',
        {},
        [
          data.target_tags || data.target_service_accounts || '--',
        ],
      );
    },
  },
  // {
  //   label: t('过滤条件'),
  //   field: '',
  // },
  {
    label: t('协议/端口'),
    render({ data }: any) {
      return h(
        'span',
        {},
        (data?.allowed || data?.denied) ? (data?.allowed || data?.denied).map((e: any) => {
          return h(
            'div',
            {},
            `${e.protocol}:${e.port}`,
          );
        }) : '--',
      );
    },
  },
  {
    label: t('优先级'),
    field: 'priority',
  },
  {
    label: t('修改时间'),
    field: 'updated_at',
    sort: true,
  },
  {
    label: t('创建时间'),
    field: 'created_at',
    sort: true,
  },
  {
    label: t('操作'),
    field: '',
    render({ data }: any) {
      return h(
        'span',
        {},
        [
          h(
            Button,
            {
              text: true,
              theme: 'primary',
              disabled: data.bk_biz_id !== -1,
              onClick() {
                router.push({
                  name: 'resourceDetail',
                  params: {
                    type: 'gcp',
                  },
                  query: {
                    id: data.id,
                  },
                });
              },
            },
            [
              t('编辑'),
            ],
          ),
          h(
            Button,
            {
              class: 'ml10',
              text: true,
              disabled: data.bk_biz_id !== -1,
              theme: 'primary',
              onClick() {
                securityHandleShowDelete([data.id]);
              },
            },
            [
              t('删除'),
            ],
          ),
        ],
      );
    },
  },
];
const types = [
  { name: 'group', label: t('安全组') },
  { name: 'gcp', label: t('GCP防火墙规则') },
];

// 方法

const handleConfirm = (bizId: number) => {
  const securityGroupIds = selections.value.map(e => e.id);
  const params = {
    security_group_ids: securityGroupIds,
    bk_biz_id: bizId,
  };
  return resourceStore
    .assignBusiness(activeType.value === 'group' ? 'security_groups' : 'vendors/gcp/firewalls/rules', params)
    .then(() => {
      Message({
        theme: 'success',
        message: t('分配成功'),
      });
    });
};

const isRowSelectEnable = ({ row }: DoublePlainObject) => {
  return row.bk_biz_id === -1;
};
</script>

<template>
  <bk-loading
    :loading="state.isLoading"
  >
    <section>
      <bk-button
        class="w100"
        theme="primary"
        @click="handleDistribution"
      >
        {{ t('分配') }}
      </bk-button>
      <bk-button
        class="w100 ml10"
        theme="primary"
        @click="securityHandleShowDelete(selections.map(e => e.id))"
      >
        {{ t('删除') }}
      </bk-button>
    </section>

    <bk-radio-group
      class="mt20"
      v-model="activeType"
      :disabled="state.isLoading"
    >
      <bk-radio-button
        v-for="item in types"
        :key="item.name"
        :label="item.name"
      >
        {{ item.label }}
      </bk-radio-button>
    </bk-radio-group>

    <bk-table
      v-if="activeType === 'group'"
      class="mt20"
      row-hover="auto"
      remote-pagination
      :pagination="state.pagination"
      :columns="groupColumns"
      :data="state.datas"
      :is-row-select-enable="isRowSelectEnable"
      @page-limit-change="state.handlePageSizeChange"
      @page-value-change="state.handlePageChange"
      @column-sort="state.handleSort"
      @selection-change="handleSelectionChange"
    />

    <bk-table
      v-if="activeType === 'gcp'"
      class="mt20"
      row-hover="auto"
      remote-pagination
      :pagination="state.pagination"
      :columns="gcpColumns"
      :data="state.datas"
      :is-row-select-enable="isRowSelectEnable"
      @page-limit-change="state.handlePageSizeChange"
      @page-value-change="state.handlePageChange"
      @column-sort="state.handleSort"
      @selection-change="handleSelectionChange"
    />

    <resource-business
      v-model:is-show="isShowDistribution"
      @handle-confirm="handleConfirm"
      :title="t(activeType === 'group' ? '安全组分配' : 'GCP防火墙分配')"
    />

    <security-delete-dialog>
      <h3 class="g-resource-tips" v-if="activeType === 'group'">
        {{ t('安全组被实例关联或者被其他安全组规则关联时不能直接删除，请删除关联关系后再进行删除') }}
        <bk-button text theme="primary">{{ t('查看关联实例') }}</bk-button>
      </h3>
      <h3 class="g-resource-tips" v-else>
        {{ t('防火墙规则被实例关联') }}<bk-button text theme="primary">{{ t('查看关联实例') }}</bk-button>
        {{ t('请注意删除防火墙规则后无法恢复，请谨慎操作') }}
      </h3>
    </security-delete-dialog>
  </bk-loading>
</template>

<style lang="scss" scoped>
.w100 {
  width: 100px;
}
.w60 {
  width: 60px;
}
.mt20 {
  margin-top: 20px;
}
</style>
