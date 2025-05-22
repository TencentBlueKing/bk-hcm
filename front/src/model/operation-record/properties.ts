import { ModelProperty } from '@/model/typings';
import { QueryRuleOPEnum } from '@/typings';

export default [
  {
    id: 'created_at',
    name: '操作时间',
    type: 'datetime',
  },
  {
    id: 'res_type',
    name: '资源类型',
    type: 'string',
    meta: {
      search: {
        filterRules(value) {
          if (value === 'load_balancer') {
            return {
              field: 'res_type',
              op: QueryRuleOPEnum.IN,
              value: ['load_balancer', 'url_rule', 'listener', 'url_rule_domain', 'target_group'],
            };
          }
          return { field: 'res_type', op: QueryRuleOPEnum.EQ, value };
        },
      },
    },
  },
  {
    id: 'res_id',
    name: '实例ID',
    type: 'string',
  },
  {
    id: 'cloud_res_id',
    name: '云资源ID',
    type: 'string',
  },
  {
    id: 'res_name',
    name: '资源名称',
    type: 'string',
    meta: {
      search: {
        filterRules(value: string | string[]) {
          if (Array.isArray(value) && value.length > 1) {
            return {
              op: QueryRuleOPEnum.OR,
              rules: value.map((val) => ({ field: 'res_name', op: QueryRuleOPEnum.CS, value: val })),
            };
          }
          if (Array.isArray(value) && value.length === 1) {
            return { field: 'res_name', op: QueryRuleOPEnum.CS, value: value[0] };
          }
          return { field: 'res_name', op: QueryRuleOPEnum.CS, value };
        },
      },
    },
  },
  {
    id: 'source',
    name: '操作来源',
    type: 'string',
  },
  {
    id: 'action',
    name: '操作方式',
    type: 'string',
    meta: { search: { op: QueryRuleOPEnum.IN } },
  },
  {
    id: 'bk_biz_id',
    name: '所属业务',
    type: 'business',
  },
  {
    id: 'operator',
    name: '操作人',
    type: 'user',
  },
  {
    id: 'rid',
    name: '请求ID',
    type: 'string',
  },
  {
    id: 'detail.data.res_flow.flow_id',
    name: '任务类型',
    type: 'string',
    meta: { search: { op: QueryRuleOPEnum.JSON_NEQ, enableEmpty: true } },
  },
] as ModelProperty[];
