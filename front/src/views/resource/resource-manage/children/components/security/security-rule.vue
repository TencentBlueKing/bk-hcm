<script lang="ts" setup>
import {
  ref,
  watch,
  h,
  reactive,
  PropType,
} from 'vue';
import {
  useI18n,
} from 'vue-i18n';

import {
  Button,
  Message,
} from 'bkui-vue';

import {
  useResourceStore,
} from '@/store/resource';

import { SecurityRuleEnum } from '@/typings';

import {
  useRouter,
} from 'vue-router';

import UseSecurityRule from '@/views/resource/resource-manage/hooks/use-security-rule';
import useQueryList from '@/views/resource/resource-manage/hooks/use-query-list';
import useColumns from '@/views/resource/resource-manage/hooks/use-columns';

const props = defineProps({
  filter: {
    type: Object as PropType<any>,
  },
  id: {
    type: String as PropType<any>,
  },
  vendor: {
    type: String as PropType<any>,
  },
});

// use hook
const {
  t,
} = useI18n();

const router = useRouter();

const {
  isShowSecurityRule,
  handleSecurityRule,
  SecurityRule,
} = UseSecurityRule();

const resourceStore = useResourceStore();

const activeType = ref('ingress');
const deleteDialogShow = ref(false);
const deleteId = ref(0);
const deleteLoading = ref(false);

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

watch(
  () => activeType.value,
  (v) => {
    state.isLoading = true;
    handleSwtichType(v);
  },
);

// 获取列表数据
const fetchList = async (fetchType: string) => {
  console.log(1111);
  // eslint-disable-next-line vue/no-mutating-props
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

// 切换tab
const handleSwtichType = async (v: any) => {
  // eslint-disable-next-line vue/no-mutating-props
  props.filter.rules[0].value = v;
  const params = {
    fetchUrl: `vendors/${props.vendor}/security_groups/${props.id}/rules`,
  };
  // eslint-disable-next-line max-len
  const { datas, pagination, isLoading, handlePageChange, handlePageSizeChange, handleSort } = await fetchList(params.fetchUrl);
  console.log('datas', datas);
  state.datas = datas;
  state.isLoading = isLoading;
  state.pagination = pagination;
  state.handlePageChange = handlePageChange;
  state.handlePageSizeChange = handlePageSizeChange;
  state.handleSort = handleSort;
};

// 确定删除
const handleDeleteConfirm = () => {
  deleteLoading.value = true;
  resourceStore
    .delete(`vendors/${props.vendor}/security_groups/${props.id}/rules`, deleteId.value)
    .then(() => {
      Message({
        theme: 'success',
        message: t('删除成功'),
      });
      handleSwtichType(activeType.value);
    })
    .finally(() => {
      deleteLoading.value = false;
    });
};

// 初始化
handleSwtichType(activeType.value);


const inColumns = [
  {
    label: t('来源'),
    render({ data }: any) {
      return h(
        'span',
        {},
        [
          data.cloud_address_group_id || data.cloud_address_id
          || data.cloud_service_group_id || data.cloud_service_id || data.cloud_target_security_group_id
          || data.ipv4_cidr || data.ipv6_cidr,
        ],
      );
    },
  },
  {
    label: t('协议端口'),
    render({ data }: any) {
      return h(
        'span',
        {},
        [
          `${data.protocol}:${data.port}`,
        ],
      );
    },
  },
  {
    label: t('策略'),
    render({ data }: any) {
      return h(
        'span',
        {},
        [
          SecurityRuleEnum[data.action],
        ],
      );
    },
  },
  {
    label: t('备注'),
    field: 'memo',
  },
  {
    label: t('修改时间'),
    field: 'updated_at',
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
              theme: 'primary',
              onClick() {
                deleteDialogShow.value = true;
                deleteId.value = data.id;
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

const outColumns = [
  {
    label: t('目标'),
    render({ data }: any) {
      return h(
        'span',
        {},
        [
          data.cloud_address_group_id || data.cloud_address_id
          || data.cloud_service_group_id || data.cloud_service_id || data.cloud_target_security_group_id
          || data.ipv4_cidr || data.ipv6_cidr,
        ],
      );
    },
  },
  {
    label: t('协议端口'),
    render({ data }: any) {
      return h(
        'span',
        {},
        [
          `${data.protocol}:${data.port}`,
        ],
      );
    },
  },
  {
    label: t('策略'),
    render({ data }: any) {
      return h(
        'span',
        {},
        [
          SecurityRuleEnum[data.action],
        ],
      );
    },
  },
  {
    label: t('备注'),
    field: 'memo',
  },
  {
    label: t('修改时间'),
    field: 'updated_at',
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
              theme: 'primary',
              onClick() {
                deleteDialogShow.value = true;
                deleteId.value = data.id;
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
// tab 信息
const types = [
  { name: 'ingress', label: t('入站规则') },
  { name: 'egress', label: t('出站规则') },
];

</script>

<template>
  <bk-loading
    :loading="state.isLoading"
  >
    <section class="mt20 rule-main">
      <bk-radio-group
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

      <bk-button theme="primary" @click="handleSecurityRule">
        {{t('新增规则')}}
      </bk-button>
    </section>

    <bk-table
      v-if="activeType === 'ingress'"
      class="mt20"
      row-hover="auto"
      remote-pagination
      :columns="inColumns"
      :data="state.datas"
      :pagination="state.pagination"
      @page-limit-change="state.handlePageSizeChange"
      @page-value-change="state.handlePageChange"
      @column-sort="state.handleSort"
    />

    <bk-table
      v-if="activeType === 'egress'"
      class="mt20"
      row-hover="auto"
      remote-pagination
      :columns="outColumns"
      :data="state.datas"
      :pagination="state.pagination"
      @page-limit-change="state.handlePageSizeChange"
      @page-value-change="state.handlePageChange"
      @column-sort="state.handleSort"
    />

  </bk-loading>

  <security-rule
    v-model:isShow="isShowSecurityRule"
    :title="t('添加入站规则')" />


  <bk-dialog
    :is-show="deleteDialogShow"
    :title="'确定删除要该条规则?'"
    :theme="'primary'"
    @closed="() => deleteDialogShow = false"
    :is-loading="deleteLoading"
    @confirm="handleDeleteConfirm()"
  >
    <span>删除后不可恢复</span>
  </bk-dialog>
</template>

<style lang="scss" scoped>
  .rule-main {
    display: flex;
    align-items: center;
    justify-content: space-between;
  }
</style>
