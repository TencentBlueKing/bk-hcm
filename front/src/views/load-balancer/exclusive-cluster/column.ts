import { Column, Model } from '@/decorator';
import { EXCLUSIVE_CLUSTER_ISP_NAME, EXCLUSIVE_CLUSTER_TYPE_NAME } from './constants';

@Model('load-balancer/exclusive-cluster-column')
export class DisplayFieldExclusiveCluster {
  @Column('string', { name: '集群ID', index: 0, width: 120 })
  cloud_id: string;

  @Column('string', { name: '集群名称', index: 1, width: 180 })
  name: string;

  @Column('enum', { name: '类型', index: 2, width: 90, option: EXCLUSIVE_CLUSTER_TYPE_NAME })
  cluster_type: string;

  @Column('string', { name: '集群标签', index: 3, width: 120 })
  cluster_tag: string;

  @Column('string', { name: '可用区', index: 4, width: 140 })
  zone: string;

  @Column('enum', { name: '运营商', index: 5, width: 100, option: EXCLUSIVE_CLUSTER_ISP_NAME })
  isp: string;

  @Column('number', { name: '最大连接数', index: 6, width: 120, sort: true })
  max_conn: number;

  @Column('number', { name: '集群内实例数', index: 7, width: 130, sort: true })
  clb_resource_count: number;

  @Column('business', { name: '业务', index: 8, width: 140 })
  bk_biz_id: number;

  @Column('string', { name: '分配状态', index: 9, width: 90 })
  assign_status: string;
}
