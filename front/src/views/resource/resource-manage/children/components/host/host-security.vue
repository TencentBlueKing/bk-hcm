<script lang="ts" setup>
import { ref, PropType, reactive, h, watch, computed, inject } from 'vue';
import { useI18n } from 'vue-i18n';
import useQueryCommonList from '@/views/resource/resource-manage/hooks/use-query-list-common';
import useColumns from '@/views/resource/resource-manage/hooks/use-columns';
import useSelection from '@/views/resource/resource-manage/hooks/use-selection';
import { useResourceStore, useAccountStore } from '@/store';
import { QueryRuleOPEnum } from '@/typings';
import { GLOBAL_BIZS_KEY, VendorEnum } from '@/common/constant';
import { AUTH_BIZ_UPDATE_IAAS_RESOURCE, AUTH_UPDATE_IAAS_RESOURCE } from '@/constants/auth-symbols';
import routerAction from '@/router/utils/action';
import bus from '@/common/bus';

import { Button, Message, OverflowTitle } from 'bkui-vue';
import SecurityGroupSelectorDialog from '@/components/security-group-selector-dialog/index.vue';
import CopyToClipboard from '@/components/copy-to-clipboard/index.vue';
import ModalFooter from '@/components/modal/modal-footer.vue';

const props = defineProps({
  data: {
    type: Object as PropType<any>,
  },
  isBindBusiness: {
    type: [Boolean, String],
  },
});

const activeType = ref('ingress');
const tableData = ref([]);
const isShow = ref(false);
const securityId = ref(0);
const fetchUrl = ref<string>(`vendors/${props.data.vendor}/security_groups/${securityId.value}/rules/list`);
const fetchFilter = reactive({
  op: QueryRuleOPEnum.AND,
  rules: [{ field: 'type', op: 'eq', value: activeType.value }],
});
const securityFetchFilter = ref<any>({
  filter: {
    op: 'and',
    rules: [
      { field: 'account_id', op: 'eq', value: props.data.account_id },
      { field: 'region', op: 'eq', value: props.data.region },
    ],
  },
});
const isListLoading = ref(false);
const unBindShow = ref(false);
const unBindLoading = ref(false);
const ids = ref([]);
const curreClickId = ref(); // 当前点击的id
const curreSelectId = ref(); // 当前选择的安全组id
const tableItem = ref();
const isBindLoading = ref(false);

const state = reactive<any>({
  datas: [],
  pagination: {
    current: 1,
    limit: 10,
    count: 0,
  },
  isLoading: false,
  handlePageChange: () => {},
  handlePageSizeChange: () => {},
  handleSort: () => {},
  columns: useColumns('securityCommon', false, props.data.vendor).columns,
});

// use hook
const { t } = useI18n();
const resourceStore = useResourceStore();
const accountStore = useAccountStore();
const authVerifyData: any = inject('authVerifyData');

const actionName = computed(() => {
  // 资源下没有业务ID
  return isResourcePage.value ? 'iaas_resource_operate' : 'biz_iaas_resource_operate';
});

const authSign = computed(() => {
  return isResourcePage.value
    ? { type: AUTH_UPDATE_IAAS_RESOURCE, relation: [props.data.account_id] }
    : { type: AUTH_BIZ_UPDATE_IAAS_RESOURCE, relation: [props.data.bk_biz_id] };
});

// 是否显示表格上方的绑定按钮
const isBindBtnShow = computed(() => {
  return props.data.vendor === 'tcloud' || props.data.vendor === 'aws' || props.data.vendor === 'huawei';
});
// 是否显示表格上方的排序按钮
const isSortBtnShow = computed(() => {
  return props.data.vendor === 'tcloud';
});

const selectorDialogState = reactive({ isShow: false, isHidden: true, sortOnly: false });

// 绑定是否多选
const multiple = computed(() => {
  const isMultiple = [VendorEnum.TCLOUD, VendorEnum.AWS];
  return isMultiple.includes(props.data.vendor);
});

// 权限弹窗 bus通知最外层弹出
const showAuthDialog = (authActionName: string) => {
  bus.$emit('auth', authActionName);
};

const isResourcePage = computed(() => {
  // 资源下没有业务ID
  return !accountStore.bizs;
});

const { selections } = useSelection();

watch(
  () => activeType.value,
  (val) => {
    fetchFilter.rules[0].value = val;
    state.columns.forEach((e: any) => {
      if (e.field === 'resource') {
        e.label = val === 'ingress' ? t('来源') : t('目标');
      }
    });
  },
);

