import { getMetadataStorage } from '@/decorator/metadata/globals';
import type { IColumnMetadata } from '@/decorator/typings';
import type { ModelPropertyGeneric } from './typings';
import type { ObjectType } from './manager';

export class Model<M> {
  target: ObjectType<M>;

  constructor(ModelClass: ObjectType<M>) {
    this.target = ModelClass;
  }

  createInstance(): M {
    return new this.target();
  }

  getProperties<T extends ModelPropertyGeneric>(): (T & { id: string & keyof M })[] {
    const columnMetadata = getMetadataStorage().columns.filter(
      (item: IColumnMetadata) => item.target === this.target || this.target.prototype instanceof (item.target as any),
    );
    const properties = columnMetadata.map((item: IColumnMetadata) => item.def);
    return properties
      .filter((item: ModelPropertyGeneric) => item.hidden !== true)
      .sort((a: ModelPropertyGeneric, b: ModelPropertyGeneric) => a.index - b.index);
  }

  getPropertiesByGroup<T extends ModelPropertyGeneric>(): Record<string, (T & { id: string & keyof M })[]> {
    const properties = this.getProperties<T>();
    return properties.reduce((acc, curr) => {
      acc[curr.group] = acc[curr.group] || [];
      acc[curr.group].push(curr);
      return acc;
    }, {} as Record<string, (T & { id: string & keyof M })[]>);
  }
}
