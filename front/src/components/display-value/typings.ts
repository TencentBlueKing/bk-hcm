import { type PropertyDisplayConfig } from '@/model/typings';

export type AppearanceType =
  | 'status'
  | 'link'
  | 'wxwork-link'
  | 'tag'
  | 'cvm-status'
  | 'clb-status'
  | 'business-assign-tag'
  | 'dynamic-status';

export type DisplayType = {
  on?: 'cell' | 'info' | 'search';
  appearance?: AppearanceType;
  showOverflowTooltip?: boolean;
} & PropertyDisplayConfig;
