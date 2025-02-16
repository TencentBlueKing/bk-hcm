<script setup lang="ts">
import type {
  DoublePlainObject,
  // PlainObject,
  FilterType,
} from '@/typings/resource';
import { GcpTypeEnum } from '@/typings';
import { Button, InfoBox, Message, Tag, bkTooltips } from 'bkui-vue';
import { useResourceStore, useAccountStore } from '@/store';
import { ref, h, PropType, watch, reactive, defineExpose, computed, withDirectives, nextTick } from 'vue';

import { useI18n } from 'vue-i18n';
import { useRouter, useRoute } from 'vue-router';
import useQueryCommonList from '@/views/resource/resource-manage/hooks/use-query-list-common';
import useColumns from '@/views/resource/resource-manage/hooks/use-columns';
import useFilter from '@/views/resource/resource-manage/hooks/use-filter';
import { useRegionsStore } from '@/store/useRegionsStore';
import { VendorEnum, VendorMap } from '@/common/constant';
import { cloneDeep } from 'lodash-es';
import { useBusinessMapStore } from '@/store/useBusinessMap';
import { useResourceAccountStore } from '@/store/useResourceAccountStore';
import useSelection from '../../hooks/use-selection';
import { BatchDistribution, DResourceType } from '@/views/resource/resource-manage/children/dialog/batch-distribution';
import { TemplateTypeMap } from '../dialog/template-dialog';
import { Senarios, useWhereAmI } from '@/hooks/useWhereAmI';
import http from '@/http';
import { timeFormatter, formatTags } from '@/common/util';
import { storeToRefs } from 'pinia';

import SecurityGroupChangeConfirmDialog from '../dialog/security-group/change-confirm.vue';
import SecurityGroupSingleDeleteDialog from '../dialog/security-group/single-delete.vue';
import SecurityGroupAssignDialog from '../dialog/security-group/assign.vue';
import SecurityGroupUpdateMgmtAttrDialog from '../dialog/security-group/update-mgmt-attr.vue';
import { MGMT_TYPE_MAP } from '@/constants/security-group';
import { ISecurityGroupOperateItem, useSecurityGroupStore, SecurityGroupManageType } from '@/store/security-group';

const { BK_HCM_AJAX_URL_PREFIX } = window.PROJECT_CONFIG;

const props = defineProps({
  filter: {
    type: Object as PropType<FilterType>,
  },
  isResourcePage: {
    type: Boolean,
  },
  authVerifyData: {
    type: Object as PropType<any>,
  },
  whereAmI: {
    type: String,
  },
});

// use hooks
const { t } = useI18n();

const { getRegionName } = useRegionsStore();
const securityGroupStore = useSecurityGroupStore();
const { getNameFromBusinessMap } = useBusinessMapStore();
const router = useRouter();
const route = useRoute();
const { whereAmI } = useWhereAmI();

const resourceAccountStore = useResourceAccountStore();
const { currentVendor, currentAccountVendor } = storeToRefs(resourceAccountStore);
const resourceStore = useResourceStore();
const accountStore = useAccountStore();

const activeType = ref('group');
const fetchUrl = ref<string>('security_groups/list');

const emit = defineEmits(['auth', 'handleSecrityType', 'edit', 'editTemplate']);
const { columns, generateColumnsSettings } = useColumns('group');

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
  columns,
  params: {
    fetchUrl: 'security_groups',
    columns: 'group',
  },
});

const templateData = ref([]);

const { searchData, searchValue, filter } = useFilter(props);

const { datas, pagination, isLoading, handlePageChange, handlePageSizeChange, getList } = useQueryCommonList(
  {
    ...props,
    filter: filter.value,
  },
  fetchUrl,
  // {
  //   handleAsyncRequest: async (data: any[]) => {
  //     // 安全组需要异步加载一些关联资源数据
  //     if (activeType.value !== 'group') return;

  //     const security_group_ids: string[] = data.map((item: any) => item.id);
  //     const [relResList, ruleCountMap] = await Promise.all([
  //       securityGroupStore.queryRelatedResources(security_group_ids),
  //       securityGroupStore.batchQueryRuleCount(security_group_ids),
  //     ]);

  //     return data.map((item) => {
  //       const relResItem = relResList.find((relRes) => relRes.id === item.id);
  //       const rel_res_count = relResItem.resources.reduce((acc, cur) => acc + cur.count, 0);
  //       const res_res_types = relResItem.resources.flatMap(({ res_name }) => res_name);
  //       const rule_count = ruleCountMap[item.id];
  //       return { ...item, rel_res_count, res_res_types, rule_count };
  //     });
  //   },
  // },
);

