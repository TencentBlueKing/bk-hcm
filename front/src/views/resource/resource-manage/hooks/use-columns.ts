// table 字段相关信息
import type {
  PlainObject,
} from '@/typings/resource';

import {
  h,
} from 'vue';

import {
  useRouter,
} from 'vue-router';

export default (type: string) => {
  const router = useRouter();

  const vpcColumns = [
    {
      type: 'selection',
      hiddenWhenDelete: true,
    },
    {
      label: 'ID',
      field: 'id',
      sort: true,
      render({ cell }: PlainObject) {
        return h(
          'span',
          {
            onClick() {
              router.push({
                name: 'resourceDetail',
                params: {
                  type: 'vpc',
                },
              });
            },
          },
          [
            cell || '--',
          ],
        );
      },
    },
    {
      label: '资源 ID',
      field: 'cid',
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
    },
    {
      label: '云区域',
      field: 'bk_cloud_id',
    },
    {
      label: '地域',
      field: 'region',
    },
    {
      label: 'IPv4 CIDR',
      field: 'ipv4_cidr',
    },
    {
      label: 'IPv6 CIDR',
      field: 'ipv6_cidr',
    },
    {
      label: '状态',
      field: 'status',
    },
    {
      label: '默认 VPC',
      field: 'is_default',
    },
    {
      label: '子网数',
      field: '',
    },
    {
      label: '创建时间',
      field: 'create_at',
      sort: true,
    },
    {
      label: '操作',
      field: '',
      hiddenWhenDelete: true,
    },
  ];

  const subnetColumns = [
    {
      type: 'selection',
      hiddenWhenDelete: true,
    },
    {
      label: 'ID',
      field: '',
      sort: true,
      render({ cell }: PlainObject) {
        return h(
          'span',
          {
            onClick() {
              router.push({
                name: 'resourceDetail',
                params: {
                  type: 'subnet',
                },
              });
            },
          },
          [
            cell || '--',
          ],
        );
      },
    },
    {
      label: '资源 ID',
      field: 'cid',
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
    },
    {
      label: '所属 VPC',
      field: 'vpc_cid',
    },
    {
      label: '可用区',
      field: 'zone',
    },
    {
      label: 'IPv4 CIDR',
      field: 'ipv4_cidr',
    },
    {
      label: 'IPv6 CIDR',
      field: 'ipv6_cidr',
    },
    {
      label: '关联路由表',
      field: '',
    },
    {
      label: '状态',
      field: 'status',
    },
    {
      label: '默认子网',
      field: 'is_default',
    },
    {
      label: '可用 IPv4 地址',
      field: '',
    },
    {
      label: '创建时间',
      field: 'create_at',
      sort: true,
    },
    {
      label: '操作',
      field: '',
      hiddenWhenDelete: true,
    },
  ];

  const groupColumns = [
    {
      type: 'selection',
      hiddenWhenDelete: true,
    },
    {
      label: 'ID',
      field: '',
      sort: true,
      render({ cell }: PlainObject) {
        return h(
          'span',
          {
            onClick() {
              router.push({
                name: 'resourceDetail',
                params: {
                  type: 'subnet',
                },
              });
            },
          },
          [
            cell || '--',
          ],
        );
      },
    },
    {
      label: '资源 ID',
      field: 'cid',
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
    },
    {
      label: '业务',
      field: 'vpc_cid',
    },
    {
      label: '业务拓扑',
      field: 'zone',
    },
    {
      label: '地域',
      field: 'ipv4_cidr',
    },
    {
      label: '描述',
      field: 'ipv6_cidr',
    },
    {
      label: '关联实例',
      field: '',
    },
  ];

  const gcpColumns = [
    {
      type: 'selection',
      hiddenWhenDelete: true,
    },
    {
      label: 'ID',
      field: '',
      sort: true,
      render({ cell }: PlainObject) {
        return h(
          'span',
          {
            onClick() {
              router.push({
                name: 'resourceDetail',
                params: {
                  type: 'subnet',
                },
              });
            },
          },
          [
            cell || '--',
          ],
        );
      },
    },
    {
      label: '资源 ID',
      field: 'cid',
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
    },
    {
      label: '业务',
      field: 'vpc_cid',
    },
    {
      label: '业务拓扑',
      field: 'zone',
    },
    {
      label: 'VPC',
      field: 'vpc',
    },
    {
      label: '描述',
      field: 'ipv6_cidr',
    },
  ];

  const columnsMap = {
    vpc: vpcColumns,
    subnet: subnetColumns,
    group: groupColumns,
    gcp: gcpColumns,
  };

  return columnsMap[type];
};
