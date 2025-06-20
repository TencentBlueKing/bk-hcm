import { ModelPropertyType, ModelPropertyGeneric } from '@/model/typings';
import { getMetadataStorage } from '../metadata/globals';
import { IColumnMetadata } from '../typings';

export function Column(
  typeOrDef?: ModelPropertyType | ModelPropertyGeneric,
  def?: Partial<ModelPropertyGeneric>,
): PropertyDecorator {
  return function (object: Object, propertyName: string | symbol) {
    // 格式化参数
    let type: ModelPropertyType | undefined;

    if (typeof typeOrDef === 'string') {
      type = typeOrDef;
    } else if (typeOrDef) {
      def = typeOrDef as ModelPropertyGeneric;
      type = typeOrDef.type;
    }

    if (!def) {
      def = {};
    }

    // 尝试自动读取type
    const reflectMetadataType =
      Reflect && (Reflect as any).getMetadata
        ? (Reflect as any).getMetadata('design:type', object, propertyName)
        : undefined;
    if (!type && reflectMetadataType) {
      type = reflectMetadataType;
    }

    if (!def.type && type) {
      def.type = type;
    }

    if (!def.type)
      throw new Error(
        `Column type for ${object.constructor.name}#${String(propertyName)} is not defined and cannot be guessed.`,
      );

    const columnMetadata: IColumnMetadata = {
      target: object.constructor,
      propertyName,
      def: {
        id: propertyName as string,
        name: def.name || propertyName.toString(),
        ...def,
      },
    };

    getMetadataStorage().columns.push(columnMetadata);
  };
}
