<!-- eslint-disable no-nested-ternary -->
<script lang="ts" setup>
import { ref, watch, h, reactive, PropType, inject, computed, withDirectives } from 'vue';
import { useI18n } from 'vue-i18n';

import { bkTooltips, Button, Message } from 'bkui-vue';

import { useResourceStore } from '@/store';

import { SecurityRuleEnum, HuaweiSecurityRuleEnum, AzureSecurityRuleEnum } from '@/typings';
import { VendorEnum } from '@/common/constant';

import UseSecurityRule from '@/views/resource/resource-manage/hooks/use-security-rule';
import useQueryCommonList from '@/views/resource/resource-manage/hooks/use-query-list-common';
import useColumns from '@/views/resource/resource-manage/hooks/use-columns';
import bus from '@/common/bus';
import { timeFormatter } from '@/common/util';
import {
  azureSourceAddressTypes,
  AzureSourceTypeArr,
  azureTargetAddressTypes,
  AzureTargetTypeArr,
} from './add-rule/vendors/azure';
import { useRoute } from 'vue-router';
import { awsSourceAddressTypes, AwsSourceTypeArr } from './add-rule/vendors/aws';
import { tcloudSourceAddressTypes, TcloudSourceTypeArr } from './add-rule/vendors/tcloud';
import { huaweiSourceAddressTypes } from './add-rule/vendors/huawei';
import RuleSort from './security-rule-sort.vue';

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
const { t } = useI18n();

const { isShowSecurityRule, handleSecurityRule, SecurityRule } = UseSecurityRule();

const resourceStore = useResourceStore();
const route = useRoute();

const activeType = ref('ingress');
const deleteDialogShow = ref(false);
const deleteId = ref(0);
const securityRuleLoading = ref(false);
const fetchUrl = ref<string>(`vendors/${route.query.vendor}/security_groups/${props.id}/rules/list`);
const dataId = ref('');
const azureDefaultList = ref([]);
const azureDefaultColumns = ref([]);
const authVerifyData: any = inject('authVerifyData');
const isResourcePage: any = inject('isResourcePage');
const show = ref<Boolean>(false);

const actionName = computed(() => {
  // 资源下没有业务ID
  return isResourcePage.value ? 'iaas_resource_operate' : 'biz_iaas_resource_operate';
});
// 排序功能目前只在公有云与业务下的自研云实现
const showSort = computed(() => {
  const { vendor } = route.query;
  return vendor === VendorEnum.TCLOUD || (vendor === VendorEnum.ZIYAN && !isResourcePage.value);
});

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
    if (route.query.vendor === 'azure') {
      getDefaultList(v);
    }
  },
);

const getDefaultList = async (type: string) => {
  const list = await resourceStore.getAzureDefaultList(type);
  azureDefaultList.value = list?.data;
};

// 获取列表数据
const { datas, pagination, isLoading, handlePageChange, handlePageSizeChange, getList } = useQueryCommonList(
  props,
  fetchUrl,
  route.query.vendor === 'tcloud' ? { sort: 'cloud_policy_index', order: 'ASC' } : '',
);

state.datas = datas;
state.isLoading = isLoading;
state.pagination = pagination;
state.handlePageChange = handlePageChange;
state.handlePageSizeChange = handlePageSizeChange;

// 切换tab
const handleSwtichType = async () => {
  if (route.query.vendor === 'azure') {
    getDefaultList(activeType.value);
  }
};

