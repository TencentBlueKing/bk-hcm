<!-- eslint-disable no-nested-ternary -->
<script lang="ts" setup>
import { ref, watch, h, PropType, inject, computed, withDirectives, ComputedRef, resolveComponent } from 'vue';
import { useI18n } from 'vue-i18n';

import { bkTooltips, Button, Message } from 'bkui-vue';

import { useResourceStore } from '@/store';

import { SecurityRuleEnum, HuaweiSecurityRuleEnum, AzureSecurityRuleEnum } from '@/typings';
import { VendorEnum } from '@/common/constant';

import UseSecurityRule from '@/views/resource/resource-manage/hooks/use-security-rule';
import useQueryCommonList from '@/views/resource/resource-manage/hooks/use-query-list-common';
import { AUTH_UPDATE_IAAS_RESOURCE, AUTH_BIZ_UPDATE_IAAS_RESOURCE } from '@/constants/auth-symbols';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import { timeFormatter } from '@/common/util';
import {
  azureSourceAddressTypes,
  AzureSourceTypeArr,
  azureTargetAddressTypes,
  AzureTargetTypeArr,
} from './add-rule/vendors/azure';

import { awsSourceAddressTypes, AwsSourceTypeArr } from './add-rule/vendors/aws';
import { tcloudSourceAddressTypes, TcloudSourceTypeArr } from './add-rule/vendors/tcloud';
import { huaweiSourceAddressTypes } from './add-rule/vendors/huawei';
import RuleSort from './security-rule-sort.vue';
import { IOverflowTooltipOption } from 'bkui-vue/lib/table/props';
import { showSort } from '../../plugin/security-group/show-sort.plugin';
import { checkVendorInResource } from '../../plugin/security-group/check-vendor-in-resource.plugin';

const props = defineProps({
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
  accountId: {
    type: String as PropType<string>,
  },
});

// use hook
const { t } = useI18n();

const { isShowSecurityRule, handleSecurityRule, SecurityRule } = UseSecurityRule();

const resourceStore = useResourceStore();

const { getBizsId } = useWhereAmI();
const bizId = computed(() => getBizsId());
const updateSign = computed(() => {
  if (bizId.value) return { type: AUTH_BIZ_UPDATE_IAAS_RESOURCE, relation: [bizId.value] };
  return { type: AUTH_UPDATE_IAAS_RESOURCE, relation: [props.accountId] };
});
const HcmAuthComp = resolveComponent('hcm-auth');

const hasEditScopeInBusiness = inject<ComputedRef<boolean>>('hasEditScopeInBusiness');
const hasEditScopeInResource = inject<ComputedRef<boolean>>('hasEditScopeInResource');
const operateTooltipsOption = inject<ComputedRef<IOverflowTooltipOption>>('operateTooltipsOption');

const activeType = ref<'ingress' | 'egress'>('ingress');
const deleteDialogShow = ref(false);
const deleteId = ref(0);
const securityRuleLoading = ref(false);
const fetchUrl = ref<string>(`vendors/${props.vendor}/security_groups/${props.id}/rules/list`);
const dataId = ref('');
const azureDefaultList = ref([]);
const azureDefaultColumns = ref([]);
const isResourcePage: any = inject('isResourcePage');
const show = ref<Boolean>(false);
const filter = ref({ op: 'and', rules: [{ field: 'type', op: 'eq', value: 'ingress' }] });

