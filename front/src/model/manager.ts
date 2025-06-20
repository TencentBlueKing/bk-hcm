import { Model } from './model';

export interface ObjectLiteral {
  [key: string]: any;
}

export type ObjectType<T> = new () => T;

export function getModel<M extends ObjectLiteral>(ModelClass: ObjectType<M>) {
  return new Model<M>(ModelClass);
}
