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

import { SecurityRuleEnum, HuaweiSecurityRuleEnum, AzureSecurityRuleEnum } from '@/typings';

import UseSecurityRule from '@/views/resource/resource-manage/hooks/use-security-rule';
import useQueryCommonList from '@/views/resource/resource-manage/hooks/use-query-list-common';
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
});

watch(
  () => activeType.value,
  (v) => {
    state.isLoading = true;
    // eslint-disable-next-line vue/no-mutating-props
    props.filter.rules[0].value = v;
  },
);

// 获取列表数据
const {
  datas,
  pagination,
  isLoading,
  handlePageChange,
  handlePageSizeChange,
  getList,
} = useQueryCommonList(props, fetchUrl);

// 切换tab
const handleSwtichType = async () => {
  state.datas = datas;
  state.isLoading = isLoading;
  state.pagination = pagination;
  state.handlePageChange = handlePageChange;
  state.handlePageSizeChange = handlePageSizeChange;
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
const handleSubmitRule = async (data: any) => {
  console.log('data', data.id, activeType.value);
  securityRuleLoading.value = true;
  const params = {
    [`${activeType.value}_rule_set`]: data,
  };
  try {
    if (data.id) {
      await resourceStore.update(`vendors/${props.vendor}/security_groups/${props.id}/rules`, data, data.id);
    } else {
      await resourceStore.add(`vendors/${props.vendor}/security_groups/${props.id}/rules/create`, params);
    }
    Message({
      message: t(data.id ? '更新成功' : '添加成功'),
      theme: 'success',
    });
    getList();
  } catch (error) {
    console.log(error);
  } finally {
    isShowSecurityRule.value = false;
    securityRuleLoading.value = false;
  }
};

const handleSecurityRuleDialog = (data: any) => {
  console.log('data', data);
  resourceStore.setSecurityRuleDetail(data);
  handleSecurityRule();
};

// 初始化
handleSwtichType();

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
          || data.source_address_prefix || data.source_address_prefixs || data.cloud_source_security_group_ids
          || data.destination_address_prefix || data.destination_address_prefixes
          || data.cloud_destination_security_group_ids || '--',
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
          `${data.protocol || '--'}:${data.port || data.to_port || '--'}`,
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
            : SecurityRuleEnum[data.action],
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
          props.vendor !== 'huawei' && h(
            Button,
            {
              text: true,
              theme: 'primary',
              onClick() {
                handleSecurityRuleDialog(data);
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
          || data.ipv4_cidr || data.ipv6_cidr || data.cloud_remote_group_id || data.remote_ip_prefix
          || data.source_address_prefix || data.source_address_prefixs || data.cloud_source_security_group_ids
          || data.destination_address_prefix || data.destination_address_prefixes
          || data.cloud_destination_security_group_ids,
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
          `${data.protocol || '--'}:${data.port || data.to_port || '--'}`,
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
            : SecurityRuleEnum[data.action],
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
          props.vendor !== 'huawei' && h(
            Button,
            {
              text: true,
              theme: 'primary',
              onClick() {
                handleSecurityRuleDialog(data);
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
}

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

      <bk-button theme="primary" @click="handleSecurityRuleDialog({})">
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
    />

  </bk-loading>

  <security-rule
    v-model:isShow="isShowSecurityRule"
    :loading="securityRuleLoading"
    dialog-width="1200"
    :active-type="activeType"
    :title="t(activeType === 'egress' ? '添加出站规则' : '添加入站规则')"
    :vendor="vendor"
    @submit="handleSubmitRule"
  />


  <bk-dialog
    :is-show="deleteDialogShow"
    :title="'确定删除要该条规则?'"
    :theme="'primary'"
    @closed="() => deleteDialogShow = false"
    :is-loading="securityRuleLoading"
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
