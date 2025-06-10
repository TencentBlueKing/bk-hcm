import { MetadataStorage } from './metadata-storage';

export function getMetadataStorage() {
  const globalScope = global as any;
  if (!globalScope.hcmMetadataStorage) {
    globalScope.hcmMetadataStorage = new MetadataStorage();
  }
  return globalScope.hcmMetadataStorage;
}
