<script lang="ts" setup>
import DetailInfo from '@/views/resource/resource-manage/common/info/detail-info';
import { useResourceStore } from '@/store';
import { useI18n } from 'vue-i18n';
import { inject, PropType, h, withDirectives, reactive, computed, ComputedRef } from 'vue';
import { bkTooltips, Message, Tag, Button } from 'bkui-vue';
import { useRegionsStore } from '@/store/useRegionsStore';
import { useBusinessMapStore } from '@/store/useBusinessMap';
import { timeFormatter, formatTags } from '@/common/util';
import { FieldList } from '../../../common/info-list/types';
import { MGMT_TYPE_MAP } from '@/constants/security-group';
import BusinessValue from '@/components/display-value/business-value.vue';
import UserValue from '@/components/display-value/user-value.vue';
import UpdateMgmtTypeDialog from '../../dialog/security-group/update-mgmt-type.vue';
import UpdateMgmtAttrSingleDialog from '../../dialog/security-group/update-mgmt-attr-single.vue';
import { SecurityGroupMgmtAttrSingleType, SecurityGroupManageType } from '@/store/security-group';
import { useVerify } from '@/hooks/useVerify';
import { Senarios, useWhereAmI } from '@/hooks/useWhereAmI';
import UsageBizValue from '@/views/resource/resource-manage/children/components/security/usage-biz-value.vue';
import { VendorEnum } from '@/common/constant';
import { IOverflowTooltipOption } from 'bkui-vue/lib/table/props';

const props = defineProps({
  id: {
    type: String as PropType<any>,
  },
  vendor: {
    type: String as PropType<any>,
  },
  loading: {
    type: Boolean as PropType<boolean>,
  },
  detail: {
    type: Object as PropType<any>,
  },
  getDetail: {
    type: Function as PropType<() => void>,
  },
});

const { t } = useI18n();

const isResourcePage = inject<ComputedRef<boolean>>('isResourcePage');
const hasEditScopeInBusiness = inject<ComputedRef<boolean>>('hasEditScopeInBusiness');
const hasEditScopeInResource = inject<ComputedRef<boolean>>('hasEditScopeInResource');
const operateTooltipsOption = inject<ComputedRef<IOverflowTooltipOption>>('operateTooltipsOption');

const resourceStore = useResourceStore();
const { getRegionName } = useRegionsStore();
const { getNameFromBusinessMap } = useBusinessMapStore();
const { whereAmI } = useWhereAmI();

const { handleAuth, authVerifyData } = useVerify();

const authAction = computed(() =>
  whereAmI.value === Senarios.business ? 'biz_iaas_resource_operate' : 'iaas_resource_operate',
);

