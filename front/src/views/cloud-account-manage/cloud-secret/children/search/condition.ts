/* eslint-disable @typescript-eslint/no-unused-vars */
import { Model, Column } from '@/decorator';
import { QueryRuleOPEnum } from '@/typings';

export const SECRET_STATUS_OPTIONS = {
  enabled: '已启用',
  disabled: '已禁用',
};

export const SUB_ACCOUNT_TYPE_OPTIONS = {
  console: '控制台账号',
  programming: '编程账号',
};

@Model('cloud-secret/search-condition')
export class SearchCondition {
  @Column('string', {
    name: '云密钥ID',
    meta: {
      search: {
        filterRules(value: string | string[]) {
          if (Array.isArray(value) && value.length > 1) {
            return {
              op: QueryRuleOPEnum.OR,
              rules: value.map((val) => ({ field: 'cloud_secret_id', op: QueryRuleOPEnum.CIS, value: val })),
            };
          }
          if (Array.isArray(value) && value.length === 1) {
            return { field: 'cloud_secret_id', op: QueryRuleOPEnum.CIS, value: value[0] };
          }
          return { field: 'cloud_secret_id', op: QueryRuleOPEnum.CIS, value };
        },
      },
    },
    index: 0,
  })
  cloud_secret_id: string;

  @Column('enum', {
    name: '密钥状态',
    option: SECRET_STATUS_OPTIONS,
    index: 1,
  })
  status: string;

  @Column('string', {
    name: '所属三级账号ID',
    op: QueryRuleOPEnum.CIS,
    index: 2,
  })
  cloud_sub_account_id: string;

  @Column('string', {
    name: '所属二级账号ID',
    op: QueryRuleOPEnum.CIS,
    index: 3,
  })
  cloud_main_account_id: string;

  @Column('user', {
    name: '三级账号负责人',
    meta: {
      search: {
        filterRules(value: string | string[]) {
          if (Array.isArray(value) && value.length > 1) {
            return {
              op: QueryRuleOPEnum.OR,
              rules: value.map((val) => ({
                field: 'sub_account_managers',
                op: QueryRuleOPEnum.JSON_CONTAINS,
                value: val,
              })),
            };
          }
          if (Array.isArray(value) && value.length === 1) {
            return { field: 'sub_account_managers', op: QueryRuleOPEnum.JSON_CONTAINS, value: value[0] };
          }
          return { field: 'sub_account_managers', op: QueryRuleOPEnum.JSON_CONTAINS, value };
        },
      },
    },
    index: 4,
  })
  sub_account_managers: string;

  @Column('user', {
    name: '二级账号负责人',
    meta: {
      search: {
        filterRules(value: string | string[]) {
          if (Array.isArray(value) && value.length > 1) {
            return {
              op: QueryRuleOPEnum.OR,
              rules: value.map((val) => ({ field: 'account_managers', op: QueryRuleOPEnum.JSON_CONTAINS, value: val })),
            };
          }
          if (Array.isArray(value) && value.length === 1) {
            return { field: 'account_managers', op: QueryRuleOPEnum.JSON_CONTAINS, value: value[0] };
          }
          return { field: 'account_managers', op: QueryRuleOPEnum.JSON_CONTAINS, value };
        },
      },
    },
    index: 5,
  })
  account_managers: string;
}