// 确定删除
const handleDeleteConfirm = () => {
  securityRuleLoading.value = true;
  resourceStore
    .delete(`vendors/${route.query.vendor}/security_groups/${props.id}/rules`, deleteId.value)
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

const handleRuleSubmit = () => {
  getList();
};

const handleSecurityRuleDialog = (data: any) => {
  dataId.value = data?.id;
  resourceStore.setSecurityRuleDetail(data);
  handleSecurityRule();
};

// 规则排序抽屉
const handleSecurityRuleSort = () => {
  show.value = true;
};

// 权限弹窗 bus通知最外层弹出
const showAuthDialog = (authActionName: string) => {
  bus.$emit('auth', authActionName);
};

const handelSortDone = () => {
  state.handlePageChange(1);
};
// 初始化
handleSwtichType();
getList();

// 入站规则列字段
const inColumns: any = computed(() =>
  [
    {
      label: t('名称'),
      field: 'name',
      isShow: route.query.vendor === 'azure',
    },
    {
      label: t('优先级'),
      field: 'priority',
      isShow: route.query.vendor === 'huawei' || route.query.vendor === 'azure',
    },
    {
      label: t('源地址类型'),
      render({ data }: any) {
        const vendor = (route.query.vendor as VendorEnum) || VendorEnum.TCLOUD;
        const sourceMap: any = {
          [VendorEnum.AWS]: {
            types: awsSourceAddressTypes,
            arr: AwsSourceTypeArr,
          },
          [VendorEnum.AZURE]: {
            types: azureSourceAddressTypes,
            arr: AzureSourceTypeArr,
          },
          [VendorEnum.HUAWEI]: {
            types: huaweiSourceAddressTypes,
            arr: TcloudSourceTypeArr,
          },
          [VendorEnum.TCLOUD]: {
            types: tcloudSourceAddressTypes,
            arr: TcloudSourceTypeArr,
          },
        };
        const { types } = sourceMap[vendor];
        const { arr } = sourceMap[vendor];
        const map = new Map(types.map((item: { value: string; label: string }) => [item.value, item.label]));
        let k = '';
        arr.forEach((type: string) => data[type] && (k = type));
        return map.get(k) || '--';
      },
      isShow: true,
    },
    {
      label: t('源地址'),
      render({ data }: any) {
        return h('span', {}, [
          data.cloud_address_group_id ||
            data.cloud_address_id ||
            data.cloud_service_group_id ||
            data.cloud_target_security_group_id ||
            data.ipv4_cidr ||
            data.ipv6_cidr ||
            data.cloud_remote_group_id ||
            data.remote_ip_prefix ||
            (data.source_address_prefix === '*' ? t('ALL') : data.source_address_prefix) ||
            data.source_address_prefixes ||
            data.cloud_source_security_group_ids ||
            data.destination_address_prefix ||
            data.destination_address_prefixes ||
            data.cloud_destination_security_group_ids ||
            (data?.ethertype === 'IPv6' ? '::/0' : '0.0.0.0/0'),
        ]);
      },
      isShow: true,
    },
    {
      label: t('源端口'),
      render({ data }: any) {
        return (data.source_port_range === '*' ? 'ALL' : data.source_port_range) || '--';
      },
      isShow: route.query.vendor === 'azure',
    },
    {
      label: t('目标地址类型'),
      render({ data }: any) {
        const map = new Map(
          azureTargetAddressTypes.map((item: { value: string; label: string }) => [item.value, item.label]),
        );
        let k = '';
        AzureTargetTypeArr.forEach((type: string) => data[type] && (k = type));
        return map.get(k) || '--';
      },
      isShow: route.query.vendor === 'azure',
    },
    {
      label: t('类型'),
      field: 'ethertype',
      isShow: route.query.vendor === 'huawei',
    },

    {
      label: t('目标地址'),
      render({ data }: any) {
        return (
          (data.destination_address_prefix === '*' ? t('ALL') : data.destination_address_prefix) ||
          data.destination_address_prefixes ||
          data.cloud_destination_security_group_ids
        );
      },
      isShow: route.query.vendor === 'azure',
    },
    {
      label: route.query.vendor === 'azure' ? t('目标端口协议类型') : t('协议'),
      render({ data }: any) {
        return h('span', {}, [
          data.cloud_service_id ||
            (route.query.vendor === 'aws' && data.protocol === '-1'
              ? t('ALL')
              : route.query.vendor === 'huawei' && !data.protocol
              ? t('ALL')
              : route.query.vendor === 'azure' && data.protocol === '*'
              ? t('ALL')
              : `${data.protocol}`),
        ]);
      },
      isShow: true,
    },
    {
      label: route.query.vendor === 'azure' ? t('目标协议端口') : t('端口'),
      render({ data }: any) {
        return h('span', {}, [
          data.cloud_service_id ||
            (route.query.vendor === 'aws' && data.to_port === -1
              ? t('ALL')
              : route.query.vendor === 'huawei' && !data.port
              ? t('ALL')
              : route.query.vendor === 'azure' && data.destination_port_range === '*'
              ? t('ALL')
              : `${data.port || data.to_port || data.destination_port_range || data.destination_port_ranges || '--'}`),
        ]);
      },
      isShow: true,
    },
    {
      label: t('策略'),
      render({ data }: any) {
        return h('span', {}, [
          route.query.vendor === 'huawei'
            ? HuaweiSecurityRuleEnum[data.action]
            : route.query.vendor === 'azure'
            ? AzureSecurityRuleEnum[data.access]
            : route.query.vendor === 'aws'
            ? t('允许')
            : SecurityRuleEnum[data.action] || '--',
        ]);
      },
      isShow: route.query.vendor !== 'aws',
    },
    {
      label: t('备注'),
      field: 'memo',
      render: ({ data }) => data.memo || '--',
      isShow: true,
    },
    {
      label: t('修改时间'),
      field: 'updated_at',
      render: ({ cell }: { cell: string }) => timeFormatter(cell),
      isShow: true,
    },
    {
      label: t('操作'),
      field: 'operate',
      render({ data }: any) {
        return h('span', {}, [
          withDirectives(
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
                    text: true,
                    theme: 'primary',
                    disabled:
                      !authVerifyData.value?.permissionAction[actionName.value] || route.query.vendor === 'huawei',
                    onClick() {
                      handleSecurityRuleDialog(data);
                    },
                  },
                  [t('编辑')],
                ),
              ],
            ),
            [
              [
                bkTooltips,
                {
                  content: '该功能当前未支持',
                  disabled: route.query.vendor !== 'huawei',
                },
              ],
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
                [t('删除')],
              ),
            ],
          ),
        ]);
      },
      isShow: true,
    },
  ].filter(({ isShow }) => !!isShow),
);