const selectSearchData = computed(() => {
  let searchName = '安全组 ID';
  switch (activeType.value) {
    case 'group': {
      searchName = '安全组 ID';
      break;
    }
    case 'gcp': {
      searchName = '防火墙 ID';
      break;
    }
    case 'template': {
      searchName = '模板 ID';
      break;
    }
  }
  return [
    {
      name: searchName,
      id: 'cloud_id',
    },
    ...searchData.value,
    ...(activeType.value === 'template'
      ? []
      : [
          {
            name: '云地域',
            id: 'region',
          },
        ]),
  ];
});

// eslint-disable-next-line max-len
state.datas = datas;
state.isLoading = isLoading;
state.pagination = pagination;
state.handlePageChange = handlePageChange;
state.handlePageSizeChange = handlePageSizeChange;

watch(
  () => datas.value,
  async (data) => {
    if (activeType.value === 'template') {
      templateData.value = data;
      const ids = data.map(({ id }) => id);
      if (!ids.length) return;
      const url = `${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud${
        whereAmI.value === Senarios.business ? `/bizs/${accountStore.bizs}` : ''
      }/argument_templates/instance/rule/list`;
      const res = await http.post(url, {
        ids,
        bk_biz_id: whereAmI.value === Senarios.business ? accountStore.bizs : undefined,
      });
      const dataMap = new Map<any, { id: any; instance_num: any; rule_num: any }>(
        res.data.map((element: { id: any; instance_num: any; rule_num: any }) => [element.id, element]),
      );
      templateData.value.forEach((item) => {
        const foundElement = dataMap.get(item.id);
        if (foundElement) {
          item.instance_num = foundElement?.instance_num;
          item.rule_num = foundElement?.rule_num;
        } else {
          item.instance_num = '--';
          item.rule_num = '--';
        }
      });
    }
  },
  {
    deep: true,
  },
);

const handleSwtichType = async (type: string) => {
  if (type === 'gcp') {
    fetchUrl.value = 'vendors/gcp/firewalls/rules/list';
    state.params.fetchUrl = 'vendors/gcp/firewalls/rules';
    state.params.columns = 'gcp';
  } else if (type === 'group') {
    fetchUrl.value = 'security_groups/list';
    state.params.fetchUrl = 'security_groups';
    state.params.columns = 'group';
  } else if (type === 'template') {
    fetchUrl.value = 'argument_templates/list';
    state.params.fetchUrl = 'argument_templates';
    state.params.columns = 'template';
  }
  emit('handleSecrityType', type);
  router.replace({ query: Object.assign({}, route.query, { type: 'security', scene: type }) });
};

// 抛出请求数据的方法，新增成功使用
const fetchComponentsData = () => {
  getList();
};

// 初始化
getList();

defineExpose({ fetchComponentsData });
const isRowSelectEnable = ({ row, isCheckAll }: DoublePlainObject) => {
  if (isCheckAll) return true;
  return isCurRowSelectEnable(row);
};
const isCurRowSelectEnable = (row: any) => {
  if (!props.isResourcePage) return true;
  if (row.id) {
    return row.bk_biz_id === -1;
  }
};
const { selections, handleSelectionChange, resetSelections } = useSelection();

