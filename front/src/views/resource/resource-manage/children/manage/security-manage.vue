<script setup lang="ts">
import type {
  DoublePlainObject,
  // PlainObject,
  FilterType,
} from '@/typings/resource';
import { GcpTypeEnum } from '@/typings';
import { Button, InfoBox, Loading, Message, Table, Tag, bkTooltips } from 'bkui-vue';
import { useResourceStore, useAccountStore } from '@/store';
import {
  ref,
  h,
  PropType,
  watch,
  reactive,
  defineExpose,
  computed,
  withDirectives,
  nextTick,
  Ref,
  Fragment,
  vShow,
  useTemplateRef,
} from 'vue';

import { useI18n } from 'vue-i18n';
import { useRouter, useRoute } from 'vue-router';
import useQueryCommonList from '@/views/resource/resource-manage/hooks/use-query-list-common';
import useColumns from '@/views/resource/resource-manage/hooks/use-columns';
import useFilter from '@/views/resource/resource-manage/hooks/use-filter';
import { useRegionsStore } from '@/store/useRegionsStore';
import { GLOBAL_BIZS_KEY, VendorEnum, VendorMap } from '@/common/constant';
import { cloneDeep } from 'lodash-es';
import { useBusinessMapStore } from '@/store/useBusinessMap';
import { useResourceAccountStore } from '@/store/useResourceAccountStore';
import useSelection from '../../hooks/use-selection';
import { BatchDistribution, DResourceType } from '@/views/resource/resource-manage/children/dialog/batch-distribution';
import { TemplateTypeMap } from '../dialog/template-dialog';
import { Senarios, useWhereAmI } from '@/hooks/useWhereAmI';
import http from '@/http';
import { timeFormatter } from '@/common/util';
import { storeToRefs } from 'pinia';

import SecurityGroupChangeConfirmDialog from '../dialog/security-group/change-confirm.vue';
import SecurityGroupSingleDeleteDialog from '../dialog/security-group/single-delete.vue';
import CloneSecurity, { IData, ICloneSecurityProps } from '../dialog/clone-security/index.vue';
import SecurityGroupAssignDialog from '../dialog/security-group/assign.vue';
import SecurityGroupUpdateMgmtAttrDialog from '../dialog/security-group/update-mgmt-attr.vue';
import SyncAccountResource from '@/components/sync-account-resource/index.vue';
import UnclaimedComp from '../components/security/unclaimed-comp/index.vue';
import { MGMT_TYPE_MAP, SecurityGroupManageType } from '@/constants/security-group';
import { ISecurityGroupOperateItem, useSecurityGroupStore } from '@/store/security-group';
import { ISearchItem } from 'bkui-vue/lib/search-select/utils';
import { useBusinessGlobalStore } from '@/store/business-global';
import UsageBizValue from '@/views/resource/resource-manage/children/components/security/usage-biz-value.vue';
import RefreshCell from '../components/security/refresh-cell/index.vue';
import { showClone } from '../plugin/security-group/show-clone.plugin';
import { checkVendorInResource } from '../plugin/security-group/check-vendor-in-resource.plugin';
import {
  AUTH_BIZ_CREATE_IAAS_RESOURCE,
  AUTH_BIZ_DELETE_IAAS_RESOURCE,
  AUTH_BIZ_UPDATE_IAAS_RESOURCE,
  AUTH_CREATE_IAAS_RESOURCE,
  AUTH_DELETE_IAAS_RESOURCE,
  AUTH_UPDATE_IAAS_RESOURCE,
} from '@/constants/auth-symbols';
import HcmAuth from '@/components/auth/auth.vue';
import { buildMultipleValueRulesItem } from '@/utils';

const props = defineProps({
  filter: {
    type: Object as PropType<FilterType>,
  },
  isResourcePage: {
    type: Boolean,
  },
  bkBizId: Number,
});

const emit = defineEmits(['handleSecrityType', 'edit', 'editTemplate']);

// use hooks
const { t } = useI18n();

