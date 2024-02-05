<script lang="ts" setup>
import {
  ref,
  watch,
  h,
  reactive,
  PropType,
  inject,
  computed,
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
} from '@/store';

import { SecurityRuleEnum, HuaweiSecurityRuleEnum, AzureSecurityRuleEnum } from '@/typings';

import UseSecurityRule from '@/views/resource/resource-manage/hooks/use-security-rule';
import useQueryCommonList from '@/views/resource/resource-manage/hooks/use-query-list-common';
import useColumns from '@/views/resource/resource-manage/hooks/use-columns';
import bus from '@/common/bus';
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
  relatedSecurityGroups: {
    type: Array as PropType<Array<any>>,
  },
  templateData: {
    type: Object as PropType<Record<string, Array<any>>>,
  },
});

// use hook
const {
  t,
} = useI18n();

const {
  isShowSecurityRule,
  handleSecurityRule,
  SecurityRule,
} = UseSecurityRule();

const resourceStore = useResourceStore();

const activeType = ref('ingress');
const deleteDialogShow = ref(false);
const deleteId = ref(0);
const securityRuleLoading = ref(false);
const fetchUrl = ref<string>(`vendors/${props.vendor}/security_groups/${props.id}/rules/list`);
const dataId = ref('');
const AllData = ref({ ALL: 'ALL', '-1': '-1', '*': '*' });
const azureDefaultList = ref([]);
const azureDefaultColumns = ref([]);
const authVerifyData: any = inject('authVerifyData');
const isResourcePage: any = inject('isResourcePage');

const actionName = computed(() => {   // 资源下没有业务ID
  return isResourcePage.value ? 'iaas_resource_operate' : 'biz_iaas_resource_operate';
});

// 权限hook
// const {
//   showPermissionDialog,
//   handlePermissionConfirm,
//   handlePermissionDialog,
//   handleAuth,
//   permissionParams,
//   authVerifyData,
// } = useVerify();

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
  columns: useColumns('group').columns,
});

watch(
  () => activeType.value,
  (v) => {
    state.isLoading = true;
    // eslint-disable-next-line vue/no-mutating-props
    props.filter.rules[0].value = v;
    if (props.vendor === 'azure') {
      getDefaultList(v);
    }
  },
);

const getDefaultList = async (type: string) => {
  const list = await resourceStore.getAzureDefaultList(type);
  azureDefaultList.value = list?.data;
};

// 获取列表数据
const {
  datas,
  pagination,
  isLoading,
  handlePageChange,
  handlePageSizeChange,
  getList,
} = useQueryCommonList(props, fetchUrl, props.vendor === 'tcloud' ? { sort: 'cloud_policy_index', order: 'ASC' } : '');

state.datas = datas;
state.isLoading = isLoading;
state.pagination = pagination;
state.handlePageChange = handlePageChange;
state.handlePageSizeChange = handlePageSizeChange;

// 切换tab
const handleSwtichType = async () => {
  if (props.vendor === 'azure') {
    getDefaultList(activeType.value);
  }
};

// 确定删除
const handleDeleteConfirm = () => {
  securityRuleLoading.value = true;
  resourceStore
    .delete(`vendors/${props.vendor}/security_groups/${props.id}/rules`, deleteId.value)
    .then(() => {
      Message({
        theme: 'success',
        message: t('删除成功'),
      });
      getList();
    })
    .finally(() => {
      securityRuleLoading.value = false;
    });
};

// 提交规则
const handleSubmitRule = async (tableData: any) => {
  const data = JSON.parse(JSON.stringify(tableData));
  securityRuleLoading.value = true;
  if (props.vendor === 'aws') {   // aws 需要from_port 、to_port
    data.forEach((e: any) => {
      if (typeof e.port === 'string' && e?.port.includes('-')) {
        // eslint-disable-next-line prefer-destructuring
        e.from_port = Number(e.port.split('-')[0]);
        // eslint-disable-next-line prefer-destructuring
        e.to_port = Number(e.port.split('-')[1]);
      } else {
        console.log('-1', e.port);
        e.from_port = e.port === 'ALL' ? -1 : Number(e.port);
        e.to_port = e.port === 'ALL' ? -1 : Number(e.port);
      }
      delete e.port;
    });
  }
  // 过滤没有值的字段
  data.forEach((e: any) => {
    if (e?.source_port_range?.includes(',')) {
      e.source_port_ranges = e.source_port_range.split(',');
      e.source_port_range = '';
    }
    if (e?.destination_port_range?.includes(',')) {
      e.destination_port_ranges = e.destination_port_range.split(',');
      e.destination_port_range = '';
    }
    e.port = AllData.value[e.protocol] ? AllData.value[e.protocol] : e.port;
    Object.keys(e).forEach((item: any) => { // 删除没有val的key
      if (!e[item] || e[item] === 'huaweiAll') {
        if (e[item] === 'huaweiAll') {
          delete e.port;
        }
        delete e[item];
      }
    });
    e.priority = +e.priority;
  });
  const params = {
    [`${activeType.value}_rule_set`]: data,
  };
  try {
    if (data[0].id) {
      await resourceStore.update(`vendors/${props.vendor}/security_groups/${props.id}/rules`, data[0], data[0].id);
    } else {
      await resourceStore.add(`vendors/${props.vendor}/security_groups/${props.id}/rules/create`, params);
    }
    Message({
      message: t(data[0].id ? t('更新成功') : t('添加成功')),
      theme: 'success',
    });
    getList();
    isShowSecurityRule.value = false;
  } catch (error) {
    console.log(error);
  } finally {
    securityRuleLoading.value = false;
  }
};