// 出站规则列字段
const outColumns: any = computed(() =>
  [
    {
      label: t('名称'),
      field: 'name',
      isShow: route.query.vendor === 'azure',
    },
    {
      label: t('优先级'),
      field: 'priority',
      isShow: route.query.vendor === 'huawei' || route.query.vendor === 'azure',
    },
    {
      label: t('源地址类型'),
      render({ data }: any) {
        const map = new Map(
          azureSourceAddressTypes.map((item: { value: string; label: string }) => [item.value, item.label]),
        );
        let k = '';
        AzureSourceTypeArr.forEach((type: string) => data[type] && (k = type));
        return map.get(k) || '--';
      },
      isShow: route.query.vendor === 'azure',
    },
    {
      label: t('源地址'),
      render({ data }: any) {
        return h('span', {}, [
          data.cloud_address_group_id ||
            data.cloud_address_id ||
            data.cloud_service_group_id ||
            data.cloud_target_security_group_id ||
            data.ipv4_cidr ||
            data.ipv6_cidr ||
            data.cloud_remote_group_id ||
            data.remote_ip_prefix ||
            (data.source_address_prefix === '*' ? t('ALL') : data.source_address_prefix) ||
            data.source_address_prefixes ||
            data.cloud_source_security_group_ids ||
            data.destination_address_prefix ||
            data.destination_address_prefixes ||
            data.cloud_destination_security_group_ids ||
            (data?.ethertype === 'IPv6' ? '::/0' : '0.0.0.0/0'),
        ]);
      },
      isShow: route.query.vendor === 'azure',
    },
    {
      label: t('源端口'),
      render({ data }: any) {
        return (data.source_port_range === '*' ? 'ALL' : data.source_port_range) || '--';
      },
      isShow: route.query.vendor === 'azure',
    },
    {
      label: t('目标地址类型'),
      render({ data }: any) {
        const vendor = route.query.vendor as VendorEnum;
        const targetMap: any = {
          [VendorEnum.AWS]: {
            types: awsSourceAddressTypes,
            arr: AwsSourceTypeArr,
          },
          [VendorEnum.AZURE]: {
            types: azureTargetAddressTypes,
            arr: AzureTargetTypeArr,
          },
          [VendorEnum.HUAWEI]: {
            types: huaweiSourceAddressTypes,
            arr: TcloudSourceTypeArr,
          },
          [VendorEnum.TCLOUD]: {
            types: tcloudSourceAddressTypes,
            arr: TcloudSourceTypeArr,
          },
        };
        const { types } = targetMap[vendor];
        const { arr } = targetMap[vendor];
        const map = new Map(types.map((item: { value: string; label: string }) => [item.value, item.label]));
        let k = '';
        arr.forEach((type: string) => data[type] && (k = type));
        return map.get(k) || '--';
      },
      isShow: true,
    },
    {
      label: t('目标地址'),
      render({ data }: any) {
        return h('span', {}, [
          data.cloud_address_group_id ||
            data.cloud_address_id ||
            data.cloud_service_group_id ||
            data.cloud_target_security_group_id ||
            data.ipv4_cidr ||
            data.ipv6_cidr ||
            data.cloud_remote_group_id ||
            data.remote_ip_prefix ||
            data.cloud_source_security_group_ids ||
            (data.destination_address_prefix === '*' ? t('ALL') : data.destination_address_prefix) ||
            data.destination_address_prefixes ||
            data.cloud_destination_security_group_ids ||
            (data?.ethertype === 'IPv6' ? '::/0' : '0.0.0.0/0'),
        ]);
      },
      isShow: true,
    },
    {
      label: t('类型'),
      field: 'ethertype',
      isShow: route.query.vendor === 'huawei',
    },
    {
      label: route.query.vendor === 'azure' ? t('目标端口协议类型') : t('协议'),
      render({ data }: any) {
        return h('span', {}, [
          data.cloud_service_id ||
            (route.query.vendor === 'aws' && data.protocol === '-1'
              ? t('ALL')
              : route.query.vendor === 'huawei' && !data.protocol
              ? t('ALL')
              : route.query.vendor === 'azure' && data.protocol === '*'
              ? t('ALL')
              : `${data.protocol}`),
        ]);
      },
      isShow: true,
    },
    {
      label: route.query.vendor === 'azure' ? t('目标协议端口') : t('端口'),
      render({ data }: any) {
        return h('span', {}, [
          data.cloud_service_id ||
            (route.query.vendor === 'aws' && data.to_port === -1
              ? t('ALL')
              : route.query.vendor === 'huawei' && !data.port
              ? t('ALL')
              : route.query.vendor === 'azure' && data.destination_port_range === '*'
              ? t('ALL')
              : `${data.port || data.to_port || data.destination_port_range || '--'}`),
        ]);
      },
      isShow: true,
    },
    {
      label: t('策略'),
      render({ data }: any) {
        return h('span', {}, [
          route.query.vendor === 'huawei'
            ? HuaweiSecurityRuleEnum[data.action]
            : route.query.vendor === 'azure'
            ? AzureSecurityRuleEnum[data.access]
            : route.query.vendor === 'aws'
            ? t('允许')
            : SecurityRuleEnum[data.action] || '--',
        ]);
      },
      isShow: route.query.vendor !== 'aws',
    },
    {
      label: t('备注'),
      field: 'memo',
      render: ({ data }) => data.memo || '--',
      isShow: true,
    },
    {
      label: t('修改时间'),
      field: 'updated_at',
      render: ({ cell }: { cell: string }) => timeFormatter(cell),
      isShow: true,
    },
    {
      label: t('操作'),
      field: 'operate',
      render({ data }: any) {
        return h('span', {}, [
          h(
            'span',
            {
              onClick() {
                showAuthDialog(actionName.value);
              },
            },
            [
              withDirectives(
                h(
                  Button,
                  {
                    text: true,
                    theme: 'primary',
                    disabled:
                      !authVerifyData.value?.permissionAction[actionName.value] || route.query.vendor === 'huawei',
                    onClick() {
                      handleSecurityRuleDialog(data);
                    },
                  },
                  [t('编辑')],
                ),
                [
                  [
                    bkTooltips,
                    {
                      content: '该功能当前未支持',
                      disabled: route.query.vendor !== 'huawei',
                    },
                  ],
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
                [t('删除')],
              ),
            ],
          ),
        ]);
      },
      isShow: true,
    },
  ].filter(({ isShow }) => !!isShow),
);

const defaultColumns = activeType.value === 'ingress' ? inColumns.value : outColumns.value;
azureDefaultColumns.value = defaultColumns.filter(
  (item: any) => item.field !== 'operate' && item.field !== 'updated_at',
);

// tab 信息
const types = [
  { name: 'ingress', label: t('入站规则') },
  { name: 'egress', label: t('出站规则') },
];
</script>

<template>
  <div>
    <bk-loading :loading="state.isLoading">
      <section class="rule-main">
        <bk-radio-group v-model="activeType" :disabled="state.isLoading">
          <bk-radio-button v-for="item in types" :key="item.name" :label="item.name">
            {{ item.label }}
          </bk-radio-button>
        </bk-radio-group>

        <div @click="showAuthDialog(actionName)">
          <bk-button
            :disabled="!authVerifyData?.permissionAction[actionName]"
            theme="primary"
            @click="handleSecurityRuleDialog({})"
          >
            {{ t('新增规则') }}
          </bk-button>
        </div>

        <bk-button icon="plus" v-if="showSort" @click="handleSecurityRuleSort">
          {{ t('规则排序') }}
        </bk-button>
      </section>

      <div v-if="route.query.vendor === 'azure'" class="mb20">
        <h4 class="mt10">Azure默认{{ activeType === 'ingress' ? t('入站') : t('出站') }}规则</h4>
        <bk-table
          class="mt10"
          row-hover="auto"
          :columns="azureDefaultColumns"
          :data="azureDefaultList"
          show-overflow-tooltip
        >
          <template #empty>
            <div class="security-empty-container">
              <bk-exception
                class="exception-wrap-item exception-part"
                type="empty"
                scene="part"
                description="无规则，默认拒绝所有流量"
              />
            </div>
          </template>
        </bk-table>
      </div>

      <h4 v-if="route.query.vendor === 'azure'" class="mt10">
        Azure{{ activeType === 'ingress' ? t('入站') : t('出站') }}规则
      </h4>
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
      >
        <template #empty>
          <div class="security-empty-container">
            <bk-exception
              class="exception-wrap-item exception-part"
              type="empty"
              scene="part"
              description="无规则，默认拒绝所有流量"
            />
          </div>
        </template>
      </bk-table>

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
      >
        <template #empty>
          <div class="security-empty-container">
            <bk-exception
              class="exception-wrap-item exception-part"
              type="empty"
              scene="part"
              description="无规则，默认拒绝所有流量"
            />
          </div>
        </template>
      </bk-table>
    </bk-loading>

    <security-rule
      v-model:isShow="isShowSecurityRule"
      :loading="securityRuleLoading"
      :id="props.id"
      dialog-width="1680"
      :active-type="activeType"
      :title="
        t(activeType === 'egress' ? `${dataId ? '编辑' : '添加'}出站规则` : `${dataId ? '编辑' : '添加'}入站规则`)
      "
      :is-edit="!!dataId"
      :vendor="vendor"
      @submit="handleRuleSubmit"
      :related-security-groups="props.relatedSecurityGroups"
      :template-data="props.templateData"
    />

    <bk-dialog
      v-model:is-show="deleteDialogShow"
      :title="'确定删除要该条规则?'"
      :theme="'primary'"
      @closed="() => (deleteDialogShow = false)"
      :is-loading="securityRuleLoading"
      @confirm="handleDeleteConfirm()"
    >
      <span>删除后不可恢复</span>
    </bk-dialog>

    <bk-sideslider v-model:isShow="show" :title="t('规则排序')" width="640" quick-close>
      <template #default>
        <rule-sort
          :id="props.id"
          :filter="props.filter"
          :type="activeType"
          v-model:show="show"
          @sort-done="handelSortDone"
        ></rule-sort>
      </template>
    </bk-sideslider>
  </div>
</template>

<style lang="scss" scoped>
.rule-main {
  display: flex;
  align-items: center;
  gap: 16px;
}

.security-empty-container {
  display: felx;
  align-items: center;
  margin: auto;
}
</style>
