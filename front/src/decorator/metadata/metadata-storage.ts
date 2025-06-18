import { IModelMetadata, IColumnMetadata } from '../typings';

export class MetadataStorage {
  readonly models: IModelMetadata[] = [];

  readonly columns: IColumnMetadata[] = [];
}
