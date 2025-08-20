import { VendorEnum } from '@/common/constant';
import { IQueryResData } from '@/typings';

// 用户网络类型
export type NetworkAccountType = 'STANDARD' | 'LEGACY';
interface AccountTypeInfo {
  NetworkAccountType: NetworkAccountType; // 用户账号的网络类型
  RequestId: string; // 请求id
}
// 用户网络类型 resp
export type NetworkAccountTypeResp = IQueryResData<AccountTypeInfo>;

// 查询当前地域下可用区列表和资源列表 - 输入参数
export interface ResourceOfCurrentRegionReqData {
  // 云账号id
  account_id: string;
  // 地域
  region: string;
  // 指定可用区
  master_zone?: string[];
  // 指定IP版本，如"IPv4"、"IPv6"、"IPv6_Nat"
  ip_version?: string[];
  // 指定运营商类型，如："BGP","CMCC","CUCC","CTCC"
  isp?: string[];
  // 返回可用区资源列表数目, 默认20, 最大值100
  limit?: number;
  // 返回可用区资源列表起始偏移量, 默认0
  offset?: number;
}
// 查询当前地域下可用区列表和资源列表 - 响应结果
export type ResourceOfCurrentRegionResp = IQueryResData<DescribeResourcesResponse>;
interface DescribeResourcesResponse {
  // 响应数据
  ZoneResourceSet: ZoneResource[];
  // 符合条件的总记录条数
  TotalCount: number;
  RequestId: string;
}
export interface ZoneResource {
  // 主可用区
  MasterZone: string;
  // 资源列表
  ResourceSet: {
    // 运营商内具体资源信息，如"CMCC", "CUCC", "CTCC", "BGP", "INTERNAL"
    Type: string[];
    // 运营商信息，如"CMCC", "CUCC", "CTCC", "BGP", "INTERNAL"
    Isp: string;
    // 可用资源
    AvailabilitySet?: {
      // 运营商内具体资源信息，如"CMCC", "CUCC", "CTCC", "BGP"
      Type: string;
      // 资源可用性，"Available"：可用，"Unavailable"：不可用
      Availability: string;
    }[];
    // 运营商类型信息
    TypeSet?: {
      // 运营商类型
      Type: string;
      // 规格可用性
      SpecAvailabilitySet?: SpecAvailability[];
    }[];
  }[];
  // 备可用区
  SlaveZone?: string;
  // ip版本（枚举值：IPv4，IPv6，IPv6_Nat）
  IPVersion: string;
  // 所属地域
  ZoneRegion: string;
  // 是否本地可用区
  LocalZone: boolean;
  // 可用区资源的类型，SHARED表示共享资源，EXCLUSIVE表示独占资源
  ZoneResourceType: string;
  // 可用区是否是EdgeZone可用区
  EdgeZone: boolean;
  // 网络出口
  Egress: string;
}
export interface SpecAvailability {
  // 规格类型 clb.c2.medium（标准型）clb.c3.small（高阶型1）clb.c3.medium（高阶型2）clb.c4.small（超强型1）clb.c4.medium（超强型2）clb.c4.large（超强型3）clb.c4.xlarge（超强型4）shared（共享型）
  SpecType?: string;
  // 规格可用性。资源可用性，"Available"：可用，"Unavailable"：不可用
  Availability?: string;
}

// 申请负载均衡 - 输入参数
export interface ApplyClbModel {
  // 业务ID
  bk_biz_id: number;
  // 账号ID
  account_id: string;
  // 地域
  region: string;
  // 网络类型: 公网 OPEN, 内网 INTERNAL
  load_balancer_type: 'OPEN' | 'INTERNAL';
  // 名称
  name: string;
  // 主可用区, 仅限公网型
  zones: string;
  // 备可用区, 目前仅广州、上海、南京、北京、中国香港、首尔地域的 IPv4 版本的 CLB 支持主备可用区。
  backup_zones?: string;
  // ip版本: IPV4, IPV6(ipv6 nat64), IPv6FullChain(ipv6)
  address_ip_version?: 'IPV4' | 'IPv6FullChain' | 'IPV6';
  // 安全组放通模式
  load_balancer_pass_to_target: boolean;
  // 云VpcID
  cloud_vpc_id: string;
  // 云子网ID, 内网型必填
  cloud_subnet_id?: string;
  // 绑定已有eip的ip地址, ipv6 nat64 不支持
  vip?: string;
  // 绑定eip id
  cloud_eip_id?: string;
  // 运营商类型(仅公网), 枚举值: CMCC, CUCC, CTCC, BGP。通过 TCloudDescribeResource 接口确定
  vip_isp?: string;
  // 网络计费模式(暂不支持包月)
  internet_charge_type?: 'TRAFFIC_POSTPAID_BY_HOUR' | 'BANDWIDTH_POSTPAID_BY_HOUR' | 'BANDWIDTH_PACKAGE';
  // 最大出带宽，单位Mbps
  internet_max_bandwidth_out?: number;
  // 带宽包id，计费模式为带宽包计费时必填
  bandwidth_package_id?: string;
  // 负载均衡规格类型: 性能容量型规格, 留空为共享型
  sla_type?: string;
  // // 按月付费自动续费(暂不支持包月)
  // auto_renew?: boolean;
  // 购买数量
  require_count: number;
  // 备注
  memo?: string;
  // 可用区类型, 0: 单可用区 2: 主备可用区（仅前端使用）
  zoneType: '0' | '1';
  // 云厂商（仅前端使用）
  vendor: VendorEnum;
  // 用户网络类型
  account_type: NetworkAccountType;
  // 负载均衡规格类型, 0：共享型 1：性能容量型（仅前端使用）
  slaType: '0' | '1';
  [key: string]: any;
}
