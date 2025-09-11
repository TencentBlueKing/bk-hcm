import { Column, Model } from '@/decorator';
import { RsInstType } from '../../constants';

@Model('load-balancer/rs-display')
export class DisplayFieldDeviceRsRule {
  @Column('string', { name: '所属监听器名称', index: 0, width: 120 })
  lbl_name: string;

  @Column('number', { name: '所属监听器端口', index: 0, width: 120 })
  lbl_port: number;

  @Column('string', { name: '所属监听器域名', index: 0, width: 120 })
  lb_domain: string;

  @Column('string', { name: '所属URL', index: 0, width: 120 })
  rule_url: string;

  @Column('enum', { name: 'RS类型', index: 0, option: RsInstType, width: 100, fixed: 'left' })
  inst_type: string;

  @Column('string', { name: '所属目标组', index: 0, width: 120 })
  target_group_name: string;

  @Column('number', { name: 'RS端口', index: 0, width: 80, fixed: 'left' })
  port: number;

  @Column('number', { name: 'RS权重', index: 0, width: 100, fixed: 'left' })
  weight: number;

  @Column('array', { name: '所属负载均衡VIP', index: 0, width: 120 })
  lb_vips: string[];

  @Column('string', { name: '所属地域', index: 0, width: 120 })
  lb_region: string;
}
