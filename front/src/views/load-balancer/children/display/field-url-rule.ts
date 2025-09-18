import { Column, Model } from '@/decorator';

@Model('load-balancer/rule-display')
export class DisplayFieldUrlRule {
  @Column('array', { name: 'VIP 域名' })
  lb_vips: string[];

  @Column('string', { name: '监听器协议' })
  lbl_protocol: string;

  @Column('number', { name: '监听器端口' })
  lbl_port: number;

  @Column('string', { name: 'URL' })
  rule_url: string;

  @Column('array', { name: '监听器域名' })
  rule_domain: string[];

  @Column('number', { name: 'RS数量' })
  target_count: number;
}
