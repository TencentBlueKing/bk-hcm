/* eslint-disable @typescript-eslint/no-unused-vars */
import { Model, Column } from '@/decorator';
import { QueryRuleOPEnum } from '@/typings';

// 资源纳管状态选项 (sync_status)
export const SYNC_STATUS_OPTIONS = {
  sync_success: '同步成功',
  sync_failed: '同步失败',
  not_sync: '未同步',
  syncing: '同步中',
};

// 站点类型选项
export const SITE_TYPE_OPTIONS = {
  china: '中国站',
  international: '国际站',
};

@Model('cloud-account-manage/search-condition')
export class SearchCondition {
  @Column('string', {
    name: '二级账号ID',
    props: {
      multiple: false,
    },
    index: 0,
  })
  id: string;

  @Column('string', {
    name: '二级账号名称',
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
    index: 2,
  })
  managers: string;

  @Column('user', {
    name: '安全负责人',
    meta: {
      search: {
        filterRules(value: string | string[]) {
          if (Array.isArray(value) && value.length > 1) {
            return {
              op: QueryRuleOPEnum.OR,
              rules: value.map((val) => ({
                field: 'security_managers',
                op: QueryRuleOPEnum.JSON_CONTAINS,
                value: val,
              })),
            };
          }
          if (Array.isArray(value) && value.length === 1) {
            return { field: 'security_managers', op: QueryRuleOPEnum.JSON_CONTAINS, value: value[0] };
          }
          return { field: 'security_managers', op: QueryRuleOPEnum.JSON_CONTAINS, value };
        },
      },
    },
    index: 3,
  })
  security_managers: string;

  @Column('string', {
    name: '邮箱',
    props: {
      multiple: false,
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
    index: 4,
  })
  email: string;

  // 以下字段暂时不显示，保留配置以备后用
  // @Column('enum', {
  //   name: '资源纳管',
  //   option: SYNC_STATUS_OPTIONS,
  //   index: 5,
  // })
  // sync_status: string;

  // @Column('enum', {
  //   name: '站点类型',
  //   option: SITE_TYPE_OPTIONS,
  //   index: 6,
  // })
  // site: string;

  // @Column('business', {
  //   name: '使用业务',
  //   index: 7,
  // })
  // usage_biz_ids: number[];
}
