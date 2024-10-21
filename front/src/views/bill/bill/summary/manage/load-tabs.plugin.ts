import i18n from '@/language/i18n';
import { baseTabs } from './tabs';
import OperationProduct from '@/views/bill/bill/summary/product';

const { t } = i18n.global;

export const getTabs = () => [...baseTabs, { name: 'product', label: t('业务'), Component: OperationProduct }];
