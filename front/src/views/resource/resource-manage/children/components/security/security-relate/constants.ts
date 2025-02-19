import i18n from '@/language/i18n';
import { ITab } from './typings';

const { t } = i18n.global;

export const RELATED_RES_KEY_MAP: Record<ITab, string> = { CVM: 'cvm', CLB: 'load_balancer' };

export const RELATED_RES_NAME_MAP: Record<ITab, string> = { CVM: t('云主机'), CLB: t('负载均衡') };
