/* eslint-disable no-nested-ternary */
// table 字段相关信息
import i18n from '@/language/i18n';
import { CloudType, SecurityRuleEnum, HuaweiSecurityRuleEnum, AzureSecurityRuleEnum } from '@/typings';
import { useAccountStore, useLoadBalancerStore } from '@/store';
import { Button } from 'bkui-vue';
import { type Settings } from 'bkui-vue/lib/table/props';
import { h, ref } from 'vue';
import type { Ref } from 'vue';
import { RouteLocationRaw, useRoute, useRouter } from 'vue-router';
import { CLB_BINDING_STATUS, CLOUD_HOST_STATUS, VendorEnum, VendorMap } from '@/common/constant';
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
import './use-columns.scss';
import { defaults } from 'lodash';
import { timeFormatter } from '@/common/util';
import { IP_VERSION_MAP, LBRouteName, LB_NETWORK_TYPE_MAP, SCHEDULER_MAP } from '@/constants/clb';
import { formatBillCost, getInstVip } from '@/utils';
import { Spinner } from 'bkui-vue/lib/icon';
import { APPLICATION_TYPE_MAP } from '@/views/service/apply-list/constants';
import dayjs from 'dayjs';
import { BILLS_ROOT_ACCOUNT_SUMMARY_STATE_MAP, BILL_TYPE__MAP_HW, CURRENCY_MAP } from '@/constants';
import { BILL_VENDORS_MAP, BILL_SITE_TYPES_MAP } from '@/views/bill/account/account-manage/constants';
import CopyToClipboard from '@/components/copy-to-clipboard/index.vue';

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

