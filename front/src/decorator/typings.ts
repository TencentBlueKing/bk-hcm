import { ModelPropertyGeneric } from '@/model/typings';

export interface IColumnMetadata {
  readonly target: Function | string;
  readonly propertyName: string | symbol;
  readonly def: Partial<ModelPropertyGeneric>;
}

export interface IModelMetadata {
  readonly target: Function | string;
  readonly name: string;
}

export interface ModelOptions {
  name?: string;
}
