<script setup lang="ts">
import type {
  PlainObject,
  DoublePlainObject,
  FilterType,
} from '@/typings/resource';
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
import { CloudType } from '@/typings';

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
  isLoading: false,
  handlePageChange: () => {},
  handlePageSizeChange: () => {},
  handleSort: () => {},
  columns: useColumns('security'),
});

let securityHandleShowDelete: any;
let SecurityDeleteDialog: any;

const {
  isShowDistribution,
  handleDistribution,
  ResourceBusiness,
} = useBusiness();


const fetchList = (fetchType: string) => {
  const {
    datas,
    pagination,
    isLoading,
    handlePageChange,
    handlePageSizeChange,
    handleSort,
  } = useQueryList(props, fetchType);
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
    [],
    fetchType,
    t(title),
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
    if (v === 'group') {
      const { datas, pagination, isLoading, handlePageChange, handlePageSizeChange, handleSort } = fetchList('security_groups');
      state.datas = [{ id: 333, vendor: 'tcloud' }] || datas;
      state.isLoading = isLoading;
      state.pagination = pagination;
      state.handlePageChange = handlePageChange;
      state.handlePageSizeChange = handlePageSizeChange;
      state.handleSort = handleSort;
      state.columns = useColumns('security');
      const { handleShowDelete, DeleteDialog } = showDeleteDialog('security_groups', '删除安全组');
      securityHandleShowDelete = handleShowDelete;
      SecurityDeleteDialog = DeleteDialog;
    } else if (v === 'gcp') {
      const { datas, pagination, isLoading, handlePageChange, handlePageSizeChange, handleSort } = fetchList('vendors/gcp/firewalls/rules');
      state.datas = [{ id: 333, vendor: 'tcloud' }] || datas;
      state.isLoading = isLoading;
      state.pagination = pagination;
      state.handlePageChange = handlePageChange;
      state.handlePageSizeChange = handlePageSizeChange;
      state.handleSort = handleSort;
      state.columns = useColumns('gcp');
      const { handleShowDelete, DeleteDialog } = showDeleteDialog('cloud/vendors/gcp/firewalls/rules', '删除防火墙规则');
      securityHandleShowDelete = handleShowDelete;
      SecurityDeleteDialog = DeleteDialog;
    }
  },
);

const handleSelection = () => {};

const groupColumns = [
  {
    type: 'selection',
  },
  {
    label: 'ID',
    field: 'id',
    sort: true,
    render({ data }: DoublePlainObject) {
      return h(
        'span',
        {
          onClick() {
            router.push({
              name: 'resourceDetail',
              params: {
                type: 'security',
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
    label: '资源 ID',
    field: 'account_id',
    sort: true,
  },
  {
    label: '名称',
    field: 'name',
    sort: true,
  },
  {
    label: '云厂商',
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
    label: '地域',
    field: 'region',
  },
  {
    label: '描述',
    field: 'memo',
  },
  {
    label: '关联模板',
    field: '',
  },
  {
    label: '修改时间',
    field: 'update_at',
    sort: true,
  },
  {
    label: '创建时间',
    field: 'create_at',
    sort: true,
  },
  {
    label: '操作',
    field: '',
    render() {
      return h(
        'span',
        {},
        [
          h(
            Button,
            {
              text: true,
              theme: 'primary',
              onClick() {
                router.push({
                  name: 'resourceDetail',
                  params: {
                    type: 'security',
                  },
                  query: {
                    activeTab: 'rule',
                  },
                });
              },
            },
            [
              '配置规则',
            ],
          ),
          h(
            Button,
            {
              class: 'ml10',
              text: true,
              theme: 'primary',
              onClick() {
                securityHandleShowDelete();
              },
            },
            [
              '删除',
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
    field: '',
    sort: true,
    render({ cell }: PlainObject) {
      return h(
        'span',
        {
          onClick() {
            router.push({
              name: 'resourceDetail',
              params: {
                type: 'gcp',
              },
            });
          },
        },
        [
          cell || '--',
        ],
      );
    },
  },
  {
    label: '资源 ID',
    field: 'account_id',
    sort: true,
  },
  {
    label: '名称',
    field: '',
    sort: true,
  },
  {
    label: '云厂商',
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
    label: 'VPC',
    field: '',
  },
  {
    label: '类型',
    field: '',
  },
  {
    label: '目标',
    field: '',
  },
  {
    label: '过滤条件',
    field: '',
  },
  {
    label: '协议/端口',
    field: '',
  },
  {
    label: '操作',
    field: '',
  },
  {
    label: '优先级',
    field: '',
  },
  {
    label: '修改时间',
    field: 'update_at',
    sort: true,
  },
  {
    label: '创建时间',
    field: 'create_at',
    sort: true,
  },
  {
    label: '操作',
    field: '',
    render() {
      return h(
        'span',
        {},
        [
          h(
            Button,
            {
              text: true,
              theme: 'primary',
              onClick() {
                router.push({
                  name: 'resourceDetail',
                  params: {
                    type: 'security',
                  },
                  query: {
                    activeTab: 'rule',
                  },
                });
              },
            },
            [
              '编辑',
            ],
          ),
          h(
            Button,
            {
              class: 'ml10',
              text: true,
              theme: 'primary',
              onClick() {
                securityHandleShowDelete();
              },
            },
            [
              '删除',
            ],
          ),
        ],
      );
    },
  },
];
const types = [
  { name: 'group', label: '安全组' },
  { name: 'gcp', label: 'GCP防火墙规则' },
];

// 方法

const handleConfirm = (bizId: number) => {
  const params = {
    security_group_ids: [1],
    bk_biz_id: bizId,
  };
  return resourceStore
    .assignBusiness('security_groups', params)
    .then(() => {
      Message({
        theme: 'success',
        message: '分配成功',
      });
    });
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
        @click="securityHandleShowDelete"
      >
        {{ t('删除') }}
      </bk-button>
    </section>

    <bk-radio-group
      class="mt20"
      v-model="activeType"
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
      :pagination="state.pagination"
      :columns="groupColumns"
      :data="state.datas"
      @page-limit-change="state.handlePageSizeChange"
      @page-value-change="state.handlePageChange"
      @column-sort="state.handleSort"
      @selection-change="handleSelection"
    />

    <bk-table
      v-if="activeType === 'gcp'"
      class="mt20"
      row-hover="auto"
      :pagination="state.pagination"
      :columns="gcpColumns"
      :data="state.datas"
      @page-limit-change="state.handlePageSizeChange"
      @page-value-change="state.handlePageChange"
      @column-sort="state.handleSort"
      @selection-change="handleSelection"
    />

    <resource-business
      v-model:is-show="isShowDistribution"
      @handle-confirm="handleConfirm"
      :title="t('安全组分配')"
    />

    <security-delete-dialog>
      <h3 class="g-resource-tips">
        {{ t('安全组被实例关联或者被其他安全组规则关联时不能直接删除，请删除关联关系后再进行删除') }}
        <bk-button text theme="primary">{{ t('查看关联实例') }}</bk-button><br />
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
