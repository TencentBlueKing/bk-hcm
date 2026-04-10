/* eslint-disable @typescript-eslint/no-unused-vars */
import { Model, Column } from '@/decorator';
import { QueryRuleOPEnum } from '@/typings';

@Model('permission-policy/search-condition')
export class SearchCondition {
  @Column('string', {
    name: '权限策略库名称',
    props: {
      multiple: false,
    },
    meta: {
      search: {
        filterRules(value: string | string[]) {
          if (Array.isArray(value) && value.length > 1) {
            return {
              op: QueryRuleOPEnum.OR,
              rules: value.map((val) => ({ field: 'name', op: QueryRuleOPEnum.CIS, value: val })),
            };
          }
          if (Array.isArray(value) && value.length === 1) {
            return { field: 'name', op: QueryRuleOPEnum.CIS, value: value[0] };
          }
          return { field: 'name', op: QueryRuleOPEnum.CIS, value };
        },
      },
    },
    index: 0,
  })
  name: string;

  @Column('string', {
    name: '权限策略库描述',
    props: {
      multiple: false,
    },
    meta: {
      search: {
        filterRules(value: string | string[]) {
          if (Array.isArray(value) && value.length > 1) {
            return {
              op: QueryRuleOPEnum.OR,
              rules: value.map((val) => ({ field: 'memo', op: QueryRuleOPEnum.CIS, value: val })),
            };
          }
          if (Array.isArray(value) && value.length === 1) {
            return { field: 'memo', op: QueryRuleOPEnum.CIS, value: value[0] };
          }
          return { field: 'memo', op: QueryRuleOPEnum.CIS, value };
        },
      },
    },
    index: 1,
  })
  memo: string;

  @Column('user', {
    name: '创建人',
    meta: {
      search: {
        filterRules(value: string | string[]) {
          if (Array.isArray(value) && value.length > 1) {
            return {
              op: QueryRuleOPEnum.OR,
              rules: value.map((val) => ({ field: 'creator', op: QueryRuleOPEnum.EQ, value: val })),
            };
          }
          if (Array.isArray(value) && value.length === 1) {
            return { field: 'creator', op: QueryRuleOPEnum.EQ, value: value[0] };
          }
          return { field: 'creator', op: QueryRuleOPEnum.EQ, value };
        },
      },
    },
    index: 2,
  })
  creator: string;

  @Column('user', {
    name: '更新人',
    meta: {
      search: {
        filterRules(value: string | string[]) {
          if (Array.isArray(value) && value.length > 1) {
            return {
              op: QueryRuleOPEnum.OR,
              rules: value.map((val) => ({ field: 'reviser', op: QueryRuleOPEnum.EQ, value: val })),
            };
          }
          if (Array.isArray(value) && value.length === 1) {
            return { field: 'reviser', op: QueryRuleOPEnum.EQ, value: value[0] };
          }
          return { field: 'reviser', op: QueryRuleOPEnum.EQ, value };
        },
      },
    },
    index: 3,
  })
  reviser: string;
}
