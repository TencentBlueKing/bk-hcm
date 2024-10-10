import { JSX } from 'vue/jsx-runtime';

type StringCell = string | ((cell: string) => string); // 字符串或者返回字符串的函数

type BaseField = {
  name: string;
  value?: any;
  cls?: StringCell;
  link?: StringCell;
  copy?: boolean;
  edit?: boolean;
  type?: string;
  prop?: string;
  tipsContent?: string;
  txtBtn?: (cell: string) => void;
};
type FieldWithRenderString = BaseField & {
  render: (value: BaseField['value']) => string | number;
  copyContent?: StringCell; // 可选
};
type FieldWithRenderJSX = BaseField & {
  render: (value: BaseField['value']) => JSX.Element | '--';
  copyContent?: StringCell; // 必填
};
type FieldWithoutRender = BaseField & {
  render?: never; // 不传
  copyContent?: StringCell; // 可选
};

export type Field = FieldWithRenderString | FieldWithRenderJSX | FieldWithoutRender;
/*
  1.如果 render 返回 JSX.Element，copyContent 必须提供。
  2.如果 render 返回 string 或 render 未定义，copyContent 是可选的。
*/
export type EnsureValidField<F extends Field> = F extends { render: (value: BaseField['value']) => JSX.Element | '--' }
  ? (F & { copy: false }) | (F & { copyContent: StringCell })
  : F;

export type FieldList = Array<EnsureValidField<Field>>;
