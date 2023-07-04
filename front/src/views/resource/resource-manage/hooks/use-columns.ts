// table 字段相关信息
import i18n from '@/language/i18n';
import { CloudType, SecurityRuleEnum, HuaweiSecurityRuleEnum, AzureSecurityRuleEnum } from '@/typings';
import { useAccountStore } from '@/store';
import {
  Button,
} from 'bkui-vue';
import {
  h,
} from 'vue';
import {
  useRoute,
  useRouter,
} from 'vue-router';
import { CLOUD_HOST_STATUS, VendorEnum } from '@/common/constant';
import { useRegionsStore } from '@/store/useRegionsStore';

export default (type: string, isSimpleShow = false, vendor?: string) => {
  const router = useRouter();
  const route = useRoute();
  const accountStore = useAccountStore();
  const { t } = i18n.global;
  const { getRegionName } = useRegionsStore();

  const getLinkField = (type: string, label = 'ID', field = 'id', idFiled = 'id', onlyShowOnList = true) => {
    return {
      label,
      field,
      sort: true,
      width: label === 'ID' ? '120' : 'auto',
      onlyShowOnList,
      render({ data }: { cell: string, data: any }) {
        if (data[idFiled] < 0 || !data[idFiled]) {
          return '--';
        }
        return h(
          Button,
          {
            text: true,
            theme: 'primary',
            onClick() {
              const routeInfo: any = {
                query: {
                  id: data[idFiled],
                  type: data.vendor,
                },
              };
              // 业务下
              if (route.path.includes('business')) {
                routeInfo.query.bizs = accountStore.bizs;
                Object.assign(
                  routeInfo,
                  {
                    name: `${type}BusinessDetail`,
                  },
                );
              } else {
                Object.assign(
                  routeInfo,
                  {
                    name: 'resourceDetail',
                    params: {
                      type,
                    },
                  },
                );
              }
              router.push(routeInfo);
            },
          },
          [
            data[field] || '--',
          ],
        );
      },
    };
  };

  const vpcColumns = [
    getLinkField('vpc'),
    {
      label: '资源 ID',
      field: 'cloud_id',
      sort: true,
      render({ cell }: { cell: string }) {
        return h(
          'span',
          [
            cell || '--',
          ],
        );
      },
    },
    {
      label: '名称',
      field: 'name',
      sort: true,
      render({ cell }: { cell: string }) {
        return h(
          'span',
          [
            cell || '--',
          ],
        );
      },
    },
    {
      label: '云厂商',
      field: 'vendor',
      render({ cell }: { cell: string }) {
        return h(
          'span',
          [
            CloudType[cell] || '--',
          ],
        );
      },
    },
    {
      label: '地域',
      field: 'region',
      render: ({ cell, row }: { cell: string, row: { vendor: VendorEnum } }) => getRegionName(row.vendor, cell),
    },
    {
      label: '管控区域 ID',
      field: 'bk_cloud_id',
      render({ cell }: { cell: number }) {
        if (cell > -1) {
          return cell;
        }
        return '--';
      },
    },
    {
      label: '更新时间',
      field: 'updated_at',
      sort: true,
    },
    {
      label: '创建时间',
      field: 'created_at',
      sort: true,
    },
  ];

  const subnetColumns = [
    getLinkField('subnet', 'ID', 'id', 'id', false),
    {
      label: '资源 ID',
      field: 'cloud_id',
      sort: true,
      render({ cell }: { cell: string }) {
        const index = cell.lastIndexOf('/') <= 0 ? 0 : cell.lastIndexOf('/') + 1;
        const value = cell.slice(index);
        return h(
          'span',
          [
            value || '--',
          ],
        );
      },
    },
    {
      label: '名称',
      field: 'name',
      sort: true,
      render({ cell }: { cell: string }) {
        return h(
          'span',
          [
            cell || '--',
          ],
        );
      },
    },
    {
      label: '云厂商',
      field: 'vendor',
      render({ cell }: { cell: string }) {
        return h(
          'span',
          [
            CloudType[cell] || '--',
          ],
        );
      },
    },
    {
      label: '地域',
      field: 'region',
      render: ({ cell, row }: { cell: string, row: { vendor: VendorEnum } }) => getRegionName(row.vendor, cell),
    },
    {
      label: '可用区',
      field: 'zone',
      render({ cell }: { cell: string }) {
        return h(
          'span',
          [
            cell || '--',
          ],
        );
      },
    },
    getLinkField('vpc', '所属 VPC', 'vpc_id', 'vpc_id', false),
    {
      label: 'IPv4 CIDR',
      field: 'ipv4_cidr',
      render({ cell }: { cell: string }) {
        return h(
          'span',
          [
            cell || '--',
          ],
        );
      },
    },
    getLinkField('route', '关联路由表', 'route_table_id', 'route_table_id', false),
    {
      label: '更新时间',
      field: 'updated_at',
      sort: true,
    },
    {
      label: '创建时间',
      field: 'created_at',
      sort: true,
    },
  ];

  const groupColumns = [
    {
      type: 'selection',
      width: '100',
      onlyShowOnList: true,
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
        return h(
          'span',
          {},
          [
            CloudType[data.vendor],
          ],
        );
      },
    },
    {
      label: '地域',
      field: 'region',
      render: ({ cell, row }: { cell: string, row: { vendor: VendorEnum } }) => getRegionName(row.vendor, cell),
    },
    {
      label: '描述',
      field: 'memo',
    },
  ];

  const gcpColumns = [
    {
      type: 'selection',
      width: '100',
      onlyShowOnList: true,
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
      width: '100',
      onlyShowOnList: true,
    },
    getLinkField('drive'),
    {
      label: '资源 ID',
      field: 'cloud_id',
      sort: true,
    },
    {
      label: '名称',
      field: 'name',
      sort: true,
    },
    {
      label: '云厂商',
      field: 'vendor',
      render({ cell }: { cell: string }) {
        return h(
          'span',
          [
            CloudType[cell] || '--',
          ],
        );
      },
    },
    {
      label: '类型',
      field: 'disk_type',
      sort: true,
      render({ cell }: { cell: string }) {
        return h(
          'span',
          [
            cell || '--',
          ],
        );
      },
    },
    {
      label: '容量(GB)',
      field: 'disk_size',
      sort: true,
      render({ cell }: { cell: string }) {
        return h(
          'span',
          [
            cell || '--',
          ],
        );
      },
    },
    {
      label: '状态',
      field: 'status',
      render({ cell }: { cell: string }) {
        return h(
          'span',
          [
            cell || '--',
          ],
        );
      },
    },
    {
      label: '可用区',
      field: 'zone',
      sort: true,
      render({ cell }: { cell: string }) {
        return h(
          'span',
          [
            cell || '--',
          ],
        );
      },
    },
    getLinkField('host', '挂载实例', 'instance_id', 'instance_id'),
    {
      label: '创建时间',
      field: 'created_at',
    },
  ];

  const imageColumns = [
    getLinkField('image'),
    {
      label: '资源 ID',
      field: 'cloud_id',
      sort: true,
    },
    {
      label: '名称',
      field: 'name',
      sort: true,
    },
    {
      label: '云厂商',
      field: 'vendor',
      sort: true,
      render({ cell }: { cell: string }) {
        return h(
          'span',
          [
            CloudType[cell] || '--',
          ],
        );
      },
    },
    {
      label: '操作系统类型',
      field: 'platform',
      sort: true,
    },
    {
      label: '架构',
      field: 'architecture',
      sort: true,
    },
    {
      label: '状态',
      field: 'state',
    },
    {
      label: '类型',
      field: 'type',
      sort: true,
    },
    {
      label: '创建时间',
      field: 'created_at',
      sort: true,
    },
  ];

  const networkInterfaceColumns = [
    getLinkField('network-interface'),
    {
      label: '名称',
      field: 'name',
      sort: true,
    },
    {
      label: '云厂商',
      field: 'vendor',
      sort: true,
      render({ cell }: { cell: string }) {
        return h(
          'span',
          [
            CloudType[cell] || '--',
          ],
        );
      },
    },
    {
      label: '地域',
      field: 'region',
      render: ({ cell, row }: { cell: string, row: { vendor: VendorEnum } }) => getRegionName(row.vendor, cell),
    },
    {
      label: '可用区域',
      field: 'zone',
      render({ cell }: { cell: string }) {
        return h(
          'span',
          [
            cell || '--',
          ],
        );
      },
    },
    {
      label: '虚拟网络',
      field: 'cloud_vpc_id',
      showOverflowTooltip: true,
      render({ cell }: { cell: string }) {
        return h(
          'span',
          [
            cell || '--',
          ],
        );
      },
    },
    {
      label: '子网',
      showOverflowTooltip: true,
      field: 'cloud_subnet_id',
      render({ cell }: { cell: string }) {
        return h(
          'span',
          [
            cell || '--',
          ],
        );
      },
    },
    {
      label: '关联的实例',
      field: 'instance_id',
      showOverflowTooltip: true,
      render({ cell }: { cell: string }) {
        return h(
          'span',
          [
            cell || '--',
          ],
        );
      },
    },
    {
      label: '内网IP',
      render({ data }: any) {
        return [
          h(
            'span',
            {},
            [
              data?.private_ipv4.join(',') || data?.private_ipv6.join(',') || '--',
            ],
          ),
        ];
      },
    },
    {
      label: '关联的公网IP地址',
      field: 'public_ip',
      render({ data }: any) {
        return [
          h(
            'span',
            {},
            [
              data?.public_ipv4.join(',') || data?.public_ipv6.join(',') || '--',
            ],
          ),
        ];
      },
    },
    {
      label: '创建时间',
      field: 'created_at',
      width: 180,
      sort: true,
    },
  ];

  const routeColumns = [
    getLinkField('route'),
    {
      label: '资源 ID',
      field: 'cloud_id',
      sort: true,
    },
    {
      label: '云厂商',
      field: 'vendor',
      render({ cell }: { cell: string }) {
        return h(
          'span',
          [
            CloudType[cell] || '--',
          ],
        );
      },
    },
    {
      label: '地域',
      field: 'region',
      sort: true,
      render: ({ cell, row }: { cell: string, row: { vendor: VendorEnum } }) => getRegionName(row.vendor, cell),
    },
    {
      label: '名称',
      field: 'name',
      sort: true,
    },
    getLinkField('vpc', '所属网络(VPC)', 'vpc_id', 'vpc_id'),
    {
      label: '关联子网',
      field: '',
      sort: true,
    },
    {
      label: '更新时间',
      field: 'updated_at',
    },
    {
      label: '创建时间',
      field: 'created_at',
    },
  ];

  const cvmsColumns = [
    {
      type: 'selection',
      width: '100',
      onlyShowOnList: true,
    },
    getLinkField('host'),
    {
      label: '资源 ID',
      field: 'cloud_id',
    },
    {
      label: '云厂商',
      onlyShowOnList: true,
      render({ data }: any) {
        return h(
          'span',
          {},
          [
            CloudType[data.vendor],
          ],
        );
      },
    },
    {
      label: '地域',
      onlyShowOnList: true,
      field: 'region',
      render: ({ cell, row }: { cell: string, row: { vendor: VendorEnum } }) => getRegionName(row.vendor, cell),
    },
    {
      label: '名称',
      field: 'name',
    },
    {
      label: '状态',
      render({ data }: any) {
        return h(
          'span',
          {},
          [
            CLOUD_HOST_STATUS[data.status] || data.status,
          ],
        );
      },
    },
    {
      label: '操作系统',
      render({ data }: any) {
        return h(
          'span',
          {},
          [
            data.os_name || '--',
          ],
        );
      },
    },
    {
      label: '管控区域 ID',
      field: 'bk_cloud_id',
      render({ data }: any) {
        return h(
          'span',
          {},
          [
            data.bk_cloud_id === -1 ? '未绑定' : data.bk_cloud_id,
          ],
        );
      },
    },
    {
      label: '内网IP',
      field: '',
      render({ data }: any) {
        return h(
          'span',
          {},
          [
            data.private_ipv4_addresses || data.private_ipv6_addresses,
          ],
        );
      },
    },
    {
      label: '公网IP',
      field: '',
      render({ data }: any) {
        return h(
          'span',
          {},
          [
            data.public_ipv4_addresses || data.public_ipv6_addresses,
          ],
        );
      },
    },
    {
      label: '创建时间',
      field: 'created_at',
    },
  ];

  const securityCommonColumns = [
    {
      label: t('来源'),
      field: 'resource',
      render({ data }: any) {
        return h(
          'span',
          {},
          [
            data.cloud_address_group_id || data.cloud_address_id
            || data.cloud_service_group_id || data.cloud_service_id || data.cloud_target_security_group_id
            || data.ipv4_cidr || data.ipv6_cidr || data.cloud_remote_group_id || data.remote_ip_prefix
            || (data.source_address_prefix === '*' ? t('任何') : data.source_address_prefix) || data.source_address_prefixes || data.cloud_source_security_group_ids
            || (data.destination_address_prefix === '*' ? t('任何') : data.destination_address_prefix) || data.destination_address_prefixes
            || data.cloud_destination_security_group_ids || '--',
          ],
        );
      },
    },
    {
      label: '协议端口',
      render({ data }: any) {
        return h(
          'span',
          {},
          [
            // eslint-disable-next-line no-nested-ternary
            vendor === 'aws' && (data.protocol === '-1' && data.to_port === -1) ? t('全部')
            // eslint-disable-next-line no-nested-ternary
              : vendor === 'huawei' && (!data.protocol && !data.port) ? t('全部')
                : vendor === 'azure' && (data.protocol === '*' && data.destination_port_range === '*') ? t('全部') :  `${data.protocol}:${data.port || data.to_port || data.destination_port_range || '--'}`,
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
            vendor === 'huawei' ? HuaweiSecurityRuleEnum[data.action] : vendor === 'azure' ? AzureSecurityRuleEnum[data.access]
              : vendor === 'aws' ? t('允许') : (SecurityRuleEnum[data.action] || '--'),
          ],
        );
      },
    },
    {
      label: '备注',
      field: 'memo',
      render({ data }: any) {
        return h(
          'span',
          {},
          [
            data.memo || '--',
          ],
        );
      },
    },
    {
      label: t('修改时间'),
      field: 'updated_at',
    },
  ];

  const eipColumns = [
    {
      type: 'selection',
      width: '100',
      onlyShowOnList: true,
    },
    getLinkField('eips'),
    {
      label: '资源 ID',
      field: 'cloud_id',
      sort: true,
    },
    {
      label: '云厂商',
      field: 'vendor',
      render({ cell }: { cell: string }) {
        return h(
          'span',
          [
            CloudType[cell] || '--',
          ],
        );
      },
    },
    {
      label: '地域',
      field: 'region',
      render: ({ cell, row }: { cell: string, row: { vendor: VendorEnum } }) => getRegionName(row.vendor, cell),
    },
    {
      label: '名称',
      field: 'name',
      sort: true,
      render({ cell }: { cell: string }) {
        return h(
          'span',
          [
            cell || '--',
          ],
        );
      },
    },
    {
      label: '公网 IP',
      field: 'public_ip',
      render({ cell }: { cell: string }) {
        return h(
          'span',
          [
            cell || '--',
          ],
        );
      },
    },
    {
      label: '状态',
      field: 'status',
      render({ cell }: { cell: string }) {
        return h(
          'span',
          [
            cell || '--',
          ],
        );
      },
    },
    getLinkField('host', '绑定资源的实例', 'cvm_id', 'cvm_id'),
    {
      label: '绑定资源的类型',
      field: 'instance_type',
      render({ cell }: { cell: string }) {
        return h(
          'span',
          [
            cell || '--',
          ],
        );
      },
    },
    {
      label: '更新时间',
      field: 'updated_at',
      sort: true,
    },
    {
      label: '创建时间',
      field: 'created_at',
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
  };

  const columns = columnsMap[type] || [];

  return columns.filter((column: any) => !isSimpleShow || !column.onlyShowOnList);
};