const groupColumns = [
  { type: 'selection', width: 30, minWidth: 30, onlyShowOnList: true },
  {
    label: '安全组 ID',
    field: 'cloud_id',
    isDefaultShow: true,
    render({ data }: any) {
      return h(
        Button,
        {
          text: true,
          theme: 'primary',
          disabled: data.bk_biz_id !== -1 && props.isResourcePage,
          onClick() {
            const routeInfo: any = { query: { ...route.query, id: data.id, vendor: data.vendor } };
            // 业务下
            if (route.path.includes('business')) {
              routeInfo.query.bizs = accountStore.bizs;
              Object.assign(routeInfo, { name: 'securityBusinessDetail' });
            } else {
              Object.assign(routeInfo, { name: 'resourceDetail', params: { type: 'security' } });
            }
            router.push(routeInfo);
          },
        },
        [data.cloud_id || '--'],
      );
    },
    width: 120,
  },
  {
    label: '名称',
    field: 'name',
    isDefaultShow: true,
    width: 120,
  },
  {
    label: t('地域'),
    field: 'region',
    isDefaultShow: true,
    render: ({ data }: any) => {
      return getRegionName(data.vendor, data.region);
    },
    width: 150,
  },
  {
    label: t('云厂商'),
    field: 'vendor',
    filter: {
      list: Object.entries(VendorMap).map(([value, text]) => ({ value, text })),
    },
    render: ({ cell }: any) => VendorMap[cell],
    width: 90,
  },
  {
    label: t('备注'),
    field: 'memo',
    isDefaultShow: true,
    render: ({ cell }: any) => (cell ? cell : '--'),
    width: 120,
  },
  {
    label: t('规则个数'),
    field: 'rule_count',
    width: 90,
  },
  {
    label: t('关联实例数'),
    field: 'rel_res_count',
    width: 120,
  },
  {
    label: t('关联的资源类型'),
    field: 'res_res_types',
    filter: true,
    width: 150,
    render: ({ cell }: { cell: string[] }) => (cell ? cell.map((res_name) => h(Tag, null, res_name)) : '--'),
  },
  {
    label: t('使用业务'),
    field: 'usage_biz_ids',
    filter: true,
    isDefaultShow: true,
    width: 100,
    render: ({ cell }: any) => cell?.join(',') ?? '--',
  },
  {
    label: t('管理类型'),
    field: 'mgmt_type',
    filter: {
      list: Object.entries(MGMT_TYPE_MAP).map(([value, text]) => ({ value, text })),
    },
    isDefaultShow: true,
    width: 100,
    render: ({ cell }: any) => {
      let theme: '' | 'info' | 'warning';
      theme = cell === 'biz' ? 'info' : '';
      if (!cell) theme = 'warning';
      return h(Tag, { theme, radius: '11px' }, MGMT_TYPE_MAP[cell]);
    },
  },
  {
    label: t('管理业务'),
    field: 'mgmt_biz_id',
    filter: true,
    isDefaultShow: true,
    width: 100,
    render: ({ cell }: any) => (cell ? getNameFromBusinessMap(cell) : '--'),
  },
  {
    label: t('主负责人'),
    field: 'manager',
    width: 100,
  },
  {
    label: t('备份负责人'),
    field: 'bak_manager',
    width: 100,
  },
  {
    label: t('标签'),
    field: 'tags',
    isDefaultShow: true,
    render: ({ cell }: any) => formatTags(cell),
    width: 100,
  },
  {
    label: '是否分配',
    field: 'bk_biz_id',
    notDisplayedInBusiness: true,
    isDefaultShow: true,
    render: ({ data, cell }: { data: any; cell: number }) => {
      const { mgmt_type } = data;

      let displayValue = cell === -1 ? t('未分配') : t('已分配');
      let theme: '' | 'success' | 'danger' = cell === -1 ? '' : 'success';

      // 不可分配的情况
      if (theme === '' && (!mgmt_type || mgmt_type === 'platform')) {
        displayValue = t('不允许分配');
        theme = 'danger';
      }

      return withDirectives(h(Tag, { theme }, displayValue), [
        [bkTooltips, { content: getNameFromBusinessMap(cell), disabled: theme !== 'success', theme: 'light' }],
      ]);
    },
    width: 120,
  },
  {
    label: t('操作'),
    field: 'operate',
    isDefaultShow: true,
    width: 120,
    fixed: 'right',
    render({ data }: any) {
      const isAssigned = data.bk_biz_id !== -1 && props.isResourcePage;

      const authMap = {
        rule: props.isResourcePage ? 'iaas_resource_operate' : 'biz_iaas_resource_operate',
        clone: props.isResourcePage ? 'iaas_resource_create' : 'biz_iaas_resource_create',
        delete: props.isResourcePage ? 'iaas_resource_delete' : 'biz_iaas_resource_delete',
      };

      const operationList = [
        {
          type: 'rule',
          name: t('配置规则'),
          auth: authMap.rule,
          disabled: !props.authVerifyData?.permissionAction[authMap.rule] || isAssigned,
          handleClick: async () => {
            isChangeEffectConfirmDialogShow.value = true;
            const resData = await securityGroupStore.queryRelatedResources([data.id]);
            const { resources } = resData[0] ?? {};
            currentSecurityGroup.value = { ...data, resources };
          },
        },
        {
          type: 'clone',
          name: t('克隆'),
          auth: authMap.clone,
          disabled: !props.authVerifyData?.permissionAction[authMap.clone] || isAssigned,
          handleClick: () => {},
          hidden: props.isResourcePage,
        },
        {
          type: 'delete',
          name: t('删除'),
          auth: authMap.delete,
          disabled: !props.authVerifyData?.permissionAction[authMap.delete] || isAssigned,
          handleClick: () => handleDeleteSG(data),
        },
      ];

      return h(
        'div',
        { class: 'operation-cell' },
        operationList.map(({ name, auth, disabled, handleClick, hidden }) => {
          if (hidden) return null;
          return h(
            'span',
            { onClick: () => emit('auth', auth) },
            h(Button, { text: true, theme: 'primary', disabled, onClick: handleClick }, name),
          );
        }),
      );
    },
  },
].filter((item) => {
  if (Senarios.business === whereAmI.value) return !item.notDisplayedInBusiness;
  return true;
});

