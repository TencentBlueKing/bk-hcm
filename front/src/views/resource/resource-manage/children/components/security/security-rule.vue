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
} from 'bkui-vue';

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

const activeType = ref('in');

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

const fetchList = async (fetchType: string) => {
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

const handleSwtichType = async (type: string) => {
  console.log('props', props.vendor);
  const params = {
    fetchUrl: `vendors/${props.vendor}/security_groups/${props.id}/rules`,
    columns: 'group',
    // dialogName: t('删除安全组'),
  };
  if (type === 'out') {
    params.fetchUrl = 'vendors/gcp/firewalls/rules';
    params.columns = 'gcp';
    // params.dialogName = t('删除防火墙规则');
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
  // const { handleShowDelete, DeleteDialog } = showDeleteDialog(params.fetchUrl, params.dialogName);
  // securityHandleShowDelete = handleShowDelete;
  // SecurityDeleteDialog = DeleteDialog;
};


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
    field: 'id',
  },
  {
    label: t('端口协议'),
    field: 'id',
  },
  {
    label: t('端口'),
    field: 'id',
  },
  {
    label: t('策略'),
    field: 'id',
  },
  {
    label: t('操作'),
    field: 'id',
  },
];
const outData = [
  {
    id: 233,
  },
];
// tab 信息
const types = [
  { name: 'in', label: t('入站规则') },
  { name: 'out', label: t('出站规则') },
];

</script>

<template>
  <section class="mt20 rule-main">
    <bk-radio-group
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

    <bk-button theme="primary" @click="handleSecurityRule">
      {{t('新增规则')}}
    </bk-button>
  </section>

  <bk-table
    v-if="activeType === 'in'"
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
    v-if="activeType === 'out'"
    class="mt20"
    row-hover="auto"
    :columns="outColumns"
    :data="outData"
  />

  <security-rule
    v-model:isShow="isShowSecurityRule"
    :title="t('添加入站规则')" />
</template>

<style lang="scss" scoped>
  .rule-main {
    display: flex;
    align-items: center;
    justify-content: space-between;
  }
</style>
