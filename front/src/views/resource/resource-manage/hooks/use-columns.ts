// table 字段相关信息
import type {
  PlainObject,
} from '@/typings/resource';
import i18n from '@/language/i18n';
import { CloudType } from '@/typings';
import {
  Button,
  InfoBox,
} from 'bkui-vue';
import {
  h,
} from 'vue';
import {
  useRoute,
  useRouter,
} from 'vue-router';
import {
  useResourceStore,
} from '@/store/resource';

export default (type: string) => {
  const resourceStore = useResourceStore();
  const router = useRouter();
  const route = useRoute();
  const { t } = i18n.global;

  const getDeleteField = (type: string) => {
    return {
      label: '操作',
      onlyShowOnList: true,
      render({ data }: any) {
        return h(
          Button,
          {
            text: true,
            theme: 'primary',
            onClick() {
              InfoBox({
                title: '请确认是否删除',
                subTitle: `将删除【${data.name}】`,
                theme: 'danger',
                headerAlign: 'center',
                footerAlign: 'center',
                contentAlign: 'center',
                onConfirm() {
                  resourceStore
                    .deleteBatch(
                      type,
                      {
                        ids: [data.id],
                      },
                    );
                },
              });
            },
          },
          [
            t('删除'),
          ],
        );
      },
    };
  };

  const getLinkField = (type: string, label: string = 'ID', field: string = 'id') => {
    return {
      label,
      field,
      sort: true,
      render({ cell }: { cell: string }) {
        return h(
          Button,
          {
            text: true,
            theme: 'primary',
            onClick() {
              const routeInfo: any = {
                query: {
                  id: cell,
                }
              }
              // 业务下
              if (route.path.includes('business')) {
                Object.assign(
                  routeInfo,
                  {
                    name: `${type}BusinessDetail`,
                  }
                )
              } else {
                Object.assign(
                  routeInfo,
                  {
                    name: 'resourceDetail',
                    params: {
                      type,
                    }
                  }
                )
              }
              router.push(routeInfo);
            },
          },
          [
            cell || '--',
          ],
        );
      },
    }
  };

  const vpcColumns = [
    {
      type: 'selection',
      onlyShowOnList: true,
    },
    getLinkField('vpc'),
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
      label: '地域',
      field: 'region',
    },
    {
      label: '云区域',
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
    getDeleteField('vpcs'),
  ];

  const subnetColumns = [
    {
      type: 'selection',
      onlyShowOnList: true,
    },
    getLinkField('subnet'),
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
      label: '地域',
      field: 'region',
    },
    {
      label: '所属 VPC',
      field: 'vpc_id',
    },
    {
      label: '关联路由表',
      field: '',
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
    getDeleteField('subnets'),
  ];

  const groupColumns = [
    {
      type: 'selection',
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
    },
    {
      label: '描述',
      field: 'memo',
    },
  ];

  const gcpColumns = [
    {
      type: 'selection',
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

  const driveColumns = [
    {
      type: 'selection',
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
    },
    {
      label: '容量(GB)',
      field: 'disk_size',
      sort: true,
    },
    {
      label: '运行状态',
      field: '',
    },
    {
      label: '可用区',
      field: 'zone',
      sort: true,
    },
    {
      label: '挂载实例',
      field: '',
      sort: true,
    },
    {
      label: '创建时间',
      field: 'created_at',
    },
    getDeleteField('disks'),
  ];

  const imageColumns = [
    getLinkField('image'),
    {
      label: '实例 ID',
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
      label: '平台',
      field: 'platform',
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
      label: '内网IP地址',
      field: 'internal_ip',
    },
    {
      label: '关联的公网IP地址',
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
    },
    {
      label: '名称',
      field: 'name',
      sort: true,
    },
    getLinkField('vpc', '所属网络(VPC)', 'vpc_id'),
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

  const columnsMap = {
    vpc: vpcColumns,
    subnet: subnetColumns,
    group: groupColumns,
    gcp: gcpColumns,
    drive: driveColumns,
    image: imageColumns,
    networkInterface: networkInterfaceColumns,
    route: routeColumns,
  };

  return columnsMap[type];
};
