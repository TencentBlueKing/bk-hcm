import i18n from '@/language/i18n';
import { ModelProperty } from '@/model/typings';
import relatedCvmsViewProperties from '@/model/security-group/related-cvms.view';
import relatedLoadbalancerViewProperties from '@/model/security-group/related-loadbalancer.view';

const { t } = i18n.global;

export enum RelatedResourceOperateType {
  BIND = 'bind',
  UNBIND = 'unbind',
}

export enum SecurityGroupManageType {
  BIZ = 'biz',
  PLATFORM = 'platform',
  UNKNOWN = '',
}

export enum SecurityGroupRelatedResourceName {
  CVM = 'CVM',
  CLB = 'CLB',
}

export const MGMT_TYPE_MAP: Record<string, string> = {
  [SecurityGroupManageType.BIZ]: t('业务管理'),
  [SecurityGroupManageType.PLATFORM]: t('平台管理'),
  [SecurityGroupManageType.UNKNOWN]: t('未确认'),
};

export const RELATED_RES_KEY_MAP: Record<SecurityGroupRelatedResourceName, string> = {
  [SecurityGroupRelatedResourceName.CVM]: 'cvm',
  [SecurityGroupRelatedResourceName.CLB]: 'load_balancer',
};

export const RELATED_RES_NAME_MAP: Record<SecurityGroupRelatedResourceName, string> = {
  [SecurityGroupRelatedResourceName.CVM]: t('云主机'),
  [SecurityGroupRelatedResourceName.CLB]: t('负载均衡'),
};

export const RELATED_RES_PROPERTIES_MAP: Record<SecurityGroupRelatedResourceName, ModelProperty[]> = {
  [SecurityGroupRelatedResourceName.CVM]: relatedCvmsViewProperties,
  [SecurityGroupRelatedResourceName.CLB]: relatedLoadbalancerViewProperties,
};

// 安全组-关联实例-操作disabled-tips
export const RELATED_RES_OPERATE_DISABLED_TIPS_MAP: Record<string, string> = {
  [RelatedResourceOperateType.BIND]: t('不支持安全组绑定负载均衡功能，请到负载均衡实例详情绑定安全组'),
  [RelatedResourceOperateType.UNBIND]: t('不支持安全组解绑负载均衡功能，请到负载均衡实例详情解绑安全组'),
};
