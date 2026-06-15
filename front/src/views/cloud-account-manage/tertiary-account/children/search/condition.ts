/* eslint-disable @typescript-eslint/no-unused-vars */
import { Model, Column } from '@/decorator';
import { QueryRuleOPEnum } from '@/typings';

export const ACCOUNT_TYPE_OPTIONS: Record<number, string> = {
  1: '控制台账号',
  0: '编程账号',
};

@Model('tertiary-account/search-condition')
export class SearchCondition {
  @Column('string', {
    name: '三级账号ID',
    props: {
      multiple: true,
    },
    meta: {
      search: {
        filterRules(value: string | string[]) {
          if (Array.isArray(value) && value.length > 1) {
            return { field: 'cloud_id', op: QueryRuleOPEnum.IN, value };
          }
          if (Array.isArray(value) && value.length === 1) {
            return { field: 'cloud_id', op: QueryRuleOPEnum.EQ, value: value[0] };
          }
          return { field: 'cloud_id', op: QueryRuleOPEnum.EQ, value };
        },
      },
    },
    index: 0,
  })
  cloud_id: string;

  @Column('string', {
    name: '三级账号名称',
    props: {
      multiple: true,
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
    index: 1,
  })
  name: string;

  @Column('string', {
    name: '所属二级账号ID',
    props: {
      multiple: true,
    },
    meta: {
      search: {
        filterRules(value: string | string[]) {
          return {
            field: 'extension.cloud_main_account_id',
            op: QueryRuleOPEnum.JSON_IN,
            value: Array.isArray(value) ? value : [value],
          };
        },
      },
    },
    index: 2,
  })
  'extension.cloud_main_account_id': string;

  @Column('enum', {
    name: '账号类型',
    option: ACCOUNT_TYPE_OPTIONS,
    props: {
      multiple: false,
    },
    meta: {
      search: {
        filterRules(value: string | number) {
          return { field: 'extension.console_login', op: QueryRuleOPEnum.JSON_EQ, value: Number(value) };
        },
      },
    },
    index: 3,
  })
  'extension.console_login': number;

  @Column('user', {
    name: '负责人',
    meta: {
      search: {
        filterRules(value: string | string[]) {
          if (Array.isArray(value) && value.length > 1) {
            return {
              op: QueryRuleOPEnum.OR,
              rules: value.map((val) => ({ field: 'managers', op: QueryRuleOPEnum.JSON_CONTAINS, value: val })),
            };
          }
          if (Array.isArray(value) && value.length === 1) {
            return { field: 'managers', op: QueryRuleOPEnum.JSON_CONTAINS, value: value[0] };
          }
          return { field: 'managers', op: QueryRuleOPEnum.JSON_CONTAINS, value };
        },
      },
    },
    index: 4,
  })
  managers: string;

  @Column('business', {
    name: '所属业务',
    meta: {
      search: {
        filterRules(value: number | number[]) {
          const values = Array.isArray(value) ? value.map(Number) : [Number(value)];
          return { field: 'bk_biz_ids', op: QueryRuleOPEnum.JSON_OVERLAPS, value: values };
        },
      },
    },
    index: 5,
  })
  bk_biz_ids: number[];

  @Column('string', {
    name: '邮箱',
    props: {
      multiple: true,
    },
    meta: {
      search: {
        filterRules(value: string | string[]) {
          if (Array.isArray(value) && value.length > 1) {
            return {
              op: QueryRuleOPEnum.OR,
              rules: value.map((val) => ({ field: 'email', op: QueryRuleOPEnum.CIS, value: val })),
            };
          }
          if (Array.isArray(value) && value.length === 1) {
            return { field: 'email', op: QueryRuleOPEnum.CIS, value: value[0] };
          }
          return { field: 'email', op: QueryRuleOPEnum.CIS, value };
        },
      },
    },
    index: 6,
  })
  email: string;

  @Column('string', {
    name: '手机号码',
    props: {
      multiple: false,
      placeholder: '请输入手机号，无国家码(地区码)前缀',
    },
    index: 7,
  })
  phone_num: string;
}
