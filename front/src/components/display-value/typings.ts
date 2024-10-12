export type AppearanceType = 'status';

export type DisplayType = {
  on?: 'cell' | 'info' | 'search';
  appearance?: AppearanceType;
  showOverflowTooltip?: boolean;
};
