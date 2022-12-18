<script setup lang="ts">
import type {
  PlainObject,
  DoublePlainObject,
  FilterType,
} from '@/typings/resource';
import {
  Button,
  InfoBox,
  Message } from 'bkui-vue';

import {
  ref,
  h,
  PropType,
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
  datas,
  pagination,
  isLoading,
  handlePageChange,
  handlePageSizeChange,
  handleSort,
} = useQueryList(props, 'security_groups');
datas.value = [{ id: 333, vendor: 'tcloud' }];

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
    sort: true,
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
    sort: true,
  },
  {
    label: '描述',
    field: 'memo',
  },
  {
    label: '关联模板',
    field: '',
    sort: true,
  },
  {
    label: '修改时间',
    field: 'update_at',
  },
  {
    label: '创建时间',
    field: 'create_at',
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
                const haveAssResource = true;
                const subTitle: any = ref('请注意删除安全组后无法恢复，请谨慎操作');
                if (haveAssResource) {
                  subTitle.value = h(
                    'span', {},
                    [
                      h(
                        'span', {},
                        [
                          '安全组被实例关联或者被其他安全组规则关联时不能直接删除，请删除关联关系后再进行删除',
                        ],
                      ),
                      h(
                        Button, {
                          text: true,
                          theme: 'primary',
                          onClick(cell) {
                            console.log('111', cell);
                          },
                        },
                        [
                          '查看这条配置规则',
                        ],
                      ),
                    ],
                  );
                }
                InfoBox({
                  title: '确认删除',
                  subTitle: subTitle.value,
                  onConfirm() {
                    return resourceStore
                      .delete('security_groups', data.id)
                      .then(() => {
                        Message({
                          theme: 'success',
                          message: '删除成功',
                        });
                      });
                  },
                  headerAlign: 'left',
                  footerAlign: 'left',
                  contentAlign: 'left',
                });
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
    label: '实例 ID',
    field: '',
    sort: true,
  },
  {
    label: '名称',
    field: '',
    sort: true,
  },
  {
    label: '云厂商',
    field: '',
    sort: true,
  },
  {
    label: 'IP',
    field: '',
    sort: true,
  },
  {
    label: '云区域',
    field: '',
  },
  {
    label: '地域',
    field: '',
    sort: true,
  },
  {
    label: 'VPC',
    field: '',
    sort: true,
  },
  {
    label: '子网',
    field: '',
    sort: true,
  },
  {
    label: '状态',
    field: '',
  },
  {
    label: '创建时间',
    field: '',
  },
  {
    label: '操作',
    field: '',
  },
];
const tableData: any[] = [{}];
const types = [
  { name: 'group', label: '安全组' },
  { name: 'gcp', label: 'GCP防火墙规则' },
];
const activeType = ref('group');

// 方法
const handleSortBy = () => {

};

const handleConfirm = () => {
  console.log(11111);
  return resourceStore
    .assignBusiness('security_groups')
    .then(() => {
      Message({
        theme: 'success',
        message: '删除成功',
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
      :data="datas"
      @page-limit-change="handlePageSizeChange"
      @page-value-change="handlePageChange"
      @column-sort="handleSort"
    />

    <bk-table
      v-if="activeType === 'gcp'"
      class="mt20"
      row-hover="auto"
      :columns="gcpColumns"
      :data="tableData"
      @column-sort="handleSortBy"
    />

    <resource-business
      v-model:is-show="isShowDistribution"
      @handle-confirm="handleConfirm"
      :title="t('安全组分配')"
    />
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
