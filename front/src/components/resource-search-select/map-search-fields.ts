import type { ISearchItem } from 'bkui-vue/lib/search-select/utils';
import type { ModelPropertySearch, ModelPropertyType } from '@/model/typings';

const ASYNC_TYPES = new Set<ModelPropertyType>(['business', 'account', 'cloud-area', 'region']);
const MULTIPLE_TYPES = new Set<ModelPropertyType>([
  'enum',
  'list',
  'user',
  'business',
  'account',
  'cloud-area',
  'region',
]);

const optionLabel = (value: unknown) => {
  if (value === null || value === undefined) return '';
  if (typeof value === 'string' || typeof value === 'number') return String(value);
  if (typeof value === 'object' && 'label' in value) return String((value as { label: unknown }).label ?? '');
  return String(value);
};

export const mapSearchFieldsToItems = (fields: ModelPropertySearch[]): ISearchItem[] => {
  return fields
    .filter((field) => !field.apiOnly && !field.hidden)
    .map((field) => {
      const item: ISearchItem = { id: field.id, name: field.name };
      const { option } = field;
      if (option && typeof option !== 'function') {
        item.children = Object.entries(option).map(([id, value]) => ({
          id: String(id),
          name: optionLabel(value),
        }));
      }
      // bkui SearchSelect 在传入 get-menu-list 时，async === undefined 也会走异步菜单（先出 loading 白块）
      const isAsync = ASYNC_TYPES.has(field.type) || typeof option === 'function';
      item.async = isAsync;
      item.children = item.children ?? [];
      if (MULTIPLE_TYPES.has(field.type)) {
        item.multiple = true;
      }
      const extra = field.meta?.search?.props;
      if (extra) Object.assign(item, extra);
      return item;
    });
};
