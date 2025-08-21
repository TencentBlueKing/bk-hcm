import { RulesItem } from '@/typings';

export interface ApplicationsType {
  label: string;
  name: string;
  Component: any;
  rules?: RulesItem[];
  props?: any;
}
