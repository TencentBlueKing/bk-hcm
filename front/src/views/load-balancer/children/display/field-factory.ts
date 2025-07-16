import { getModel } from '@/model/manager';
import { DisplayFieldClb } from './field-clb';
import { DisplayFieldListener } from './field-listener';
import { DisplayFieldRs } from './field-rs';

export enum DisplayFieldType {
  CLB = 'clb',
  LISTENER = 'listener',
  RS = 'rs',
}

export class DisplayFieldFactory {
  static createModel(key: DisplayFieldType) {
    switch (key) {
      case DisplayFieldType.CLB:
        return getModel(DisplayFieldClb);
      case DisplayFieldType.LISTENER:
        return getModel(DisplayFieldListener);
      case DisplayFieldType.RS:
        return getModel(DisplayFieldRs);
    }
  }
}