const mgmtAttrFields: FieldList = [
  {
    name: '管理类型',
    prop: 'mgmt_type',
    render: (val: string) => {
      let theme: '' | 'info' | 'warning';
      theme = val === 'biz' ? 'info' : '';
      if (!val) theme = 'warning';
      return h('div', [
        h(Tag, { theme, radius: '11px' }, MGMT_TYPE_MAP[val]),
        !val &&
          h(
            Button,
            {
              text: true,
              theme: 'primary',
              size: 'small',
              onClick: handleUpdateMgmtAttr,
              style: { marginLeft: '10px', fontSize: '12px' },
            },
            '去确认',
          ),
      ]);
    },
    copy: false,
  },
  {
    name: t('管理业务'),
    prop: 'mgmt_biz_id',
    render: (val: number) => {
      const { bk_biz_id, mgmt_type } = props.detail;
      const editEnabled = bk_biz_id === -1 && mgmt_type === SecurityGroupManageType.BIZ;
      return h('div', [
        val === -1 ? '--' : h(BusinessValue, { value: val }),
        editEnabled &&
          h('i', {
            class: 'icon hcm-icon bkhcm-icon-bianji edit-icon',
            onclick: () => handleUpdateMgmtAttrSingle('mgmt_biz_id'),
          }),
      ]);
    },
    copy: false,
  },
  {
    name: t('分配状态'),
    prop: 'bk_biz_id',
    render: (val: number) => {
      const { mgmt_type } = props.detail;
      if (mgmt_type === SecurityGroupManageType.PLATFORM) {
        return h(Tag, { theme: 'danger' }, '不允许分配');
      }
      if (val === -1) {
        return h(Tag, '未分配');
      }
      return withDirectives(h(Tag, { theme: 'success' }, '已分配'), [
        [bkTooltips, { content: getNameFromBusinessMap(val), disabled: mgmt_type === undefined, theme: 'light' }],
      ]);
    },
    copy: false,
  },
  {
    name: t('使用业务'),
    prop: 'usage_biz_ids',
    render: (val: number[]) => {
      const { bk_biz_id } = props.detail;
      const unassigned = bk_biz_id === -1;
      return h('div', { style: { display: 'flex', alignItems: 'center', width: '100%' } }, [
        h(UsageBizValue, { value: val, style: { width: 'calc(100% - 24px)', flex: 0, whiteSpace: 'nowrap' } }),
        unassigned &&
          h('i', {
            class: 'icon hcm-icon bkhcm-icon-bianji edit-icon',
            onclick: () => handleUpdateMgmtAttrSingle('usage_biz_ids'),
          }),
      ]);
    },
    copy: false,
    showOverflowTips: false,
  },
  {
    name: t('主负责人'),
    prop: 'manager',
    render: (val: string) => {
      const { bk_biz_id } = props.detail;
      const unassigned = bk_biz_id === -1;
      return h('div', [
        h(UserValue, { value: val }),
        unassigned &&
          h('i', {
            class: 'icon hcm-icon bkhcm-icon-bianji edit-icon',
            onclick: () => handleUpdateMgmtAttrSingle('manager'),
          }),
      ]);
    },
    copy: true,
    copyContent: (val: string) => val,
  },
  {
    name: t('备份负责人'),
    prop: 'bak_manager',
    render: (val: string) => {
      const { bk_biz_id } = props.detail;
      const unassigned = bk_biz_id === -1;
      return h('div', [
        h(UserValue, { value: val }),
        unassigned &&
          h('i', {
            class: 'icon hcm-icon bkhcm-icon-bianji edit-icon',
            onclick: () => handleUpdateMgmtAttrSingle('bak_manager'),
          }),
      ]);
    },
    copy: true,
    copyContent: (val: string) => val,
  },
];

const businessMgmtAttrFields: FieldList = [
  {
    name: t('使用业务'),
    prop: 'usage_biz_ids',
    render: (val: number[]) => h(UsageBizValue, { value: val }),
    copy: false,
  },
  {
    name: '管理类型',
    prop: 'mgmt_type',
    render: (val: string) => {
      let theme: '' | 'info' | 'warning';
      theme = val === 'biz' ? 'info' : '';
      if (!val) theme = 'warning';
      return h('div', [h(Tag, { theme, radius: '11px' }, MGMT_TYPE_MAP[val])]);
    },
    copy: false,
  },
  {
    name: t('管理业务'),
    prop: 'mgmt_biz_id',
    render: (val: number) => h(BusinessValue, { value: val }),
    copy: false,
  },
  {
    name: t('分配状态'),
    prop: 'bk_biz_id',
    render: (val: number) => {
      const { mgmt_type } = props.detail;
      if (mgmt_type === SecurityGroupManageType.PLATFORM) {
        return h(Tag, { theme: 'danger' }, '不允许分配');
      }
      if (val === -1) {
        return h(Tag, '未分配');
      }
      return withDirectives(h(Tag, { theme: 'success' }, '已分配'), [
        [bkTooltips, { content: getNameFromBusinessMap(val), disabled: mgmt_type === undefined, theme: 'light' }],
      ]);
    },
    copy: false,
  },
];

const settingInfo = computed(() => {
  const fields: FieldList = [
    { name: 'ID', prop: 'id' },
    { name: t('账号 ID'), prop: 'account_id' },
    {
      name: t('资源名称'),
      prop: 'name',
      edit: !isResourcePage.value
        ? hasEditScopeInBusiness.value && !['azure', 'aws'].includes(props.vendor)
        : hasEditScopeInResource.value && !['azure', 'aws'].includes(props.vendor),
    },
    { name: t('云资源ID'), prop: 'cloud_id' },
    { name: t('云厂商'), prop: 'vendorName' },
    { name: t('地域'), prop: 'region', render: () => getRegionName(props.vendor, props.detail?.region) },
    { name: t('创建时间'), prop: 'created_at', render: (val: string) => timeFormatter(val) },
    { name: t('修改时间'), prop: 'updated_at', render: (val: string) => timeFormatter(val) },
    { name: t('标签'), prop: 'tags', render: (val: any) => formatTags(val) },
    {
      name: t('备注'),
      prop: 'memo',
      edit: !isResourcePage.value
        ? hasEditScopeInBusiness.value && props.vendor !== 'aws'
        : hasEditScopeInResource.value && props.vendor !== 'aws',
    },
  ];
  if ([VendorEnum.TCLOUD, VendorEnum.AWS, VendorEnum.HUAWEI].includes(props.vendor)) {
    fields.splice(8, 0, { name: t('关联CVM实例数'), prop: 'cvm_count' });
    if (VendorEnum.AWS === props.vendor) {
      fields.splice(9, 0, { name: t('所属VPC'), prop: 'vpc_id' }, { name: t('所属云VPC'), prop: 'cloud_vpc_id' });
    }
  } else if (VendorEnum.AZURE === props.vendor) {
    fields.splice(
      7,
      0,
      { name: t('关联网络接口数'), prop: 'network_interface_count' },
      { name: t('关联子网数'), prop: 'subnet_count' },
    );
  }

  return fields;
});

