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
    name: '权限策略库名称',
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
    meta: {
      search: {
        filterRules(value: string | string[]) {
          if (Array.isArray(value) && value.length > 1) {
            return {
              op: QueryRuleOPEnum.OR,
              rules: value.map((val) => ({ field: 'desc', op: QueryRuleOPEnum.CIS, value: val })),
            };
          }
          if (Array.isArray(value) && value.length === 1) {
            return { field: 'desc', op: QueryRuleOPEnum.CIS, value: value[0] };
          }
          return { field: 'desc', op: QueryRuleOPEnum.CIS, value };
        },
      },
    },
    index: 1,
  })
  desc: string;

  @Column('user', {
    name: '创建人',
    meta: {
      search: {
        filterRules(value: string | string[]) {
          if (Array.isArray(value) && value.length > 1) {
            return {
              op: QueryRuleOPEnum.OR,
              rules: value.map((val) => ({ field: 'creator', op: QueryRuleOPEnum.JSON_CONTAINS, value: val })),
            };
          }
          if (Array.isArray(value) && value.length === 1) {
            return { field: 'creator', op: QueryRuleOPEnum.JSON_CONTAINS, value: value[0] };
          }
          return { field: 'creator', op: QueryRuleOPEnum.JSON_CONTAINS, value };
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
              rules: value.map((val) => ({
                field: 'update',
                op: QueryRuleOPEnum.JSON_CONTAINS,
                value: val,
              })),
            };
          }
          if (Array.isArray(value) && value.length === 1) {
            return { field: 'update', op: QueryRuleOPEnum.JSON_CONTAINS, value: value[0] };
          }
          return { field: 'update', op: QueryRuleOPEnum.JSON_CONTAINS, value };
        },
      },
    },
    index: 3,
  })
  update: string;
}