watch(
  () => activeType.value,
  (v) => {
    // eslint-disable-next-line vue/no-mutating-props
    filter.value.rules[0].value = v;
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
const { datas, pagination, isLoading, handlePageChange, handlePageSizeChange, getList } = useQueryCommonList(
  { filter: filter.value },
  fetchUrl,
  props.vendor === 'tcloud' ? { sort: 'cloud_policy_index', order: 'ASC' } : '',
);

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

const handelSortDone = () => {
  handlePageChange(1);
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
      isShow: props.vendor === 'azure',
    },
    {
      label: t('优先级'),
      field: 'priority',
      isShow: props.vendor === 'huawei' || props.vendor === 'azure',
    },
    {
      label: t('源地址类型'),
      render({ data }: any) {
        const vendor = (props.vendor as VendorEnum) || VendorEnum.TCLOUD;
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
      isShow: props.vendor === 'azure',
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
      isShow: props.vendor === 'azure',
    },
    {
      label: t('类型'),
      field: 'ethertype',
      isShow: props.vendor === 'huawei',
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
      isShow: props.vendor === 'azure',
    },
    {
      label: props.vendor === 'azure' ? t('目标端口协议类型') : t('协议'),
      render({ data }: any) {
        return h('span', {}, [
          data.cloud_service_id ||
            (props.vendor === 'aws' && data.protocol === '-1'
              ? t('ALL')
              : props.vendor === 'huawei' && !data.protocol
              ? t('ALL')
              : props.vendor === 'azure' && data.protocol === '*'
              ? t('ALL')
              : `${data.protocol}`),
        ]);
      },
      isShow: true,
    },
    {
      label: props.vendor === 'azure' ? t('目标协议端口') : t('端口'),
      render({ data }: any) {
        return h('span', {}, [
          data.cloud_service_id ||
            (props.vendor === 'aws' && data.to_port === -1
              ? t('ALL')
              : props.vendor === 'huawei' && !data.port
              ? t('ALL')
              : props.vendor === 'azure' && data.destination_port_range === '*'
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
          props.vendor === 'huawei'
            ? HuaweiSecurityRuleEnum[data.action]
            : props.vendor === 'azure'
            ? AzureSecurityRuleEnum[data.access]
            : props.vendor === 'aws'
            ? t('允许')
            : SecurityRuleEnum[data.action] || '--',
        ]);
      },
      isShow: props.vendor !== 'aws',
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
        return h(
          HcmAuthComp,
          { sign: updateSign.value, tag: 'span', style: { display: 'flex', gap: '8px' } },
          {
            default: ({ noPerm }: { noPerm: boolean }) => [
              withDirectives(
                h(
                  Button,
                  {
                    text: true,
                    theme: 'primary',
                    disabled:
                      noPerm ||
                      props.vendor === 'huawei' ||
                      (isResourcePage.value && !hasEditScopeInResource.value) ||
                      (!isResourcePage.value && !hasEditScopeInBusiness.value),
                    onClick() {
                      handleSecurityRuleDialog(data);
                    },
                  },
                  [t('编辑')],
                ),
                [
                  [
                    bkTooltips,
                    props.vendor === 'huawei'
                      ? { content: '该功能当前未支持', disabled: props.vendor !== 'huawei' }
                      : operateTooltipsOption.value,
                  ],
                ],
              ),
              withDirectives(
                h(
                  Button,
                  {
                    text: true,
                    theme: 'primary',
                    disabled:
                      noPerm ||
                      (isResourcePage.value && !hasEditScopeInResource.value) ||
                      (!isResourcePage.value && !hasEditScopeInBusiness.value),
                    onClick() {
                      deleteDialogShow.value = true;
                      deleteId.value = data.id;
                    },
                  },
                  [t('删除')],
                ),
                [[bkTooltips, operateTooltipsOption.value]],
              ),
            ],
          },
        );
      },
      isShow: !checkVendorInResource(props.vendor),
      showOverflowTooltip: false,
    },
  ].filter(({ isShow }) => !!isShow),
);

// 出站规则列字段
const outColumns: any = computed(() =>
  [
    {
      label: t('名称'),
      field: 'name',
      isShow: props.vendor === 'azure',
    },
    {
      label: t('优先级'),
      field: 'priority',
      isShow: props.vendor === 'huawei' || props.vendor === 'azure',
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
      isShow: props.vendor === 'azure',
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
      isShow: props.vendor === 'azure',
    },
    {
      label: t('源端口'),
      render({ data }: any) {
        return (data.source_port_range === '*' ? 'ALL' : data.source_port_range) || '--';
      },
      isShow: props.vendor === 'azure',
    },
    {
      label: t('目标地址类型'),
      render({ data }: any) {
        const vendor = props.vendor as VendorEnum;
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
      isShow: props.vendor === 'huawei',
    },
    {
      label: props.vendor === 'azure' ? t('目标端口协议类型') : t('协议'),
      render({ data }: any) {
        return h('span', {}, [
          data.cloud_service_id ||
            (props.vendor === 'aws' && data.protocol === '-1'
              ? t('ALL')
              : props.vendor === 'huawei' && !data.protocol
              ? t('ALL')
              : props.vendor === 'azure' && data.protocol === '*'
              ? t('ALL')
              : `${data.protocol}`),
        ]);
      },
      isShow: true,
    },
    {
      label: props.vendor === 'azure' ? t('目标协议端口') : t('端口'),
      render({ data }: any) {
        return h('span', {}, [
          data.cloud_service_id ||
            (props.vendor === 'aws' && data.to_port === -1
              ? t('ALL')
              : props.vendor === 'huawei' && !data.port
              ? t('ALL')
              : props.vendor === 'azure' && data.destination_port_range === '*'
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
          props.vendor === 'huawei'
            ? HuaweiSecurityRuleEnum[data.action]
            : props.vendor === 'azure'
            ? AzureSecurityRuleEnum[data.access]
            : props.vendor === 'aws'
            ? t('允许')
            : SecurityRuleEnum[data.action] || '--',
        ]);
      },
      isShow: props.vendor !== 'aws',
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
        return h(
          HcmAuthComp,
          { sign: updateSign.value, tag: 'span', style: { display: 'flex', gap: '8px' } },
          {
            default: ({ noPerm }: { noPerm: boolean }) => [
              withDirectives(
                h(
                  Button,
                  {
                    text: true,
                    theme: 'primary',
                    disabled:
                      noPerm ||
                      props.vendor === 'huawei' ||
                      (isResourcePage.value && !hasEditScopeInResource.value) ||
                      (!isResourcePage.value && !hasEditScopeInBusiness.value),
                    onClick() {
                      handleSecurityRuleDialog(data);
                    },
                  },
                  [t('编辑')],
                ),
                [
                  [
                    bkTooltips,
                    props.vendor === 'huawei'
                      ? { content: '该功能当前未支持', disabled: props.vendor !== 'huawei' }
                      : operateTooltipsOption.value,
                  ],
                ],
              ),
              withDirectives(
                h(
                  Button,
                  {
                    text: true,
                    theme: 'primary',
                    disabled:
                      noPerm ||
                      (isResourcePage.value && !hasEditScopeInResource.value) ||
                      (!isResourcePage.value && !hasEditScopeInBusiness.value),
                    onClick() {
                      deleteDialogShow.value = true;
                      deleteId.value = data.id;
                    },
                  },
                  [t('删除')],
                ),
                [[bkTooltips, operateTooltipsOption.value]],
              ),
            ],
          },
        );
      },
      isShow: !checkVendorInResource(props.vendor),
      showOverflowTooltip: false,
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
    <section class="rule-main">
      <bk-radio-group v-model="activeType" :disabled="isLoading">
        <bk-radio-button v-for="item in types" :key="item.name" :label="item.name">
          {{ item.label }}
        </bk-radio-button>
      </bk-radio-group>

      <hcm-auth v-if="!checkVendorInResource(props.vendor)" :sign="updateSign" tag="span" v-slot="{ noPerm }">
        <bk-button
          :disabled="
            noPerm || (isResourcePage && !hasEditScopeInResource) || (!isResourcePage && !hasEditScopeInBusiness)
          "
          v-bk-tooltips="operateTooltipsOption"
          theme="primary"
          @click="handleSecurityRuleDialog({})"
        >
          {{ t('新增规则') }}
        </bk-button>
      </hcm-auth>

      <bk-button
        v-if="showSort(props.vendor)"
        icon="plus"
        :disabled="(isResourcePage && !hasEditScopeInResource) || (!isResourcePage && !hasEditScopeInBusiness)"
        v-bk-tooltips="operateTooltipsOption"
        @click="handleSecurityRuleSort"
      >
        {{ t('规则排序') }}
      </bk-button>
    </section>

    <div v-if="props.vendor === 'azure'" class="mb20">
      <h4 class="mt10">Azure默认{{ activeType === 'ingress' ? t('入站') : t('出站') }}规则</h4>
      <bk-table
        class="mt10"
        row-hover="auto"
        :columns="azureDefaultColumns"
        :data="azureDefaultList"
        show-overflow-tooltip
        v-bkloading="{ loading: isLoading }"
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

    <h4 v-if="props.vendor === 'azure'" class="mt10">
      Azure{{ activeType === 'ingress' ? t('入站') : t('出站') }}规则
    </h4>

    <bk-loading :loading="isLoading">
      <bk-table
        v-if="activeType === 'ingress'"
        class="mt20"
        row-hover="auto"
        remote-pagination
        :columns="inColumns"
        :data="datas"
        :pagination="pagination"
        show-overflow-tooltip
        @page-limit-change="handlePageSizeChange"
        @page-value-change="handlePageChange"
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
        :data="datas"
        :pagination="pagination"
        show-overflow-tooltip
        @page-limit-change="handlePageSizeChange"
        @page-value-change="handlePageChange"
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
      v-model:is-show="isShowSecurityRule"
      :loading="securityRuleLoading"
      :id="props.id"
      dialog-width="90vw"
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

    <bk-sideslider v-model:is-show="show" :title="t('规则排序')" width="640" quick-close>
      <template #default>
        <rule-sort
          :id="props.id"
          :filter="filter"
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
