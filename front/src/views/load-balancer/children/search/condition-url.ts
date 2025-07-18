import { Column, Model } from '@/decorator';

@Model('load-balancer/url-condition')
export class SearchConditionUrl {
  @Column('string', { name: '域名' })
  domain: string;
}
