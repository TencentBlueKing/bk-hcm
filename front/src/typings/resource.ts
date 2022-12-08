// define
export type PlainObject = {
  [k: string]: string | boolean | number
};

export type DoublePlainObject = {
  [k: string]: PlainObject
};

export type FilterType = {
  op: 'and' | 'or';
  rules: {
    field: string;
    op: 'eq';
    value: string | number | string[];
  }[]
};
