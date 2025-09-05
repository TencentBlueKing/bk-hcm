import { getModel } from '@/model/manager';
import { DisplayFieldClb } from './field-clb';
import { DisplayFieldListener } from './field-listener';
import { DisplayFieldRs } from './field-rs';
import { DisplayFieldRule } from './field-rule';
import { DisplayFieldUrlRule } from './field-url-rule';
import { DisplayFieldDeviceRsRule } from './field-device-rs';

export enum DisplayFieldType {
  CLB = 'clb',
  LISTENER = 'listener',
  RS = 'rs',
  Rule = 'rule',
  URL = 'url-rule',
  DeviceRs = 'device-rs',
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
      case DisplayFieldType.Rule:
        return getModel(DisplayFieldRule);
      case DisplayFieldType.URL:
        return getModel(DisplayFieldUrlRule);
      case DisplayFieldType.DeviceRs:
        return getModel(DisplayFieldDeviceRsRule);
    }
  }
}