const groupSettings = generateColumnsSettings(groupColumns);

const gcpColumns = [
  { type: 'selection', width: 30, minWidth: 30, onlyShowOnList: true },
  {
    label: '防火墙 ID	',
    field: 'cloud_id',
    width: '120',
    sort: true,
    isDefaultShow: true,
    render({ data }: any) {
      return h(
        Button,
        {
          text: true,
          theme: 'primary',
          onClick() {
            const routeInfo: any = {
              query: {
                ...route.query,
                id: data.id,
              },
            };
            // 业务下
            if (route.path.includes('business')) {
              Object.assign(routeInfo, {
                name: 'gcpBusinessDetail',
              });
            } else {
              Object.assign(routeInfo, {
                name: 'resourceDetail',
                params: {
                  type: 'gcp',
                },
              });
            }
            router.push(routeInfo);
          },
        },
        [data.cloud_id || '--'],
      );
    },
  },
  // {
  //   label: t('资源 ID'),
  //   field: 'account_id',
  //   sort: true,
  // },
  {
    label: '防火墙名称',
    field: 'name',
    sort: true,
    isDefaultShow: true,
  },
  {
    label: t('云厂商'),
    field: 'vendor',
    sort: true,
    isDefaultShow: true,
    render() {
      return h('span', {}, [t('谷歌云')]);
    },
  },
  {
    label: '所属VPC',
    field: 'vpc_id',
    sort: true,
    isDefaultShow: true,
  },
  {
    label: t('优先级'),
    field: 'priority',
    sort: true,
    isDefaultShow: true,
  },
  {
    label: '流量方向',
    field: 'type',
    sort: true,
    isDefaultShow: true,
    render({ data }: any) {
      return h('span', {}, [GcpTypeEnum[data.type]]);
    },
  },
  {
    label: t('目标'),
    field: 'target_tags',
    sort: true,
    isDefaultShow: true,
    render({ data }: any) {
      return h('span', {}, [data.target_tags || data.target_service_accounts || '--']);
    },
  },
  // {
  //   label: t('过滤条件'),
  //   field: '',
  // },
  {
    label: t('协议/端口'),
    field: 'allowed_denied',
    sort: true,
    isDefaultShow: true,
    render({ data }: any) {
      return h(
        'span',
        {},
        data?.allowed || data?.denied
          ? (data?.allowed || data?.denied).map((e: any) => {
              return h('div', {}, `${e.protocol}:${e.port}`);
            })
          : '--',
      );
    },
  },
  {
    label: '是否分配',
    field: 'bk_biz_id',
    sort: true,
    notDisplayedInBusiness: true,
    isDefaultShow: true,
    render: ({ data, cell }: { data: { bk_biz_id: number }; cell: number }) => {
      return withDirectives(
        h(
          Tag,
          {
            theme: data.bk_biz_id === -1 ? false : 'success',
          },
          [data.bk_biz_id === -1 ? '未分配' : '已分配'],
        ),
        [
          [
            bkTooltips,
            {
              content: getNameFromBusinessMap(cell),
              disabled: !cell || cell === -1,
              theme: 'light',
            },
          ],
        ],
      );
    },
  },
  {
    label: '所属业务',
    field: 'bk_biz_id2',
    notDisplayedInBusiness: true,
    render({ data }: any) {
      return h('span', {}, [data.bk_biz_id === -1 ? t('未分配') : getNameFromBusinessMap(data.bk_biz_id)]);
    },
  },
  {
    label: t('创建时间'),
    field: 'created_at',
    sort: true,
    render: ({ cell }: { cell: string }) => timeFormatter(cell),
  },
  {
    label: t('修改时间'),
    field: 'updated_at',
    sort: true,
    render: ({ cell }: { cell: string }) => timeFormatter(cell),
  },
  {
    label: t('操作'),
    field: 'operator',
    isDefaultShow: true,
    render({ data }: any) {
      return h('span', {}, [
        h(
          'span',
          {
            onClick() {
              emit('auth', props.isResourcePage ? 'iaas_resource_operate' : 'biz_iaas_resource_operate');
            },
          },
          [
            h(
              Button,
              {
                text: true,
                theme: 'primary',
                disabled:
                  !props.authVerifyData?.permissionAction[
                    props.isResourcePage ? 'iaas_resource_operate' : 'biz_iaas_resource_operate'
                  ] ||
                  (data.bk_biz_id !== -1 && props.isResourcePage),
                onClick() {
                  emit('edit', cloneDeep(data));
                },
              },
              [t('编辑')],
            ),
          ],
        ),
        h(
          'span',
          {
            onClick() {
              emit('auth', props.isResourcePage ? 'iaas_resource_operate' : 'biz_iaas_resource_operate');
            },
          },
          [
            h(
              Button,
              {
                class: 'ml10',
                text: true,
                disabled:
                  !props.authVerifyData?.permissionAction[
                    props.isResourcePage ? 'iaas_resource_delete' : 'biz_iaas_resource_delete'
                  ] ||
                  (data.bk_biz_id !== -1 && props.isResourcePage),
                theme: 'primary',
                onClick() {
                  securityHandleShowDelete(data);
                },
              },
              [t('删除')],
            ),
          ],
        ),
      ]);
    },
  },
].filter((item) => {
  if (Senarios.business === whereAmI.value) return !item.notDisplayedInBusiness;
  return true;
});

