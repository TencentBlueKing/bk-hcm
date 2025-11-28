/* eslint-disable no-nested-ternary */
// table 字段相关信息
import { useAccountStore } from '@/store';
import { Info, Spinner, Share, InfoLine } from 'bkui-vue/lib/icon';
import { Button, Tag, OverflowTitle } from 'bkui-vue';
import i18n from '@/language/i18n';
import { type Settings } from 'bkui-vue/lib/table/props';
import { ref } from 'vue';
import type { Ref } from 'vue';
import { RouteLocationRaw, useRoute, useRouter } from 'vue-router';
import routerAction from '@/router/utils/action';

import {
  CLOUD_HOST_STATUS,
  VendorEnum,
  VendorMap,
  RESOURCE_PLAN_STATUSES_MAP,
  GLOBAL_BIZS_KEY,
} from '@/common/constant';
import { useRegionsStore } from '@/store/useRegionsStore';
import { Senarios, useWhereAmI } from '@/hooks/useWhereAmI';
import { useBusinessMapStore } from '@/store/useBusinessMap';
import { useCloudAreaStore } from '@/store/useCloudAreaStore';
import StatusAbnormal from '@/assets/image/Status-abnormal.png';
import StatusNormal from '@/assets/image/Status-normal.png';
import StatusUnknown from '@/assets/image/Status-unknown.png';
import StatusSuccess from '@/assets/image/success-account.png';
import StatusLoading from '@/assets/image/status_loading.png';
import StatusFailure from '@/assets/image/failed-account.png';
import { HOST_RUNNING_STATUS, HOST_SHUTDOWN_STATUS } from '../common/table/HostOperations';
import cssModule from './use-scr-columns.module.scss';
import { defaults } from 'lodash';
import { timeFormatter } from '@/common/util';
import { capacityLevel } from '@/utils/scr';
import { formatDisplayNumber, getResourceTypeName, getReturnPlanName } from '@/utils';
import {
  getRecycleTaskStatusLabel,
  getBusinessNameById,
  getPrecheckStatusLabel,
} from '@/views/ziyanScr/host-recycle/field-dictionary';
import { getRegionCn, getZoneCn } from '@/views/ziyanScr/cvm-web/transform';
import { getCvmProduceStatus, getTypeCn } from '@/views/ziyanScr/cvm-produce/transform';
import { getImageName } from '@/views/ziyanScr/cvm-produce/component/property-display/transform';
import { useApplyStages } from '@/views/ziyanScr/hooks/use-apply-stages';
import { transformAntiAffinityLevels } from '@/views/ziyanScr/hostApplication/components/transform';
import useCvmChargeType from '@/views/ziyanScr/hooks/use-cvm-charge-type';
import { ICvmSystemDisk } from '@/views/ziyanScr/components/cvm-system-disk/typings';

import WName from '@/components/w-name';
import { SCR_POOL_PHASE_MAP, SCR_RECALL_DETAIL_STATUS_MAP } from '@/constants';
import CopyToClipboard from '@/components/copy-to-clipboard/index.vue';
import { ResourcesDemandsStatus, type IResourcesTicketItem } from '@/typings/resourcePlan';
import { ChargeType, ChargeTypeMap } from '@/typings/plan';
import { RESOURCE_DEMANDS_STATUS_NAME, RESOURCE_DEMANDS_STATUS_CLASSES } from '@/components/resource-plan/constants';
import QcloudZoneValue from '@/views/ziyanScr/components/qcloud-resource/zone-value.vue';
import QcloudRegionValue from '@/views/ziyanScr/components/qcloud-resource/region-value.vue';
import CvmSystemDiskDisplay from '@/views/ziyanScr/components/cvm-system-disk/display.vue';
import CvmDataDiskDisplay from '@/views/ziyanScr/components/cvm-data-disk/display.vue';
import isEqual from 'lodash/isEqual';
import { RES_ASSIGN_TYPE } from '@/components/device-type-selector/constants';
import { MENU_BUSINESS_TICKET_RESOURCE_PLAN_DETAILS } from '@/constants/menu-symbol';

interface LinkFieldOptions {
  type: string; // 资源类型
  label?: string; // 显示文本
  field?: string; // 字段
  idFiled?: string; // id字段
  onlyShowOnList?: boolean; // 只在列表中显示
  linkable?: boolean | ((data: any) => boolean); // 可链接性
  render?: (data: any) => any; // 自定义渲染内容
  renderSuffix?: (data: any) => any; // 自定义后缀渲染内容
  contentClass?: string; // 内容class
  sort?: boolean; // 是否支持排序
}

