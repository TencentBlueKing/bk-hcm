import { type PropertyDisplayConfig } from '@/model/typings';

export type AppearanceType = 'status' | 'link' | 'wxwork-link' | 'tag' | 'cvm-status' | 'clb-status';

export type DisplayType = {
  on?: 'cell' | 'info' | 'search';
  appearance?: AppearanceType;
  showOverflowTooltip?: boolean;
} & PropertyDisplayConfig;