watch(
  () => selections.value,
  (val) => {
    const [id] = val.map((e: any) => e.id);
    curreSelectId.value = id;
  },
  { deep: true },
);

const columns: any = [
  {
    label: '安全组ID',
    showOverflowTooltip: false,
    render({ data }: any) {
      const isAzureVendor = data.vendor === 'azure';
      const operateId = isAzureVendor ? data.extension.security_group_id : data.id;
      const displayId = isAzureVendor ? data.extension.cloud_security_group_id : data.cloud_id || '--';
      const show = () => {
        securityId.value = operateId;
        showRuleDialog();
      };
      const jump = () => {
        if (isResourcePage.value) {
          routerAction.open({
            path: '/resource/resource',
            // accountId用于左侧账号列表定位账号
            query: { type: 'security', accountId: data.account_id },
          });
        } else {
          routerAction.open({
            path: '/business/security',
            query: { [GLOBAL_BIZS_KEY]: accountStore.bizs, type: 'security', scene: 'group' },
          });
        }
      };

      return h('div', { class: 'with-operate-cell' }, [
        h(OverflowTitle, { class: 'display-wrap text-link', onClick: show }, displayId),
        h(CopyToClipboard, { class: 'operate-btn', content: displayId }),
        h(
          Button,
          { class: 'operate-btn', text: true, theme: 'primary', onClick: jump },
          h('i', { class: 'hcm-icon bkhcm-icon-jump-fill' }),
        ),
      ]);
    },
  },
  {
    label: '安全组名称',
    render({ data }: any) {
      return h('span', {}, [data.vendor === 'azure' ? data.extension.resource_group_name : data.name || '--']);
    },
  },
  {
    label: t('操作'),
    render({ data }: any) {
      return h('span', {}, [
        data.vendor === 'azure' &&
          h(
            Button,
            {
              text: true,
              theme: 'primary',
              class: 'mr10',
              disabled: data.vendor === 'azure' && data.extension?.cloud_security_group_id, // 如果有安全组id 就不可以绑定
              onClick() {
                if (data.vendor === 'azure') {
                  securityId.value = data.extension.security_group_id;
                  curreClickId.value = data.id;
                } else {
                  securityId.value = data.id;
                }
                handleSecurityBind();
              },
            },
            ['绑定'],
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
                text: true,
                theme: 'primary',
                disabled:
                  (data.vendor === 'azure' && !data.extension?.cloud_security_group_id) ||
                  !authVerifyData.value?.permissionAction[actionName.value], // 如果没有安全组id 就不可以解绑
                onClick() {
                  if (data.vendor === 'azure') {
                    securityId.value = data.extension.security_group_id;
                    curreClickId.value = data.id;
                  } else {
                    securityId.value = data.id;
                  }
                  unBind(data);
                },
              },
              ['解绑'],
            ),
          ],
        ),
      ]);
    },
  },
];

if (props.data.vendor === 'azure') {
  columns.unshift({
    label: t('网络接口名称'),
    field: 'name',
    showOverflowTooltip: false,
    render({ data }: any) {
      const jump = () => {
        if (isResourcePage.value) {
          routerAction.open({
            path: '/resource/resource',
            // accountId用于左侧账号列表定位账号
            query: { type: 'network-interface', accountId: data.account_id },
          });
        } else {
          routerAction.open({ path: '/business/network-interface', query: { [GLOBAL_BIZS_KEY]: accountStore.bizs } });
        }
      };

      return h('div', { class: 'with-operate-cell' }, [
        h(OverflowTitle, { class: 'display-wrap' }, data.name),
        h(CopyToClipboard, { class: 'operate-btn', content: data.name }),
        h(
          Button,
          { class: 'operate-btn', text: true, theme: 'primary', onClick: jump },
          h('i', { class: 'hcm-icon bkhcm-icon-jump-fill' }),
        ),
      ]);
    },
  });
}

// tab 信息
const types = [
  { name: 'ingress', label: t('入站规则') },
  { name: 'egress', label: t('出站规则') },
];

// 主机中安全组的列表
const getSecurityGroupsList = async () => {
  isListLoading.value = true;
  try {
    let res: any = {};
    if (props.data.vendor === 'azure') {
      res = await resourceStore.getNetworkList(props.data.vendor, props.data.id);
    } else {
      res = await resourceStore.getSecurityGroupsListByCvmId(props.data.id);
    }
    tableData.value = res.data;
  } finally {
    isListLoading.value = false;
  }
};