const gcpSettings = generateColumnsSettings(gcpColumns);

const templateColumns = [
  { type: 'selection', width: 30, minWidth: 30, onlyShowOnList: true },
  {
    label: '模板ID',
    field: 'cloud_id',
    isDefaultShow: true,
    render: ({ data }: any) => {
      return h(
        Button,
        {
          text: true,
          theme: 'primary',
          onClick() {
            const routeInfo: any = {
              query: {
                ...route.query,
                id: data.cloud_id,
              },
            };
            if (route.path.includes('business')) {
              Object.assign(routeInfo, {
                name: 'templateBusinessDetail',
              });
            } else {
              Object.assign(routeInfo, {
                name: 'resourceDetail',
                params: {
                  type: 'template',
                },
              });
            }
            router.push(routeInfo);
          },
        },
        [data.cloud_id],
      );
    },
  },
  {
    label: '模板名称',
    field: 'name',
    isDefaultShow: true,
  },
  {
    label: '云厂商',
    field: 'vendor',
    render: ({ cell }: any) => VendorMap[cell],
    isDefaultShow: true,
  },
  {
    label: '类型',
    field: 'type',
    render: ({ cell }: any) => TemplateTypeMap[cell],
    isDefaultShow: true,
  },
  {
    label: '关联实例数',
    field: 'instance_num',
    isDefaultShow: true,
  },
  {
    label: '规则数',
    field: 'rule_num',
    isDefaultShow: true,
  },
  {
    label: '是否分配',
    field: 'bk_biz_id',
    isDefaultShow: true,
    notDisplayedInBusiness: true,
    render: ({ data }: { data: { bk_biz_id: number }; cell: number }) => {
      return withDirectives(
        h(
          Tag,
          {
            theme: data.bk_biz_id === -1 ? false : 'success',
          },
          [data.bk_biz_id === -1 ? '未分配' : '已分配'],
        ),
        [
          [
            bkTooltips,
            {
              content: getNameFromBusinessMap(data.bk_biz_id),
              disabled: !data.bk_biz_id || data.bk_biz_id === -1,
              theme: 'light',
            },
          ],
        ],
      );
    },
  },
  {
    field: 'actions',
    label: '操作',
    isDefaultShow: true,
    render({ data }: any) {
      return h('span', {}, [
        h(
          Button,
          {
            text: true,
            theme: 'primary',
            onClick() {
              emit('editTemplate', {
                type: data.type,
                templates: data.templates,
                group_templates: data.group_templates,
                name: data.name,
                bk_biz_id: data.bk_biz_id,
                id: data.id,
                account_id: data.account_id,
              });
            },
          },
          ['编辑'],
        ),
        h(
          Button,
          {
            class: 'ml10',
            text: true,
            theme: 'primary',
            onClick() {
              securityHandleShowDelete(data);
            },
          },
          [t('删除')],
        ),
      ]);
    },
  },
]
  .filter(
    ({ field }) =>
      (whereAmI.value === Senarios.resource && !['actions'].includes(field)) || whereAmI.value !== Senarios.resource,
  )
  .filter((item) => {
    if (Senarios.business === whereAmI.value) return !item.notDisplayedInBusiness;
    return true;
  });

