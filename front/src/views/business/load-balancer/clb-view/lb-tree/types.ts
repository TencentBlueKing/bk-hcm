import { VendorEnum } from '@/common/constant';
import { IListResData, IQueryResData } from '@/typings';

export type ResourceTypeStr = 'all' | 'lb' | 'listener' | 'domain' | 'loading';

interface ITreeNodeConfig<T> {
  /**
   * 资源类型
   */
  type: ResourceTypeStr;
  /**
   * todo: Tree组件节点唯一标识。考虑到 domain 接口中没有 cloud_id，因此需要手动构建
   */
  nodeKey: string;
  /**
   * 显示值
   */
  displayValue: string;
  /**
   * Tree组件异步加载配置项
   */
  async: boolean;
  /**
   * 下级资源分页接口的起始偏移量
   */
  start?: number;
  /**
   * 下级资源总数
   */
  count?: number;
  /**
   * 下级资源列表数据
   */
  children: Array<T>;
}

// 负载均衡
export interface LoadBalancer extends ITreeNodeConfig<Listener> {
  id: string; // 资源ID
  cloud_id: string; // 云资源ID
  name: string; // 名称
  vendor: string; // 供应商
  account_id: string; // 账号ID
  bk_biz_id: number; // 业务ID
  ip_version: string; // ip版本
  lb_type: 'OPEN' | 'INTERNAL'; // 网络类型
  region: string; // 地域
  zones: string[]; // 可用区
  backup_zones: any[]; // 备可用区
  vpc_id: string; // vpcID
  cloud_vpc_id: string; // 云vpcID
  subnet_id: string; // 子网ID
  cloud_subnet_id: string; // 云子网ID
  private_ipv4_addresses: any[]; // 内网ipv4地址
  private_ipv6_addresses: any[]; // 内网ipv6地址
  public_ipv4_addresses: string[]; // 外网ipv4地址
  public_ipv6_addresses: any[]; // 外网ipv6地址
  domain: string; // 域名
  status: string; // 状态
  cloud_created_time: string; // 负载均衡在云上创建时间
  cloud_status_time: string; // 负载均衡状态变更时间
  cloud_expired_time: string; // 负载均衡过期时间
  memo: null; // 备注
  creator: string; // 创建者
  reviser: string; // 修改者
  created_at: string; // 创建时间
  updated_at: string; // 修改时间
  delete_protect: boolean; // 是否开启删除保护
  listenerNum: number; // todo: 监听器数量。异步接口请求获得。
}
export type LoadBalancers = Array<LoadBalancer>;
export type LoadBalancerListResData = IListResData<LoadBalancers>;

// 监听器
export interface Listener extends ITreeNodeConfig<Domain> {
  id: string; // 资源ID
  cloud_id: string; // 云资源ID
  name: string; // 名称
  vendor: VendorEnum; // 供应商
  bk_biz_id: number; // 业务ID
  account_id: string; // 账号ID
  lb_id: string; // 负载均衡ID
  cloud_lb_id: string; // 云负载均衡ID
  protocol: string; // 协议
  port: string; // 端口
  end_port: string; // 终止端口
  default_domain: string; // 默认域名
  zones: string[]; // 可用区数组
  target_group_id: string; // 目标组ID
  scheduler: string[]; // 负载均衡方式数组
  domain_num: number; // 域名数量
  url_num: number; // URL数量
  binding_status: 'success' | 'failed' | 'binding' | 'partial_failed'; // 绑定状态
  memo: string; // 备注
  creator: string; // 创建者
  reviser: string; // 修改者
  created_at: string; // 创建时间
  updated_at: string; // 修改时间
}
export type Listeners = Array<Listener>;
export type ListenerListResData = IListResData<Listeners>;

// 域名
export interface Domain extends ITreeNodeConfig<never> {
  domain: string; // 监听的域名
  url_count: number; // url数量
  id: string; // todo: 资源ID，用于路由跳转。接口中没有该字段，需要在程序中构建该关系。
  isDefault: boolean; // todo: 是否为默认域名。接口中没有该字段，需要在程序中构建该关系
  listener_id: string; // todo: 所属监听器ID。接口中没有该字段，需要在程序中构建该关系
}
export type Domains = Array<Domain>;
export type DomainListResData = IQueryResData<{
  default_domain: string; // 默认域名
  domain_list: Domains; // 域名信息列表
}>;

export type ResourceType = LoadBalancer | Listener | Domain;
export type ResourcesType = LoadBalancers | Listeners | Domains;
