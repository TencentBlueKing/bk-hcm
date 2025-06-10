import { getMetadataStorage } from '@/decorator/metadata/globals';
import type { IColumnMetadata } from '@/decorator/typings';
import type { ModelPropertyGeneric } from './typings';
import type { ObjectType } from './manager';

export class Model<M> {
  instance: M;

  target: Function;

  constructor(ModelClass: ObjectType<M>) {
    this.target = ModelClass;
    this.instance = new ModelClass();
  }

  getProperties<T extends ModelPropertyGeneric>(): T[] {
    const columnMetadata = getMetadataStorage().columns.filter(
      (item: IColumnMetadata) => item.target === this.target || this.target.prototype instanceof (item.target as any),
    );
    const properties = columnMetadata.map((item: IColumnMetadata) => item.def);
    return properties.sort((a: ModelPropertyGeneric, b: ModelPropertyGeneric) => a.index - b.index);
  }
}
