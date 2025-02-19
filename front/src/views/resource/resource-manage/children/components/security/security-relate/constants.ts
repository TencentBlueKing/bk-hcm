import i18n from '@/language/i18n';
import { RelatedResourceType } from './typings';
import { ModelProperty } from '@/model/typings';
import relatedCvmsViewProperties from '@/model/security-group/related-cvms.view';
import relatedLoadbalancerViewProperties from '@/model/security-group/related-loadbalancer.view';

const { t } = i18n.global;

export const RELATED_RES_KEY_MAP: Record<RelatedResourceType, string> = { CVM: 'cvm', CLB: 'load_balancer' };

export const RELATED_RES_NAME_MAP: Record<RelatedResourceType, string> = { CVM: t('云主机'), CLB: t('负载均衡') };

export const RELATED_RES_PROPERTIES_MAP: Record<RelatedResourceType, ModelProperty[]> = {
  CVM: relatedCvmsViewProperties,
  CLB: relatedLoadbalancerViewProperties,
};
