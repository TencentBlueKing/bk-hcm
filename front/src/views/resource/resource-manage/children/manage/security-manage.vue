<script setup lang="ts">
import type {
  PlainObject,
} from '@/typings/resource';
import {
  Button,
  InfoBox } from 'bkui-vue';

import {
  ref,
  h,
} from 'vue';

import {
  useI18n,
} from 'vue-i18n';
import {
  useRouter,
} from 'vue-router';
import useBusiness from '../../hooks/use-business';

// use hooks
const {
  t,
} = useI18n();

const router = useRouter();

const {
  isShowDistribution,
  handleDistribution,
  ResourceBusiness,
} = useBusiness();

const groupColumns = [
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
                type: 'security',
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
              onClick(cell) {
                console.log('111', cell);
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
                    Button, {
                      text: true,
                      theme: 'primary',
                      onClick(cell) {
                        console.log('111', cell);
                      },
                    },
                    [
                      '配置规则',
                    ],
                  );
                }
                InfoBox({
                  title: '确认删除',
                  subTitle: h(
                    Button, {
                      text: true,
                      theme: 'primary',
                      onClick(cell) {
                        console.log('111', cell);
                      },
                    },
                    [
                      '配置规则',
                    ],
                  ),
                  onConfirm() { },
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
</script>

<template>
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
    :columns="groupColumns"
    :data="tableData"
    @column-sort="handleSortBy"
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
    :title="t('安全组分配')"
  />
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