const mgmtAttrDialogState = reactive({
  isShow: false,
  isHidden: true,
});
const handleUpdateMgmtAttr = () => {
  mgmtAttrDialogState.isShow = true;
  mgmtAttrDialogState.isHidden = false;
};

const mgmtAttrSingleDialogState = reactive<{
  isShow: boolean;
  isHidden: boolean;
  field: SecurityGroupMgmtAttrSingleType;
}>({
  isShow: false,
  isHidden: true,
  field: undefined,
});
const handleUpdateMgmtAttrSingle = (field: SecurityGroupMgmtAttrSingleType) => {
  if (!authVerifyData.value?.permissionAction?.[authAction.value]) {
    handleAuth(authAction.value);
    return;
  }
  mgmtAttrSingleDialogState.isShow = true;
  mgmtAttrSingleDialogState.isHidden = false;
  mgmtAttrSingleDialogState.field = field;
};
const handleUpdateMgmtAttrSuccess = () => {
  props.getDetail();
};

const handleChange = async (val: any) => {
  try {
    await resourceStore.updateSecurityInfo(props.id, val);
    Message({
      theme: 'success',
      message: t('更新成功'),
    });
    props.getDetail();
  } catch (error) {}
};
</script>

<template>
  <bk-loading :loading="props.loading">
    <h3 class="info-title">安全组信息</h3>
    <div class="wrap-info">
      <detail-info
        :fields="settingInfo"
        :detail="props.detail"
        @change="handleChange"
        label-width="130px"
        global-copyable
      />
    </div>
    <template v-if="isResourcePage">
      <h3 class="info-title">资产归属</h3>
      <div class="wrap-info">
        <detail-info :fields="mgmtAttrFields" :detail="props.detail" @change="handleChange" label-width="130px" />
      </div>
    </template>
    <template v-else>
      <h3 class="info-title">
        使用范围
        <bk-button
          text
          size="small"
          theme="primary"
          style="font-weight: 400; margin-left: 12px"
          :disabled="!hasEditScopeInBusiness"
          :class="{ 'hcm-no-permision-text-btn': !authVerifyData?.permissionAction?.[authAction] }"
          v-bk-tooltips="operateTooltipsOption"
          @click="handleUpdateMgmtAttrSingle('usage_biz_ids')"
        >
          <i class="icon hcm-icon bkhcm-icon-bianji edit-icon" />
          <span style="margin-left: 4px">编辑</span>
        </bk-button>
      </h3>
      <div class="wrap-info">
        <detail-info
          :fields="businessMgmtAttrFields"
          :detail="props.detail"
          @change="handleChange"
          :col="1"
          label-width="130px"
        />
      </div>
    </template>
  </bk-loading>
  <!-- 确认管理类型 -->
  <template v-if="isResourcePage && !mgmtAttrDialogState.isHidden">
    <update-mgmt-type-dialog
      v-model="mgmtAttrDialogState.isShow"
      :detail="props.detail"
      @hidden="mgmtAttrDialogState.isHidden = true"
      @success="handleUpdateMgmtAttrSuccess"
    />
  </template>
  <!-- 编辑单个管理属性 -->
  <template v-if="!mgmtAttrSingleDialogState.isHidden">
    <update-mgmt-attr-single-dialog
      v-model="mgmtAttrSingleDialogState.isShow"
      :detail="props.detail"
      :field="mgmtAttrSingleDialogState.field"
      @hidden="mgmtAttrSingleDialogState.isHidden = true"
      @success="handleUpdateMgmtAttrSuccess"
    />
  </template>
</template>
