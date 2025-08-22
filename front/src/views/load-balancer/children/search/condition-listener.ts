import { Column, Model } from '@/decorator';
import { LISTENER_PROTOCOL_NAME } from '../../constants';
import { QueryRuleOPEnum } from '@/typings';
import { buildFilterRulesWithSearchSelect } from '@/utils/search';
import { toArray } from '@/common/util';

@Model('load-balancer/listener-condition')
export class SearchConditionListener {
  @Column('string', { name: '监听器ID' })
  id: string;

  @Column('string', { name: '监听器ID' })
  cloud_id: string;

  @Column('string', {
    name: '监听器名称',
    meta: {
      search: {
        filterRules(value) {
          return buildFilterRulesWithSearchSelect(value, 'name', QueryRuleOPEnum.CS);
        },
      },
    },
  })
  name: string;

  @Column('enum', { name: '协议', option: LISTENER_PROTOCOL_NAME })
  protocol: string;

  @Column('array', {
    name: '端口',
    format: (value) => {
      const values = toArray(value).map((val: string) => Number(val));
      return values.filter((val: number) => !isNaN(val));
    },
  })
  port: number[];
}
