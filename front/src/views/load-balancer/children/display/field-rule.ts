import { Column, Model } from '@/decorator';
import { SCHEDULER_NAME } from '../../constants';

@Model('load-balancer/rule-display')
export class DisplayFieldRule {
  @Column('string', { name: '域名' })
  domain: string;

  @Column('string', { name: 'URL路径' })
  url: string[];

  @Column('enum', { name: '均衡方式', option: SCHEDULER_NAME })
  scheduler: string;

  @Column('string', { name: '目标组' })
  target_group_id: string;

  @Column('number', { name: 'RS数量' })
  rs_num: number;

  @Column('string', { name: '同步状态' })
  binding_status: string;
}