export default (type: string, isSimpleShow = false, vendor?: string) => {
  const router = useRouter();
  const route = useRoute();
  const accountStore = useAccountStore();
  const loadBalancerStore = useLoadBalancerStore();
  const { t } = i18n.global;
  const { getRegionName } = useRegionsStore();
  const { whereAmI } = useWhereAmI();
  const businessMapStore = useBusinessMapStore();
  const cloudAreaStore = useCloudAreaStore();

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
   * todo: 更换实现方式, 取消使用stopPropagation
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

  const vpcColumns = [
    { type: 'selection', width: 30, minWidth: 30, onlyShowOnList: true },
    getLinkField({ type: 'vpc', label: 'VPC ID', field: 'cloud_id' }),
    // {
    //   label: '资源 ID',
    //   field: 'cloud_id',
    //   sort: true,
    //   isDefaultShow: true,
    //   render({ cell }: { cell: string }) {
    //     return h('span', [cell || '--']);
    //   },
    // },
    {
      label: 'VPC 名称',
      field: 'name',
      sort: true,
      isDefaultShow: true,
      render({ cell }: { cell: string }) {
        return h('span', [cell || '--']);
      },
    },
    {
      label: '云厂商',
      field: 'vendor',
      sort: true,
      isDefaultShow: true,
      render({ cell }: { cell: string }) {
        return h('span', [CloudType[cell] || '--']);
      },
    },
    {
      label: '地域',
      field: 'region',
      sort: true,
      isDefaultShow: true,
      render: ({ cell, row }: { cell: string; row: { vendor: VendorEnum } }) => getRegionName(row.vendor, cell) || '--',
    },
    {
      label: '是否分配',
      field: 'bk_biz_id',
      sort: true,
      isOnlyShowInResource: true,
      isDefaultShow: true,
      render: ({ data, cell }: { data: { bk_biz_id: number }; cell: number }) => (
        <bk-tag
          v-bk-tooltips={{
            content: businessMapStore.businessMap.get(cell),
            disabled: !cell || cell === -1,
            theme: 'light',
          }}
          theme={data.bk_biz_id === -1 ? false : 'success'}>
          {data.bk_biz_id === -1 ? '未分配' : '已分配'}
        </bk-tag>
      ),
    },
    {
      label: '所属业务',
      field: 'bk_biz_id2',
      isOnlyShowInResource: true,
      render: ({ data }: any) => businessMapStore.businessMap.get(data.bk_biz_id) || '未分配',
    },
    {
      label: '管控区域',
      field: 'bk_cloud_id',
      isDefaultShow: true,
      render({ cell }: { cell: number }) {
        if (cell !== -1) {
          return `[${cell}] ${cloudAreaStore.getNameFromCloudAreaMap(cell)}`;
        }
        return '--';
      },
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

  const subnetColumns = [
    { type: 'selection', width: 30, minWidth: 30, onlyShowOnList: true },
    getLinkField({ type: 'subnet', label: '子网 ID', field: 'cloud_id', idFiled: 'id', onlyShowOnList: false }),
    // {
    //   label: '资源 ID',
    //   field: 'cloud_id',
    //   sort: true,
    //   render({ cell }: { cell: string }) {
    //     const index =          cell.lastIndexOf('/') <= 0 ? 0 : cell.lastIndexOf('/') + 1;
    //     const value = cell.slice(index);
    //     return h('span', [value || '--']);
    //   },
    // },
    {
      label: '子网名称',
      field: 'name',
      sort: true,
      render({ cell }: { cell: string }) {
        return h('span', [cell || '--']);
      },
    },
    {
      label: '云厂商',
      field: 'vendor',
      sort: true,
      isDefaultShow: true,
      render({ cell }: { cell: string }) {
        return h('span', [CloudType[cell] || '--']);
      },
    },
    {
      label: '地域',
      field: 'region',
      sort: true,
      isDefaultShow: true,
      render: ({ cell, row }: { cell: string; row: { vendor: VendorEnum } }) => getRegionName(row.vendor, cell),
    },
    {
      label: '可用区',
      field: 'zone',
      sort: true,
      isDefaultShow: true,
      render({ cell }: { cell: string }) {
        return h('span', [cell || '--']);
      },
    },
    getLinkField({ type: 'vpc', label: '所属 VPC', field: 'cloud_vpc_id', idFiled: 'vpc_id', onlyShowOnList: false }),
    {
      label: 'IPv4 CIDR',
      field: 'ipv4_cidr',
      isDefaultShow: true,
      render({ cell }: { cell: string }) {
        return h('span', [cell || '--']);
      },
    },
    // getLinkField(
    //   'route',
    //   '关联路由表',
    //   'route_table_id',
    //   'route_table_id',
    //   false,
    // ),
    // {
    //   label: '可用IPv4地址数',
    //   field: 'count_of_ipv4_cidr',
    //   isDefaultShow: true,
    //   render({ data }: any) {
    //     return data.ipv4_cidr.length;
    //   },
    // },
    {
      label: '是否分配',
      field: 'bk_biz_id',
      sort: true,
      isOnlyShowInResource: true,
      isDefaultShow: true,
      render: ({ data, cell }: { data: { bk_biz_id: number }; cell: number }) => (
        <bk-tag
          v-bk-tooltips={{
            content: businessMapStore.businessMap.get(cell),
            disabled: !cell || cell === -1,
            theme: 'light',
          }}
          theme={data.bk_biz_id === -1 ? false : 'success'}>
          {data.bk_biz_id === -1 ? '未分配' : '已分配'}
        </bk-tag>
      ),
    },
    {
      label: '所属业务',
      field: 'bk_biz_id2',
      isOnlyShowInResource: true,
      render: ({ data }: any) => businessMapStore.businessMap.get(data.bk_biz_id) || '未分配',
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

  const groupColumns = [
    { type: 'selection', width: 30, minWidth: 30, onlyShowOnList: true },
    getLinkField({ type: 'subnet' }),
    {
      label: '资源 ID',
      field: 'account_id',
      sort: true,
    },
    {
      label: '名称',
      field: 'name',
      sort: true,
    },
    {
      label: t('云厂商'),
      render({ data }: any) {
        return h('span', {}, [CloudType[data.vendor]]);
      },
    },
    {
      label: '地域',
      field: 'region',
      render: ({ cell, row }: { cell: string; row: { vendor: VendorEnum } }) => getRegionName(row.vendor, cell),
    },
    {
      label: '描述',
      field: 'memo',
    },
  ];

  const gcpColumns = [
    { type: 'selection', width: 30, minWidth: 30, onlyShowOnList: true },
    getLinkField({ type: 'subnet' }),
    {
      label: '资源 ID',
      field: 'account_id',
      sort: true,
    },
    {
      label: '名称',
      field: 'name',
      sort: true,
    },
    // {
    //   label: '业务',
    //   render({ cell }: any) {
    //     return h(
    //       'span',
    //       {},
    //       [
    //         cell,
    //       ],
    //     );
    //   },
    // },
    // {
    //   label: '业务拓扑',
    //   field: 'zone',
    // },
    {
      label: 'VPC',
      field: 'vpc_id',
    },
    {
      label: '描述',
      field: 'memo',
    },
  ];

  const driveColumns: any[] = [
    { type: 'selection', width: 30, minWidth: 30, onlyShowOnList: true },
    getLinkField({ type: 'drive', label: '云硬盘ID', field: 'cloud_id' }),
    // {
    //   label: '资源 ID',
    //   field: 'cloud_id',
    //   sort: true,
    // },
    {
      label: '云硬盘名称',
      field: 'name',
      sort: true,
      isDefaultShow: true,
      render: ({ cell }: any) => cell || '--',
    },
    {
      label: '云厂商',
      field: 'vendor',
      sort: true,
      isDefaultShow: true,
      render({ cell }: { cell: string }) {
        return h('span', [CloudType[cell] || '--']);
      },
    },
    {
      label: '地域',
      field: 'region',
      sort: true,
      isDefaultShow: true,
      render: ({ cell, row }: { cell: string; row: { vendor: VendorEnum } }) => getRegionName(row.vendor, cell),
    },
    {
      label: '可用区',
      field: 'zone',
      sort: true,
      isDefaultShow: true,
      render({ cell }: { cell: string }) {
        return h('span', [cell || '--']);
      },
    },
    {
      label: '云硬盘状态',
      field: 'status',
      sort: true,
      isDefaultShow: true,
      render({ cell }: { cell: string }) {
        return h('span', [cell || '--']);
      },
    },
    {
      label: '硬盘分类',
      field: 'is_system_disk',
      sort: true,
      isDefaultShow: true,
      render({ cell }: { cell: boolean }) {
        return h('span', [cell ? '系统盘' : '数据盘']);
      },
    },
    {
      label: '类型',
      field: 'disk_type',
      sort: true,
      isDefaultShow: true,
      render({ cell }: { cell: string }) {
        return h('span', [cell || '--']);
      },
    },
    {
      label: '容量(GB)',
      field: 'disk_size',
      sort: true,
      isDefaultShow: true,
      render({ cell }: { cell: string }) {
        return h('span', [cell || '--']);
      },
    },
    getLinkField({ type: 'host', label: '挂载的主机', field: 'instance_id', idFiled: 'instance_id' }),
    {
      label: '是否分配',
      field: 'bk_biz_id',
      sort: true,
      isOnlyShowInResource: true,
      isDefaultShow: true,
      render: ({ data, cell }: { data: { bk_biz_id: number }; cell: number }) => (
        <bk-tag
          v-bk-tooltips={{
            content: businessMapStore.businessMap.get(cell),
            disabled: !cell || cell === -1,
            theme: 'light',
          }}
          theme={data.bk_biz_id === -1 ? false : 'success'}>
          {data.bk_biz_id === -1 ? '未分配' : '已分配'}
        </bk-tag>
      ),
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

  const imageColumns = [
    getLinkField({ type: 'image', label: '镜像ID', field: 'cloud_id', idFiled: 'id' }),
    // {
    //   label: '资源 ID',
    //   field: 'cloud_id',
    //   sort: true,
    // },
    {
      label: '镜像名称',
      field: 'name',
      sort: true,
      isDefaultShow: true,
    },
    {
      label: '云厂商',
      field: 'vendor',
      sort: true,
      isDefaultShow: true,
      render({ cell }: { cell: string }) {
        return h('span', [CloudType[cell] || '--']);
      },
    },
    {
      label: '操作系统类型',
      field: 'platform',
      sort: true,
      isDefaultShow: true,
    },
    {
      label: '架构',
      field: 'architecture',
      sort: true,
      isDefaultShow: true,
    },
    {
      label: '状态',
      field: 'state',
      sort: true,
      isDefaultShow: true,
    },
    {
      label: '类型',
      field: 'type',
      sort: true,
      isDefaultShow: true,
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

  const networkInterfaceColumns = [
    getLinkField({ type: 'network-interface', label: '接口 ID', field: 'cloud_id', idFiled: 'id' }),
    {
      label: '接口名称',
      field: 'name',
      sort: true,
      isDefaultShow: true,
    },
    {
      label: '云厂商',
      field: 'vendor',
      sort: true,
      isDefaultShow: true,
      render({ cell }: { cell: string }) {
        return h('span', [CloudType[cell] || '--']);
      },
    },
    {
      label: '地域',
      field: 'region',
      sort: true,
      isDefaultShow: true,
      render: ({ cell, row }: { cell: string; row: { vendor: VendorEnum } }) => getRegionName(row.vendor, cell),
    },
    {
      label: '可用区',
      field: 'zone',
      render({ cell }: { cell: string }) {
        return h('span', [cell || '--']);
      },
    },
    {
      label: '所属VPC',
      field: 'cloud_vpc_id',
      sort: true,
      isDefaultShow: true,
      showOverflowTooltip: true,
      render({ cell }: { cell: string }) {
        return h('span', [cell || '--']);
      },
    },
    {
      label: '所属子网',
      showOverflowTooltip: true,
      field: 'cloud_subnet_id',
      sort: true,
      isDefaultShow: true,
      render({ cell }: { cell: string }) {
        return h('span', [cell || '--']);
      },
    },
    // {
    //   label: '关联的实例',
    //   field: 'instance_id',
    //   showOverflowTooltip: true,
    //   render({ cell }: { cell: string }) {
    //     return h('span', [cell || '--']);
    //   },
    // },
    {
      label: '内网IP',
      field: 'private_ipv4_or_ipv6',
      isDefaultShow: true,
      render({ data }: any) {
        return [h('span', {}, [data?.private_ipv4.join(',') || data?.private_ipv6.join(',') || '--'])];
      },
    },
    {
      label: '关联的公网IP地址',
      field: 'public_ip',
      // 目前公网IP地址不支持排序
      // sort: true,
      isDefaultShow: true,
      render({ data }: any) {
        return [h('span', {}, [data?.public_ipv4.join(',') || data?.public_ipv6.join(',') || '--'])];
      },
    },
    {
      label: '所属业务',
      field: 'bk_biz_id',
      sort: true,
      isOnlyShowInResource: true,
      render: ({ data }: any) => businessMapStore.businessMap.get(data.bk_biz_id) || '未分配',
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

  const routeColumns = [
    getLinkField({ type: 'route', label: '路由表ID', field: 'cloud_id', idFiled: 'id' }),
    // {
    //   label: '资源 ID',
    //   field: 'cloud_id',
    //   sort: true,
    // },
    {
      label: '路由表名称',
      field: 'name',
      sort: true,
      isDefaultShow: true,
      render: ({ cell }: any) => cell || '--',
    },
    {
      label: '云厂商',
      field: 'vendor',
      sort: true,
      isDefaultShow: true,
      render({ cell }: { cell: string }) {
        return h('span', [CloudType[cell] || '--']);
      },
    },
    {
      label: '地域',
      field: 'region',
      sort: true,
      isDefaultShow: true,
      render: ({ cell, row }: { cell: string; row: { vendor: VendorEnum } }) => getRegionName(row.vendor, cell),
    },
    getLinkField({ type: 'vpc', label: '所属网络(VPC)', field: 'vpc_id', idFiled: 'vpc_id' }),
    // {
    //   label: '关联子网',
    //   field: '',
    //   sort: true,
    // },
    {
      label: '所属业务',
      field: 'bk_biz_id',
      isOnlyShowInResource: true,
      sort: true,
      render: ({ data }: any) => businessMapStore.businessMap.get(data.bk_biz_id) || '未分配',
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

  const cvmsColumns = [
    { type: 'selection', width: 30, minWidth: 30, onlyShowOnList: true },
    //   移除 ID 搜索条件
    // {
    //   label: 'ID',
    //   field: 'id',
    //   isDefaultShow: false,
    //   onlyShowOnList: true,
    // },
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
        return <CopyToClipboard content={ips} class={['copy-icon', 'ml4']} />;
      },
      contentClass: 'cell-private-ip',
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
          <div class={'cell-public-ip'}>
            <span>{ips}</span>
            <CopyToClipboard content={ips} class={['copy-icon', 'ml4']} />
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
      render({ data }: any) {
        return h('span', {}, [CloudType[data.vendor]]);
      },
    },
    {
      label: '地域',
      onlyShowOnList: true,
      field: 'region',
      sort: true,
      isDefaultShow: true,
      render: ({ cell, row }: { cell: string; row: { vendor: VendorEnum } }) => getRegionName(row.vendor, cell),
    },
    {
      label: '主机名称',
      field: 'name',
      sort: true,
      isDefaultShow: true,
    },
    {
      label: '主机状态',
      field: 'status',
      sort: true,
      isDefaultShow: true,
      render({ data }: any) {
        // return h('span', {}, [CLOUD_HOST_STATUS[data.status] || data.status]);
        return (
          <div class={'cvm-status-container'}>
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
            <span>{CLOUD_HOST_STATUS[data.status] || data.status}</span>
          </div>
        );
      },
    },
    {
      label: '是否分配',
      field: 'bk_biz_id',
      sort: true,
      isOnlyShowInResource: true,
      isDefaultShow: true,
      render: ({ data, cell }: { data: { bk_biz_id: number }; cell: number }) => (
        <bk-tag
          v-bk-tooltips={{
            content: businessMapStore.businessMap.get(cell),
            disabled: !cell || cell === -1,
            theme: 'light',
          }}
          theme={data.bk_biz_id === -1 ? false : 'success'}>
          {data.bk_biz_id === -1 ? '未分配' : '已分配'}
        </bk-tag>
      ),
    },

    {
      label: '所属业务',
      field: 'bk_biz_id2',
      isOnlyShowInResource: true,
      render: ({ data }: any) => businessMapStore.businessMap.get(data.bk_biz_id) || '未分配',
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
      label: '实例规格',
      field: 'machine_type',
      sort: true,
      isOnlyShowInResource: true,
    },
    {
      label: '操作系统',
      field: 'os_name',
      render({ data }: any) {
        return h('span', {}, [data.os_name || '--']);
      },
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

  const securityCommonColumns = [
    {
      label: t('来源'),
      field: 'resource',
      render({ data }: any) {
        return h('span', {}, [
          data.cloud_address_group_id ||
            data.cloud_address_id ||
            data.cloud_service_group_id ||
            data.cloud_service_id ||
            data.cloud_target_security_group_id ||
            data.ipv4_cidr ||
            data.ipv6_cidr ||
            data.cloud_remote_group_id ||
            data.remote_ip_prefix ||
            (data.source_address_prefix === '*' ? t('任何') : data.source_address_prefix) ||
            data.source_address_prefixes ||
            data.cloud_source_security_group_ids ||
            (data.destination_address_prefix === '*' ? t('任何') : data.destination_address_prefix) ||
            data.destination_address_prefixes ||
            data.cloud_destination_security_group_ids ||
            '--',
        ]);
      },
    },
    {
      label: '协议端口',
      render({ data }: any) {
        return h('span', {}, [
          // eslint-disable-next-line no-nested-ternary
          vendor === 'aws' && data.protocol === '-1' && data.to_port === -1
            ? t('全部')
            : vendor === 'huawei' && !data.protocol && !data.port
            ? t('全部')
            : vendor === 'azure' && data.protocol === '*' && data.destination_port_range === '*'
            ? t('全部')
            : `${data.protocol}:${data.port || data.to_port || data.destination_port_range || '--'}`,
        ]);
      },
    },
    {
      label: t('策略'),
      render({ data }: any) {
        return h('span', {}, [
          // eslint-disable-next-line no-nested-ternary
          vendor === 'huawei'
            ? HuaweiSecurityRuleEnum[data.action]
            : vendor === 'azure'
            ? AzureSecurityRuleEnum[data.access]
            : vendor === 'aws'
            ? t('允许')
            : SecurityRuleEnum[data.action] || '--',
        ]);
      },
    },
    {
      label: '备注',
      field: 'memo',
      render({ data }: any) {
        return h('span', {}, [data.memo || '--']);
      },
    },
    {
      label: t('修改时间'),
      field: 'updated_at',
      render: ({ cell }: { cell: string }) => timeFormatter(cell),
    },
  ];

  const eipColumns = [
    { type: 'selection', width: 30, minWidth: 30, onlyShowOnList: true },
    getLinkField({ type: 'eips', label: 'IP资源ID', field: 'cloud_id', idFiled: 'id' }),
    // {
    //   label: '资源 ID',
    //   field: 'cloud_id',
    //   sort: true,
    // },
    {
      label: 'IP名称',
      field: 'name',
      sort: true,
      isDefaultShow: true,
      render({ cell }: { cell: string }) {
        return h('span', [cell || '--']);
      },
    },
    {
      label: '云厂商',
      field: 'vendor',
      sort: true,
      isDefaultShow: true,
      render({ cell }: { cell: string }) {
        return h('span', [CloudType[cell] || '--']);
      },
    },
    {
      label: '地域',
      field: 'region',
      sort: true,
      isDefaultShow: true,
      render: ({ cell, row }: { cell: string; row: { vendor: VendorEnum } }) => getRegionName(row.vendor, cell),
    },
    {
      label: '公网 IP',
      field: 'public_ip',
      sort: true,
      isDefaultShow: true,
      render({ cell }: { cell: string }) {
        return h('span', [cell || '--']);
      },
    },
    // {
    //   label: '状态',
    //   field: 'status',
    //   render({ cell }: { cell: string }) {
    //     return h('span', [cell || '--']);
    //   },
    // },
    getLinkField({
      type: 'host',
      label: '绑定的资源实例',
      field: 'cvm_id',
      idFiled: 'cvm_id',
      onlyShowOnList: false,
      render: (data) => data.host,
      sort: false,
    }),
    {
      label: '绑定的资源类型',
      field: 'instance_type',
      isDefaultShow: true,
      render({ cell }: { cell: string }) {
        return h('span', [cell || '--']);
      },
    },
    {
      label: '是否分配',
      field: 'bk_biz_id',
      sort: true,
      isOnlyShowInResource: true,
      isDefaultShow: true,
      render: ({ data, cell }: { data: { bk_biz_id: number }; cell: number }) => (
        <bk-tag
          v-bk-tooltips={{
            content: businessMapStore.businessMap.get(cell),
            disabled: !cell || cell === -1,
            theme: 'light',
          }}
          theme={data.bk_biz_id === -1 ? false : 'success'}>
          {data.bk_biz_id === -1 ? '未分配' : '已分配'}
        </bk-tag>
      ),
    },
    {
      label: '所属业务',
      field: 'bk_biz_id2',
      isOnlyShowInResource: true,
      render: ({ data }: any) => businessMapStore.businessMap.get(data.bk_biz_id) || '未分配',
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

  const operationRecordColumns = [
    {
      label: '操作时间',
      field: 'created_at',
      isDefaultShow: true,
      sort: true,
      render: ({ cell }: { cell: string }) => timeFormatter(cell),
    },
    {
      label: '资源类型',
      field: 'res_type',
    },
    {
      label: '资源名称',
      field: 'res_name',
      isDefaultShow: true,
    },
    // {
    //   label: '云资源ID',
    //   field: 'cloud_res_id',
    // },
    {
      label: '操作方式',
      field: 'action',
      isDefaultShow: true,
      filter: true,
    },
    {
      label: '操作来源',
      field: 'source',
      isDefaultShow: true,
      filter: true,
    },
    {
      label: '所属业务',
      field: 'bk_biz_id',
      isOnlyShowInResource: true,
      render: ({ cell }: { cell: number }) => businessMapStore.businessMap.get(cell) || '未分配',
    },
    // {
    //   label: '云厂商',
    //   field: 'vendor',
    // },
    {
      label: '云账号',
      field: 'account_id',
    },
    {
      label: '操作人',
      field: 'operator',
      isDefaultShow: true,
    },
  ];

  const lbColumns = [
    { type: 'selection', width: 30, minWidth: 30, onlyShowOnList: true },
    getLinkField({
      type: 'lb',
      label: '负载均衡名称',
      field: 'name',
      linkable: () => whereAmI.value === Senarios.business,
      render: (data) => (
        <Button
          text
          theme='primary'
          onClick={renderFieldPushState(
            {
              name: LBRouteName.lb,
              params: { id: data.id },
              query: { ...route.query, type: 'detail' },
            },
            () => {
              loadBalancerStore.setLbTreeSearchTarget({ ...data, searchK: 'lb_name', searchV: data.name, type: 'lb' });
            },
          )}>
          {data.name || '--'}
        </Button>
      ),
    }),
    {
      label: () => (
        <span v-bk-tooltips={{ content: '用户通过该域名访问负载均衡流量', placement: 'top' }}>负载均衡域名</span>
      ),
      field: 'domain',
      isDefaultShow: true,
      render: ({ cell }: { cell: string }) => cell || '--',
    },
    {
      label: '负载均衡VIP',
      field: 'vip',
      isDefaultShow: true,
      render: ({ data }: any) => {
        return getInstVip(data);
      },
    },
    {
      label: '网络类型',
      field: 'lb_type',
      isDefaultShow: true,
      sort: true,
      filter: {
        list: [
          { text: LB_NETWORK_TYPE_MAP.OPEN, value: LB_NETWORK_TYPE_MAP.OPEN },
          { text: LB_NETWORK_TYPE_MAP.INTERNAL, value: LB_NETWORK_TYPE_MAP.INTERNAL },
        ],
      },
    },
    {
      label: '监听器数量',
      field: 'listenerNum',
      isDefaultShow: true,
      render: ({ cell }: { cell: number }) => cell || '0',
    },
    {
      label: '分配状态',
      field: 'bk_biz_id',
      isDefaultShow: true,
      isOnlyShowInResource: true,
      render: ({ cell }: { cell: number }) => (
        <bk-tag
          v-bk-tooltips={{
            content: businessMapStore.businessMap.get(cell),
            disabled: !cell || cell === -1,
          }}
          theme={cell === -1 ? false : 'success'}>
          {cell === -1 ? '未分配' : '已分配'}
        </bk-tag>
      ),
    },
    {
      label: '删除保护',
      field: 'delete_protect',
      isDefaultShow: true,
      render: ({ cell }: { cell: boolean }) => (cell ? <bk-tag theme='success'>开启</bk-tag> : <bk-tag>关闭</bk-tag>),
      filter: {
        list: [
          { text: '开启', value: true },
          { text: '关闭', value: false },
        ],
      },
    },
    {
      label: 'IP版本',
      field: 'ip_version',
      isDefaultShow: true,
      render: ({ cell }: { cell: string }) => IP_VERSION_MAP[cell],
      sort: true,
    },
    {
      label: '云厂商',
      field: 'vendor',
      render({ cell }: { cell: string }) {
        return h('span', [CloudType[cell] || '--']);
      },
      sort: true,
    },
    {
      label: '地域',
      field: 'region',
      render: ({ cell, row }: { cell: string; row: { vendor: VendorEnum } }) => getRegionName(row.vendor, cell) || '--',
      sort: true,
    },
    {
      label: '可用区域',
      field: 'zones',
      render: ({ cell }: { cell: string[] }) => cell?.join(','),
      sort: true,
    },
    {
      label: '状态',
      field: 'status',
      sort: true,
      render: ({ cell }: { cell: string }) => {
        let icon = StatusSuccess;
        switch (cell) {
          case '创建中':
            icon = StatusLoading;
            break;
          case '正常运行':
            icon = StatusSuccess;
            break;
        }
        return cell ? (
          <div class='status-column-cell'>
            <img class={`status-icon${cell === 'binding' ? ' spin-icon' : ''}`} src={icon} alt='' />
            <span>{cell}</span>
          </div>
        ) : (
          '--'
        );
      },
    },
    {
      label: '所属vpc',
      field: 'cloud_vpc_id',
      sort: true,
    },
  ];

  const listenerColumns = [
    getLinkField({
      type: 'listener',
      label: '监听器名称',
      field: 'name',
      render: (data) => (
        <Button
          text
          theme='primary'
          onClick={renderFieldPushState(
            {
              name: LBRouteName.listener,
              params: { id: data.id },
              query: { ...route.query, type: 'detail', protocol: data.protocol },
            },
            () => {
              loadBalancerStore.setLbTreeSearchTarget({
                ...data,
                searchK: 'listener_name',
                searchV: data.name,
                type: 'listener',
              });
            },
          )}>
          {data.name || '--'}
        </Button>
      ),
    }),
    {
      label: '监听器ID',
      field: 'cloud_id',
    },
    {
      label: '协议',
      field: 'protocol',
      isDefaultShow: true,
    },
    {
      label: '端口',
      field: 'port',
      isDefaultShow: true,
      render: ({ data, cell }: any) => `${cell}${data.end_port ? `-${data.end_port}` : ''}`,
    },
    {
      label: '均衡方式',
      field: 'scheduler',
      isDefaultShow: true,
      render: ({ cell }: { cell: string }) => SCHEDULER_MAP[cell] || '--',
    },
    {
      label: '域名数量',
      field: 'domain_num',
      isDefaultShow: true,
    },
    {
      label: 'URL数量',
      field: 'url_num',
      isDefaultShow: true,
    },
    {
      label: '同步状态',
      field: 'binding_status',
      isDefaultShow: true,
      render: ({ cell }: { cell: string }) => {
        let icon = StatusSuccess;
        switch (cell) {
          case 'binding':
            icon = StatusLoading;
            break;
          case 'success':
            icon = StatusSuccess;
            break;
        }
        return cell ? (
          <div class='status-column-cell'>
            <img class={`status-icon${cell === 'binding' ? ' spin-icon' : ''}`} src={icon} alt='' />
            <span>{CLB_BINDING_STATUS[cell]}</span>
          </div>
        ) : (
          '--'
        );
      },
    },
  ];

  const targetGroupColumns = [
    { type: 'selection', width: 30, minWidth: 30, onlyShowOnList: true },
    getLinkField({
      type: 'name',
      label: '目标组名称',
      field: 'name',
      idFiled: 'name',
      onlyShowOnList: false,
      render: ({ id, name }) => (
        <Button
          text
          theme='primary'
          onClick={renderFieldPushState(
            {
              name: LBRouteName.tg,
              params: { id },
              query: { ...route.query, type: 'detail' },
            },
            () => {
              loadBalancerStore.setTgSearchTarget(name);
            },
          )}>
          {name}
        </Button>
      ),
    }),
    {
      label: '关联的负载均衡',
      field: 'lb_name',
      isDefaultShow: true,
      render({ cell }: any) {
        return cell?.trim() || '--';
      },
    },
    {
      label: '绑定监听器数量',
      field: 'listener_num',
      isDefaultShow: true,
    },
    {
      label: '协议',
      field: 'protocol',
      render({ cell }: any) {
        return cell?.trim() || '--';
      },
      isDefaultShow: true,
      sort: true,
      filter: {
        list: [
          { value: 'TCP', text: 'TCP' },
          { value: 'UDP', text: 'UDP' },
          { value: 'HTTP', text: 'HTTP' },
          { value: 'HTTPS', text: 'HTTPS' },
        ],
      },
    },
    {
      label: '端口',
      field: 'port',
      isDefaultShow: true,
      sort: true,
    },
    {
      label: '健康检查',
      field: 'health_check.health_switch',
      isDefaultShow: true,
      filter: {
        list: [
          { value: 1, text: '已开启' },
          { value: 0, text: '未开启' },
        ],
      },
      render({ cell }: { cell: Number }) {
        return cell ? <bk-tag theme='success'>已开启</bk-tag> : <bk-tag>未开启</bk-tag>;
      },
    },
    {
      label: '云厂商',
      field: 'vendor',
      render({ cell }: { cell: string }) {
        return h('span', [CloudType[cell] || '--']);
      },
      sort: true,
      filter: {
        list: [{ value: VendorEnum.TCLOUD, text: VendorMap[VendorEnum.TCLOUD] }],
      },
    },
    {
      label: '地域',
      field: 'region',
      render: ({ cell, row }: { cell: string; row: { vendor: VendorEnum } }) => getRegionName(row.vendor, cell) || '--',
      sort: true,
    },
    {
      label: '所属VPC',
      field: 'cloud_vpc_id',
      sort: true,
    },
    {
      label: '健康检查端口',
      field: 'health_check',
      render: ({ cell }: any) => {
        const { health_num, un_health_num } = cell;
        const total = health_num + un_health_num;
        if (!health_num || !un_health_num) return '--';
        return (
          <div class='port-status-col'>
            <span class={un_health_num ? 'un-health' : total ? 'health' : 'special-health'}>{un_health_num}</span>/
            <span>{health_num + un_health_num}</span>
          </div>
        );
      },
    },
  ];

  const rsConfigColumns = [
    {
      label: '内网IP',
      field: 'private_ip_address',
      isDefaultShow: true,
      render: ({ data }: any) => {
        return [
          ...(data.private_ipv4_addresses || []),
          ...(data.private_ipv6_addresses || []),
          // 更新目标组detail中的rs字段
          ...(data.private_ip_address || []),
        ].join(',');
      },
    },
    {
      label: '公网IP',
      field: 'public_ip_address',
      render: ({ data }: any) => {
        return (
          [
            ...(data.public_ipv4_addresses || []),
            ...(data.public_ipv6_addresses || []),
            // 更新目标组detail中的rs字段
            ...(data.public_ip_address || []),
          ].join(',') || '--'
        );
      },
    },
    {
      label: '名称',
      field: 'name',
      isDefaultShow: true,
      render: ({ data }: any) => {
        return data.name || data.inst_name;
      },
    },
    {
      label: '地域',
      field: 'region',
      render: ({ cell }: { cell: string }) => getRegionName(VendorEnum.TCLOUD, cell) || '--',
    },
    {
      label: '资源类型',
      field: 'inst_type',
      render: ({ data }: any) => {
        return data.machine_type || data.inst_type;
      },
    },
    {
      label: '所属VPC',
      field: 'cloud_vpc_ids',
      isDefaultShow: true,
      render: ({ cell }: { cell: string[] }) => cell?.join(','),
    },
  ];

  const domainColumns = [
    {
      label: 'URL数量',
      field: 'url_count',
      isDefaultShow: true,
      sort: true,
    },
    {
      label: '同步状态',
      field: 'sync_status',
      isDefaultShow: true,
      render: ({ cell }: { cell: string }) => {
        let icon = StatusSuccess;
        switch (cell) {
          case 'binding':
            icon = StatusLoading;
            break;
          case 'success':
            icon = StatusSuccess;
            break;
        }
        return cell ? (
          <div class='status-column-cell'>
            <img class={`status-icon${cell === 'binding' ? ' spin-icon' : ''}`} src={icon} alt='' />
            <span>{CLB_BINDING_STATUS[cell]}</span>
          </div>
        ) : (
          '--'
        );
      },
    },
  ];

  const targetGroupListenerColumns = [
    getLinkField({
      type: 'targetGroup',
      label: '绑定的监听器',
      field: 'lbl_name',
      render: ({ lbl_id, lbl_name, protocol }: any) => (
        <Button
          text
          theme='primary'
          onClick={renderFieldPushState({
            name: LBRouteName.listener,
            params: { id: lbl_id },
            query: {
              ...route.query,
              type: 'detail',
              protocol,
            },
          })}>
          {lbl_name}
        </Button>
      ),
    }),
    {
      label: '关联的负载均衡',
      field: 'lb_name',
      isDefaultShow: true,
      width: 300,
      render: ({ data }: any) => {
        const vip = getInstVip(data);
        const { lb_name } = data;
        return `${lb_name}（${vip}）`;
      },
    },
    {
      label: '关联的URL',
      field: 'url',
      isDefaultShow: true,
      render: ({ cell }: { cell: string }) => cell || '--',
    },
    {
      label: '协议',
      field: 'protocol',
      isDefaultShow: true,
      filter: {
        list: [
          { value: 'TCP', text: 'TCP' },
          { value: 'UDP', text: 'UDP' },
          { value: 'HTTP', text: 'HTTP' },
          { value: 'HTTPS', text: 'HTTPS' },
        ],
      },
    },
    {
      label: '端口',
      field: 'port',
      isDefaultShow: true,
      render: ({ data, cell }: any) => `${cell}${data.end_port ? `-${data.end_port}` : ''}`,
    },
    {
      label: '异常端口数',
      field: 'healthCheck',
      isDefaultShow: true,
      render: ({ cell }: any) => {
        if (!cell) return '--';
        const { health_num, un_health_num } = cell;
        return (
          <div class='port-status-col'>
            <span class={un_health_num ? 'un-health' : 'health'}>{un_health_num}</span>/
            <span>{health_num + un_health_num}</span>
          </div>
        );
      },
    },
  ];

  const urlColumns = [
    { type: 'selection', width: 30, minWidth: 30, onlyShowOnList: true },
    {
      label: 'URL路径',
      field: 'url',
      isDefaultShow: true,
      sort: true,
    },
    {
      label: '轮询方式',
      field: 'scheduler',
      isDefaultShow: true,
      render: ({ cell }: { cell: string }) => SCHEDULER_MAP[cell] || '--',
      sort: true,
    },
  ];

  const certColumns = [
    {
      label: '证书名称',
      field: 'name',
    },
    {
      label: '资源ID',
      field: 'cloud_id',
    },
    {
      label: '云厂商',
      field: 'vendor',
      render({ cell }: { cell: string }) {
        return h('span', [CloudType[cell] || '--']);
      },
    },
    {
      label: '证书类型',
      field: 'cert_type',
      filter: {
        list: [
          {
            text: '服务器证书',
            value: '服务器证书',
          },
          {
            text: '客户端CA证书',
            value: '客户端CA证书',
          },
        ],
      },
    },
    {
      label: '域名',
      field: 'domain',
      render: ({ cell }: { cell: string[] }) => {
        return cell?.join(';') || '--';
      },
    },
    {
      label: '上传时间',
      field: 'cloud_created_time',
      sort: true,
      render: ({ cell }: { cell: string }) => {
        // 由于云上返回的是(UTC+8)时间, 所以先转零时区
        const utcTime = dayjs(cell).subtract(8, 'hour');
        return timeFormatter(utcTime);
      },
    },
    {
      label: '过期时间',
      field: 'cloud_expired_time',
      sort: true,
      render: ({ cell }: { cell: string }) => {
        // 由于云上返回的是(UTC+8)时间, 所以先转零时区
        const utcTime = dayjs(cell).subtract(8, 'hour');
        return timeFormatter(utcTime);
      },
    },
    {
      label: '证书状态',
      field: 'cert_status',
      filter: {
        list: [
          {
            text: '正常',
            value: '正常',
          },
          {
            text: '已过期',
            value: '已过期',
          },
        ],
      },
      render: ({ cell }: { cell: string }) => {
        let icon;
        switch (cell) {
          case '正常':
            icon = StatusNormal;
            break;
          case '已过期':
            icon = StatusAbnormal;
            break;
        }
        return (
          <div class='status-column-cell'>
            <img class='status-icon' src={icon} alt='' />
            <span>{cell}</span>
          </div>
        );
      },
    },
    {
      label: '分配状态',
      field: 'bk_biz_id',
      isOnlyShowInResource: true,
      isDefaultShow: true,
      render: ({ data, cell }: { data: { bk_biz_id: number }; cell: number }) => (
        <bk-tag
          v-bk-tooltips={{
            content: businessMapStore.businessMap.get(cell),
            disabled: !cell || cell === -1,
          }}
          theme={data.bk_biz_id === -1 ? false : 'success'}>
          {data.bk_biz_id === -1 ? '未分配' : '已分配'}
        </bk-tag>
      ),
    },
  ];

  const firstAccountColumns = [
    {
      label: '一级帐号ID',
      field: 'cloud_id',
    },
    {
      label: '云厂商',
      field: 'vendor',
      render: ({ cell }: any) => BILL_VENDORS_MAP[cell] || '--',
    },
    {
      label: '帐号邮箱',
      field: 'email',
    },
    {
      label: '主负责人',
      field: 'managers',
      render: ({ cell }: any) => cell.join(','),
    },
    // {
    //   label: '组织架构',
    //   field: 'dept_id',
    // },
    {
      label: '备注',
      field: 'memo',
    },
  ];

  const secondaryAccountColumns = [
    {
      label: '二级账号ID',
      field: 'cloud_id',
    },
    {
      label: '所属一级帐号',
      field: 'parent_account_name',
    },
    {
      label: '云厂商',
      field: 'vendor',
      render: ({ cell }: any) => BILL_VENDORS_MAP[cell] || '--',
    },
    {
      label: '站点类型',
      field: 'site',
      render: ({ cell }: any) => BILL_SITE_TYPES_MAP[cell],
    },
    {
      label: '帐号邮箱',
      field: 'email',
    },
    {
      label: '主负责人',
      field: 'managers',
      render: ({ cell }: any) => cell.join(','),
    },
    {
      label: '业务',
      field: 'op_product_id',
    },
    {
      label: '备注',
      field: 'memo',
    },
  ];

  const myApplyColumns = [
    // {
    //   label: '申请ID',
    //   field: 'id',
    // },
    // {
    //   label: '来源',
    //   field: 'source',
    // },
    {
      label: '申请类型',
      field: 'type',
      render: ({ cell }: { cell: string }) => APPLICATION_TYPE_MAP[cell],
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
          <div class={'cvm-status-container'}>
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

  const billsRootAccountSummaryColumns = [
    {
      label: '一级账号ID',
      field: 'root_account_id',
      isDefaultShow: true,
    },
    {
      label: '一级账号名称',
      field: 'root_account_name',
      isDefaultShow: true,
    },
    {
      label: '账号状态',
      field: 'state',
      isDefaultShow: true,
      render: ({ cell }: any) => BILLS_ROOT_ACCOUNT_SUMMARY_STATE_MAP[cell],
    },
    {
      label: '账单同步（人民币-元）当月',
      field: 'current_month_rmb_cost_synced',
      isDefaultShow: true,
      render: ({ cell }: any) => formatBillCost(cell),
      sort: true,
    },
    {
      label: '账单同步（人民币-元）上月',
      field: 'last_month_rmb_cost_synced',
      isDefaultShow: true,
      render: ({ cell }: any) => formatBillCost(cell),
      sort: true,
    },
    {
      label: '账单同步（美金-美元）当月',
      field: 'current_month_cost_synced',
      isDefaultShow: true,
      render: ({ cell }: any) => formatBillCost(cell),
      sort: true,
    },
    {
      label: '账单同步（美金-美元）上月',
      field: 'last_month_cost_synced',
      isDefaultShow: true,
      render: ({ cell }: any) => formatBillCost(cell),
      sort: true,
    },
    {
      label: '账单同步环比',
      field: 'month_on_month_value',
      isDefaultShow: true,
    },
    {
      label: '当前账单人民币（元）',
      field: 'current_month_rmb_cost',
      isDefaultShow: true,
      render: ({ cell }: any) => formatBillCost(cell),
      sort: true,
    },
    {
      label: '当前账单美金（美元）',
      field: 'current_month_cost',
      isDefaultShow: true,
      render: ({ cell }: any) => formatBillCost(cell),
      sort: true,
    },
    {
      label: '调账人民币（元）',
      field: 'adjustment_rmb_cost',
      isDefaultShow: true,
      render: ({ cell }: any) => formatBillCost(cell),
      sort: true,
    },
    {
      label: '调账美金（美元）',
      field: 'adjustment_cost',
      isDefaultShow: true,
      render: ({ cell }: any) => formatBillCost(cell),
      sort: true,
    },
  ];

  const billsMainAccountSummaryColumns = [
    {
      label: '二级账号ID',
      field: 'main_account_cloud_id',
      isDefaultShow: true,
    },
    {
      label: '二级账号名称',
      field: 'main_account_name',
      isDefaultShow: true,
    },
    {
      label: '一级账号名称',
      field: 'root_account_name',
      isDefaultShow: true,
    },
    {
      label: '业务名称',
      field: 'bk_biz_id',
      isDefaultShow: true,
      render: ({ data }: any) => businessMapStore.businessMap.get(data.bk_biz_id) || '未分配',
    },
    {
      label: '已确认账单人民币（元）',
      field: 'current_month_rmb_cost_synced',
      isDefaultShow: true,
      render: ({ cell }: any) => formatBillCost(cell),
      sort: true,
    },
    {
      label: '已确认账单美金（美元）',
      field: 'current_month_cost_synced',
      isDefaultShow: true,
      render: ({ cell }: any) => formatBillCost(cell),
      sort: true,
    },
    {
      label: '当前账单人民币（元）',
      field: 'current_month_rmb_cost',
      isDefaultShow: true,
      render: ({ cell }: any) => formatBillCost(cell),
      sort: true,
    },
    {
      label: '当前账单美金（美元）',
      field: 'current_month_cost',
      isDefaultShow: true,
      render: ({ cell }: any) => formatBillCost(cell),
      sort: true,
    },
    {
      label: '调账人民币（元）',
      field: 'adjustment_rmb_cost',
      isDefaultShow: true,
      render: ({ cell }: any) => formatBillCost(cell),
      sort: true,
    },
    {
      label: '调账美金（美元）',
      field: 'adjustment_cost',
      isDefaultShow: true,
      render: ({ cell }: any) => formatBillCost(cell),
      sort: true,
    },
  ];

  const billDetailAwsColumns = [
    {
      label: '核算日期',
      field: 'bill_date',
      render: ({ data: { bill_year, bill_month, bill_day } }: any) =>
        dayjs(new Date(bill_year, bill_month - 1, bill_day)).format('YYYYMMDD'),
    },
    {
      label: 'ID',
      field: 'id',
      isDefaultShow: true,
    },
    {
      label: '一级账号ID',
      field: 'root_account_id',
      isDefaultShow: true,
    },
    {
      label: '二级账号ID',
      field: 'main_account_id',
      isDefaultShow: true,
    },
    {
      label: '云厂商',
      field: 'vendor',
      isDefaultShow: true,
      render: ({ cell }: { cell: VendorEnum }) => VendorMap[cell],
    },
    {
      label: '业务名称',
      field: 'bk_biz_id',
      isDefaultShow: true,
      render: ({ cell }: { cell: number }) => businessMapStore.businessMap.get(cell) || '未分配',
    },
    {
      label: '币种',
      field: 'currency',
      isDefaultShow: true,
      render: ({ cell }: any) => CURRENCY_MAP[cell],
    },
    {
      label: '本期应付金额',
      field: 'cost',
      isDefaultShow: true,
      render: ({ cell }: any) => formatBillCost(cell),
      sort: true,
    },
    {
      label: '资源类型编码',
      field: 'hc_product_code',
      isDefaultShow: true,
    },
    {
      label: '产品名称',
      field: 'hc_product_name',
      isDefaultShow: true,
    },
    {
      label: '预留实例使用量',
      field: 'res_amount',
      isDefaultShow: true,
    },
    {
      label: '预留实例使用单位',
      field: 'res_amount_unit',
      isDefaultShow: true,
    },
  ];

  const billDetailAzureColumns = [
    {
      label: '核算日期',
      field: 'bill_date',
      render: ({ data: { bill_year, bill_month, bill_day } }: any) =>
        dayjs(new Date(bill_year, bill_month - 1, bill_day)).format('YYYYMMDD'),
    },
    {
      label: 'ID',
      field: 'id',
      isDefaultShow: true,
    },
    {
      label: '一级账号ID',
      field: 'root_account_id',
      isDefaultShow: true,
    },
    {
      label: '二级账号ID',
      field: 'main_account_id',
      isDefaultShow: true,
    },
    {
      label: '云厂商',
      field: 'vendor',
      isDefaultShow: true,
      render: ({ cell }: { cell: VendorEnum }) => VendorMap[cell],
    },
    {
      label: '业务名称',
      field: 'bk_biz_id',
      isDefaultShow: true,
      render: ({ cell }: { cell: number }) => businessMapStore.businessMap.get(cell) || '未分配',
    },
    {
      label: '币种',
      field: 'currency',
      isDefaultShow: true,
      render: ({ cell }: any) => CURRENCY_MAP[cell],
    },
    {
      label: '本期应付金额',
      field: 'cost',
      isDefaultShow: true,
      render: ({ cell }: any) => formatBillCost(cell),
      sort: true,
    },
    {
      label: '资源类型编码',
      field: 'hc_product_code',
      isDefaultShow: true,
    },
    {
      label: '产品名称',
      field: 'hc_product_name',
      isDefaultShow: true,
    },
    {
      label: '预留实例使用量',
      field: 'res_amount',
      isDefaultShow: true,
    },
    {
      label: '预留实例使用单位',
      field: 'res_amount_unit',
      isDefaultShow: true,
    },
  ];

  const billDetailGcpColumns = [
    {
      label: '核算日期',
      field: 'bill_date',
      render: ({ data: { bill_year, bill_month, bill_day } }: any) =>
        dayjs(new Date(bill_year, bill_month - 1, bill_day)).format('YYYYMMDD'),
    },
    {
      label: 'ID',
      field: 'id',
      isDefaultShow: true,
    },
    {
      label: '一级账号ID',
      field: 'root_account_id',
      isDefaultShow: true,
    },
    {
      label: '二级账号ID',
      field: 'main_account_id',
      isDefaultShow: true,
    },
    {
      label: '云厂商',
      field: 'vendor',
      isDefaultShow: true,
      render: ({ cell }: { cell: VendorEnum }) => VendorMap[cell],
    },
    {
      label: '业务名称',
      field: 'bk_biz_id',
      isDefaultShow: true,
      render: ({ cell }: { cell: number }) => businessMapStore.businessMap.get(cell) || '未分配',
    },
    {
      label: '币种',
      field: 'currency',
      isDefaultShow: true,
      render: ({ cell }: any) => CURRENCY_MAP[cell],
    },
    {
      label: '本期应付金额',
      field: 'cost',
      isDefaultShow: true,
      render: ({ cell }: any) => formatBillCost(cell),
      sort: true,
    },
    {
      label: '资源类型编码',
      field: 'hc_product_code',
      isDefaultShow: true,
    },
    {
      label: '产品名称',
      field: 'hc_product_name',
      isDefaultShow: true,
    },
    {
      label: '预留实例使用量',
      field: 'res_amount',
      isDefaultShow: true,
    },
    {
      label: '预留实例使用单位',
      field: 'res_amount_unit',
      isDefaultShow: true,
    },
  ];

  const billDetailHuaweiColumns = [
    {
      label: '核算日期',
      field: 'bill_date',
      render: ({ data: { bill_year, bill_month, bill_day } }: any) =>
        dayjs(new Date(bill_year, bill_month - 1, bill_day)).format('YYYYMMDD'),
    },
    {
      label: 'ID',
      field: 'id',
      isDefaultShow: true,
    },
    {
      label: '一级账号ID',
      field: 'root_account_id',
      isDefaultShow: true,
    },
    {
      label: '二级账号ID',
      field: 'main_account_id',
      isDefaultShow: true,
    },
    {
      label: '云厂商',
      field: 'vendor',
      isDefaultShow: true,
      render: ({ cell }: { cell: VendorEnum }) => VendorMap[cell],
    },
    {
      label: '业务名称',
      field: 'bk_biz_id',
      isDefaultShow: true,
      render: ({ cell }: { cell: number }) => businessMapStore.businessMap.get(cell) || '未分配',
    },
    {
      label: '币种',
      field: 'currency',
      isDefaultShow: true,
      render: ({ cell }: any) => CURRENCY_MAP[cell],
    },
    {
      label: '本期应付金额',
      field: 'cost',
      isDefaultShow: true,
      render: ({ cell }: any) => formatBillCost(cell),
      sort: true,
    },
    {
      label: '资源类型编码',
      field: 'hc_product_code',
      isDefaultShow: true,
    },
    {
      label: '产品名称',
      field: 'hc_product_name',
      isDefaultShow: true,
    },
    {
      label: '预留实例使用量',
      field: 'res_amount',
      isDefaultShow: true,
    },
    {
      label: '预留实例使用单位',
      field: 'res_amount_unit',
      isDefaultShow: true,
    },
    {
      label: '使用量类型',
      field: 'extension.usage_type',
    },
    {
      label: '使用量',
      field: 'extension.usage',
    },
    {
      label: '使用量度量单位',
      field: 'extension.unit',
    },
    {
      label: '云服务类型编码',
      field: 'extension.cloud_service_type',
    },
    {
      label: '云服务类型名称',
      field: 'extension.cloud_service_type_name',
    },
    {
      label: '云服务区编码',
      field: 'extension.region',
    },
    {
      label: '云服务区名称',
      field: 'extension.region_name',
    },
    {
      label: '资源类型编码',
      field: 'extension.resource_type',
    },
    {
      label: '资源类型名称',
      field: 'extension.resource_type_name',
    },
    {
      label: '计费模式',
      field: 'extension.charge_mode',
    },
    {
      label: '账单类型',
      field: 'extension.bill_type',
      render: ({ cell }: any) => BILL_TYPE__MAP_HW[cell],
    },
  ];

  const billDetailZenlayerColumns = [
    {
      label: '核算日期',
      field: 'bill_date',
      render: ({ data: { bill_year, bill_month, bill_day } }: any) =>
        dayjs(new Date(bill_year, bill_month - 1, bill_day)).format('YYYYMMDD'),
    },
    {
      label: 'ID',
      field: 'id',
      isDefaultShow: true,
    },
    {
      label: '一级账号ID',
      field: 'root_account_id',
      isDefaultShow: true,
    },
    {
      label: '二级账号ID',
      field: 'main_account_id',
      isDefaultShow: true,
    },
    {
      label: '云厂商',
      field: 'vendor',
      isDefaultShow: true,
      render: ({ cell }: { cell: VendorEnum }) => VendorMap[cell],
    },
    {
      label: '业务名称',
      field: 'bk_biz_id',
      isDefaultShow: true,
      render: ({ cell }: { cell: number }) => businessMapStore.businessMap.get(cell) || '未分配',
    },
    {
      label: '币种',
      field: 'currency',
      isDefaultShow: true,
      render: ({ cell }: any) => CURRENCY_MAP[cell],
    },
    {
      label: '本期应付金额',
      field: 'cost',
      isDefaultShow: true,
      render: ({ cell }: any) => formatBillCost(cell),
      sort: true,
    },
    {
      label: '资源类型编码',
      field: 'hc_product_code',
      isDefaultShow: true,
    },
    {
      label: '产品名称',
      field: 'hc_product_name',
      isDefaultShow: true,
    },
    {
      label: '预留实例使用量',
      field: 'res_amount',
      isDefaultShow: true,
    },
    {
      label: '预留实例使用单位',
      field: 'res_amount_unit',
      isDefaultShow: true,
    },
  ];

  const billsSummaryOperationRecordColumns = [
    {
      label: '操作时间',
      field: 'created_at',
      render: ({ cell }: any) => timeFormatter(cell),
    },
    {
      label: '完成时间',
      field: 'updated_at',
      render: ({ cell }: any) => timeFormatter(cell),
    },
    {
      label: '状态',
      field: 'state',
    },
    {
      label: '账单月份',
      field: 'bill_month',
      render: ({ data }: any) => dayjs(new Date(data.bill_year, data.bill_month - 1)).format('YYYY-MM'),
    },
    {
      label: '云厂商',
      field: 'vendor',
      isDefaultShow: true,
      render: ({ cell }: { cell: VendorEnum }) => VendorMap[cell],
    },
    {
      label: '操作人',
      field: 'operator',
    },
    {
      label: '人民币（元）',
      field: 'rmb_cost',
      render: ({ cell }: any) => formatBillCost(cell),
      sort: true,
    },
    {
      label: '美金（美元）',
      field: 'cost',
      render: ({ cell }: any) => formatBillCost(cell),
      sort: true,
    },
  ];

  const columnsMap = {
    vpc: vpcColumns,
    subnet: subnetColumns,
    group: groupColumns,
    gcp: gcpColumns,
    drive: driveColumns,
    image: imageColumns,
    networkInterface: networkInterfaceColumns,
    route: routeColumns,
    cvms: cvmsColumns,
    securityCommon: securityCommonColumns,
    eips: eipColumns,
    operationRecord: operationRecordColumns,
    lb: lbColumns,
    listener: listenerColumns,
    targetGroup: targetGroupColumns,
    rsConfig: rsConfigColumns,
    domain: domainColumns,
    url: urlColumns,
    targetGroupListener: targetGroupListenerColumns,
    cert: certColumns,
    firstAccount: firstAccountColumns,
    secondaryAccount: secondaryAccountColumns,
    myApply: myApplyColumns,
    billsRootAccountSummary: billsRootAccountSummaryColumns,
    billsMainAccountSummary: billsMainAccountSummaryColumns,
    billDetailAws: billDetailAwsColumns,
    billDetailAzure: billDetailAzureColumns,
    billDetailGcp: billDetailGcpColumns,
    billDetailHuawei: billDetailHuaweiColumns,
    billDetailZenlayer: billDetailZenlayerColumns,
    billsSummaryOperationRecord: billsSummaryOperationRecordColumns,
  };

  let columns = (columnsMap[type] || []).filter((column: any) => !isSimpleShow || !column.onlyShowOnList);
  if (whereAmI.value !== Senarios.resource) columns = columns.filter((column: any) => !column.isOnlyShowInResource);

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
          isOnlyShowInResource: !!column.isOnlyShowInResource,
        });
      }
    }
    if (whereAmI.value !== Senarios.resource) {
      fields = fields.filter((field) => !field.isOnlyShowInResource);
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