const handleSecurityRuleDialog = (data: any) => {
  console.log('data', data);
  dataId.value = data?.id;
  resourceStore.setSecurityRuleDetail(data);
  handleSecurityRule();
};

// 权限弹窗 bus通知最外层弹出
const showAuthDialog = (authActionName: string) => {
  bus.$emit('auth', authActionName);
};

// 初始化
handleSwtichType();
getList();

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
          || data.ipv4_cidr || data.ipv6_cidr || data.cloud_remote_group_id || data.remote_ip_prefix
          || (data.source_address_prefix === '*' ? t('任何') : data.source_address_prefix) || data.source_address_prefixes || data.cloud_source_security_group_ids
          || data.destination_address_prefix || data.destination_address_prefixes
          || data.cloud_destination_security_group_ids || (data?.ethertype === 'IPv6' ? '::/0' : '0.0.0.0/0'),
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
          // eslint-disable-next-line no-nested-ternary
          props.vendor === 'aws' && (data.protocol === '-1' && data.to_port === -1) ? t('全部')
            // eslint-disable-next-line no-nested-ternary
            : props.vendor === 'huawei' && (!data.protocol && !data.port) ? t('全部')
              : props.vendor === 'azure' && (data.protocol === '*' && data.destination_port_range === '*') ? t('全部') :  `${data.protocol}:${data.port || data.to_port || data.destination_port_range || data.destination_port_ranges || '--'}`,
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
          // eslint-disable-next-line no-nested-ternary
          props.vendor === 'huawei' ? HuaweiSecurityRuleEnum[data.action] : props.vendor === 'azure' ? AzureSecurityRuleEnum[data.access]
            : props.vendor === 'aws' ? t('允许') : (SecurityRuleEnum[data.action] || '--'),
        ],
      );
    },
  },
  {
    label: t('备注'),
    field: 'memo',
    render: ({ data }) => data.memo || '--',
  },
  {
    label: t('修改时间'),
    field: 'updated_at',
  },
  {
    label: t('操作'),
    field: 'operate',
    render({ data }: any) {
      return h(
        'span',
        {},
        [
          props.vendor !== 'huawei' && h(
            'span',
            {
              onClick() {
                showAuthDialog(actionName.value);
              },
            },
            [
              h(
                Button,
                {
                  text: true,
                  theme: 'primary',
                  disabled: !authVerifyData.value?.permissionAction[actionName.value],
                  onClick() {
                    handleSecurityRuleDialog(data);
                  },
                },
                [
                  t('编辑'),
                ],
              ),
            ],

          ),
          h(
            'span',
            {
              onClick() {
                showAuthDialog(actionName.value);
              },
            },
            [
              h(
                Button,
                {
                  class: 'ml10',
                  text: true,
                  theme: 'primary',
                  disabled: !authVerifyData.value?.permissionAction[actionName.value],
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
          || data.ipv4_cidr || data.ipv6_cidr || data.cloud_remote_group_id || data.remote_ip_prefix
          || data.cloud_source_security_group_ids
          || (data.destination_address_prefix === '*' ? t('任何') : data.destination_address_prefix) || data.destination_address_prefixes
          || data.cloud_destination_security_group_ids || (data?.ethertype === 'IPv6' ? '::/0' : '0.0.0.0/0'),
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
          // eslint-disable-next-line no-nested-ternary
          props.vendor === 'aws' && (data.protocol === '-1' && data.to_port === -1) ? t('全部')
            // eslint-disable-next-line no-nested-ternary
            : props.vendor === 'huawei' && (!data.protocol && !data.port) ? t('全部')
              : props.vendor === 'azure' && (data.protocol === '*' && data.destination_port_range === '*') ? t('全部') :  `${data.protocol}:${data.port || data.to_port || data.destination_port_range || '--'}`,
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
          // eslint-disable-next-line no-nested-ternary
          props.vendor === 'huawei' ? HuaweiSecurityRuleEnum[data.action] : props.vendor === 'azure' ? AzureSecurityRuleEnum[data.access]
            : props.vendor === 'aws' ? t('允许') : (SecurityRuleEnum[data.action] || '--'),
        ],
      );
    },
  },
  {
    label: t('备注'),
    field: 'memo',
    render: ({ data }) => data.memo || '--',
  },
  {
    label: t('修改时间'),
    field: 'updated_at',
  },
  {
    label: t('操作'),
    field: 'operate',
    render({ data }: any) {
      return h(
        'span',
        {},
        [
          props.vendor !== 'huawei' && h(
            'span',
            {
              onClick() {
                showAuthDialog(actionName.value);
              },
            },
            [
              h(
                Button,
                {
                  text: true,
                  theme: 'primary',
                  disabled: !authVerifyData.value?.permissionAction[actionName.value],
                  onClick() {
                    handleSecurityRuleDialog(data);
                  },
                },
                [
                  t('编辑'),
                ],
              ),
            ],

          ),
          h(
            'span',
            {
              onClick() {
                showAuthDialog(actionName.value);
              },
            },
            [
              h(
                Button,
                {
                  class: 'ml10',
                  text: true,
                  theme: 'primary',
                  disabled: !authVerifyData.value?.permissionAction[actionName.value],
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

if (props.vendor === 'huawei') {
  inColumns.unshift({
    label: t('优先级'),
    field: 'priority',
  }, {
    label: t('类型'),
    field: 'ethertype',
  });
  outColumns.unshift({
    label: t('优先级'),
    field: 'priority',
  }, {
    label: t('类型'),
    field: 'ethertype',
  });
} else if (props.vendor === 'azure')  {
  inColumns.unshift({
    label: t('名称'),
    field: 'name',
  }, {
    label: t('优先级'),
    field: 'priority',
  }, {
    label: t('目标'),
    render({ data }: any) {
      return (data.destination_address_prefix === '*' ? t('任何') : data.destination_address_prefix) || data.destination_address_prefixes || data.cloud_destination_security_group_ids;
    },
  });
  outColumns.unshift({
    label: t('名称'),
    field: 'name',
  }, {
    label: t('优先级'),
    field: 'priority',
  }, {
    label: t('来源'),
    render({ data }: any) {
      return (data.source_address_prefix === '*' ? t('任何') : data.source_address_prefix) || data.source_address_prefixes || data.cloud_source_security_group_ids;
    },
  });
  const defaultColumns = activeType.value === 'ingress' ? inColumns : outColumns;
  azureDefaultColumns.value = defaultColumns.filter((item: any) => item.field !== 'operate' && item.field !== 'updated_at');   // azure默认规则没有操作和修改时间
}

</script>

<template>
  <div>
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

        <div @click="showAuthDialog(actionName)">
          <bk-button
            :disabled="!authVerifyData?.
              permissionAction[actionName]"
            theme="primary" @click="handleSecurityRuleDialog({})">
            {{t('新增规则')}}
          </bk-button>
        </div>
      </section>

      <div v-if="props.vendor === 'azure'" class="mb20">
        <h4 class="mt10">Azure默认{{activeType === 'ingress' ? t('入站') : t('出站')}}规则</h4>
        <bk-table
          class="mt10"
          row-hover="auto"
          :columns="azureDefaultColumns"
          :data="azureDefaultList"
          show-overflow-tooltip
        />
      </div>

      <h4 v-if="props.vendor === 'azure'" class="mt10">Azure{{activeType === 'ingress' ? t('入站') : t('出站')}}规则</h4>
      <bk-table
        v-if="activeType === 'ingress'"
        class="mt20"
        row-hover="auto"
        remote-pagination
        :columns="inColumns"
        :data="state.datas"
        :pagination="state.pagination"
        show-overflow-tooltip
        @page-limit-change="state.handlePageSizeChange"
        @page-value-change="state.handlePageChange"
      />

      <bk-table
        v-if="activeType === 'egress'"
        class="mt20"
        row-hover="auto"
        remote-pagination
        :columns="outColumns"
        :data="state.datas"
        :pagination="state.pagination"
        show-overflow-tooltip
        @page-limit-change="state.handlePageSizeChange"
        @page-value-change="state.handlePageChange"
      />

    </bk-loading>

    <security-rule
      v-model:isShow="isShowSecurityRule"
      :loading="securityRuleLoading"
      dialog-width="1680"
      :active-type="activeType"
      :title="t(activeType === 'egress' ? `${dataId ? '编辑' : '添加'}出站规则` : `${dataId ? '编辑' : '添加'}入站规则`)"
      :is-edit="!!dataId"
      :vendor="vendor"
      @submit="handleSubmitRule"
      :related-security-groups="props.relatedSecurityGroups"
      :template-data="props.templateData"
    />


    <bk-dialog
      v-model:is-show="deleteDialogShow"
      :title="'确定删除要该条规则?'"
      :theme="'primary'"
      @closed="() => deleteDialogShow = false"
      :is-loading="securityRuleLoading"
      @confirm="handleDeleteConfirm()"
    >
      <span>删除后不可恢复</span>
    </bk-dialog>
  </div>
</template>

<style lang="scss" scoped>
  .rule-main {
    display: flex;
    align-items: center;
    justify-content: space-between;
  }
</style>
