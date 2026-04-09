import { TAB_PANELS } from './constants';

export type TabName = (typeof TAB_PANELS)[number]['name'];

export interface SwitchTabOptions {
  tab: TabName;
  filter?: Record<string, any>;
  detailCloudId?: string;
}

export type SwitchTabFn = (opts: SwitchTabOptions) => void;
