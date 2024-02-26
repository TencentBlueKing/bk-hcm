/* eslint-disable no-nested-ternary */
// table 字段相关信息
import i18n from '@/language/i18n';
import {
  CloudType,
  SecurityRuleEnum,
  HuaweiSecurityRuleEnum,
  AzureSecurityRuleEnum,
} from '@/typings';
import { useAccountStore } from '@/store';
import { Button } from 'bkui-vue';
import type { Settings } from 'bkui-vue/lib/table/props';
import { h, ref } from 'vue';
import type { Ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { CLOUD_HOST_STATUS, VendorEnum } from '@/common/constant';
import { useRegionsStore } from '@/store/useRegionsStore';
import { Senarios, useWhereAmI } from '@/hooks/useWhereAmI';
import { useBusinessMapStore } from '@/store/useBusinessMap';
import { useCloudAreaStore } from '@/store/useCloudAreaStore';
import StatusAbnormal from '@/assets/image/Status-abnormal.png';
import StatusNormal from '@/assets/image/Status-normal.png';
import StatusUnknown from '@/assets/image/Status-unknown.png';
import StatusSuccess from '@/assets/image/success-account.png';
import StatusFailure from '@/assets/image/failed-account.png';
import StatusPartialSuccess from '@/assets/image/result-waiting.png';

import {
  HOST_RUNNING_STATUS,
  HOST_SHUTDOWN_STATUS,
} from '../common/table/HostOperations';
import './use-columns.scss';
import { timeFormatter } from '@/common/util';

export default (type: string, isSimpleShow = false, vendor?: string) => {
  const router = useRouter();
  const route = useRoute();
  const accountStore = useAccountStore();
  const { t } = i18n.global;
  const { getRegionName } = useRegionsStore();
  const { whereAmI } = useWhereAmI();
  const businessMapStore = useBusinessMapStore();
  const cloudAreaStore = useCloudAreaStore();

  const getLinkField = (
    type: string,
    label = 'ID',
    field = 'id',
    idFiled = 'id',
    onlyShowOnList = true,
    render: (data: any) => Element | string = undefined,
    sort = true,
  ) => {
    return {
      label,
      field,
      sort,
      width: label === 'ID' ? '120' : 'auto',
      onlyShowOnList,
      isDefaultShow: true,
      render({ data }: { cell: string; data: any }) {
        if (data[idFiled] < 0 || !data[idFiled]) return '--';
        return (
          <Button
            text
            theme='primary'
            onClick={() => {
              const routeInfo: any = {
                query: {
                  ...route.query,
                  id: data[idFiled],
                  type: data.vendor,
                },
              };
              // 业务下
              if (route.path.includes('business')) {
                routeInfo.query.bizs = accountStore.bizs;
                Object.assign(routeInfo, {
                  name: `${type}BusinessDetail`,
                });
              } else {
                Object.assign(routeInfo, {
                  name: 'resourceDetail',
                  params: {
                    type,
                  },
                });
              }
              router.push(routeInfo);
            }}>
            {render ? render(data) : data[field] || '--'}
          </Button>
        );
      },
    };
  };

  const vpcColumns = [
    {
      type: 'selection',
      width: 32,
      minWidth: 32,
      onlyShowOnList: true,
      align: 'right',
    },
    getLinkField('vpc', 'VPC ID', 'cloud_id'),
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
      render: ({
        data,
        cell,
      }: {
        data: { bk_biz_id: number };
        cell: number;
      }) => (
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
      render: ({ cell }: { cell: string }) =>  timeFormatter(cell),
    },
    {
      label: '更新时间',
      field: 'updated_at',
      sort: true,
      render: ({ cell }: { cell: string }) =>  timeFormatter(cell),
    },
  ];

  const subnetColumns = [
    {
      type: 'selection',
      width: 32,
      minWidth: 32,
      onlyShowOnList: true,
      align: 'right',
    },
    getLinkField('subnet', '子网 ID', 'cloud_id', 'id', false),
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
    getLinkField('vpc', '所属 VPC', 'cloud_vpc_id', 'vpc_id', false),
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
      render: ({
        data,
        cell,
      }: {
        data: { bk_biz_id: number };
        cell: number;
      }) => (
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
      render: ({ cell }: { cell: string }) =>  timeFormatter(cell),
    },
    {
      label: '更新时间',
      field: 'updated_at',
      sort: true,
      render: ({ cell }: { cell: string }) =>  timeFormatter(cell),
    },
  ];

  const groupColumns = [
    {
      type: 'selection',
      width: 32,
      minWidth: 32,
      onlyShowOnList: true,
      align: 'right',
    },
    getLinkField('subnet'),
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
    {
      type: 'selection',
      width: 32,
      minWidth: 32,
      onlyShowOnList: true,
      align: 'right',
    },
    getLinkField('subnet'),
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
    {
      type: 'selection',
      width: 32,
      minWidth: 32,
      onlyShowOnList: true,
      align: 'right',
    },
    getLinkField('drive', '云硬盘ID', 'cloud_id'),
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
    getLinkField('host', '挂载的主机', 'instance_id', 'instance_id'),
    {
      label: '是否分配',
      field: 'bk_biz_id',
      sort: true,
      isOnlyShowInResource: true,
      isDefaultShow: true,
      render: ({
        data,
        cell,
      }: {
        data: { bk_biz_id: number };
        cell: number;
      }) => (
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
      render: ({ cell }: { cell: string }) =>  timeFormatter(cell),
    },
    {
      label: '更新时间',
      field: 'updated_at',
      sort: true,
      render: ({ cell }: { cell: string }) =>  timeFormatter(cell),
    },
  ];

  const imageColumns = [
    getLinkField('image', '镜像ID', 'cloud_id', 'id'),
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
      render: ({ cell }: { cell: string }) =>  timeFormatter(cell),
    },
    {
      label: '更新时间',
      field: 'updated_at',
      sort: true,
      render: ({ cell }: { cell: string }) =>  timeFormatter(cell),
    },
  ];

  const networkInterfaceColumns = [
    getLinkField('network-interface', '接口 ID', 'cloud_id', 'id'),
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
        return [
          h('span', {}, [
            data?.private_ipv4.join(',')
              || data?.private_ipv6.join(',')
              || '--',
          ]),
        ];
      },
    },
    {
      label: '关联的公网IP地址',
      field: 'public_ip',
      // 目前公网IP地址不支持排序
      // sort: true,
      isDefaultShow: true,
      render({ data }: any) {
        return [
          h('span', {}, [
            data?.public_ipv4.join(',') || data?.public_ipv6.join(',') || '--',
          ]),
        ];
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
      render: ({ cell }: { cell: string }) =>  timeFormatter(cell),
    },
    {
      label: '更新时间',
      field: 'updated_at',
      sort: true,
      render: ({ cell }: { cell: string }) =>  timeFormatter(cell),
    },
  ];

  const routeColumns = [
    getLinkField('route', '路由表ID', 'cloud_id', 'id'),
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
    getLinkField('vpc', '所属网络(VPC)', 'vpc_id', 'vpc_id'),
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
      render: ({ cell }: { cell: string }) =>  timeFormatter(cell),
    },
    {
      label: '更新时间',
      field: 'updated_at',
      sort: true,
      render: ({ cell }: { cell: string }) =>  timeFormatter(cell),
    },
  ];

  const cvmsColumns = [
    {
      type: 'selection',
      width: 32,
      minWidth: 32,
      onlyShowOnList: true,
      align: 'right',
    },
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
    getLinkField(
      'host',
      '内网IP',
      'private_ipv4_addresses',
      'id',
      false,
      data => [...data.private_ipv4_addresses, ...data.private_ipv6_addresses].join(','),
      false,
    ),
    {
      label: '公网IP',
      field: 'vendor',
      isDefaultShow: true,
      onlyShowOnList: true,
      render: ({ data }: any) => [...data.public_ipv4_addresses, ...data.public_ipv6_addresses].join(',') || '--',
    },
    {
      label: '所属VPC',
      field: 'cloud_vpc_ids',
      isDefaultShow: true,
      onlyShowOnList: true,
      render: ({ data }: any) => data.cloud_vpc_ids.join(',') || '--',
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
      render: ({
        data,
        cell,
      }: {
        data: { bk_biz_id: number };
        cell: number;
      }) => (
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
      render: ({ cell }: { cell: string }) =>  timeFormatter(cell),
    },
    {
      label: '更新时间',
      field: 'updated_at',
      sort: true,
      render: ({ cell }: { cell: string }) =>  timeFormatter(cell),
    },
  ];

  const securityCommonColumns = [
    {
      label: t('来源'),
      field: 'resource',
      render({ data }: any) {
        return h('span', {}, [
          data.cloud_address_group_id
            || data.cloud_address_id
            || data.cloud_service_group_id
            || data.cloud_service_id
            || data.cloud_target_security_group_id
            || data.ipv4_cidr
            || data.ipv6_cidr
            || data.cloud_remote_group_id
            || data.remote_ip_prefix
            || (data.source_address_prefix === '*'
              ? t('任何')
              : data.source_address_prefix)
            || data.source_address_prefixes
            || data.cloud_source_security_group_ids
            || (data.destination_address_prefix === '*'
              ? t('任何')
              : data.destination_address_prefix)
            || data.destination_address_prefixes
            || data.cloud_destination_security_group_ids
            || '--',
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
              : vendor === 'azure'
              && data.protocol === '*'
              && data.destination_port_range === '*'
                ? t('全部')
                : `${data.protocol}:${
                  data.port || data.to_port || data.destination_port_range || '--'
                }`,
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
      render: ({ cell }: { cell: string }) =>  timeFormatter(cell),
    },
  ];

  const eipColumns = [
    {
      type: 'selection',
      width: 32,
      minWidth: 32,
      onlyShowOnList: true,
      align: 'right',
    },
    getLinkField('eips', 'IP资源ID', 'cloud_id', 'id'),
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
    getLinkField('host', '绑定的资源实例', 'cvm_id', 'cvm_id', false, data => data.host, false),
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
      render: ({
        data,
        cell,
      }: {
        data: { bk_biz_id: number };
        cell: number;
      }) => (
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
      render: ({ cell }: { cell: string }) =>  timeFormatter(cell),
    },
    {
      label: '更新时间',
      field: 'updated_at',
      sort: true,
      render: ({ cell }: { cell: string }) =>  timeFormatter(cell),
    },
  ];

  const operationRecordColumns = [
    {
      label: '操作时间',
      field: 'created_at',
      isDefaultShow: true,
      sort: true,
      render: ({ cell }: { cell: string }) =>  timeFormatter(cell),
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
      label: '任务状态',
      field: 'task_status',
      isDefaultShow: true,
      filter: true,
      render: ({ cell }: { cell: string }) => {
        if (!cell) return '--';
        let icon;
        switch (cell) {
          case 'success':
            icon = StatusSuccess;
            break;
          case 'fail':
            icon = StatusFailure;
            break;
          case 'partial_success':
            icon = StatusPartialSuccess;
            break;
        }
        return (
          <div class='status-column-cell'>
            <img class='status-icon' src={icon} alt='' />
            <span>{cell === 'success' ? '成功' : cell === 'fail' ? '失败' : '部分成功'}</span>
          </div>
        );
      },
    },
    {
      label: '操作人',
      field: 'operator',
      isDefaultShow: true,
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
      fields = fields.filter(field => !field.isOnlyShowInResource);
    }
    const settings: Ref<Settings> = ref({
      fields,
      checked: fields
        .filter(field => field.isDefaultShow)
        .map(field => field.field),
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
