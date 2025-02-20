export type AppearanceType = 'status' | 'cvm-status' | 'clb-status';

export type DisplayType = {
  on?: 'cell' | 'info' | 'search';
  appearance?: AppearanceType;
  showOverflowTooltip?: boolean;
};
