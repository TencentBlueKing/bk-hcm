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
  useRouter,
} from 'vue-router';
import {
  useResourceStore,
} from '@/store/resource';

export default (type: string) => {
  const resourceStore = useResourceStore();
  const router = useRouter();
  const { t } = i18n.global;

  const getDeleteField = (type: string) => {
    return {
      label: '操作',
      hiddenWhenDelete: true,
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

  const vpcColumns = [
    {
      type: 'selection',
      hiddenWhenDelete: true,
    },
    {
      label: 'ID',
      field: 'id',
      sort: true,
      render({ cell }: { cell: string }) {
        return h(
          Button,
          {
            text: true,
            theme: 'primary',
            onClick() {
              router.push({
                name: 'resourceDetail',
                params: {
                  type: 'vpc',
                },
                query: {
                  id: cell,
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
      hiddenWhenDelete: true,
    },
    {
      label: 'ID',
      field: 'id',
      sort: true,
      render({ cell }: { cell: string }) {
        return h(
          Button,
          {
            text: true,
            theme: 'primary',
            onClick() {
              router.push({
                name: 'resourceDetail',
                params: {
                  type: 'subnet',
                },
                query: {
                  id: cell,
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
    // {
    //   label: '关联实例',
    //   field: '',
    //   render() {
    //     h(
    //       Button,
    //       {
    //         text: true,
    //         theme: 'primary',
    //         onClick() {
    //           router.push({
    //             name: 'resourceDetail',
    //             params: {
    //               type: 'security',
    //             },
    //             query: {
    //               activeTab: 'rule',
    //             },
    //           });
    //         },
    //       },
    //       [
    //         t('配置规则'),
    //       ],
    //     );
    //   },
    // },
  ];

  const gcpColumns = [
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
    {
      label: 'ID',
      field: 'id',
      sort: true,
      render({ cell }: { cell: string }) {
        return h(
          Button,
          {
            text: true,
            theme: 'primary',
            onClick() {
              router.push({
                name: 'resourceDetail',
                params: { type: 'drive' },
                query: {
                  id: cell,
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

  const columnsMap = {
    vpc: vpcColumns,
    subnet: subnetColumns,
    group: groupColumns,
    gcp: gcpColumns,
    drive: driveColumns,
  };

  return columnsMap[type];
};