const { datas, pagination, isLoading, handlePageChange, handlePageSizeChange, handleSort } = useQueryCommonList(
  {
    filter: fetchFilter,
  },
  fetchUrl,
);

state.datas = datas;
state.isLoading = isLoading;
state.pagination = pagination;
state.handlePageChange = handlePageChange;
state.handlePageSizeChange = handlePageSizeChange;
state.handleSort = handleSort;
state.columns = useColumns('securityCommon', false, props.data.vendor).columns;

watch(
  () => tableData.value,
  (val) => {
    // 修改filterrules
    if (
      (props.data.vendor === 'aws' || props.data.vendor === 'tcloud' || props.data.vendor === 'huawei') &&
      val?.length
    ) {
      ids.value = val.map((e: any) => e.id);
      securityFetchFilter.value.filter.rules = securityFetchFilter.value.filter.rules.filter((e) => e.field !== 'id');
      securityFetchFilter.value.filter.rules.push({ field: 'id', op: 'nin', value: ids.value });
    }
  },
  { deep: true, immediate: true },
);

if (props.data.vendor === 'aws') {
  securityFetchFilter.value.filter.rules.push({
    field: 'extension.vpc_id',
    op: 'json_eq',
    value: props.data.vpc_ids[0],
  });
}
if (isResourcePage.value) {
  securityFetchFilter.value.filter.rules.push({ field: 'bk_biz_id', op: 'eq', value: -1 }); // 资源下才需要查未绑定的数据
}

const showRuleDialog = async () => {
  isShow.value = true;
  // 获取列表数据
  fetchUrl.value = `vendors/${props.data.vendor}/security_groups/${securityId.value}/rules/list`;
  fetchFilter.rules = [{ field: 'type', op: 'eq', value: activeType.value }];
  if (props.data.vendor === 'huawei') {
    const huaweiColummns = [
      {
        label: t('优先级'),
        field: 'priority',
      },
      {
        label: t('类型'),
        field: 'ethertype',
      },
    ];
    state.columns.unshift(...huaweiColummns);
  } else if (props.data.vendor === 'azure') {
    const awsColummns = [
      {
        label: t('优先级'),
        field: 'priority',
      },
      {
        label: t('名称'),
        field: 'name',
      },
    ];
    state.columns.unshift(...awsColummns);
  }
};

const handleSecurityBind = () => {
  selectorDialogState.isHidden = false;
  selectorDialogState.isShow = true;
  selectorDialogState.sortOnly = false;
};

const handleSecuritySort = () => {
  selectorDialogState.isHidden = false;
  selectorDialogState.isShow = true;
  selectorDialogState.sortOnly = true;
};

// 安全组绑定主机
const handleSecurityConfirm = async (security_group_id: string) => {
  // 暂时只支持一个一个绑定 后期会修改成绑定多个
  let type = 'cvms';
  let params: any = { security_group_id, cvm_id: props.data.id };
  if (props.data.vendor === 'azure') {
    type = 'network_interfaces';
    params = { security_group_id: securityId.value, network_interface_id: curreClickId.value };
  }
  await resourceStore.bindSecurityInfo(type, params);
};

// 解绑弹窗
const unBind = async (dataItem: any) => {
  unBindShow.value = true;
  tableItem.value = dataItem;
};

// 确认解绑
const handleConfirmUnBind = async () => {
  if (tableData.value.length === 1) {
    // 只有一条主机时不能解绑
    unBindShow.value = false;
    return;
  }
  unBindLoading.value = true;
  let type = 'cvms';
  let params: any = { security_group_id: securityId.value, cvm_id: props.data.id };
  if (props.data.vendor === 'azure') {
    type = 'network_interfaces';
    params = { security_group_id: securityId.value, network_interface_id: curreClickId.value };
  }
  try {
    await resourceStore.unBindSecurityInfo(type, params);
    unBindShow.value = false;
    Message({
      message: t('解绑成功'),
      theme: 'success',
    });
    getSecurityGroupsList();
  } finally {
    unBindLoading.value = false;
  }
};

// 关闭弹窗
const handleClose = () => {
  unBindShow.value = false;
};

const handleSelectSecurity = async (security_group_ids: string[]) => {
  try {
    isBindLoading.value = true;
    if (multiple.value) {
      await resourceStore.batchBindSecurityInfo({
        security_group_ids,
        cvm_id: props.data.id,
      });
    } else {
      await handleSecurityConfirm(security_group_ids[0]);
    }

    selectorDialogState.isShow = false;
  } finally {
    isBindLoading.value = false;
  }

  Message({ message: t('绑定成功'), theme: 'success' });
  getSecurityGroupsList();
};
getSecurityGroupsList();
</script>