export default (type: string, isSimpleShow = false) => {
  const router = useRouter();
  const route = useRoute();
  const { t } = i18n.global;
  const accountStore = useAccountStore();
  const { getRegionName } = useRegionsStore();
  const { whereAmI } = useWhereAmI();
  const businessMapStore = useBusinessMapStore();
  const cloudAreaStore = useCloudAreaStore();
  const { transformApplyStages } = useApplyStages();

  const { cvmChargeTypes, cvmChargeTypeNames, getMonthName } = useCvmChargeType();

  const getLinkField = (options: LinkFieldOptions) => {
    // 设置options的默认值
    defaults(options, {
      label: 'ID',
      field: 'id',
      idFiled: 'id',
      onlyShowOnList: true,
      linkable: true,
      render: undefined,
      sort: true,
    });

    const { type, label, field, idFiled, onlyShowOnList, linkable, render, renderSuffix, contentClass, sort } = options;

    return {
      label,
      field,
      sort,
      width: label === 'ID' ? '120' : 'auto',
      onlyShowOnList,
      isDefaultShow: true,
      render({ data }: { cell: string; data: any }) {
        if (data[idFiled] < 0 || !data[idFiled]) return '--';
        // 是否可链接
        if (!(typeof linkable === 'function' ? linkable(data) : linkable)) {
          return (
            <div class={contentClass}>
              {data[field] || '--'}
              {renderSuffix?.(data)}
            </div>
          );
        }

        const defaultClickHandler = () => {
          const routeInfo: any = { query: { ...route.query, id: data[idFiled], type: data.vendor } };
          // 业务下
          if (route.path.includes('business')) {
            routeInfo.query.bizs = accountStore.bizs;
            Object.assign(routeInfo, { name: `${type}BusinessDetail` });
          } else {
            Object.assign(routeInfo, { name: 'resourceDetail', params: { type } });
          }
          router.push(routeInfo);
        };

        return (
          <div class={contentClass}>
            <Button text theme='primary' onClick={defaultClickHandler}>
              {render ? render(data) : data[field] || '--'}
            </Button>
            {renderSuffix?.(data)}
          </div>
        );
      },
    };
  };

  /**
   * 自定义 render field 的 push 导航
   * @param to 目标路由信息
   */
  const renderFieldPushState = (to: RouteLocationRaw, cb?: (...args: any) => any) => {
    return (e: Event) => {
      // 阻止事件冒泡
      e.stopPropagation();
      // 导航
      router.push(to);
      // 执行回调
      typeof cb === 'function' && cb();
    };
  };

  const hIColumns = [
    {
      label: '需求类型',
      field: 'require_type',
      render: ({ row }: any) => getTypeCn(row.require_type),
    },
    {
      label: '实例族',
      field: 'label.device_group',
    },
    {
      label: '机型',
      field: 'device_type',
    },
    {
      label: 'CPU(核)',
      field: 'cpu',
      sort: true,
    },
    {
      label: '内存(G)',
      field: 'mem',
      sort: true,
    },
    {
      label: '地域',
      field: 'region',
      render: ({ row }: any) => getRegionCn(row.region),
    },
    {
      label: '园区',
      field: 'zone',
      render: ({ row }: any) => getZoneCn(row.zone),
    },
    {
      label: '库存情况',
      field: 'capacity_flag',
      sort: { value: 'desc' },
      render({ cell }: { cell: string }) {
        const { class: theClass, text } = capacityLevel(cell);
        return <span class={cssModule[`${theClass}`]}>{text}</span>;
      },
    },
  ];
  const CRSOcolumns = [
    { type: 'selection', width: 30, minWidth: 30, isDefaultShow: true, align: 'center' },
    {
      label: '机型',
      field: 'spec.device_type',
      width: 200,
      showOverflowTooltip: false,
      render: ({ row }: any) => (
        <div style={{ display: 'flex', width: '100%', gap: '4px' }}>
          {row.original.spec.device_type !== row.spec.device_type && (
            <Info
              style={{ flex: 'none', color: '#e9a24c' }}
              v-bk-tooltips={{ content: `原始值：${row.original.spec.device_type}` }}
            />
          )}
          <OverflowTitle
            type='tips'
            resizeable={true}
            style={{ flex: 1, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
            {row.spec.device_type}
          </OverflowTitle>
          {Senarios.service === whereAmI.value && row.source === 'purchase_to_resource_pool' && (
            <div>
              <bk-tag theme='info'>资源池</bk-tag>
            </div>
          )}
        </div>
      ),
    },
    {
      label: t('计费模式'),
      field: 'spec.charge_type',
      width: 120,
      render: ({ data, cell }: any) => {
        if (cvmChargeTypes.PREPAID === cell) {
          return `${cvmChargeTypeNames[cell]}(${getMonthName(data.spec.charge_months)})`;
        }
        if (cvmChargeTypes.POSTPAID_BY_HOUR === cell) {
          return `${cvmChargeTypeNames[cell]}`;
        }
      },
    },
    {
      label: '状态',
      field: 'stage',
      width: 120,
      render: ({ row }: any) => transformApplyStages(row.stage),
    },
    {
      label: '地域',
      field: 'spec.region',
      width: 120,
      render: ({ row }: any) => getRegionCn(row.spec.region),
    },
    {
      label: '园区',
      field: 'spec.zone',
      width: 160,
      showOverflowTooltip: false,
      render: ({ row }: any) => {
        const getZoneName = (zones: string[]) => {
          if (zones?.[0] === 'all') {
            return '全部可用区';
          }
          return zones?.map((zone: string) => getZoneCn(zone))?.join('，') ?? '--';
        };
        return (
          <div style={{ display: 'flex', width: '100%', gap: '4px' }}>
            {!isEqual(row.original.spec.zones, row.spec.zones) && (
              <Info
                style={{ flex: 'none', color: '#e9a24c' }}
                v-bk-tooltips={{ content: `原始值：${getZoneName(row.original.spec.zones)}` }}
              />
            )}
            <OverflowTitle
              type='tips'
              resizeable={true}
              style={{ flex: 1, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
              {getZoneName(row.spec.zones)}
            </OverflowTitle>
          </div>
        );
      },
    },
    {
      label: '资源分布方式',
      field: 'spec.res_assign',
      render: ({ row }: any) => RES_ASSIGN_TYPE[row.spec.res_assign as keyof typeof RES_ASSIGN_TYPE]?.label ?? '--',
      width: 120,
    },
    {
      label: '反亲和性',
      field: 'anti_affinity_level',
      width: 90,
      render: ({ row }: any) => transformAntiAffinityLevels(row.anti_affinity_level),
    },
    {
      label: '镜像',
      field: 'spec.image_id',
      width: 200,
      render: ({ row }: any) => getImageName(row.spec.image_id),
    },
    {
      label: 'VPC',
      field: 'spec.vpc',
      width: 120,
      showOverflowTooltip: false,
      render: ({ row }: any) => (
        <div style={{ display: 'flex', width: '100%', gap: '4px' }}>
          {row.original.spec.vpc !== row.spec.vpc && (
            <Info
              style={{ flex: 'none', color: '#e9a24c' }}
              v-bk-tooltips={{ content: `原始值：${row.original.spec.vpc || '--'}` }}
            />
          )}
          <OverflowTitle
            type='tips'
            resizeable={true}
            style={{ flex: 1, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
            {row.spec.vpc || '--'}
          </OverflowTitle>
        </div>
      ),
    },
    {
      label: '子网',
      field: 'spec.subnet',
      width: 120,
      showOverflowTooltip: false,
      render: ({ row }: any) => (
        <div style={{ display: 'flex', width: '100%', gap: '4px' }}>
          {row.original.spec.subnet !== row.spec.subnet && (
            <Info
              style={{ flex: 'none', color: '#e9a24c' }}
              v-bk-tooltips={{ content: `原始值：${row.original.spec.subnet || '--'}` }}
            />
          )}
          <OverflowTitle
            type='tips'
            resizeable={true}
            style={{ flex: 1, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
            {row.spec.subnet || '--'}
          </OverflowTitle>
        </div>
      ),
    },
    {
      label: '系统盘',
      field: 'spec.system_disk',
      width: 130,
      render: ({ cell }: { cell: ICvmSystemDisk }) => <CvmSystemDiskDisplay systemDisk={cell} />,
    },
    {
      label: '数据盘',
      field: 'spec.data_disk',
      width: 150,
      render: ({ cell, row }: { cell: any; row: any }) => (
        <CvmDataDiskDisplay dataDiskList={cell} diskType={row.spec.disk_type} diskSize={row.spec.disk_size} />
      ),
    },
    {
      label: '备注',
      field: 'remark',
      render: ({ cell }: { cell: string }) => cell || '--',
    },
  ];
  const PRSOcolumns = [
    {
      label: '地域',
      field: 'spec.region',
      render: ({ row }: any) => getRegionCn(row.spec.region),
    },
    {
      label: '园区',
      field: 'spec.zone',
      showOverflowTooltip: false,
      render: ({ row }: any) => (
        <div style={{ display: 'flex', width: '100%', gap: '4px' }}>
          {row.original.spec.zone !== row.spec.zone && (
            <Info
              style={{ flex: 'none', color: '#e9a24c' }}
              v-bk-tooltips={{ content: `原始值：${getZoneCn(row.original.spec.zone)}` }}
            />
          )}
          <OverflowTitle
            type='tips'
            resizeable={true}
            style={{ flex: 1, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
            {getZoneCn(row.spec.zone)}
          </OverflowTitle>
        </div>
      ),
    },
    {
      label: '反亲和性',
      field: 'anti_affinity_level',
      width: 90,
      render: ({ row }: any) => transformAntiAffinityLevels(row.anti_affinity_level),
    },
    {
      label: '操作系统',
      field: 'spec.os_type',
    },
    {
      label: '数据盘大小',
      field: 'spec.disk_size',
      width: 100,
    },
    {
      label: 'RAID类型',
      field: 'spec.raid_type',
      width: 100,
    },
    {
      label: '备注',
      field: 'remark',
      render: ({ cell }: { cell: string }) => cell || '--',
    },
    {
      label: '状态',
      field: 'stage',
      render: ({ row }: any) => transformApplyStages(row.stage),
    },
  ];
  const CHColumns = [
    {
      label: '机型',
      field: 'spec.device_type',
      minWidth: 150,
      isDefaultShow: true,
    },
    {
      label: '计费模式',
      field: 'spec.charge_type',
      minWidth: 120,
      isDefaultShow: true,
      render: ({ cell }: any) => ChargeTypeMap[cell as ChargeType] || '--',
    },
    {
      label: '需求数量',
      field: 'replicas',
      minWidth: 90,
      isDefaultShow: true,
    },
    {
      label: '地域',
      field: 'spec.region',
      minWidth: 150,
      render: ({ cell }: { cell: string }) => getRegionName(VendorEnum.TCLOUD, cell) || '--',
      isDefaultShow: true,
    },
    {
      label: '园区',
      field: 'spec.zones',
      minWidth: 150,
      render: ({ row }: any) => {
        const { zones } = row.spec;
        if (zones?.[0] === 'all') {
          return '全部可用区';
        }
        return zones?.map((zone: string) => getZoneCn(zone))?.join('，') ?? '--';
      },
      isDefaultShow: true,
    },
    {
      label: '资源分布方式',
      field: 'spec.res_assign',
      render: ({ row }: any) => RES_ASSIGN_TYPE[row.spec.res_assign as keyof typeof RES_ASSIGN_TYPE]?.label ?? '--',
      minWidth: 120,
      isDefaultShow: true,
    },
    {
      label: '镜像',
      field: 'spec.image_id',
      render: ({ row }: any) => getImageName(row.spec.image_id),
      minWidth: 190,
      isDefaultShow: true,
    },
    {
      label: '系统盘',
      field: 'spec.system_disk',
      width: 130,
      isDefaultShow: true,
      render: ({ cell }: { cell: ICvmSystemDisk }) => <CvmSystemDiskDisplay systemDisk={cell} />,
    },
    {
      label: '数据盘',
      field: 'spec.data_disk',
      width: 150,
      isDefaultShow: true,
      render: ({ cell, row }: { cell: any; row: any }) => (
        <CvmDataDiskDisplay dataDiskList={cell} diskType={row.spec.disk_type} diskSize={row.spec.disk_size} />
      ),
    },
    {
      label: '私有网络',
      field: 'spec.vpc',
      minWidth: 150,
      isDefaultShow: true,
    },
    {
      label: '私有子网',
      field: 'spec.subnet',
      minWidth: 150,
      isDefaultShow: true,
    },
    {
      label: '备注',
      field: 'remark',
    },
  ];
  const PMColumns = [
    {
      label: '机型',
      field: 'spec.device_type',
      width: 150,
    },
    {
      label: '需求数量',
      field: 'replicas',
      width: 90,
    },
    {
      label: '地域',
      field: 'spec.region',
      width: 150,
      render: ({ cell }: { cell: string }) => getRegionName(VendorEnum.TCLOUD, cell) || '--',
    },
    {
      label: '园区',
      field: 'spec.zone',
      width: 150,
    },
    {
      label: 'RAID 类型',
      field: 'spec.raid_type',
      width: 110,
    },
    {
      label: '操作系统',
      field: 'spec.os_type',
    },
    {
      label: '备注',
      field: 'remark',
    },
  ];
  const RRColumns = [
    { type: 'selection', width: 30, minWidth: 30, isDefaultShow: true },
    {
      label: '状态',
      field: 'recyclable',
      render: ({ cell, data }: any) => (
        <span
          class={cssModule[cell ? 'c-success' : 'c-danger']}
          v-bk-tooltips={{ content: data.message, disabled: cell, placement: 'right', theme: 'light' }}>
          {cell ? t('可回收') : t('不可回收')}
        </span>
      ),
    },
    {
      label: '固资号',
      field: 'asset_id',
    },
    {
      label: '内网IP',
      field: 'ip',
    },
    {
      label: '所属业务',
      field: 'bk_biz_name',
    },
    {
      label: '所属模块',
      field: 'topo_module',
    },
    {
      label: '维护人',
      field: 'operator',
    },
    {
      label: '备份维护人',
      field: 'bak_operator',
    },
    {
      label: '机型',
      field: 'device_type',
    },
    {
      label: '主机状态',
      field: 'state',
    },
    {
      label: '入库时间',
      field: 'input_time',
    },
  ];
  const BSAColumns = [
    { type: 'selection', width: 30, minWidth: 30, isDefaultShow: true },
    {
      label: '固资号',
      field: 'asset_id',
    },
    {
      label: '内网IP',
      field: 'ip',
    },
    {
      label: '机型',
      field: 'device_type',
    },
    {
      label: '园区',
      field: 'sub_zone',
    },
  ];
  const RTColumns = [
    {
      label: '固资号',
      field: 'asset_id',
    },
    {
      label: '内网IP',
      field: 'ip',
    },
    {
      label: '机型',
      field: 'device_type',
    },
    {
      label: '园区',
      field: 'sub_zone',
    },
    {
      label: '维护人',
      field: 'operator',
    },
    {
      label: '备份维护人',
      field: 'bak_operator',
    },
  ];

  // 主机申请-设备视角
  const HostApplyDeviceColumns = [
    { type: 'selection', width: 30, minWidth: 30, isDefaultShow: true },
    {
      label: '业务',
      field: 'bk_biz_id',
      render({ cell }: any) {
        return businessMapStore.getNameFromBusinessMap(cell);
      },
      notDisplayedInBusiness: true,
    },
    {
      label: '单号',
      field: 'order_id',
      width: 80,
      render: ({ data, cell }: any) => {
        return (
          <Button
            text
            theme='primary'
            onClick={() => {
              const to = {
                name: 'host-application-detail',
                params: { id: data.order_id },
                query: { ...route.query, creator: data.bk_username, bkBizId: data.bk_biz_id },
              };
              if (Senarios.business === whereAmI.value) {
                // 业务下
                Object.assign(to, { name: 'HostApplicationsDetail', [GLOBAL_BIZS_KEY]: data.bk_biz_id });
              }
              routerAction.redirect(to, { history: true });
            }}>
            {cell}
          </Button>
        );
      },
    },
    { label: '子单号', field: 'suborder_id', width: 80 },
    {
      label: '需求类型',
      field: 'require_type',
      render: ({ row }: any) => getTypeCn(row.require_type),
    },
    {
      label: '申请人',
      field: 'bk_username',
      render({ cell }: any) {
        return <WName name={cell} />;
      },
    },
    {
      label: '交付人',
      field: 'deliverer',
      width: 150,
      render({ cell }: any) {
        const name = cell === 'icr' ? 'ICR' : cell;
        const alias = cell === 'icr' ? 'ICR（IEG资源服务助手）' : cell;
        return <WName name={name} alias={alias} />;
      },
    },
    { label: '内网IP', field: 'ip' },
    { label: '固资号', field: 'asset_id' },
    { label: '资源类型', field: 'resource_type' },
    { label: '机型', field: 'device_type' },
    { label: '园区', field: 'zone_name' },
    { label: '所在母机IP', field: 'owner_ip', render: ({ row }: any) => row.owner_ip || '--' },
    { label: '交付时间', field: 'update_at', width: 160, render: ({ cell }: any) => timeFormatter(cell) },
    { label: '申请时间', field: 'create_at', width: 160, render: ({ cell }: any) => timeFormatter(cell) },
    {
      label: '备注信息',
      field: 'remark',
      render({ data }: any) {
        return `${data.description}${data.description && data.remark && '/'}${data.remark}` || '--';
      },
    },
  ];
  const CAcolumns = [
    {
      label: '需求类型',
      field: 'require_type',
      render: ({ row }: any) => getTypeCn(row.require_type),
    },
    {
      label: '实例族',
      field: 'label.device_group',
    },
    {
      label: '机型',
      field: 'device_type',
    },
    {
      label: 'CPU核',
      field: 'cpu',
      sort: true,
    },
    {
      label: '内存(G)',
      field: 'mem',
      sort: true,
    },
    {
      label: '地域',
      field: 'region',
      render: ({ cell }: { cell: string }) => getRegionCn(cell) || '--',
    },
    {
      label: '园区',
      field: 'zone',
      render: ({ row }: any) => getZoneCn(row.zone),
    },
    {
      label: '库存情况',
      field: 'capacity_flag',
      sort: {
        value: 'desc',
      },
      render({ cell }: { cell: string }) {
        const { class: theClass, text } = capacityLevel(cell);
        return <span class={cssModule[`${theClass}`]}>{text}</span>;
      },
    },
  ];
  // 预检详情状态 render
  const getPrecheckStatusView = (value: string) => {
    const label = getPrecheckStatusLabel(value);
    if (value === 'SUCCESS') {
      return <span class={cssModule['c-success']}>{label}</span>;
    }
    if (value === 'RUNNING') {
      return (
        <>
          <Spinner />
          <span>{label}</span>
        </>
      );
    }
    if (value === 'FAILED') {
      return (
        <div>
          <div class={[cssModule['c-danger'], cssModule['fail-flex']]}>
            <span>{label}</span>
            <a
              target='_blank'
              href='https://iwiki.woa.com/pages/viewpage.action?pageId=349178371'
              v-bk-tooltips={{
                content:
                  '请查看解决方案，根据提示处理完成后，请返回“回收单据列表”或者跳转到“单据详情”进行“重试”或者“去除预检失败IP提交”',
              }}>
              <Share />
            </a>
          </div>
        </div>
      );
    }
    return <span>{label}</span>;
  };
  const ERcolumns = [
    {
      label: '步骤',
      field: 'step_desc',
      width: 300,
    },
    {
      label: '状态',
      field: 'status',
      width: 80,
      render: ({ row }: any) => {
        return getPrecheckStatusView(row.status);
      },
    },
    {
      label: '状态描述',
      field: 'message',
    },
    {
      label: '开始时间',
      field: 'create_at',
      width: 160,
      render: ({ cell }: any) => timeFormatter(cell),
    },
    {
      label: '结束时间',
      field: 'end_at',
      width: 160,
      render: ({ cell }: any) => timeFormatter(cell),
    },
    {
      label: '执行日志',
      field: 'log',
    },
  ];
  const PDcolumns = [
    {
      label: '单号',
      field: 'order_id',
      width: 80,
    },
    {
      label: '子单号',
      field: 'suborder_id',
      width: 80,
    },
    {
      label: '状态',
      field: 'status',
      render: ({ row }: any) => {
        return getPrecheckStatusView(row.status);
      },
      exportFormatter: (row: any) => getPrecheckStatusLabel(row.status),
    },
    {
      label: '已执行/总数',
      field: 'mem',
      render: ({ row }: any) => {
        return (
          <div>
            <span class={cssModule[row.success_num > 0 ? 'c-success' : '']}>{row.success_num}</span>
            <span>/</span>
            <span>{row.total_num}</span>
          </div>
        );
      },
      exportFormatter: (row: any) => {
        return `${row.success_num}/${row.total_num}`;
      },
    },
    {
      label: '更新时间',
      field: 'update_at',
      width: 160,
      render: ({ cell }: any) => timeFormatter(cell),
      formatter: ({ update_at }: any) => timeFormatter(update_at),
    },
    {
      label: '创建时间',
      field: 'create_at',
      width: 160,
      render: ({ cell }: any) => timeFormatter(cell),
      formatter: ({ create_at }: any) => timeFormatter(create_at),
    },
  ];
  const getRecycleTaskStatusView = (value: string) => {
    const label = getRecycleTaskStatusLabel(value);
    if (value === 'DONE') {
      return (
        <>
          <span class={cssModule['c-success']}>{label}</span>
        </>
      );
    }
    if (value.includes('ING')) {
      return (
        <>
          <span>{label}</span>
          <Spinner />
        </>
      );
    }
    if (value === 'DETECT_FAILED') {
      return (
        <bk-badge
          class={cssModule['c-danger']}
          v-bk-tooltips={{ content: '请到“预检详情”查看失败原因，或者点击“去除预检失败IP提交”' }}
          dot>
          {label}
        </bk-badge>
      );
    }
    if (value.includes('FAILED')) {
      return <span class={cssModule['c-danger']}>{label}</span>;
    }
    return <span>{label}</span>;
  };

  // 主机回收-单据视角
  const HostRecycleApplicationColumns = [
    { type: 'selection', width: 30, minWidth: 30, isDefaultShow: true },
    {
      label: '业务',
      field: 'bk_biz_id',
      render: ({ row }: any) => {
        return getBusinessNameById(row.bk_biz_id);
      },
      formatter: ({ bk_biz_id }: any) => {
        return getBusinessNameById(bk_biz_id);
      },
      notDisplayedInBusiness: true,
    },
    {
      label: '资源类型',
      field: 'resource_type',
      width: 120,
      render: ({ row }: any) => {
        return <span>{getResourceTypeName(row.resource_type)}</span>;
      },
      formatter: ({ resource_type }: any) => {
        return getResourceTypeName(resource_type);
      },
    },
    {
      label: '回收类型',
      field: 'return_plan',
      render: ({ row }: any) => {
        return <span>{getReturnPlanName(row.return_plan, row.resource_type)}</span>;
      },
      formatter: ({ return_plan, resource_type }: any) => {
        return getReturnPlanName(return_plan, resource_type);
      },
    },
    {
      label: '回收成本',
      field: 'cost_concerned',
      render: ({ row }: any) => {
        return <span>{row.cost_concerned ? '涉及' : '不涉及'}</span>;
      },
      formatter: ({ cost_concerned }: any) => {
        return cost_concerned ? '涉及' : '不涉及';
      },
    },
    {
      label: '状态',
      field: 'status',
      width: 100,
      render: ({ row }: any) => {
        return getRecycleTaskStatusView(row.status);
      },
      exportFormatter: (row: any) => getRecycleTaskStatusLabel(row.status),
    },
    {
      label: '当前处理人',
      field: 'handler',
      width: 100,
      render: ({ row }: any) => {
        return row.handler !== 'AUTO' ? (
          <Button
            text
            theme='primary'
            onClick={() => {
              window.open(`wxwork://message?username=${row.handler}`);
            }}>
            {row.handler}
          </Button>
        ) : (
          <span class={cssModule['cell-font-color']}>{row.handler}</span>
        );
      },
    },
    {
      label: '总数/成功/失败',
      width: 120,
      render: ({ row }: any) => {
        return (
          <div>
            <span>{row.total_num}</span>
            <span>/</span>
            <span class={cssModule[row.success_num > 0 ? 'c-success' : '']}>{row.success_num}</span>
            <span>/</span>
            <span class={cssModule[row.failed_num > 0 ? 'c-danger' : '']}>{row.failed_num}</span>
          </div>
        );
      },
      exportFormatter: (row: any) => {
        return `${row.success_num}/${row.failed_num}/${row.total_num}`;
      },
    },
    {
      label: '回收人',
      field: 'bk_username',
      render: ({ row }: any) => {
        return (
          <Button
            text
            onClick={() => {
              window.open(`wxwork://message?username=${row.bk_username}`);
            }}>
            {row.bk_username}
          </Button>
        );
      },
    },
    {
      label: '回收时间',
      field: 'create_at',
      width: 160,
      render: ({ cell }: any) => timeFormatter(cell),
      formatter: ({ create_at }: any) => timeFormatter(create_at),
    },
    {
      label: '描述',
      field: 'remark',
      showOverflowTooltip: true,
    },
    {
      label: 'OBS项目类型',
      field: 'recycle_type',
      width: 120,
    },
  ];

  // 主机回收-设备视角
  const HostRecycleDeviceColumns = [
    {
      label: '单号',
      field: 'order_id',
      width: 80,
    },
    {
      label: '固资号',
      field: 'asset_id',
    },
    {
      label: '机型',
      field: 'device_type',
    },
    {
      label: '内网IP',
      field: 'ip',
    },
    {
      label: '回收业务',
      field: 'bk_biz_name',
      notDisplayedInBusiness: true,
    },
    {
      label: '地域',
      field: 'bk_zone_name',
    },
    {
      label: '园区',
      field: 'sub_zone',
    },
    {
      label: 'Module名称',
      field: 'module_name',
    },
    {
      label: '标记',
      field: 'return_tag',
    },
    {
      label: '成本分摊比例',
      field: 'return_cost_rate',
      render: ({ row }: any) => {
        return row.return_cost_rate ? `${Math.ceil(row.return_cost_rate * 100)}%` : '-';
      },
    },
    {
      label: '状态',
      field: 'status',
      render: ({ row }: any) => getRecycleTaskStatusView(row.status),
      exportFormatter: (row: any) => getRecycleTaskStatusLabel(row.status),
    },
    {
      label: '回收人',
      field: 'bk_username',
      render: ({ row }: any) => {
        return (
          <Button
            text
            theme='primary'
            onClick={() => {
              window.open(`wxwork://message?username=${row.bk_username}`);
            }}>
            {row.bk_username}
          </Button>
        );
      },
    },
    {
      label: '创建时间',
      field: 'create_at',
      width: 160,
      render: ({ cell }: any) => timeFormatter(cell),
      formatter: ({ create_at }: any) => timeFormatter(create_at),
    },
    {
      label: '完成时间',
      field: 'return_time',
      width: 160,
      render: ({ cell }: any) => timeFormatter(cell),
    },
    {
      label: '备注',
      field: 'remark',
    },
  ];
  // 资源 - 主机回收 - 单据详情 设备销毁列表
  const deviceDestroyColumns = [
    { type: 'selection', width: 30, minWidth: 30, isDefaultShow: true },
    {
      label: '固资号',
      field: 'asset_id',
    },
    {
      label: '实例ID',
      field: 'instance_id',
    },
    {
      label: '机型',
      field: 'device_type',
    },
    {
      label: '园区',
      field: 'sub_zone',
    },
    {
      label: 'Module名称',
      field: 'module_name',
    },
    {
      label: '维护人',
      field: 'operator',
      render: ({ row }: any) => {
        return (
          <Button
            text
            theme='primary'
            onClick={() => {
              window.open(`wxwork://message?username=${row.operator}`);
            }}>
            {row.operator}
          </Button>
        );
      },
    },
    {
      label: '备份维护人',
      field: 'bak_operator',
      render: ({ row }: any) => {
        return (
          <Button
            text
            theme='primary'
            onClick={() => {
              window.open(`wxwork://message?username=${row.bak_operator}`);
            }}>
            {row.bak_operator}
          </Button>
        );
      },
    },
    {
      label: '标记',
      field: 'return_tag',
    },
    {
      label: '成本分摊比例',
      field: 'return_cost_rate',
      render: ({ row }: any) => {
        return row.return_cost_rate ? `${Math.ceil(row.return_cost_rate * 100)}%` : '-';
      },
    },
    {
      label: '校验结果',
      field: 'return_plan_msg',
      showOverflowTooltip: true,
      render: ({ row }: any) => {
        return (
          <bk-link type='info' v-clipboard={row.return_plan_msg} underline={false}>
            {row.return_plan_msg}
          </bk-link>
        );
      },
    },
    {
      label: '上架时间',
      field: 'input_time',
      width: 160,
      render: ({ cell }: any) => timeFormatter(cell),
      formatter: ({ input_time }: any) => timeFormatter(input_time),
    },
    {
      label: '销毁时间',
      field: 'return_time',
      width: 160,
      render: ({ cell }: any) => timeFormatter(cell),
      formatter: ({ return_time }: any) => timeFormatter(return_time),
    },
    {
      label: '回收单号',
      field: 'return_id',
      render: ({ row }: any) => {
        return (
          <Button
            text
            theme='primary'
            onClick={() => {
              window.open(row.return_link);
            }}>
            {row.return_id}
          </Button>
        );
      },
    },
    {
      label: '状态',
      field: 'status',
      render: ({ row }: any) => getRecycleTaskStatusView(row.status),
      exportFormatter: (row: any) => getRecycleTaskStatusLabel(row.status),
    },
  ];

  const scrResourceOnlineColumns = [
    getLinkField({
      type: 'scrResourceOnlineTask',
      label: '单号',
      field: 'id',
      render: ({ id }) => (
        <Button
          text
          theme='primary'
          onClick={renderFieldPushState({
            name: 'scrResourceManageDetail',
            params: { id },
            query: { type: 'online' },
          })}>
          {id}
        </Button>
      ),
    }),
    {
      label: '创建人',
      field: 'bk_username',
      render: ({ cell }: any) => <WName name={cell} />,
    },
    {
      label: '创建时间',
      field: 'create_at',
      render: ({ cell }: any) => timeFormatter(cell),
    },
    {
      label: '上架数量',
      field: 'total_num',
      render: ({ data }: any) => data?.status?.total_num || 0,
    },
    {
      label: '完成数量',
      field: 'success_num',
      render: ({ data }: any) => data?.status?.success_num || 0,
    },
    {
      label: '单据状态',
      field: 'phase',
      render: ({ data }: any) => {
        const phase = data?.status?.phase;
        const desc = SCR_POOL_PHASE_MAP[phase];

        if (phase === 'INIT') return <span class={cssModule['c-info']}>{desc}</span>;
        if (phase === 'RUNNING')
          return (
            <span class={cssModule['status-column-cell']}>
              <img class={[cssModule['status-icon'], cssModule['spin-icon']]} src={StatusLoading} alt='' />
              {desc}
            </span>
          );
        if (phase === 'SUCCESS') return <span class={cssModule['c-success']}>{desc}</span>;
        if (phase === 'FAILED') return <span class={cssModule['c-danger']}>{desc}</span>;

        return phase;
      },
    },
  ];

  const scrResourceOfflineColumns = [
    getLinkField({
      type: 'scrResourceOnlineTask',
      label: '单号',
      field: 'id',
      render: ({ id }) => (
        <Button
          text
          theme='primary'
          onClick={renderFieldPushState({
            name: 'scrResourceManageDetail',
            params: { id },
            query: { type: 'offline' },
          })}>
          {id}
        </Button>
      ),
    }),
    {
      label: '创建人',
      field: 'bk_username',
      render: ({ cell }: any) => <WName name={cell} />,
    },
    {
      label: '创建时间',
      field: 'create_at',
      render: ({ cell }: any) => timeFormatter(cell),
    },
    {
      label: '下架数量',
      field: 'total_num',
      render: ({ data }: any) => data?.status?.total_num || 0,
    },
    {
      label: '完成数量',
      field: 'success_num',
      render: ({ data }: any) => data?.status?.success_num || 0,
    },
    {
      label: '单据状态',
      field: 'phase',
      render: ({ data }: any) => {
        const phase = data?.status?.phase;
        const desc = SCR_POOL_PHASE_MAP[phase];

        if (phase === 'INIT') return <span class={cssModule['c-info']}>{desc}</span>;
        if (phase === 'RUNNING')
          return (
            <span class={cssModule['status-column-cell']}>
              <img class={[cssModule['status-icon'], cssModule['spin-icon']]} src={StatusLoading} alt='' />
              {desc}
            </span>
          );
        if (phase === 'SUCCESS') return <span class={cssModule['c-success']}>{desc}</span>;
        if (phase === 'FAILED') return <span class={cssModule['c-danger']}>{desc}</span>;

        return phase;
      },
    },
  ];

  const scrResourceOnlineHostColumns = [
    {
      label: '内网IP',
      field: 'ip',
      render: ({ data }: any) => data?.labels?.ip,
      exportFormatter: (data: any) => data?.labels?.ip,
    },
    {
      label: '固资号',
      field: 'bk_asset_id',
      render: ({ data }: any) => data?.labels?.bk_asset_id,
      exportFormatter: (data: any) => data?.labels?.bk_asset_id,
    },
    {
      label: '设备类型',
      field: 'device_type',
      render: ({ data }: any) => data?.labels?.device_type,
      exportFormatter: (data: any) => data?.labels?.device_type,
    },
    {
      label: '状态',
      field: 'phase',
      render: ({ cell }: any) => {
        const desc = SCR_POOL_PHASE_MAP[cell];

        if (cell === 'INIT') return <span class={cssModule['c-info']}>{desc}</span>;
        if (cell === 'RUNNING')
          return (
            <span class={cssModule['status-column-cell']}>
              <img class={[cssModule['status-icon'], cssModule['spin-icon']]} src={StatusLoading} alt='' />
              {desc}
            </span>
          );
        if (cell === 'SUCCESS') return <span class={cssModule['c-success']}>{desc}</span>;
        if (cell === 'FAILED') return <span class={cssModule['c-danger']}>{desc}</span>;

        return cell;
      },
      exportFormatter: ({ phase }: any) => SCR_POOL_PHASE_MAP[phase],
    },
    {
      label: '开始时间',
      field: 'create_at',
      render: ({ cell }: any) => timeFormatter(cell),
      exportFormatter: ({ create_at }: any) => timeFormatter(create_at),
    },
    {
      label: '结束时间',
      field: 'update_at',
      render: ({ cell }: any) => timeFormatter(cell),
      exportFormatter: ({ update_at }: any) => timeFormatter(update_at),
    },
    {
      label: '信息',
      field: 'message',
      render: ({ cell }: any) => cell || '--',
    },
  ];

  const scrResourceOfflineHostColumns = [
    {
      label: '内网IP',
      field: 'ip',
      render: ({ data }: any) => data?.labels?.ip,
      exportFormatter: (data: any) => data?.labels?.ip,
    },
    {
      label: '固资号',
      field: 'bk_asset_id',
      render: ({ data }: any) => data?.labels?.bk_asset_id,
      exportFormatter: (data: any) => data?.labels?.bk_asset_id,
    },
    {
      label: '系统重装任务',
      field: 'reinstall_link',
      render: ({ data }: any) => (
        <Button
          text
          theme='primary'
          onClick={() => {
            window.open(data.reinstall_link, '_blank');
          }}>
          {data.reinstall_id}
        </Button>
      ),
      exportFormatter: ({ reinstall_id }: any) => reinstall_id,
    },
    {
      label: '配置检查任务',
      field: 'conf_check_link',
      render: ({ data }: any) => (
        <Button
          text
          theme='primary'
          onClick={() => {
            window.open(data.conf_check_link, '_blank');
          }}>
          {data.conf_check_id}
        </Button>
      ),
      exportFormatter: ({ conf_check_id }: any) => conf_check_id,
    },
    {
      label: '状态',
      field: 'status',
      render: ({ cell }: any) => {
        const desc = SCR_RECALL_DETAIL_STATUS_MAP[cell];

        if (cell === 'TERMINATE') return <span class={cssModule['c-info']}>{desc}</span>;
        if (cell === 'REINSTALLING') {
          return (
            <span class={cssModule['status-column-cell']}>
              <img class={[cssModule['status-icon'], cssModule['spin-icon']]} src={StatusLoading} alt='' />
              {desc}
            </span>
          );
        }
        if (cell === 'RETURNED' || cell === 'DONE') return <span class={cssModule['c-success']}>{desc}</span>;
        if (cell === 'REINSTALL_FAILED') return <span class={cssModule['c-danger']}>{desc}</span>;

        return cell;
      },
      exportFormatter: ({ status }: any) => SCR_RECALL_DETAIL_STATUS_MAP[status],
    },
    {
      label: '开始时间',
      field: 'create_at',
      render: ({ cell }: any) => timeFormatter(cell),
      exportFormatter: ({ create_at }: any) => timeFormatter(create_at),
    },
    {
      label: '结束时间',
      field: 'update_at',
      render: ({ cell }: any) => timeFormatter(cell),
      exportFormatter: ({ update_at }: any) => timeFormatter(update_at),
    },
    {
      label: '信息',
      field: 'message',
      render: ({ cell }: any) => cell || '--',
    },
  ];

  const scrResourceOnlineCreateColumns = [
    { type: 'selection', width: 30, minWidth: 30, isDefaultShow: true },
    {
      label: '固资号',
      field: 'asset_id',
    },
    {
      label: '内网 IP',
      field: 'ip',
    },
    {
      label: '机型',
      field: 'device_type',
    },
    {
      label: '操作系统',
      field: 'os_type',
    },
    {
      label: '机架号',
      field: 'equipment',
    },
    {
      label: '园区',
      field: 'zone',
    },
    {
      label: '模块',
      field: 'module',
    },
    {
      label: 'IDC 单元',
      field: 'idc_unit',
    },
    {
      label: '逻辑区域',
      field: 'idc_logic_area',
    },
    {
      label: '入库时间',
      field: 'input_time',
      render: ({ cell }: any) => timeFormatter(cell),
    },
  ];

  const scrResourceOfflineCreateColumns = [
    {
      label: '机型',
      field: 'device_type',
    },
    {
      label: '地域',
      field: 'bk_cloud_region',
      render: ({ data }: any) => <QcloudRegionValue value={data.bk_cloud_region} />,
    },
    {
      label: '园区',
      field: 'bk_cloud_zone',
      render: ({ data }: any) => <QcloudZoneValue value={data.bk_cloud_zone} />,
    },
    {
      label: '数量',
      field: 'amount',
    },
  ];
  // 资源配置管理-CVM子网
  const cvmWebColumns = [
    { type: 'selection', width: 30, minWidth: 30, isDefaultShow: true },
    {
      label: 'VPC',
      field: 'vpc_name',
      render: ({ row }: any) => {
        return (
          <div class={cssModule['cvm-cell-height']}>
            <div>{row.vpc_name}</div>
            <div>{row.vpc_id}</div>
          </div>
        );
      },
    },
    {
      label: 'Subnet',
      field: 'subnet_name',
      render: ({ row }: any) => {
        return (
          <div class={cssModule['cvm-cell-height']}>
            <div>{row.subnet_name}</div>
            <div>{row.subnet_id}</div>
          </div>
        );
      },
    },
    {
      label: '地域',
      field: 'region',
      render: ({ row }: any) => {
        return (
          <div class={cssModule['cvm-cell-height']}>
            <div> {getRegionCn(row.region)}</div>
            <div>{row.region}</div>
          </div>
        );
      },
    },
    {
      label: '园区',
      field: 'zone',
      render: ({ row }: any) => {
        return (
          <div class={cssModule['cvm-cell-height']}>
            <div> {getZoneCn(row.zone)}</div>
            <div>{row.zone}</div>
          </div>
        );
      },
    },
  ];

  const ApplicationListColumns = [
    {
      label: '申请时间',
      field: 'create_at',
      width: 160,
      render: ({ cell }: any) => timeFormatter(cell),
    },
    {
      label: '期望交付时间',
      field: 'expect_time',
      width: 160,
      render: ({ cell }: any) => timeFormatter(cell),
    },
    {
      label: '备注信息',
      field: 'remark',
      width: 300,
      render: ({ data }: any) => (
        <div>
          {data.description}
          {data.description && data.remark && '/'}
          {data.remark}
        </div>
      ),
    },
  ];
  const cvmModelColumns = [
    { type: 'selection', width: 30, minWidth: 30, isDefaultShow: true },
    {
      label: '机型',
      field: 'device_type',
      width: 200,
    },
    {
      label: '需求类型',
      field: 'require_type',
      render: ({ row }: any) => getTypeCn(row.require_type),
    },
    {
      label: '地域',
      field: 'region',
      render: ({ row }: any) => getRegionCn(row.region),
    },
    {
      label: '园区',
      field: 'zone',
      render: ({ row }: any) => getZoneCn(row.zone),
    },
    {
      label: '实例族',
      field: 'label.device_group',
    },
    {
      label: 'CPU(核)',
      field: 'cpu',
      width: 50,
    },
    {
      label: '内存(G)',
      field: 'mem',
      width: 50,
    },
    {
      label: '其他信息',
      field: 'remark',
      render: ({ cell }: any) => cell || '--',
    },
    {
      label: '可查询容量',
      field: 'enable_capacity',
      render: ({ cell }: any) => (cell ? '是' : '否'),
    },
    {
      label: '可申请',
      field: 'enable_apply',
      render: ({ cell }: any) => (cell ? <span style={'color:#67c23a'}>是</span> : <span>否</span>),
    },
    {
      label: '推荐分数',
      field: 'score',
    },
    {
      label: '备注',
      field: 'comment',
      render: ({ cell }: any) => cell || '--',
    },
  ];
  const firstAccountColumns = [
    {
      label: '一级帐号ID',
      field: 'primaryAccountId',
    },
    {
      label: '云厂商',
      field: 'cloudProvider',
    },
    {
      label: '帐号邮箱',
      field: 'accountEmail',
    },
    {
      label: '主负责人',
      field: 'mainResponsiblePerson',
    },
    {
      label: '组织架构',
      field: 'organizationalStructure',
    },
    {
      label: '二级帐号个数',
      field: 'secondaryAccountCount',
    },
  ];

  const secondaryAccountColumns = [
    {
      label: '二级帐号ID',
      field: 'secondaryAccountId',
    },
    {
      label: '所属一级帐号',
      field: 'parentPrimaryAccount',
    },
    {
      label: '云厂商',
      field: 'cloudProvider',
    },
    {
      label: '站点类型',
      field: 'siteType',
    },
    {
      label: '帐号邮箱',
      field: 'accountEmail',
    },
    {
      label: '主负责人',
      field: 'mainResponsiblePerson',
    },
    {
      label: '运营产品',
      field: 'operatingProduct',
    },
  ];

  const myApplyColumns = [
    {
      label: '申请类型',
      field: 'type',
    },
    {
      label: '单据状态',
      field: 'status',
      render({ data }: any) {
        let icon = StatusAbnormal;
        let txt = '审批拒绝';
        switch (data.status) {
          case 'pending':
          case 'delivering':
            icon = StatusLoading;
            txt = '审批中';
            break;
          case 'pass':
          case 'completed':
          case 'deliver_partial':
            icon = StatusSuccess;
            txt = '审批通过';
            break;
          case 'rejected':
          case 'cancelled':
          case 'deliver_error':
            icon = StatusFailure;
            txt = '审批拒绝';
            break;
        }
        return (
          <div class={cssModule['cvm-status-container']}>
            {txt === '审批中' ? (
              <Spinner fill='#3A84FF' class={'mr6'} width={14} height={14} />
            ) : (
              <img src={icon} class={'mr6'} width={14} height={14} />
            )}

            {txt}
          </div>
        );
      },
    },
    {
      label: '申请人',
      field: 'applicant',
    },
    {
      label: '创建时间',
      field: 'created_at',
      render({ cell }: any) {
        return timeFormatter(cell);
      },
    },
    {
      label: '更新时间',
      field: 'updated_at',
      render({ cell }: any) {
        return timeFormatter(cell);
      },
    },
    {
      label: '备注',
      field: 'memo',
      render({ cell }: any) {
        return cell || '--';
      },
    },
  ];

  // 资源预测列表
  const resourceForecastColumns = [
    {
      label: '业务',
      field: 'bk_biz_name',
      fixed: 'left',
      minWidth: 110,
      isDefaultShow: true,
    },
    {
      label: '运营产品',
      minWidth: 110,
      field: 'op_product_name',
      fixed: 'left',
    },
    {
      label: '预测类型',
      field: 'demand_class',
      fixed: 'left',
      minWidth: 90,
      isDefaultShow: true,
    },
    {
      label: '期望到货时间',
      field: 'expect_time',
      fixed: 'left',
      minWidth: 110,
      isDefaultShow: true,
      render: ({ cell }: { cell: string }) => timeFormatter(cell, 'YYYY-MM-DD'),
    },
    {
      label: '可申领时间',
      field: 'can_apply_time',
      fixed: 'left',
      minWidth: 100,
      isDefaultShow: true,
      render: ({ cell }: { cell: string }) => timeFormatter(cell, 'YYYY-MM-DD'),
    },
    {
      label: '截止申领时间',
      field: 'expired_time',
      fixed: 'left',
      minWidth: 110,
      isDefaultShow: true,
      render: ({ cell }: { cell: string }) => timeFormatter(cell, 'YYYY-MM-DD'),
    },
    {
      label: '机型类型',
      field: 'device_class',
      minWidth: 100,
      fixed: 'left',
    },
    {
      label: '机型规格',
      field: 'device_type',
      fixed: 'left',
      minWidth: 120,
      isDefaultShow: true,
    },
    {
      label: '实例需求数',
      field: 'total_os',
      minWidth: 100,
      isDefaultShow: true,
      render: ({ cell }: { cell: string }) => formatDisplayNumber(cell),
    },
    {
      label: '实例已申请数',
      field: 'applied_os',
      minWidth: 110,
      isDefaultShow: true,
      render: ({ cell }: { cell: string }) => formatDisplayNumber(cell),
    },
    {
      label: '实例可申请数',
      field: 'remained_os',
      minWidth: 120,
      isDefaultShow: true,
      render: ({ cell }: { cell: string }) => formatDisplayNumber(cell),
    },
    {
      label: 'CPU需求核数',
      field: 'total_cpu_core',
      minWidth: 120,
      isDefaultShow: true,
    },
    {
      label: 'CPU已申请核数',
      field: 'applied_cpu_core',
      minWidth: 120,
      isDefaultShow: true,
    },
    {
      label: 'CPU可申请核数',
      field: 'remained_cpu_core',
      minWidth: 120,
      isDefaultShow: true,
    },
    {
      label: '总内存(GB)',
      minWidth: 90,
      field: 'total_memory',
    },
    {
      label: '已申请内存(GB)',
      minWidth: 120,
      field: 'applied_memory',
    },
    {
      label: '可申请内存(GB)',
      minWidth: 120,
      field: 'remained_memory',
    },
    {
      label: '云盘总量',
      minWidth: 90,
      field: 'total_disk_size',
    },
    {
      label: '云盘已申请数',
      minWidth: 110,
      field: 'applied_disk_size',
    },
    {
      label: '云盘可申请数',
      minWidth: 110,
      field: 'remained_disk_size',
    },
    {
      label: '城市',
      field: 'region_name',
      minWidth: 90,
      isDefaultShow: true,
    },
    {
      label: '可用区',
      minWidth: 110,
      field: 'zone_name',
      isDefaultShow: true,
    },
    {
      label: '申请人',
      field: 'creator',
      minWidth: 110,
      isDefaultShow: true,
    },
    {
      label: '计划类型',
      field: 'plan_type',
      fixed: 'right',
      minWidth: 90,
      isDefaultShow: true,
      render: ({ data }: any) => (
        <Tag theme={data.plan_type === '预测内' ? 'success' : 'warning'}>{data.plan_type}</Tag>
      ),
    },
    {
      label: '项目类型',
      field: 'obs_project',
      fixed: 'right',
      minWidth: 100,
      isDefaultShow: true,
    },
    {
      label: '机型代次',
      field: 'generation_type',
      fixed: 'right',
    },
    {
      label: '机型族',
      field: 'device_family',
      fixed: 'right',
    },
    {
      label: '云磁盘类型',
      field: 'disk_type_name',
      fixed: 'right',
    },
    {
      label: '单实例磁盘IO(MB/s)',
      field: 'disk_io',
      fixed: 'right',
    },
    {
      label: '状态',
      field: 'status',
      fixed: 'right',
      minWidth: 110,
      isDefaultShow: true,
      render: ({ cell: status, row }: { cell: ResourcesDemandsStatus; row: any }) => {
        const statusClass = cssModule[RESOURCE_DEMANDS_STATUS_CLASSES[status]];
        const { href } = router.resolve({
          name: MENU_BUSINESS_TICKET_RESOURCE_PLAN_DETAILS,
          query: {
            id: row?.ticket_id,
            bizs: row?.bk_biz_id,
          },
        });

        const tips = row?.ticket_id ? (
          <span>
            预测处于变更中，
            <bk-link href={href} theme='primary' target='_blank' style='font-size: 12px;'>
              点击链接
            </bk-link>
            查看单据
          </span>
        ) : (
          '预测处于变更过程中，请等待状态刷新'
        );

        return (
          <div class={cssModule['flex-container']}>
            <span class={statusClass}>{RESOURCE_DEMANDS_STATUS_NAME[status]}</span>
            {ResourcesDemandsStatus.LOCKED === status && (
              <InfoLine
                class={[statusClass, 'ml5']}
                v-bk-tooltips={{
                  content: tips,
                }}
              />
            )}
          </div>
        );
      },
    },
  ];

  // 资源预测批量取消列表
  const resourceForecastBatchCancelColumns = [
    {
      label: '预测ID',
      field: 'demand_id',
      isDefaultShow: true,
    },
    {
      label: '期望到货时间',
      field: 'expect_time',
      isDefaultShow: true,
      render: ({ cell }: { cell: string }) => timeFormatter(cell),
    },
    {
      label: '机型类型',
      field: 'device_class',
      isDefaultShow: true,
    },
    {
      label: '机型规格',
      field: 'device_type',
      isDefaultShow: true,
    },
    {
      label: '实例总数',
      field: 'total_os',
      isDefaultShow: true,
      render: ({ cell }: { cell: string }) => formatDisplayNumber(cell),
    },
    {
      label: '城市',
      field: 'region_name',
    },
    {
      label: '可用区',
      field: 'zone_name',
    },
    {
      label: '项目类型',
      field: 'obs_project',
      isDefaultShow: true,
    },
    {
      label: '云磁盘类型',
      field: 'disk_type_name',
    },
    {
      label: '单实例磁盘IO(MB/s)',
      field: 'disk_io',
    },
    {
      label: '云盘总量',
      field: 'total_disk_size',
    },
  ];

  // 服务请求 - 资源预测
  const forecastDemandColumns = [
    {
      label: '业务',
      field: 'bk_biz_name',
      isDefaultShow: true,
    },
    {
      label: '单据状态',
      field: 'status_name',
      isDefaultShow: true,
    },
    {
      label: '运营产品',
      field: 'bk_product_name',
    },
    {
      label: '规划产品',
      field: 'plan_product_name',
    },
    {
      label: 'CPU总核心数',
      field: 'cpu_core',
      isDefaultShow: true,
    },
    {
      label: '内存总量(GB)',
      field: 'memory',
      isDefaultShow: true,
    },
    {
      label: '云硬盘总量(GB)',
      field: 'disk_size',
      isDefaultShow: true,
    },
    {
      label: '提单人',
      field: 'applicant',
      isDefaultShow: true,
    },
    {
      label: '备注',
      field: 'remark',
    },
    {
      label: '创建时间',
      field: 'created_at',
      render: ({ cell }: { cell: string }) => timeFormatter(cell),
    },
    {
      label: '提单时间',
      field: 'submitted_at',
      isDefaultShow: true,
      render: ({ cell }: { cell: string }) => timeFormatter(cell),
    },
  ];

  // 单据下的资源预测
  const receiptForecastDemandColumns = [
    {
      label: '审批状态',
      field: 'status_name',
      isDefaultShow: true,
      render: ({ cell, data }: any) => {
        const { class: className, color } = RESOURCE_PLAN_STATUSES_MAP[data.status] || {};

        return (
          <span>
            <i
              class={`${className} ${cssModule['resource-plan-status-icon']}  ${
                cssModule[`resource-plan-status-icon-${color}`]
              }`}></i>
            {cell}
          </span>
        );
      },
    },
    {
      label: '类型',
      field: 'ticket_type_name',
      isDefaultShow: true,
    },
    {
      label: 'CPU核数',
      field: 'updated_info.cvm.cpu_core',
      isDefaultShow: true,
      render: ({ cell, data }: { cell: number; data: IResourcesTicketItem }) => {
        const value = cell - data.original_info.cvm.cpu_core;
        if (isNaN(value)) {
          return '--';
        }
        let prefix = value > 0 ? '+' : '';
        if (value === 0) {
          prefix = data.ticket_type === 'delete' ? '-' : '+';
        }
        return `${prefix}${value}`;
      },
    },
    {
      label: '内存量(GB)',
      field: 'updated_info.cvm.memory',
      isDefaultShow: true,
      render: ({ cell, data }: { cell: number; data: IResourcesTicketItem }) => {
        const value = cell - data.original_info.cvm.memory;
        if (isNaN(value)) {
          return '--';
        }
        let prefix = value > 0 ? '+' : '';
        if (value === 0) {
          prefix = data.ticket_type === 'delete' ? '-' : '+';
        }
        return `${prefix}${value}`;
      },
    },
    {
      label: '云硬盘量(GB)',
      field: 'updated_info.cbs.disk_size',
      isDefaultShow: true,
      render: ({ cell, data }: { cell: number; data: IResourcesTicketItem }) => {
        const value = cell - data.original_info.cbs.disk_size;
        if (isNaN(value)) {
          return '--';
        }
        let prefix = value > 0 ? '+' : '';
        if (value === 0) {
          prefix = data.ticket_type === 'delete' ? '-' : '+';
        }
        return `${prefix}${value}`;
      },
    },
    {
      label: '提单人',
      field: 'applicant',
      isDefaultShow: true,
    },
    {
      label: '备注',
      field: 'remark',
    },
    {
      label: '创建时间',
      field: 'created_at',
      render: ({ cell }: { cell: string }) => timeFormatter(cell),
    },
    {
      label: '提单时间',
      field: 'submitted_at',
      isDefaultShow: true,
      render: ({ cell }: { cell: string }) => timeFormatter(cell),
    },
    {
      label: '完成时间',
      field: 'completed_at',
      isDefaultShow: true,
      render: ({ cell }: { cell: string }) => timeFormatter(cell),
    },
  ];

  // 资源预测详情
  const adjustmentEntryColums = [
    {
      label: '期望到货日期',
      field: 'expect_time',
      align: 'center',
      minWidth: 120,
    },
    {
      label: '项目类型',
      field: 'obs_project',
      align: 'center',
      minWidth: 150,
    },
    {
      label: '城市',
      field: 'region_name',
      align: 'center',
      minWidth: 130,
    },
    {
      label: '可用区',
      field: 'zone_name',
      align: 'center',
      minWidth: 100,
    },
    {
      label: '实例规格',
      field: 'device_type',
      align: 'center',
      minWidth: 150,
    },
    {
      label: '实例数变更值',
      minWidth: 130,
      align: 'center',
      render: ({ data }: any) => <span>{data.change_cvm_amount}</span>,
    },
    {
      label: 'CPU核数变更值',
      minWidth: 130,
      align: 'center',
      render: ({ data }: any) => <span>{data.change_core_amount}</span>,
    },
    {
      label: '磁盘数(G)变更值',
      minWidth: 130,
      align: 'center',
      render: ({ data }: any) => <span>{data.changed_disk_amount}</span>,
    },
    {
      label: '变更类型',
      field: 'demand_source',
      minWidth: 150,
      align: 'center',
    },
    {
      label: 'CRP订单号',
      field: 'crp_sn',
      minWidth: 200,
      align: 'center',
      render: ({ cell }: { cell: string }) => (
        <Button
          theme='primary'
          text
          onClick={() => window.open(`https://crp.woa.com/crp-outside/yunti/orders/iaasplan/${cell}`, '_blank')}>
          {cell}
        </Button>
      ),
    },
    {
      label: '预测单号',
      field: 'ticket_id',
      minWidth: 200,
      align: 'center',
      // 服务下需要跳转至业务下单据详情，render由业务方决定
    },
    {
      label: '子订单号',
      field: 'suborder_id',
      minWidth: 200,
      align: 'center',
      render: ({ cell }: { cell: string }) => cell || '--',
    },
    {
      label: '备注',
      field: 'remark',
      align: 'center',
      minWidth: 150,
    },
  ];

  // 资源预测详情
  const forecastDemandDetailColums = [
    {
      label: '机型规格',
      field: 'cvm.device_type',
      isDefaultShow: true,
    },
    {
      label: '总CPU核数',
      field: 'cvm.cpu_core',
      isDefaultShow: true,
    },
    {
      label: '总内存(G)',
      field: 'cvm.memory',
      isDefaultShow: true,
    },
    {
      label: '总云盘大小(G)',
      field: 'cbs.disk_size',
      isDefaultShow: true,
    },
    {
      label: '项目类型',
      field: 'obs_project',
      isDefaultShow: true,
    },
    {
      label: '地域',
      field: 'area_name',
      isDefaultShow: true,
    },
    {
      label: '城市',
      field: 'region_name',
      isDefaultShow: true,
    },
    {
      label: '可用区',
      field: 'zone_name',
      isDefaultShow: true,
    },
    {
      label: '资源模式',
      field: 'cvm.res_mode',
      isDefaultShow: true,
    },
    {
      label: '期望到货时间',
      field: 'expect_time',
      isDefaultShow: true,
    },
    {
      label: '机型族',
      field: 'cvm.device_family',
    },
    {
      label: '机型类型',
      field: 'cvm.device_class',
    },
    {
      label: '资源池',
      field: 'cvm.res_pool',
    },
    {
      label: '核心类型',
      field: 'cvm.core_type',
    },
    {
      label: '实例数',
      field: 'cvm.os',
    },
    {
      label: '单例磁盘IO(MB/s)',
      field: 'cbs.disk_io',
      isDefaultShow: true,
    },
    {
      label: '云磁盘类型',
      field: 'cbs.disk_type_name',
      isDefaultShow: true,
    },
    {
      label: '备注',
      field: 'remark',
    },
  ];

  // 预测清单
  const forecastListColums = [
    {
      label: '机型规格',
      field: 'cvm.device_type',
      isDefaultShow: true,
    },
    {
      label: 'CPU总核数',
      field: 'cvm.cpu_core',
      isDefaultShow: true,
    },
    {
      label: '内存总量(G)',
      field: 'cvm.memory',
      isDefaultShow: true,
    },
    {
      label: '云盘总量(G)',
      field: 'cbs.disk_size',
      isDefaultShow: true,
    },
    {
      label: '项目类型',
      field: 'obs_project',
      isDefaultShow: true,
    },
    {
      label: '期望到货时间',
      field: 'expect_time',
      isDefaultShow: true,
    },
    {
      label: '城市',
      field: 'region_name',
      isDefaultShow: true,
    },
    {
      label: '可用区',
      field: 'zone_name',
      isDefaultShow: true,
    },
    {
      label: '资源模式',
      field: 'cvm.res_mode',
      isDefaultShow: true,
    },
    {
      label: '机型类型',
      field: 'cvm.device_class',
    },
    {
      label: '单实例磁盘IO(MB/s)',
      field: 'cbs.disk_io',
      isDefaultShow: true,
    },
    {
      label: '云磁盘类型',
      field: 'cbs.disk_type_name',
      isDefaultShow: true,
    },
  ];

  // 单据管理 - 账号
  const accountColums = [
    {
      label: '单号',
      field: 'order_number',
      isDefaultShow: true,
    },
    {
      label: '资源类型',
      field: 'resource_type',
      isDefaultShow: true,
    },
    {
      label: '单据状态',
      field: 'document_status',
      isDefaultShow: true,
    },
    {
      label: '申请人',
      field: 'applicant',
      isDefaultShow: true,
    },
    {
      label: '申请时间',
      field: 'application_time',
    },
    {
      label: ' 结束时间',
      field: 'end_time',
    },
    {
      label: '备注',
      field: 'remarks',
    },
  ];

  const producingColumns = [
    {
      field: 'task_id',
      label: '任务ID',
      render: ({ data }: any) => {
        return (
          <Button
            theme='primary'
            text
            onClick={() => {
              window.open(data.task_link, '_blank');
            }}>
            {data.generate_id}
          </Button>
        );
      },
    },
    {
      field: 'status',
      label: '状态',
      render: ({ row }: any) => {
        if (row.status === -1) return <span class={cssModule['c-disabled']}>未执行</span>;
        if (row.status === 0) return <span class={cssModule['c-success']}>成功</span>;
        if (row.status === 1)
          return (
            <span>
              <Spinner />
              执行中
            </span>
          );
        return <span class={cssModule['c-danger']}>失败</span>;
      },
    },
    {
      label: '成功台数/总台数',
      render: ({ row }: any) => {
        return (
          <div>
            <span class={cssModule['c-success']}>{row.success_num}</span>
            <span>/</span>
            <span>{row.total_num}</span>
          </div>
        );
      },
    },
    {
      field: 'message',
      label: '状态说明',
      showOverflowTooltip: true,
    },
    {
      field: 'start_at',
      label: '开始时间',
      render: ({ data }: any) => (data.status === -1 ? '-' : timeFormatter(data.start_at)),
    },
    {
      field: 'end_at',
      label: '结束时间',
      render: ({ data }: any) => (![0, 2].includes(data.status) ? '-' : timeFormatter(data.end_at)),
    },
  ];

  const initialColumns = [
    {
      field: 'ip',
      label: '内网 IP',
    },
    {
      field: 'status',
      label: '状态',
      render: ({ data }: any) => {
        if (data.status === -1) return <span class={cssModule['c-disabled']}>未执行</span>;
        if (data.status === 0) return <span class={cssModule['c-success']}>成功</span>;
        if (data.status === 1)
          return (
            <span>
              <i class='el-icon-loading mr-2'></i>执行中
            </span>
          );
        return <span class={cssModule['c-danger']}>失败</span>;
      },
    },
    {
      field: 'message',
      label: '状态说明',
      showOverflowTooltip: true,
    },
    {
      field: 'task_id',
      label: '关联初始化单',
      render: ({ data }: any) => {
        return (
          <Button
            theme='primary'
            text
            onClick={() => {
              window.open(data.task_link, '_blank');
            }}>
            {data.task_id}
          </Button>
        );
      },
    },
    {
      field: 'start_at',
      label: '开始时间',
      render: ({ data }: any) => (data.status === -1 ? '-' : timeFormatter(data.start_at)),
    },
    {
      field: 'end_at',
      label: '结束时间',
      render: ({ data }: any) => (![0, 2].includes(data.status) ? '-' : timeFormatter(data.end_at)),
    },
  ];

  const deliveryColumns = [
    {
      field: 'ip',
      label: '内网 IP',
    },
    {
      field: 'asset_id',
      label: '固资号',
    },
    {
      field: 'status',
      label: '状态',
      render: ({ data }: any) => {
        if (data.status === -1) return <span class={cssModule['c-disabled']}>未执行</span>;
        if (data.status === 0) return <span class={cssModule['c-success']}>成功</span>;
        if (data.status === 1)
          return (
            <span>
              <i class='el-icon-loading mr-2'></i>执行中
            </span>
          );
        return <span class={cssModule['c-danger']}>失败</span>;
      },
    },
    {
      field: 'message',
      label: '状态说明',
    },
    {
      field: 'deliverer',
      label: '匹配人',
      render: ({ data }: any) => <WName name={data.deliverer}></WName>,
    },
    {
      field: 'generate_task_id',
      label: '关联生产单',
      render: ({ data }: any) => {
        return (
          <Button
            theme='primary'
            text
            onClick={() => {
              window.open(data.generate_task_link, '_blank');
            }}>
            {data.generate_task_id}
          </Button>
        );
      },
    },
    {
      field: 'init_task_id',
      label: '关联初始化单',
      render: ({ data }: any) => {
        return (
          <Button
            theme='primary'
            text
            onClick={() => {
              window.open(data.init_task_link, '_blank');
            }}>
            {data.init_task_id}
          </Button>
        );
      },
    },
    {
      field: 'start_at',
      label: '开始时间',
      render: ({ data }: any) => (data.status === -1 ? '-' : timeFormatter(data.start_at)),
    },
    {
      field: 'end_at',
      label: '结束时间',
      render: ({ data }: any) => (![0, 2].includes(data.status) ? '-' : timeFormatter(data.end_at)),
    },
  ];

  const decommissionDetailsColumns = [
    {
      label: '固资号',
      field: 'server_asset_id',
      width: 170,
      isDefaultShow: true,
    },
    {
      label: '内网IP',
      field: 'ip',
      isDefaultShow: true,
      render({ cell }: any) {
        return (
          <Button
            text
            theme='primary'
            onClick={() => {
              window.open(`https://tmp.woa.com/host/home/${cell}`, '_blank');
            }}>
            {cell}
          </Button>
        );
      },
    },
    {
      label: '公网IP',
      field: 'outer_ip',
      isDefaultShow: true,
    },
    {
      label: '业务名称',
      field: 'app_name',
      isDefaultShow: true,
    },
    {
      label: '业务模块',
      field: 'module',
      isDefaultShow: true,
    },
    {
      label: 'SCM设备类型',
      field: 'device_type',
      isDefaultShow: true,
    },
    {
      label: '服务器类型',
      field: 'svr_type_name',
      isDefaultShow: true,
    },
    {
      label: '裁撤模块名称',
      field: 'module_name',
      isDefaultShow: true,
    },
    {
      label: '存放机房管理单元',
      field: 'idc_unit_name',
      isDefaultShow: true,
    },
    {
      label: '操作系统',
      field: 'sfw_name_version',
      isDefaultShow: true,
    },
    {
      label: '上架时间',
      field: 'go_up_date',
      isDefaultShow: true,
    },
    {
      label: 'RAID结构',
      field: 'raid_name',
      isDefaultShow: true,
    },
    {
      label: '逻辑区域',
      field: 'logic_area',
      isDefaultShow: true,
    },
    {
      label: '维护人',
      field: 'server_operator',
      isDefaultShow: true,
    },
    {
      label: '备份维护人',
      field: 'server_bak_operator',
    },
    {
      label: '设备技术分类',
      field: 'device_layer',
    },
    {
      label: 'CPU得分',
      field: 'cpu_score',
    },
    {
      label: '内存得分',
      field: 'mem_score',
    },
    {
      label: '内网流量得分',
      field: 'inner_net_traffic_score',
    },
    {
      label: '磁盘IO得分',
      field: 'disk_io_score',
    },
    {
      label: '磁盘IO使用率得分',
      field: 'disk_util_score',
    },
    {
      label: '是否达标',
      field: 'is_pass',
    },
    {
      label: '内存使用量(G)',
      field: 'mem4linux',
    },
    {
      label: '内网流量(Mb/s)',
      field: 'inner_net_traffic',
    },
    {
      label: '外网流量(Mb/s)',
      field: 'outer_net_traffic',
    },
    {
      label: '磁盘IO(Blocks/s)',
      field: 'disk_io',
    },
    {
      label: '磁盘IO使用率',
      field: 'disk_util',
    },
    {
      label: '磁盘总量(G)',
      field: 'disk_total',
    },
    {
      label: 'CPU核数',
      field: 'max_cpu_core_amount',
    },
    {
      label: '运维小组',
      field: 'group_name',
    },
    {
      label: '业务中心',
      field: 'center',
    },
  ];

  // CVM虚拟机 - CVM生产
  const cvmProduceColumns = [
    { type: 'expand', width: 40, minWidth: 40, showOverflowTooltip: false },
    { label: '单号', field: 'order_id', width: 60 },
    {
      label: '云梯单号',
      field: 'task_id',
      width: 210,
      showOverflowTooltip: () => ({ theme: 'light' }),
      render: ({ row }: any) => {
        if (row.task_id)
          return (
            <Button
              text
              theme='primary'
              onClick={() => {
                window.open(row.task_link, '_blank');
              }}>
              {row.task_id}
            </Button>
          );
        return '-';
      },
    },
    {
      label: '需求类型',
      field: 'require_type',
      width: 100,
      render: ({ row }: any) => getTypeCn(row.require_type),
    },
    {
      label: '状态',
      field: 'status',
      width: 80,
      render: ({ row }: any) => {
        const desc = getCvmProduceStatus(row.status);

        if (row.status === 'INIT') return <span class={cssModule['c-info']}>{desc}</span>;
        if (row.status === 'RUNNING')
          return (
            <span>
              <i class='el-icon-loading mr-2'></i>
              {desc}
            </span>
          );
        if (row.status === 'SUCCESS') return <span class={cssModule['c-success']}>{desc}</span>;
        if (row.status === 'FAILED') return <span class={cssModule['c-danger']}>{desc}</span>;

        return row.status;
      },
    },
    {
      label: '状态描述',
      field: 'message',
      width: 120,
    },
    {
      label: '地域',
      field: 'spec.region',
      width: 120,
      render: ({ row }: any) => getRegionCn(row.spec.region),
    },
    {
      label: '园区',
      field: 'spec.zone',
      width: 150,
      render: ({ row }: any) => getZoneCn(row.spec.zone),
    },
    {
      label: '机型',
      field: 'spec.device_type',
      width: 120,
    },
    {
      label: '创建人',
      field: 'bk_username',
      width: 100,
    },
    {
      label: '创建时间',
      field: 'create_at',
      sort: {
        value: 'desc',
      },
      width: 160,
      render: ({ cell }: any) => timeFormatter(cell),
    },
    {
      label: '结束时间',
      field: 'update_at',
      width: 160,
      render: ({ cell }: any) => timeFormatter(cell),
    },
    {
      label: '备注',
      field: 'remark',
    },
    {
      label: '生产情况-待交付',
      field: 'pending_num',
      width: 130,
      fixed: 'right',
    },
    {
      label: '生产情况-失败',
      field: 'failed_num',
      width: 120,
      fixed: 'right',
      render: ({ row }: any) => <span class={cssModule['c-danger']}>{row.failed_num}</span>,
    },
    {
      label: '生产情况-总数',
      field: 'total_num',
      width: 120,
      fixed: 'right',
    },
  ];

  // CVM虚拟机 - CVM生产 - 快速生产
  const cvmFastProduceColumns = [
    {
      field: 'require_type',
      label: '需求类型',
      width: 100,
      render: ({ row }: any) => getTypeCn(row.require_type),
    },
    {
      field: 'label.device_group',
      label: '实例族',
      width: 120,
    },
    {
      field: 'device_type',
      label: '机型',
    },
    {
      field: 'cpu',
      label: 'CPU(核)',
      sort: true,
      width: 100,
    },
    {
      field: 'mem',
      label: '内存(G)',
      sort: true,
      width: 100,
    },
    {
      field: 'region',
      label: '地域',
      render: ({ row }: any) => getRegionCn(row.region),
    },
    {
      field: 'zone',
      label: '园区',
      render: ({ row }: any) => getZoneCn(row.zone),
    },
    {
      field: 'capacity_flag',
      label: '库存情况',
      width: 140,
      sort: { value: 'desc' },
      render: ({ row }: any) => {
        const { class: theClass, text } = capacityLevel(row.capacity_flag);
        return <span class={cssModule[`${theClass}`]}>{text}</span>;
      },
    },
  ];

  // CVM虚拟机 - CVM生产 - 详情
  const cvmProduceDetailColumns = [
    {
      label: '固资号',
      field: 'asset_id',
    },
    {
      label: '内网 IP',
      field: 'ip',
    },
    {
      label: '实例 ID',
      field: 'cvm_inst_id',
    },
    {
      label: '生产时间',
      field: 'update_at',
      width: 160,
      render: ({ cell }: any) => timeFormatter(cell),
    },
  ];

  const billsRootAccountSummaryColumns = [
    {
      label: '一级账号ID',
      field: 'root_account_id',
    },
    {
      label: '一级账号名称',
      field: 'root_account_name',
    },
    {
      label: '账号状态',
      field: 'state',
    },
    {
      label: '账单同步（人民币-元）当月',
      field: 'current_month_rmb_cost_synced',
    },
    {
      label: '账单同步（人民币-元）上月',
      field: 'last_month_rmb_cost_synced',
    },
    {
      label: '账单同步（美金-美元）当月',
      field: 'current_month_cost_synced',
    },
    {
      label: '账单同步（美金-美元）上月',
      field: 'last_month_cost_synced',
    },
    {
      label: '账单同步环比',
      field: 'month_on_month_value',
    },
    {
      label: '当前账单人民币（元）',
      field: 'current_month_rmb_cost',
    },
    {
      label: '当前账单美金（美元）',
      field: 'current_month_cost',
    },
    {
      label: '调账人民币（元）',
      field: 'adjustment_cost',
    },
    {
      label: '调账美金（美元）',
      field: 'adjustment_cost',
    },
  ];

  const billsMainAccountSummaryColumns = [
    {
      label: '二级账号ID',
      field: 'main_account_id',
    },
    {
      label: '二级账号名称',
      field: 'main_account_name',
    },
    {
      label: '运营产品',
      field: 'product_name',
    },
    {
      label: '已确认账单人民币（元）',
      field: 'current_month_rmb_cost_synced',
    },
    {
      label: '已确认账单美金（美元）',
      field: 'current_month_cost_synced',
    },
    {
      label: '当前账单人民币（元）',
      field: 'current_month_rmb_cost',
    },
    {
      label: '当前账单美金（美元）',
      field: 'current_month_cost',
    },
  ];

  const billsSummaryOperationRecordColumns = [
    {
      label: '操作时间',
      field: 'operationTime',
    },
    {
      label: '状态',
      field: 'status',
    },
    {
      label: '账单月份',
      field: 'billingMonth',
    },
    {
      label: '云厂商',
      field: 'cloudVendor',
    },
    {
      label: '一级账号ID',
      field: 'primaryAccountId',
    },
    {
      label: '操作人',
      field: 'operator',
    },
    {
      label: '人民币（元）',
      field: 'rmbAmount',
    },
    {
      label: '美金（美元）',
      field: 'usdAmount',
    },
  ];

  const businessHostColumns = [
    { type: 'selection', width: 30, minWidth: 30, onlyShowOnList: true },
    {
      label: '主机ID',
      field: 'cloud_id',
      isDefaultShow: false,
      onlyShowOnList: true,
    },
    getLinkField({
      type: 'host',
      label: '内网IP',
      field: 'private_ipv4_addresses',
      idFiled: 'id',
      onlyShowOnList: false,
      render: (data) =>
        [...(data.private_ipv4_addresses || []), ...(data.private_ipv6_addresses || [])].join(',') || '--',
      renderSuffix: (data) => {
        const ips = [...(data.private_ipv4_addresses || []), ...(data.private_ipv6_addresses || [])].join(',') || '--';
        return <CopyToClipboard content={ips} class={[cssModule['copy-icon'], 'ml4']} />;
      },
      contentClass: cssModule['cell-private-ip'],
      sort: false,
    }),
    {
      label: '公网IP',
      field: 'public_ipv4_addresses',
      isDefaultShow: true,
      onlyShowOnList: true,
      render: ({ data }: any) => {
        const ips = [...(data.public_ipv4_addresses || []), ...(data.public_ipv6_addresses || [])].join(',') || '--';
        return (
          <div class={cssModule['cell-public-ip']}>
            <span>{ips}</span>
            <CopyToClipboard content={ips} class={[cssModule['copy-icon'], 'ml4']} />
          </div>
        );
      },
    },
    {
      label: '固资号',
      field: 'bk_asset_id',
      isDefaultShow: true,
      onlyShowOnList: true,
      render: ({ data }: { data: { bk_asset_id: string } }) => {
        const text = data?.bk_asset_id || '--';
        return (
          <div class={'cell-public-ip'}>
            <span>{text}</span>
            <CopyToClipboard content={text} class={['copy-icon', 'ml4']} />
          </div>
        );
      },
    },
    {
      label: '所属VPC',
      field: 'cloud_vpc_ids',
      isDefaultShow: true,
      onlyShowOnList: true,
      render: ({ data }: any) => data.cloud_vpc_ids?.join(',') || '--',
    },
    {
      label: '云厂商',
      field: 'vendor',
      sort: true,
      onlyShowOnList: true,
      isDefaultShow: true,
      render: ({ data }: any) => <span>{VendorMap[data.vendor]}</span>,
    },
    {
      label: '地域',
      onlyShowOnList: true,
      field: 'region',
      sort: true,
      isDefaultShow: true,
      render: ({ cell, row }: { cell: string; row: { vendor: VendorEnum } }) => getRegionName(row.vendor, cell) || '--',
    },
    {
      label: '实例名称',
      field: 'name',
      sort: true,
      isDefaultShow: true,
      render: ({ cell }: { cell: string }) => cell || '--',
    },
    {
      label: '主机状态',
      field: 'status',
      sort: true,
      isDefaultShow: true,
      render({ data }: any) {
        return (
          <div class={cssModule['cvm-status-container']}>
            {HOST_SHUTDOWN_STATUS.includes(data.status) ? (
              data.status.toLowerCase() === 'stopped' ? (
                <img src={StatusUnknown} class={'mr6'} width={14} height={14}></img>
              ) : (
                <img src={StatusAbnormal} class={'mr6'} width={14} height={14}></img>
              )
            ) : HOST_RUNNING_STATUS.includes(data.status) ? (
              <img src={StatusNormal} class={'mr6'} width={14} height={14}></img>
            ) : (
              <img src={StatusUnknown} class={'mr6'} width={14} height={14}></img>
            )}
            <span>{CLOUD_HOST_STATUS[data.status] || data.status || t('未获取')}</span>
          </div>
        );
      },
    },
    {
      label: '管控区域',
      field: 'bk_cloud_id',
      sort: true,
      render({ cell }: { cell: number }) {
        if (cell !== -1) {
          return `[${cell}] ${cloudAreaStore.getNameFromCloudAreaMap(cell)}`;
        }
        return '--';
      },
    },
    {
      label: '操作系统',
      field: 'os_name',
      render: ({ data }: any) => <span>{data.os_name || '--'}</span>,
    },
    {
      label: '创建时间',
      field: 'created_at',
      sort: true,
      render: ({ cell }: { cell: string }) => timeFormatter(cell),
    },
    {
      label: '更新时间',
      field: 'updated_at',
      sort: true,
      render: ({ cell }: { cell: string }) => timeFormatter(cell),
    },
  ];

  const planDemandModColumns = [
    { type: 'selection', width: 30, minWidth: 30, onlyShowOnList: true },
    {
      label: '预测ID',
      field: 'demand_id',
      isDefaultShow: true,
    },
    {
      label: '类型',
      field: 'demand_class',
      isDefaultShow: true,
    },
    {
      label: '机型规格',
      field: 'device_type',
      isDefaultShow: true,
    },
    {
      label: '期望到货时间',
      field: 'expect_time',
      isDefaultShow: true,
    },
    {
      label: '实例剩余数',
      field: 'remained_os',
      isDefaultShow: true,
      render: ({ cell }: { cell: string }) => formatDisplayNumber(cell),
    },
    {
      label: 'CPU剩余核数',
      field: 'remained_cpu_core',
      isDefaultShow: true,
    },
    {
      label: '内存剩余量(GB)',
      field: 'remained_memory',
      isDefaultShow: true,
    },
    {
      label: '云盘剩余量(GB)',
      field: 'remained_disk_size',
      isDefaultShow: true,
    },
    {
      label: '城市',
      field: 'region_name',
    },
    {
      label: '可用区',
      field: 'zone_name',
    },
    {
      label: '项目类型',
      field: 'obs_project',
      isDefaultShow: true,
    },
    {
      label: '云磁盘类型',
      field: 'disk_type_name',
    },
    {
      label: '单实例磁盘IO(MB/s)',
      field: 'disk_io',
    },
  ];

  const columnsMap = {
    hostInventor: hIColumns,
    CloudHost: CHColumns,
    cloudRequirementSubOrder: CRSOcolumns,
    physicalRequirementSubOrder: PRSOcolumns,
    PhysicalMachine: PMColumns,
    RecyclingResources: RRColumns,
    BusinessSelection: BSAColumns,
    ResourcesTotal: RTColumns,
    hostRecycleApplication: HostRecycleApplicationColumns,
    hostRecycleDevice: HostRecycleDeviceColumns,
    deviceDestroy: deviceDestroyColumns,
    hostApplyDevice: HostApplyDeviceColumns,
    pdExecutecolumns: PDcolumns,
    ExecutionRecords: ERcolumns,
    scrResourceOnline: scrResourceOnlineColumns,
    scrResourceOffline: scrResourceOfflineColumns,
    resourceForecast: resourceForecastColumns,
    resourceForecastBatchCancel: resourceForecastBatchCancelColumns,
    receiptForecastDemand: receiptForecastDemandColumns,
    forecastDemand: forecastDemandColumns,
    adjustmentEntry: adjustmentEntryColums,
    forecastDemandDetail: forecastDemandDetailColums,
    forecastList: forecastListColums,
    account: accountColums,
    CVMApplication: CAcolumns,
    scrResourceOnlineHost: scrResourceOnlineHostColumns,
    scrResourceOfflineHost: scrResourceOfflineHostColumns,
    scrResourceOnlineCreate: scrResourceOnlineCreateColumns,
    scrResourceOfflineCreate: scrResourceOfflineCreateColumns,
    cvmModel: cvmModelColumns,
    cvmWebQuery: cvmWebColumns,
    applicationList: ApplicationListColumns,
    firstAccount: firstAccountColumns,
    secondaryAccount: secondaryAccountColumns,
    myApply: myApplyColumns,
    scrProduction: producingColumns,
    scrInitial: initialColumns,
    scrDelivery: deliveryColumns,
    decommissionDetails: decommissionDetailsColumns,
    cvmProduceQuery: cvmProduceColumns,
    cvmFastProduceQuery: cvmFastProduceColumns,
    cvmProduceDetailQuery: cvmProduceDetailColumns,
    billsRootAccountSummary: billsRootAccountSummaryColumns,
    billsMainAccountSummary: billsMainAccountSummaryColumns,
    billsSummaryOperationRecord: billsSummaryOperationRecordColumns,
    businessHostColumns,
    planDemandModColumns,
  };

  let columns = (columnsMap[type] || []).filter((column: any) => !isSimpleShow || !column.onlyShowOnList);
  if (whereAmI.value === Senarios.business) columns = columns.filter((column: any) => !column.notDisplayedInBusiness);

  type ColumnsType = typeof columns;
  const generateColumnsSettings = (columns: ColumnsType) => {
    let fields = [];
    for (const column of columns) {
      if (column.field && column.label) {
        fields.push({
          label: column.label,
          field: column.field,
          disabled: type !== 'cvms' && column.field === 'id',
          isDefaultShow: !!column.isDefaultShow,
          notDisplayedInBusiness: !!column.notDisplayedInBusiness,
        });
      }
    }
    if (whereAmI.value === Senarios.business) {
      fields = fields.filter((field) => !field.notDisplayedInBusiness);
    }
    const settings: Ref<Settings> = ref({
      fields,
      checked: fields.filter((field) => field.isDefaultShow).map((field) => field.field),
    });

    return settings;
  };

  const settings = generateColumnsSettings(columns);

  return {
    columns,
    settings,
    generateColumnsSettings,
  };
};
