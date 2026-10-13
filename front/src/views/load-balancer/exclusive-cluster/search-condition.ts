import { Column, Model } from '@/decorator';
import { QueryRuleOPEnum } from '@/typings';
import { buildFilterRulesWithSearchSelect } from '@/utils/search';
import { EXCLUSIVE_CLUSTER_ISP_NAME, EXCLUSIVE_CLUSTER_TYPE_NAME } from './constants';

// 顺序对齐 column.ts 的表格列序，其中「最大连接数」「集群内实例数」不支持筛选
@Model('load-balancer/exclusive-cluster-search')
export class SearchConditionExclusiveCluster {
  @Column('string', {
    name: '集群ID',
    index: 0,
    meta: {
      search: {
        filterRules(value: string | string[]) {
          return buildFilterRulesWithSearchSelect(value, 'cloud_id', QueryRuleOPEnum.CS);
        },
      },
    },
  })
  cloud_id: string;

  @Column('string', {
    name: '集群名称',
    index: 1,
    meta: {
      search: {
        filterRules(value: string | string[]) {
          return buildFilterRulesWithSearchSelect(value, 'name', QueryRuleOPEnum.CS);
        },
      },
    },
  })
  name: string;

  @Column('enum', { name: '类型', index: 2, option: EXCLUSIVE_CLUSTER_TYPE_NAME })
  cluster_type: string;

  @Column('string', {
    name: '集群标签',
    index: 3,
    meta: {
      search: {
        filterRules(value: string | string[]) {
          return buildFilterRulesWithSearchSelect(value, 'cluster_tag', QueryRuleOPEnum.CS);
        },
      },
    },
  })
  cluster_tag: string;

  @Column('string', {
    name: '可用区',
    index: 4,
    meta: {
      search: {
        filterRules(value: string | string[]) {
          return buildFilterRulesWithSearchSelect(value, 'zone', QueryRuleOPEnum.CS);
        },
      },
    },
  })
  zone: string;

  @Column('enum', { name: '运营商', index: 5, option: EXCLUSIVE_CLUSTER_ISP_NAME })
  isp: string;

  @Column('business', { name: '业务', index: 6 })
  bk_biz_id: number;
}