<template>
  <div class="host-security-container">
    <div class="toolbar" v-if="isBindBtnShow || isSortBtnShow">
      <template v-if="!selectorDialogState.isHidden">
        <security-group-selector-dialog
          v-model="selectorDialogState.isShow"
          :loading="isBindLoading"
          :title="selectorDialogState.sortOnly ? t('安全组排序') : t('选择安全组')"
          :checked="tableData.map((item) => item.id)"
          :biz-id="accountStore.bizs"
          :account-id="props.data.account_id"
          :region="props.data.region"
          :multiple="multiple"
          :vendor="props.data.vendor"
          :sort-only="selectorDialogState.sortOnly"
          @hidden="selectorDialogState.isHidden = true"
          @confirm="handleSelectSecurity"
        />
      </template>
      <hcm-auth :sign="authSign" v-slot="{ noPerm }">
        <bk-button class="button" theme="primary" :disabled="isBindBusiness || noPerm" @click="handleSecurityBind">
          {{ t('绑定') }}
        </bk-button>
      </hcm-auth>
      <hcm-auth :sign="authSign" v-slot="{ noPerm }" v-if="isSortBtnShow">
        <bk-button
          class="button"
          :theme="isBindBtnShow ? '' : 'primary'"
          :disabled="isBindBusiness || noPerm"
          @click="handleSecuritySort"
        >
          {{ t('排序') }}
        </bk-button>
      </hcm-auth>
    </div>
    <bk-table
      class="security-list-table"
      v-bkloading="{ loading: isListLoading }"
      row-hover="auto"
      :columns="columns"
      :data="tableData"
      show-overflow-tooltip
    />
    <bk-dialog
      v-model:is-show="isShow"
      :title="activeType === 'ingress' ? '入站规则' : '出站规则'"
      width="1200"
      :theme="'primary'"
      :dialog-type="'show'"
    >
      <section class="mt20">
        <bk-radio-group v-model="activeType">
          <bk-radio-button v-for="item in types" :key="item.name" :label="item.name">
            {{ item.label }}
          </bk-radio-button>
        </bk-radio-group>
      </section>
      <bk-loading :loading="state.isLoading">
        <bk-table
          class="mt20"
          row-hover="auto"
          :columns="state.columns"
          :data="state.datas"
          remote-pagination
          show-overflow-tooltip
          :pagination="state.pagination"
          @page-limit-change="state.handlePageSizeChange"
          @page-value-change="state.handlePageChange"
          @column-sort="state.handleSort"
        />
      </bk-loading>
    </bk-dialog>

    <bk-dialog :is-show="unBindShow" :title="'确定解绑'" :theme="'primary'" @closed="handleClose">
      <!-- <div>{{ t('确定解绑') }}</div> -->
      <span v-if="tableData.length === 1">
        <span class="error-text">解绑被限制,</span>
        <span>您的主机当前只绑定了1个安全组，为了确保您的主机安全，</span>
        <span class="error-text">请至少保留1个以上的安全组，并确保安全组规则有效</span>
      </span>
      <span v-else>
        <span>
          安全组
          {{ tableItem.vendor === 'azure' ? tableItem.extension.resource_group_name : tableItem.name }} 将从主机上解绑
        </span>
        <span class="error-text">请确保主机上绑定的其他安全组是有效的，避免出现主机安全风险</span>
      </span>

      <template #footer>
        <modal-footer
          :disabled="tableData.length === 1"
          :loading="unBindLoading"
          @confirm="handleConfirmUnBind"
          @closed="handleClose"
        />
      </template>
    </bk-dialog>
  </div>
</template>

<style lang="scss" scoped>
.host-security-container {
  .security-list-table {
    max-height: 100% !important;

    :deep(.with-operate-cell) {
      display: flex;
      align-items: center;
      gap: 4px;

      .display-wrap {
        max-width: calc(100% - 32px);
      }

      .operate-btn {
        width: 12px;
        display: none;
      }

      &:hover {
        .operate-btn {
          display: inline-flex;
        }
      }
    }
  }

  .toolbar {
    display: flex;
    align-items: center;
    gap: 12px;

    .button {
      min-width: 88px;
    }

    ~ .security-list-table {
      margin-top: 16px;
      max-height: calc(100% - 48px) !important;
    }
  }
}

.security-head {
  display: flex;
  align-items: center;
}

.error-text {
  color: #ea3636;
}
</style>
