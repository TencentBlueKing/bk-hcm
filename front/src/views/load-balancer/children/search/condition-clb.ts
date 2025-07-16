import { Column, Model } from '@/decorator';
import { CLB_STATUS_NAME, IP_VERSION_DISPLAY_NAME, LB_TYPE_NAME } from '../../constants';
import { LB_ISP, VendorEnum, VendorMap } from '@/common/constant';
import { QueryRuleOPEnum } from '@/typings';
import { buildVIPFilterRules, buildFilterRulesWithSearchSelect } from '@/utils/search';

@Model('load-balancer/clb-condition')
export class SearchConditionClb {
  @Column('string', {
    name: '负载均衡名称',
    meta: {
      search: {
        filterRules(value) {
          return buildFilterRulesWithSearchSelect(value, 'name', QueryRuleOPEnum.CS);
        },
      },
    },
  })
  name: string;

  @Column('string', { name: '负载均衡ID' })
  cloud_id: string;

  @Column('string', {
    name: '负载均衡域名',
    meta: {
      search: {
        filterRules(value) {
          return buildFilterRulesWithSearchSelect(value, 'domain', QueryRuleOPEnum.JSON_CONTAINS);
        },
      },
    },
  })
  domain: string;

  @Column('string', {
    name: '负载均衡VIP',
    meta: {
      search: {
        filterRules(value) {
          return buildVIPFilterRules(value);
        },
      },
    },
  })
  lb_vip: string;

  @Column('enum', { name: '网络类型', option: LB_TYPE_NAME })
  lb_type: string;

  @Column('enum', { name: 'IP版本', option: IP_VERSION_DISPLAY_NAME })
  ip_version: string;

  @Column('enum', { name: '运营商', option: LB_ISP })
  isp: string;

  @Column('number', { name: '带宽' })
  bandwidth: number;

  // @Column('enum', { name: '云厂商',  option: VendorMap })
  @Column('enum', { name: '云厂商', option: { [VendorEnum.TCLOUD]: VendorMap[VendorEnum.TCLOUD] } })
  vendor: string;

  @Column('region', { name: '地域' })
  region: string;

  @Column('string', {
    name: '可用区域',
    meta: {
      search: {
        filterRules(value) {
          return buildFilterRulesWithSearchSelect(value, 'zones', QueryRuleOPEnum.JSON_CONTAINS);
        },
      },
    },
  })
  zones: string;

  @Column('enum', { name: '状态', option: CLB_STATUS_NAME })
  status: string;

  @Column('string', { name: '所属VPC' })
  cloud_vpc_id: string;

  @Column('bool', { name: '删除保护', option: { true: '开启', false: '关闭' } })
  delete_protect: boolean;
}