const { getRegionName } = useRegionsStore();
const securityGroupStore = useSecurityGroupStore();
const { getNameFromBusinessMap } = useBusinessMapStore();
const router = useRouter();
const route = useRoute();
const { whereAmI, getBizsId } = useWhereAmI();

const resourceAccountStore = useResourceAccountStore();
const { selectedAccountId, vendorInResourcePage } = storeToRefs(resourceAccountStore);
const resourceStore = useResourceStore();
const accountStore = useAccountStore();
const businessGlobalStore = useBusinessGlobalStore();

const authTypeMap = computed(() => {
  if (props.isResourcePage) {
    return { create: AUTH_CREATE_IAAS_RESOURCE, update: AUTH_UPDATE_IAAS_RESOURCE, delete: AUTH_DELETE_IAAS_RESOURCE };
  }
  return {
    create: AUTH_BIZ_CREATE_IAAS_RESOURCE,
    update: AUTH_BIZ_UPDATE_IAAS_RESOURCE,
    delete: AUTH_BIZ_DELETE_IAAS_RESOURCE,
  };
});

const activeType = ref<string>();
const URL_MAP: Record<string, string> = {
  gcp: 'vendors/gcp/firewalls/rules/list',
  group: 'security_groups/list',
  template: 'argument_templates/list',
};
const fetchUrl = ref<string>(URL_MAP.group);

const { generateColumnsSettings } = useColumns('group');

const cloneSecurityData = reactive<ICloneSecurityProps>({
  isShow: false,
  data: {},
});

const templateData = ref([]);

const { searchData, searchValue, filter } = useFilter(props, {
  convertValueCallbacks: { mgmt_type: (value) => (value === 'unknown' ? '' : value) },
  conditionFormatterMapper: {
    cloud_id: (value: string) => {
      return buildMultipleValueRulesItem('cloud_id', value);
    },
  },
});

const { datas, pagination, isLoading, handlePageChange, handlePageSizeChange, handleSort, getList } =
  useQueryCommonList(
    {
      ...props,
      filter: filter.value,
    },
    fetchUrl,
    {
      asyncRequestApiMethod: (datalist: any[], datalistRef: Ref<any[]>) => {
        // 安全组需要异步加载一些关联资源数
        if (activeType.value !== 'group' || !datalist.length) return [];

        fetchSecurityGroupExtraFields(datalist, datalistRef);
      },
    },
  );

// 异步加载安全组字段：关联资源、规则数、负责人信息
const fetchSecurityGroupExtraFields = async (
  datalist: ISecurityGroupOperateItem[],
  datalistRef: Ref<ISecurityGroupOperateItem[]>,
) => {
  const sgIds = datalist.map((item) => item.id);
  const unclaimedSgIds = datalist
    .filter((sg) => sg.mgmt_type === 'biz' && sg.bk_biz_id === -1 && sg.usage_biz_ids?.[0] !== -1)
    .map((sg) => sg.id);

  // 并行发起所有请求
  const ruleCountPromise = securityGroupStore.batchQueryRuleCount(sgIds).then((ruleRes) => {
    datalistRef.value.forEach((sg) => {
      sg.rule_count = ruleRes[sg.id];
    });
  });

  const relatedResourcesPromise = fetchSecurityGroupRelatedResourcesCount(sgIds, datalistRef);

  const maintainerPromise =
    whereAmI.value === Senarios.business && unclaimedSgIds.length
      ? securityGroupStore.queryUsageBizMaintainers(unclaimedSgIds).then((maintainers) => {
          const maintainerMap = new Map(maintainers.map((item) => [item.id, item]));
          datalistRef.value.forEach((sg) => {
            const entry = maintainerMap.get(sg.id);
            if (entry) {
              sg.account_managers = entry.managers;
              sg.usage_biz_infos = entry.usage_biz_infos;
            }
          });
        })
      : Promise.resolve();

  // 等待所有请求完成（但每个请求完成时会立即更新对应字段）
  Promise.allSettled([ruleCountPromise, relatedResourcesPromise, maintainerPromise]);
};
const fetchSecurityGroupRelatedResourcesCount = async (
  ids: string[],
  datalistRef: Ref<ISecurityGroupOperateItem[]>,
) => {
  return securityGroupStore.queryRelatedResourcesCount(ids).then((relatedResourcesList) => {
    const resMap = new Map(relatedResourcesList.map((item) => [item.id, item]));
    datalistRef.value.forEach((sg) => {
      const entry = resMap.get(sg.id);
      if (entry) {
        sg.rel_res_count = entry.resources?.reduce((acc, cur) => acc + cur.count, 0);
        sg.rel_res = entry.resources?.filter(({ count }) => count > 0);
        sg.error = entry.error;
      }
    });
  });
};

