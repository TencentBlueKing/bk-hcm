import { ISecurityGroupRelResCountItem } from '@/store/security-group';

export enum SecurityGroupManageType {
  platform = 'platform',
  biz = 'biz',
}

export type RelatedResourceType = ISecurityGroupRelResCountItem['resources'][1]['res_name'];
