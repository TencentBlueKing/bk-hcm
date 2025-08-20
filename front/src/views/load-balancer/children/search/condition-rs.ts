import { Column, Model } from '@/decorator';
import { RsInstType } from '../../constants';

@Model('load-balancer/rs-condition')
export class SearchConditionRs {
  @Column('array', { name: '内网IP' })
  private_ip_address: string[];

  @Column('array', { name: '公网IP' })
  public_ip_address: string[];

  @Column('string', { name: '名称' })
  inst_name: string;

  @Column('string', { name: '可用区' })
  zone: string;

  @Column('enum', { name: '资源类型', option: RsInstType })
  inst_type: string;

  @Column('array', { name: '所属VPC' })
  cloud_vpc_ids: string[];

  @Column('number', { name: '端口' })
  port: number;

  @Column('number', { name: '权重' })
  weight: number;
}