const selectSearchData = computed(() => {
  const map: Record<string, { idName: string; searchData: ISearchItem[] }> = {
    group: {
      idName: t('安全组ID'),
      searchData: [
        {
          name: t('使用业务'),
          id: 'usage_biz_id',
          children: businessGlobalStore.businessFullList.map(({ id, name }) => ({ id, name })),
        },
        {
          name: t('管理类型'),
          id: 'mgmt_type',
          children: [
            { id: SecurityGroupManageType.BIZ, name: t('业务管理') },
            { id: SecurityGroupManageType.PLATFORM, name: t('平台管理') },
            { id: 'unknown', name: t('未确认') },
          ],
          multiple: true,
        },
        {
          name: t('管理业务'),
          id: 'mgmt_biz_id',
          children: businessGlobalStore.businessFullList.map(({ id, name }) => ({ id, name })),
        },
      ],
    },
    gcp: {
      idName: t('防火墙ID'),
      searchData: [],
    },
    template: {
      idName: t('模板ID'),
      searchData: [],
    },
  };

  const baseSearchData = [{ name: map[activeType.value].idName, id: 'cloud_id' }, ...searchData.value];

  return [...baseSearchData, ...map[activeType.value].searchData];
});

watch(
  () => datas.value,
  async (data) => {
    if (activeType.value === 'template') {
      templateData.value = data;
      const ids = data.map(({ id }) => id);
      if (!ids.length) return;
      const url = `/api/v1/cloud${
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

// 抛出请求数据的方法，新增成功使用
const fetchComponentsData = () => {
  getList();
};

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

const refreshRowKeySet = ref(new Set<string>());
const getRefreshLoading = (id: string) => {
  return refreshRowKeySet.value.size === 0 || refreshRowKeySet.value.has(id);
};
const groupColumns = [
  { type: 'selection', width: 30, minWidth: 30, onlyShowOnList: true, notDisplayedInBusiness: true },
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
    render: ({ cell }: any) => cell || '--',
    width: 120,
  },
  {
    label: t('规则个数'),
    field: 'rule_count',
    width: 90,
    isDefaultShow: true,
    render: ({ cell }: any) => {
      return h(Fragment, null, [
        withDirectives(h(Loading, { loading: true, theme: 'primary', mode: 'spin', size: 'mini' }), [
          [vShow, securityGroupStore.isBatchQueryRuleCountLoading],
        ]),
        withDirectives(h('div', null, cell), [[vShow, !securityGroupStore.isBatchQueryRuleCountLoading]]),
      ]);
    },
  },
  {
    label: t('关联实例数'),
    field: 'rel_res_count',
    width: 110,
    isDefaultShow: true,
    render: ({ cell, row }: any) => {
      const { error, id } = row || {};

      const loading = getRefreshLoading(id) && securityGroupStore.isQueryRelatedResourcesCountLoading;

      return h(
        RefreshCell,
        {
          error,
          loading,
          showError: !!error,
          onClick: async () => {
            // 记录当前刷新的安全组id
            refreshRowKeySet.value.add(id);
            // 刷新关联资源数、关联的资源类型
            await fetchSecurityGroupRelatedResourcesCount([id], datas);
            // 清除当前刷新的安全组id
            refreshRowKeySet.value.delete(id);
          },
        },
        cell,
      );
    },
  },
  {
    label: t('关联的资源类型'),
    field: 'rel_res',
    width: 200,
    isDefaultShow: true,
    render: ({ cell, row }: { cell: { res_name: string; count: number }[]; row: any }) => {
      const { id } = row || {};

      const displayValue =
        cell?.length > 0
          ? cell.map(({ res_name, count }) =>
              withDirectives(h(Tag, { class: 'mr4' }, res_name), [[bkTooltips, { content: String(count) }]]),
            )
          : '--';
      const loading = getRefreshLoading(id) && securityGroupStore.isQueryRelatedResourcesCountLoading;
      return h(Fragment, null, [
        withDirectives(h(Loading, { loading: true, theme: 'primary', mode: 'spin', size: 'mini' }), [[vShow, loading]]),
        withDirectives(h('div', null, displayValue), [[vShow, !loading]]),
      ]);
    },
  },
  {
    label: t('使用业务'),
    field: 'usage_biz_ids',
    isDefaultShow: true,
    width: 100,
    showOverflowTooltip: false,
    render: ({ cell }: any) => h(UsageBizValue, { value: cell }),
  },
  {
    label: t('管理类型'),
    field: 'mgmt_type',
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
    isDefaultShow: true,
    width: 100,
    render: ({ cell, data }: any) => {
      if (!cell || cell === -1) return '--';
      const { mgmt_type, bk_biz_id } = data;
      if (mgmt_type === SecurityGroupManageType.BIZ && bk_biz_id === -1 && whereAmI.value === Senarios.business) {
        return h(UnclaimedComp, { data });
      }
      return getNameFromBusinessMap(cell);
    },
  },
  {
    label: t('主负责人'),
    field: 'manager',
    width: 100,
    render: ({ cell }: any) => cell || '--',
  },
  {
    label: t('备份负责人'),
    field: 'bak_manager',
    width: 100,
    render: ({ cell }: any) => cell || '--',
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
    width: 150,
    fixed: 'right',
    showOverflowTooltip: false,
    render({ data }: IData) {
      // 资源下：状态=未分配，才可以操作
      // 业务下：管理业务=当前业务 && 状态=已分配，才可以配置规则、删除
      const isAssigned = data.bk_biz_id !== -1;
      const isPlatformManage = data.mgmt_type === SecurityGroupManageType.PLATFORM;
      const isCurrentBizManage = data.mgmt_biz_id === getBizsId();
      const { isResourcePage } = props;

      const handleClick = (type: string) => {
        if (type === 'clone') {
          securityHandleShowClone(data);
        } else {
          handleFillCurrentSecurityGroup(data, type);
        }
      };

      const isCheckVendorInResource = checkVendorInResource(data.vendor);
      const isCloneShow = showClone(data.vendor);
      const getCommonTooltipsOption = () => {
        if (isResourcePage && isAssigned) {
          return { content: t('安全组已分配，请到业务下操作'), disabled: !isAssigned };
        }
        if (!isResourcePage && isPlatformManage) {
          return {
            content: t('该安全组的管理类型为平台管理，不允许在业务下操作'),
            disabled: isResourcePage || !isPlatformManage,
          };
        }
        if (!isResourcePage && !isAssigned) {
          return {
            content: t('该安全组当前处于未分配状态，不允许在业务下进行管理配置安全组规则等操作'),
            disabled: isResourcePage || isAssigned,
          };
        }
        if (!isResourcePage && !isCurrentBizManage) {
          return { content: t('该安全组不在当前业务管理，不允许操作'), disabled: isResourcePage || isCurrentBizManage };
        }
        return { disabled: true };
      };
      const operationList = [
        {
          type: 'rule',
          authType: authTypeMap.value.update,
          name: t('配置规则'),
          resourcePageDisabled: isAssigned || isCheckVendorInResource,
          businessPageDisabled: !(isCurrentBizManage && isAssigned),
          getTooltipsOption: getCommonTooltipsOption,
        },
        {
          type: 'clone',
          authType: authTypeMap.value.create,
          name: t('克隆'),
          hidden: isResourcePage,
          businessPageDisabled: !isCloneShow,
          getTooltipsOption() {
            if (!isCloneShow) return { content: t('该云厂商暂未支持克隆功能'), disabled: isCloneShow };
            return { disabled: true };
          },
        },
        {
          type: 'delete',
          authType: authTypeMap.value.delete,
          name: t('删除'),
          resourcePageDisabled: isAssigned || isCheckVendorInResource,
          businessPageDisabled: !(isCurrentBizManage && isAssigned),
          getTooltipsOption: getCommonTooltipsOption,
        },
      ];

      return h(
        'div',
        { class: 'operation-cell' },
        operationList.map(
          ({ type, authType, name, resourcePageDisabled, businessPageDisabled, hidden, getTooltipsOption }) => {
            if (hidden) return null;
            const disabled = isResourcePage ? resourcePageDisabled : businessPageDisabled;
            const tooltipsOption = getTooltipsOption?.() || { disabled: true };

            return h(
              HcmAuth,
              { sign: { type: authType, relation: [props.bkBizId] } },
              {
                default: ({ noPerm }: { noPerm: boolean }) =>
                  withDirectives(
                    h(
                      Button,
                      { text: true, theme: 'primary', disabled: noPerm || disabled, onClick: () => handleClick(type) },
                      name,
                    ),
                    [[bkTooltips, tooltipsOption]],
                  ),
              },
            );
          },
        ),
      );
    },
  },
].filter((item) => {
  if (Senarios.business === whereAmI.value) return !item.notDisplayedInBusiness;
  return true;
});

const groupSettings = generateColumnsSettings(groupColumns);

const gcpColumns = [
  { type: 'selection', width: 30, minWidth: 30, onlyShowOnList: true, notDisplayedInBusiness: true },
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
      return h('span', [
        h(
          HcmAuth,
          { sign: { type: authTypeMap.value.update, relation: [props.bkBizId] } },
          {
            default: ({ noPerm }: { noPerm: boolean }) =>
              h(
                Button,
                {
                  text: true,
                  theme: 'primary',
                  disabled: noPerm || (data.bk_biz_id !== -1 && props.isResourcePage),
                  onClick() {
                    emit('edit', cloneDeep(data));
                  },
                },
                [t('编辑')],
              ),
          },
        ),
        h(
          HcmAuth,
          { sign: { type: authTypeMap.value.update, relation: [props.bkBizId] } },
          {
            default: ({ noPerm }: { noPerm: boolean }) =>
              h(
                Button,
                {
                  class: 'ml8',
                  text: true,
                  disabled: noPerm || (data.bk_biz_id !== -1 && props.isResourcePage),
                  theme: 'primary',
                  onClick() {
                    securityHandleShowDelete(data);
                  },
                },
                [t('删除')],
              ),
          },
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
  { type: 'selection', width: 30, minWidth: 30, onlyShowOnList: true, notDisplayedInBusiness: true },
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

const securityType = { name: 'group', label: t('安全组') };
const gcpType = { name: 'gcp', label: t('GCP防火墙规则') };
const templateType = { name: 'template', label: '参数模板' };
const isAllVendor = computed(() => !vendorInResourcePage.value);
const isGcpVendor = computed(() => VendorEnum.GCP === vendorInResourcePage.value);
const isTcloudVendor = computed(() => VendorEnum.TCLOUD === vendorInResourcePage.value);
const types = computed(() => {
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
watch(
  types,
  () => {
    // GCP特殊处理
    if (isGcpVendor.value) {
      activeType.value = 'gcp';
      return;
    }
    // 提取场景参数
    const scene = route.query.scene as string;
    // 腾讯云或未选择账号的情况下，检查返回场景（其他云只有安全组一种场景）
    if (scene && (isTcloudVendor.value || isAllVendor.value)) {
      activeType.value = isTcloudVendor.value && scene === 'gcp' ? 'group' : scene;
      return;
    }
    // 默认情况
    activeType.value = 'group';
  },
  { immediate: true },
);
// 状态保持
watch(
  () => activeType.value,
  (v) => {
    fetchUrl.value = URL_MAP[v] || '';

    searchValue.value = [];
    resetSelections();
    // 清空刷新行key，避免切换tab时只有一行有loading效果
    refreshRowKeySet.value.clear();
    emit('handleSecrityType', v);

    // 准备路由参数。这里使用明确的路由参数进行跳转，避免连续两次路由跳转时的参数错误
    const isResourcePage = whereAmI.value === Senarios.resource;
    const isBusinessPage = whereAmI.value === Senarios.business;
    const path = isResourcePage ? '/resource/resource' : '/business/security';
    const bizId = isBusinessPage ? getBizsId() : undefined;
    const accountId = isResourcePage && selectedAccountId.value ? selectedAccountId.value : undefined;

    // 更新路由
    router.replace({ path, query: { [GLOBAL_BIZS_KEY]: bizId, type: 'security', scene: v, accountId } });
  },
  {
    immediate: true,
  },
);

// 配置规则&删除安全组
const currentSecurityGroup = ref<ISecurityGroupOperateItem>(null);
const isChangeEffectConfirmDialogShow = ref(false);
const securityGroupDeleteDialogState = reactive({
  isShow: false,
  isHidden: true,
});
const handleFillCurrentSecurityGroup = async (rowData: ISecurityGroupOperateItem, type: string) => {
  if (type === 'rule') {
    isChangeEffectConfirmDialogShow.value = true;
  } else {
    securityGroupDeleteDialogState.isShow = true;
    securityGroupDeleteDialogState.isHidden = false;
  }

  const { ruleCountMap, relatedResourcesList } = await securityGroupStore.queryRuleCountAndRelatedResources([
    rowData.id,
  ]);
  const rule_count = ruleCountMap[rowData.id] ?? 0;
  const { resources } = relatedResourcesList[0] ?? {};
  currentSecurityGroup.value = { ...rowData, resources, rule_count };
};
const handleChangeEffectConfirm = () => {
  isChangeEffectConfirmDialogShow.value = false;
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

const securityHandleShowClone = (data: IData) => {
  cloneSecurityData.isShow = true;
  cloneSecurityData.data = data;
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

const securityGroupTableRef = useTemplateRef<typeof Table>('security-group-table');
const handleClearSecurityGroupSelections = () => {
  nextTick(() => {
    resetSelections();
    securityGroupTableRef.value?.clearSelection();
  });
};
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
    handleClearSecurityGroupSelections();
  });
};

const syncDialogState = reactive({ isShow: false, isHidden: true, initialModel: null });
const handleSync = () => {
  syncDialogState.isShow = true;
  syncDialogState.isHidden = false;
  if (resourceAccountStore.resourceAccount) {
    const { id, vendor } = resourceAccountStore.resourceAccount;
    syncDialogState.initialModel = { account_id: id, vendor };
  }
};
const handleSyncError = (error: any) => {
  let { message } = error || {};
  if (error.code === 2000024) {
    message = '同步任务在进行中';
  }
  Message({ theme: 'error', message });
};

// 当table数据整个替换时, 需要清空勾选项, 确保勾选态及selections数据正确
watch(
  () => datas.value,
  () => {
    handleClearSecurityGroupSelections();
  },
);

watch(
  searchValue,
  () => {
    // 清空刷新行key，避免切换tab时只有一行有loading效果
    refreshRowKeySet.value.clear();
  },
  { deep: true },
);

defineExpose({ fetchComponentsData });
</script>

<template>
  <div class="security-manager-page">
    <div class="toolbar">
      <bk-radio-group v-model="activeType" :disabled="isLoading">
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
      <bk-button :disabled="selections.length > 0" @click="handleSync">{{ t('同步安全组') }}</bk-button>
      <bk-search-select
        class="search-filter search-selector-container"
        clearable
        :conditions="[]"
        :data="selectSearchData"
        v-model="searchValue"
        value-behavior="need-key"
      />
    </div>

    <bk-loading :key="activeType" :loading="isLoading" opacity="1">
      <bk-table
        v-if="activeType === 'group'"
        ref="security-group-table"
        :settings="groupSettings"
        row-hover="auto"
        remote-pagination
        :pagination="pagination"
        :columns="groupColumns"
        :data="datas"
        show-overflow-tooltip
        :is-row-select-enable="isRowSelectEnable"
        @selection-change="(selections: any) => handleSelectionChange(selections, isCurRowSelectEnable)"
        @select-all="(selections: any) => handleSelectionChange(selections, isCurRowSelectEnable, true)"
        @page-limit-change="handlePageSizeChange"
        @page-value-change="handlePageChange"
        @column-sort="handleSort"
      />

      <bk-table
        v-else-if="activeType === 'gcp'"
        :settings="gcpSettings"
        row-hover="auto"
        remote-pagination
        :pagination="pagination"
        :columns="gcpColumns"
        :data="datas"
        show-overflow-tooltip
        :is-row-select-enable="isRowSelectEnable"
        @selection-change="(selections: any) => handleSelectionChange(selections, isCurRowSelectEnable)"
        @select-all="(selections: any) => handleSelectionChange(selections, isCurRowSelectEnable, true)"
        @page-limit-change="handlePageSizeChange"
        @page-value-change="handlePageChange"
        @column-sort="handleSort"
      />

      <bk-table
        v-else-if="activeType === 'template'"
        :settings="templateSettings"
        row-hover="auto"
        remote-pagination
        :pagination="pagination"
        :columns="templateColumns"
        :data="templateData"
        show-overflow-tooltip
        :is-row-select-enable="isRowSelectEnable"
        @selection-change="(selections: any) => handleSelectionChange(selections, isCurRowSelectEnable)"
        @select-all="(selections: any) => handleSelectionChange(selections, isCurRowSelectEnable, true)"
        @page-limit-change="handlePageSizeChange"
        @page-value-change="handlePageChange"
        @column-sort="handleSort"
      />
    </bk-loading>

    <!-- 变更影响确认 -->
    <security-group-change-confirm-dialog
      v-model="isChangeEffectConfirmDialogShow"
      :loading="securityGroupStore.isQueryRuleCountAndRelatedResourcesLoading"
      :detail="currentSecurityGroup"
      @confirm="handleChangeEffectConfirm"
    />

    <!-- 删除安全组 -->
    <template v-if="!securityGroupDeleteDialogState.isHidden">
      <security-group-single-delete-dialog
        v-model="securityGroupDeleteDialogState.isShow"
        :loading="securityGroupStore.isQueryRuleCountAndRelatedResourcesLoading"
        :detail="currentSecurityGroup"
        @hidden="securityGroupDeleteDialogState.isHidden = true"
        @success="handleSecurityGroupOperationSuccess"
      />
    </template>

    <!-- 克隆安全组弹窗 -->
    <template v-if="cloneSecurityData.isShow">
      <CloneSecurity
        v-model:is-show="cloneSecurityData.isShow"
        :data="cloneSecurityData.data"
        @success="handleSecurityGroupOperationSuccess"
      />
    </template>

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

    <template v-if="!syncDialogState.isHidden">
      <sync-account-resource
        v-model="syncDialogState.isShow"
        title="同步安全组"
        :desc="`从云上同步安全组、安全组规则、关联的实例信息等`"
        :business-id="props.bkBizId"
        resource-name="security_group"
        :initial-model="syncDialogState.initialModel"
        :error-handler="handleSyncError"
        @hidden="
          () => {
            syncDialogState.isHidden = true;
            syncDialogState.initialModel = null;
          }
        "
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

.search-filter {
  width: 500px;
}

.search-selector-container {
  margin-left: auto;
}

.toolbar {
  display: flex;
  align-items: center;
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
