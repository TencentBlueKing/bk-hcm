import { ISecurityGroupRelResCountItem } from '@/store/security-group';

export enum SecurityGroupManageType {
  platform = 'platform',
  business = 'business',
}

export type RelatedResourceType = ISecurityGroupRelResCountItem['resources'][1]['res_name'];