const templateSettings = generateColumnsSettings(templateColumns);

const isAllVendor = computed(() => {
  return !currentVendor.value && !currentAccountVendor.value;
});
const isGcpVendor = computed(() => {
  return [currentVendor.value, currentAccountVendor.value].includes(VendorEnum.GCP);
});
const isTcloudVendor = computed(() => {
  return [currentVendor.value, currentAccountVendor.value].includes(VendorEnum.TCLOUD);
});
const types = computed(() => {
  const securityType = { name: 'group', label: t('安全组') };
  const gcpType = { name: 'gcp', label: t('GCP防火墙规则') };
  const templateType = { name: 'template', label: '参数模板' };
  if (whereAmI.value === Senarios.business || isAllVendor.value) {
    return [securityType, gcpType, templateType];
  }
  if (isGcpVendor.value) {
    return [gcpType];
  }
  if (isTcloudVendor.value) {
    return [securityType, templateType];
  }
  return [securityType];
});
watch(types, () => {
  if (isGcpVendor.value) {
    activeType.value = 'gcp';
  } else {
    activeType.value = 'group';
  }
});

// 状态保持
watch(
  () => activeType.value,
  (v) => {
    state.isLoading = true;
    state.pagination.current = 1;
    state.pagination.limit = 10;
    handleSwtichType(v);
    resetSelections();
  },
  {
    immediate: true,
  },
);

const isChangeEffectConfirmDialogShow = ref(false);
const handleChangeEffectConfirm = () => {
  const routeInfo: any = {
    query: { activeTab: 'rule', id: currentSecurityGroup.value.id, vendor: currentSecurityGroup.value.vendor },
  };
  // 业务下
  if (route.path.includes('business')) {
    Object.assign(routeInfo, { name: 'securityBusinessDetail' });
  } else {
    Object.assign(routeInfo, { name: 'resourceDetail', params: { type: 'security' } });
  }
  router.push(routeInfo);
};

const currentSecurityGroup = ref<ISecurityGroupOperateItem>(null);

const isSecurityGroupSingleDeleteDialogShow = ref(false);
const handleDeleteSG = async (rowData: any) => {
  isSecurityGroupSingleDeleteDialogShow.value = true;
  const [relResList, ruleCountMap] = await Promise.all([
    securityGroupStore.queryRelatedResources([rowData.id]),
    securityGroupStore.batchQueryRuleCount([rowData.id]),
  ]);
  const rule_count = ruleCountMap[rowData.id] ?? 0;
  const { resources } = relResList[0] ?? {};
  currentSecurityGroup.value = { ...rowData, resources, rule_count };
};

