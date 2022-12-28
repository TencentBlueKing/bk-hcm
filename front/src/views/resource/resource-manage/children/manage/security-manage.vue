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
import useDeleteSecurity from '../../hooks/use-delete-security';
import useQueryList from '../../hooks/use-query-list';
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

const {
  isShowDistribution,
  handleDistribution,
  ResourceBusiness,
} = useBusiness();

const {
  isShowSecurity,
  handleShowDeleteSecurity,
  DeleteSecurity,
} = useDeleteSecurity();

const tableData = ref<any>([]);

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


const {
  datas,
  pagination,
  isLoading,
  handlePageChange,
  handlePageSizeChange,
  handleSort,
} = useQueryList(props, 'security_groups');
datas.value = [{ id: 333, vendor: 'tcloud' }];

const activeType = ref('group');

// 状态保持
watch(
  () => activeType.value,
  (v) => {
    if (v === 'group') {
      const { datas } = fetchList('security_groups');
      tableData.value = [{ id: 333, vendor: 'tcloud' }] || datas;
    } else if (v === 'gcp') {
      const { datas } = fetchList('vendors/gcp/firewalls/rules');
      tableData.value = [{ id: 333, vendor: 'tcloud' }] || datas;
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
    render({ data }: DoublePlainObject) {
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
    render({ data }: DoublePlainObject) {
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
                handleShowDeleteSecurity();
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
    render({ data }: DoublePlainObject) {
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
    render({ data }: DoublePlainObject) {
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
                handleShowDeleteSecurity();
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
    :loading="isLoading"
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
        @click="handleShowDeleteSecurity"
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
      :pagination="pagination"
      :columns="groupColumns"
      :data="tableData"
      @page-limit-change="handlePageSizeChange"
      @page-value-change="handlePageChange"
      @column-sort="handleSort"
      @selection-change="handleSelection"
    />

    <bk-table
      v-if="activeType === 'gcp'"
      class="mt20"
      row-hover="auto"
      :pagination="pagination"
      :columns="gcpColumns"
      :data="tableData"
      @page-limit-change="handlePageSizeChange"
      @page-value-change="handlePageChange"
      @column-sort="handleSort"
      @selection-change="handleSelection"
    />

    <resource-business
      v-model:is-show="isShowDistribution"
      @handle-confirm="handleConfirm"
      :title="t('安全组分配')"
    />

    <delete-security
      v-model:is-show="isShowSecurity"
    ></delete-security>
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
