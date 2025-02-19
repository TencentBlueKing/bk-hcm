import { ISecurityGroupRelResCountItem } from '@/store/security-group';

export enum SecurityGroupManageType {
  platform = 'platform',
  business = 'business',
}

export type ITab = ISecurityGroupRelResCountItem['resources'][1]['res_name'];