const securityHandleShowDelete = (data: any) => {
  InfoBox({
    title: '请确认是否删除',
    subTitle: `将删除【${data.cloud_id}${data.name ? ` - ${data.name}` : ''}】`,
    theme: 'danger',
    headerAlign: 'center',
    footerAlign: 'center',
    contentAlign: 'center',
    extCls: 'delete-resource-infobox',
    async onConfirm() {
      let type = '';
      switch (activeType.value) {
        case 'group': {
          type = 'security_groups';
          break;
        }
        case 'gcp': {
          type = 'vendors/gcp/firewalls/rules';
          break;
        }
        case 'template': {
          type = 'argument_templates';
          break;
        }
      }
      await resourceStore.deleteBatch(type, { ids: [data.id] });
      getList();
      Message({
        message: t('删除成功'),
        theme: 'success',
      });
    },
  });
};

const securityGroupSelectedState = computed(() => {
  const state = {
    bizTypeCount: 0,
    unknownTypeCount: 0,
    accountUnique: true,
    mgmtAttrEmptyCount: 0,
  };
  selections.value.forEach((item) => {
    state.bizTypeCount += item.mgmt_type === SecurityGroupManageType.BIZ ? 1 : 0;
    state.unknownTypeCount += item.mgmt_type === SecurityGroupManageType.UNKNOWN ? 1 : 0;
    state.mgmtAttrEmptyCount +=
      item.manager || item.bak_manager || item.usage_biz_id || item.mgmt_biz_id !== -1 ? 0 : 1;
    if (state.accountUnique) {
      state.accountUnique = item.account_id === selections.value[0].account_id;
    }
  });
  return state;
});
const isAllBizType = computed(() => securityGroupSelectedState.value.bizTypeCount === selections.value.length);
const assignButtonDisabled = computed(() => !selections.value.length || !isAllBizType.value);
const isAllUnknownType = computed(() => securityGroupSelectedState.value.unknownTypeCount === selections.value.length);
const isAllMgmtAttrEmpty = computed(
  () => securityGroupSelectedState.value.mgmtAttrEmptyCount === selections.value.length,
);
const isSameAccount = computed(() => securityGroupSelectedState.value.accountUnique);
const updateMgmtAttrButtonDisabled = computed(
  () => !selections.value.length || !isAllUnknownType.value || !isAllMgmtAttrEmpty.value || !isSameAccount.value,
);

const securityGroupAssignDialogState = reactive({
  isShow: false,
  isHidden: true,
});
const handleSecurityGroupAssign = () => {
  securityGroupAssignDialogState.isShow = true;
  securityGroupAssignDialogState.isHidden = false;
};

const securityGroupMgmtAttrEditDialogState = reactive({
  isShow: false,
  isHidden: true,
});
const handleSecurityGroupUpdateMgmtAttr = () => {
  securityGroupMgmtAttrEditDialogState.isShow = true;
  securityGroupMgmtAttrEditDialogState.isHidden = false;
};

const handleSecurityGroupOperationSuccess = () => {
  getList();

  // 确保dialog销毁后再清空selections数据，避免dialog中依赖selections的逻辑被非预期执行
  nextTick(() => {
    resetSelections();
  });
};
</script>

