import { Column, Model } from '@/decorator';
import { CLB_STATUS_NAME, IP_VERSION_DISPLAY_NAME, LB_TYPE_NAME, LOAD_BALANCER_ISP_NAME } from '../../constants';
import { formatTags } from '@/common/util';
import { VendorMap } from '@/common/constant';

@Model('load-balancer/clb-display')
export class DisplayFieldClb {
  @Column('string', { name: '负载均衡名称', index: 0, width: 150, sort: true })
  name: string;

  // TODO-CLB: search-select多值搜索
  @Column('string', { name: '负载均衡ID', index: 0, width: 120 })
  cloud_id: string;

  @Column('string', { name: '负载均衡域名', index: 0, width: 120 })
  domain: string;

  @Column('string', { name: '负载均衡VIP', index: 0, width: 130 })
  lb_vip: string;

  @Column('region', { name: '地域', index: 0, width: 120, sort: true })
  region: string;

  @Column('array', { name: '可用区域', index: 0, width: 120, sort: true })
  zones: string;

  @Column('enum', { name: '网络类型', index: 0, width: 120, option: LB_TYPE_NAME, sort: true })
  lb_type: string;

  @Column('number', { name: '监听器数量', index: 0, width: 100 })
  listener_count: number;

  @Column('business', {
    name: '分配状态',
    index: 0,
    width: 100,
    meta: { display: { appearance: 'business-assign-tag' } },
  })
  bk_biz_id: string;

  @Column('bool', { name: '删除保护', index: 0, width: 100 })
  delete_protect: boolean;

  @Column('enum', { name: 'IP版本', index: 0, width: 80, option: IP_VERSION_DISPLAY_NAME, sort: true })
  ip_version: string;

  @Column('string', {
    name: '标签',
    index: 0,
    width: 80,
    meta: { display: { format: (val) => formatTags(val) } },
    defaultHidden: true,
  })
  tags: string;

  @Column('enum', { name: '云厂商', index: 0, width: 80, option: VendorMap, sort: true })
  vendor: string;

  @Column('enum', {
    name: '状态',
    index: 0,
    width: 120,
    option: CLB_STATUS_NAME,
    meta: { display: { appearance: 'clb-status' } },
    sort: true,
  })
  status: string;

  @Column('string', { name: '所属VPC', index: 0, width: 120, sort: true })
  cloud_vpc_id: string;

  @Column('enum', { name: '运营商', index: 0, width: 80, option: LOAD_BALANCER_ISP_NAME })
  isp: string;

  @Column('number', { name: '带宽', index: 0, width: 80 })
  bandwidth: number;

  @Column('string', { name: '数据同步时间', index: 0, width: 120 })
  sync_time: string;
}
