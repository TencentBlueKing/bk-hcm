import { ModelProperty } from '@/model/typings';
import { CLB_STATUS_MAP, IP_VERSION_MAP, LB_NETWORK_TYPE_MAP } from '@/constants';
import { QueryRuleOPEnum } from '@/typings';

export default [
  {
    id: 'name',
    name: '负载均衡名称',
    type: 'string',
    meta: {
      search: {
        filterRules(value: string | string[]) {
          if (Array.isArray(value) && value.length > 1) {
            return {
              op: QueryRuleOPEnum.OR,
              rules: value.map((val) => ({ field: 'name', op: QueryRuleOPEnum.CS, value: val })),
            };
          }
          if (Array.isArray(value) && value.length === 1) {
            return { field: 'name', op: QueryRuleOPEnum.CS, value: value[0] };
          }
          return { field: 'name', op: QueryRuleOPEnum.CS, value };
        },
      },
    },
  },
  {
    id: 'domain',
    name: '负载均衡域名',
    type: 'string',
    meta: {
      search: {
        filterRules(value) {
          if (Array.isArray(value) && value.length > 1) {
            return {
              op: QueryRuleOPEnum.OR,
              rules: value.map((val) => ({ field: 'zones', op: QueryRuleOPEnum.JSON_CONTAINS, value: val })),
            };
          }
          if (Array.isArray(value) && value.length === 1) {
            return { field: 'zones', op: QueryRuleOPEnum.JSON_CONTAINS, value: value[0] };
          }
          return { field: 'zones', op: QueryRuleOPEnum.JSON_CONTAINS, value };
        },
      },
    },
  },
  {
    id: 'lb_vip',
    name: '负载均衡VIP',
    type: 'string', // getInstVip(data)
  },
  {
    id: 'lb_type',
    name: '网络类型',
    type: 'enum',
    option: LB_NETWORK_TYPE_MAP,
  },
  {
    // *：异步加载的数据
    id: 'listener_num',
    name: '监听器数量',
    type: 'number',
  },
  {
    id: 'is_assigned',
    name: '分配状态',
    type: 'boolean', // 配合columnConfig render展示tag
  },
  {
    id: 'delete_protect',
    name: '删除保护',
    type: 'boolean', // 配合columnConfig render展示tag
  },
  {
    id: 'ip_version',
    name: 'IP版本',
    type: 'enum',
    option: IP_VERSION_MAP,
  },
  {
    id: 'tags',
    name: '标签',
    type: 'string', // formatTags(cell)
  },
  {
    id: 'region',
    name: '地域',
    type: 'region',
  },
  {
    id: 'zones',
    name: '可用区域',
    type: 'array',
    meta: {
      search: {
        filterRules(value) {
          if (Array.isArray(value) && value.length > 1) {
            return {
              op: QueryRuleOPEnum.OR,
              rules: value.map((val) => ({ field: 'zones', op: QueryRuleOPEnum.JSON_CONTAINS, value: val })),
            };
          }
          if (Array.isArray(value) && value.length === 1) {
            return { field: 'zones', op: QueryRuleOPEnum.JSON_CONTAINS, value: value[0] };
          }
          return { field: 'zones', op: QueryRuleOPEnum.JSON_CONTAINS, value };
        },
      },
    },
  },
  {
    id: 'status',
    name: '状态',
    type: 'enum',
    option: CLB_STATUS_MAP,
    meta: {
      display: {
        appearance: 'clb-status',
      },
    },
  },
  {
    id: 'cloud_vpc_id',
    name: '所属vpc',
    type: 'string',
    meta: {
      search: {
        op: QueryRuleOPEnum.IN,
      },
    },
  },
  {
    id: 'bk_biz_id',
    name: '所属业务',
    type: 'business',
  },
] as ModelProperty[];
