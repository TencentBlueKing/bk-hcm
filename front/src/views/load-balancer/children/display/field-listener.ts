import { Column, Model } from '@/decorator';
import { LISTENER_PROTOCOL_NAME, SCHEDULER_NAME, SSL_MODE_NAME } from '../../constants';

@Model('load-balancer/listener-display')
export class DisplayFieldListener {
  @Column('string', { name: '监听器名称', index: 0, width: 100 })
  name: string;

  @Column('string', { name: '监听器ID', index: 0 })
  cloud_id: string;

  @Column('enum', { name: '协议', index: 0, option: LISTENER_PROTOCOL_NAME })
  protocol: string;

  @Column('number', { name: '端口', index: 0, sort: true })
  port: number;

  @Column('enum', { name: '均衡方式', index: 0, option: SCHEDULER_NAME })
  scheduler: string;

  @Column('number', { name: 'SNI', index: 0 })
  sni_switch: number;

  @Column('number', { name: 'RS数量', index: 0 })
  rs_num: number;

  @Column('number', { name: '域名数量', index: 0 })
  domain_num: number;

  @Column('number', { name: 'URL数量', index: 0 })
  url_num: number;

  @Column('string', { name: '同步状态', index: 0 })
  binding_status: string;

  // 详情展示字段
  @Column('string', { name: '协议端口', index: 0 })
  protocol_and_port: string;

  @Column('datetime', { name: '创建时间', index: 0 })
  created_at: string;

  @Column('number', { name: '会话时间', index: 0 })
  session_expire_time: number;

  @Column('string', {
    name: '健康探测源IP',
    index: 0,
    meta: { display: { format: (value) => (value === 0 ? '负载均衡 VIP' : '100.64.0.0/10网段') } },
  })
  'health_check.source_ip_type': string;

  @Column('string', { name: '检查方式', index: 0 })
  'health_check.check_type': string;

  @Column('number', { name: '检查端口', index: 0 })
  'health_check.check_port': number;

  @Column('string', { name: '检查选型', index: 0 })
  'health_check.check_scheme': string;

  @Column('enum', { name: '认证方式', index: 0, option: SSL_MODE_NAME })
  'certificate.ssl_mode': string;

  @Column('string', { name: '服务器证书', index: 0 })
  'certificate.ca_cloud_id': string;

  @Column('array', { name: 'CA证书', index: 0 })
  'certificate.cert_cloud_ids': string[];

  // 组合字段
  @Column('json', { name: 'RS权重不为0数 / RS总数' })
  rs_weight_stat: object;

  // 负载均衡相关字段
  @Column('string', { name: '负载均衡VIP', width: 100 })
  lb_vip: string;

  @Column('string', { name: '负载均衡ID', width: 100 })
  lb_cloud_id: string;

  @Column('string', { name: '网络类型' })
  lb_network_types: string;

  @Column('string', { name: '负载均衡VIP' })
  lb_vips: string;

  @Column('string', { name: '负载均衡ID' })
  lb_id: string;

  @Column('string', { name: '网络类型', width: 90 })
  lb_network_type: string;

  @Column('number', { name: 'RS数量', index: 0 })
  target_count: number;

  @Column('number', { name: '域名数量', index: 0 })
  rule_domain_count: number;

  @Column('number', { name: 'URL数量', index: 0 })
  url_count: number;
}
