import { CLOUD_HOST_STATUS } from '@/common/constant';
import { ModelProperty } from '@/model/typings';

export default [
  {
    id: 'cloud_id',
    name: '主机ID',
    type: 'string',
  },
  {
    id: 'private_ip',
    name: '内网IP',
    type: 'string',
  },
  {
    id: 'public_ip',
    name: '公网IP',
    type: 'string',
  },
  {
    id: 'cloud_vpc_ids',
    name: '所属vpc',
    type: 'array',
  },
  {
    id: 'region',
    name: '地域',
    type: 'region',
  },
  {
    id: 'zone',
    name: '可用区',
    type: 'string',
  },
  {
    id: 'name',
    name: '主机名称',
    type: 'string',
  },
  {
    id: 'status',
    name: '主机状态',
    type: 'enum',
    option: CLOUD_HOST_STATUS,
    meta: {
      display: {
        appearance: 'cvm-status',
      },
    },
  },
  {
    id: 'is_assigned',
    name: '是否分配',
    type: 'boolean', // 配合columnConfig render展示tag
  },
  // TODO：这些公共的properties后续应该可以优化到一个properties中
  {
    id: 'bk_biz_id',
    name: '所属业务',
    type: 'business',
  },
  {
    id: 'bk_cloud_id',
    name: '管控区域',
    type: 'number', // TODO：cloud-area
  },
  {
    id: 'machine_type',
    name: '实例规格',
    type: 'string',
  },
  {
    id: 'os_name',
    name: '操作系统',
    type: 'string',
  },
  {
    id: 'created_at',
    name: '创建时间',
    type: 'datetime',
  },
  {
    id: 'updated_at',
    name: '更新时间',
    type: 'datetime',
  },
] as ModelProperty[];
