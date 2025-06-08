export type DisplayAppearanceType = 'select' | 'cert' | 'ca' | 'tag-input';
export type DisplayOnType = 'default' | 'cell';

export type DisplayType = { on?: DisplayOnType; appearance?: DisplayAppearanceType; showOverflowTooltip?: boolean };
