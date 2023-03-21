<script setup lang="ts">
import type {
  // PlainObject,
  FilterType,
} from '@/typings/resource';
import { GcpTypeEnum, CloudType } from '@/typings';
import {
  Button,
  InfoBox,
  Message,
} from 'bkui-vue';
import {
  useResourceStore,
} from '@/store/resource';
import {
  ref,
  h,
  PropType,
  watch,
  reactive,
  defineExpose,
} from 'vue';

import {
  useI18n,
} from 'vue-i18n';
import {
  useRouter,
  useRoute,
} from 'vue-router';
import useQueryCommonList from '@/views/resource/resource-manage/hooks/use-query-list-common';
import useColumns from '@/views/resource/resource-manage/hooks/use-columns';

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

const route = useRoute();

const activeType = ref('group');
const fetchUrl = ref<string>('security_groups/list');
const resourceStore = useResourceStore();

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
  columns: useColumns('group'),
  params: {
    fetchUrl: 'security_groups',
    columns: 'group',
  },
});

const {
  datas,
  pagination,
  isLoading,
  handlePageChange,
  handlePageSizeChange,
  getList,
} = useQueryCommonList(props, fetchUrl);

// 状态保持
watch(
  () => activeType.value,
  (v) => {
    state.isLoading = true;
    state.pagination.current = 1;
    state.pagination.limit = 10;
    handleSwtichType(v);
  },
);


const handleSwtichType = async (type: string) => {
  if (type === 'gcp') {
    fetchUrl.value = 'vendors/gcp/firewalls/rules/list';
    state.params.fetchUrl = 'vendors/gcp/firewalls/rules';
    state.params.columns = 'gcp';
  } else {
    fetchUrl.value = 'security_groups/list';
    state.params.fetchUrl = 'security_groups';
    state.params.columns = 'group';
  }
  // eslint-disable-next-line max-len
  state.datas = datas;
  state.isLoading = isLoading;
  state.pagination = pagination;
  state.handlePageChange = handlePageChange;
  state.handlePageSizeChange = handlePageSizeChange;
};

// 抛出请求数据的方法，新增成功使用
const fetchComponentsData = () => {
  getList();
};

handleSwtichType(activeType.value);

defineExpose({ fetchComponentsData });

const groupColumns = [
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
            const routeInfo: any = {
              query: {
                id: data.id,
                vendor: data.vendor,
              },
            };
            // 业务下
            if (route.path.includes('business')) {
              Object.assign(
                routeInfo,
                {
                  name: 'securityBusinessDetail',
                },
              );
            } else {
              Object.assign(
                routeInfo,
                {
                  name: 'resourceDetail',
                  params: {
                    type: 'security',
                  },
                },
              );
            }
            router.push(routeInfo);
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
                securityHandleShowDelete(data);
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
            const routeInfo: any = {
              query: {
                id: data.id,
              },
            };
            // 业务下
            if (route.path.includes('business')) {
              Object.assign(
                routeInfo,
                {
                  name: 'gcpBusinessDetail',
                },
              );
            } else {
              Object.assign(
                routeInfo,
                {
                  name: 'resourceDetail',
                  params: {
                    type: 'gcp',
                  },
                },
              );
            }
            router.push(routeInfo);
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
                securityHandleShowDelete(data);
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


const securityHandleShowDelete = (data: any) => {
  InfoBox({
    title: '请确认是否删除',
    subTitle: `将删除【${data.name}】`,
    theme: 'danger',
    headerAlign: 'center',
    footerAlign: 'center',
    contentAlign: 'center',
    async onConfirm() {
      try {
        await resourceStore
          .deleteBatch(
            activeType.value === 'group' ? 'security_groups' : 'vendors/gcp/firewalls/rules',
            { ids: [data.id] },
          );
        getList();
        Message({
          message: t('删除成功'),
          theme: 'success',
        });
      } catch (error) {
        console.log(error);
      }
    },
  });
};
</script>

<template>
  <bk-loading
    :loading="state.isLoading"
  >
    <section>
      <slot>
      </slot>
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
      @page-limit-change="state.handlePageSizeChange"
      @page-value-change="state.handlePageChange"
      @column-sort="state.handleSort"
    />

    <bk-table
      v-if="activeType === 'gcp'"
      class="mt20"
      row-hover="auto"
      remote-pagination
      :pagination="state.pagination"
      :columns="gcpColumns"
      :data="state.datas"
      @page-limit-change="state.handlePageSizeChange"
      @page-value-change="state.handlePageChange"
      @column-sort="state.handleSort"
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
