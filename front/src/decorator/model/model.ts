import type { IModelMetadata, ModelOptions } from '../typings';
import { getMetadataStorage } from '../metadata/globals';

export function Model(nameOrOptions?: string | ModelOptions, maybeOptions?: ModelOptions): ClassDecorator {
  const options = (nameOrOptions !== null && typeof nameOrOptions === 'object' ? nameOrOptions : maybeOptions) || {};
  const name = typeof nameOrOptions === 'string' ? nameOrOptions : options.name;

  return function (target) {
    const model: IModelMetadata = { target, name };
    getMetadataStorage().models.push(model);
  };
}