<template>
  <div class="security-manager-page">
    <div class="flex-row align-items-center toolbar">
      <bk-radio-group v-model="activeType" :disabled="state.isLoading">
        <bk-radio-button v-for="item in types" :key="item.name" :label="item.name">
          {{ item.label }}
        </bk-radio-button>
      </bk-radio-group>
      <slot></slot>
      <template v-if="isResourcePage">
        <bk-button
          v-if="activeType === 'group'"
          :disabled="assignButtonDisabled"
          v-bk-tooltips="{
            content: '管理类型需全部为“业务管理”',
            disabled: !selections.length || !assignButtonDisabled,
          }"
          @click="handleSecurityGroupAssign"
        >
          批量分配
        </bk-button>
        <BatchDistribution
          v-else
          :selections="selections"
          :type="activeType === 'template' ? DResourceType.templates : DResourceType.firewall"
          :get-data="
            () => {
              getList();
              resetSelections();
            }
          "
        />
        <bk-button
          v-if="activeType === 'group'"
          :disabled="updateMgmtAttrButtonDisabled"
          v-bk-tooltips="{
            content: !isAllUnknownType
              ? '管理类型需全部为“未确认”'
              : !isAllMgmtAttrEmpty
              ? '资产归属字段需全为空'
              : '必须属于同一账号',
            disabled: !selections.length || !updateMgmtAttrButtonDisabled,
          }"
          @click="handleSecurityGroupUpdateMgmtAttr"
        >
          批量添加资产归属
        </bk-button>
      </template>
      <bk-search-select
        class="search-filter search-selector-container"
        clearable
        :conditions="[]"
        :data="selectSearchData"
        v-model="searchValue"
      />
    </div>

    <bk-loading :key="activeType" :loading="state.isLoading" opacity="1">
      <bk-table
        v-if="activeType === 'group'"
        :settings="groupSettings"
        row-hover="auto"
        remote-pagination
        :pagination="state.pagination"
        :columns="groupColumns"
        :data="state.datas"
        show-overflow-tooltip
        :is-row-select-enable="isRowSelectEnable"
        @selection-change="(selections: any) => handleSelectionChange(selections, isCurRowSelectEnable)"
        @select-all="(selections: any) => handleSelectionChange(selections, isCurRowSelectEnable, true)"
        @page-limit-change="state.handlePageSizeChange"
        @page-value-change="state.handlePageChange"
        @column-sort="state.handleSort"
      />

      <bk-table
        v-else-if="activeType === 'gcp'"
        :settings="gcpSettings"
        row-hover="auto"
        remote-pagination
        :pagination="state.pagination"
        :columns="gcpColumns"
        :data="state.datas"
        show-overflow-tooltip
        :is-row-select-enable="isRowSelectEnable"
        @selection-change="(selections: any) => handleSelectionChange(selections, isCurRowSelectEnable)"
        @select-all="(selections: any) => handleSelectionChange(selections, isCurRowSelectEnable, true)"
        @page-limit-change="state.handlePageSizeChange"
        @page-value-change="state.handlePageChange"
        @column-sort="state.handleSort"
      />

      <bk-table
        v-else-if="activeType === 'template'"
        :settings="templateSettings"
        row-hover="auto"
        remote-pagination
        :pagination="state.pagination"
        :columns="templateColumns"
        :data="templateData"
        show-overflow-tooltip
        :is-row-select-enable="isRowSelectEnable"
        @selection-change="(selections: any) => handleSelectionChange(selections, isCurRowSelectEnable)"
        @select-all="(selections: any) => handleSelectionChange(selections, isCurRowSelectEnable, true)"
        @page-limit-change="state.handlePageSizeChange"
        @page-value-change="state.handlePageChange"
        @column-sort="state.handleSort"
      />
    </bk-loading>

    <!-- 变更影响确认 -->
    <security-group-change-confirm-dialog
      v-model="isChangeEffectConfirmDialogShow"
      :loading="securityGroupStore.isQueryRelatedResourcesLoading"
      :detail="currentSecurityGroup"
      @confirm="handleChangeEffectConfirm"
    />

    <!-- 删除安全组 -->
    <security-group-single-delete-dialog
      v-model="isSecurityGroupSingleDeleteDialogShow"
      :loading="securityGroupStore.isQueryRelatedResourcesLoading"
      :detail="currentSecurityGroup"
    />

    <!-- 批量分配 -->
    <template v-if="!securityGroupAssignDialogState.isHidden">
      <security-group-assign-dialog
        v-model="securityGroupAssignDialogState.isShow"
        :selections="selections"
        @hidden="securityGroupAssignDialogState.isHidden = true"
        @success="handleSecurityGroupOperationSuccess"
      />
    </template>

    <!-- 批量添加资产归属 -->
    <template v-if="!securityGroupMgmtAttrEditDialogState.isHidden">
      <security-group-update-mgmt-attr-dialog
        v-model="securityGroupMgmtAttrEditDialogState.isShow"
        :selections="selections"
        @hidden="securityGroupMgmtAttrEditDialogState.isHidden = true"
        @success="handleSecurityGroupOperationSuccess"
      />
    </template>
  </div>
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

.search-filter {
  width: 500px;
}

.search-selector-container {
  margin-left: auto;
}

.ml10 {
  margin-left: 10px;
}

.toolbar {
  gap: 10px;
}

.security-manager-page {
  height: 100%;

  :deep(.bk-nested-loading) {
    margin-top: 16px;
    height: calc(100% - 100px);

    .bk-table {
      max-height: 100%;

      .operation-cell {
        display: flex;
        align-self: center;
        gap: 8px;
      }
    }
  }
}
</style>
