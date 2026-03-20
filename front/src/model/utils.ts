import { ModelPropertyGeneric, ModelPropertyColumn } from '@/model/typings';

export const findProperty = (
  id: ModelPropertyGeneric['id'],
  properties: ModelPropertyGeneric[],
  key?: keyof ModelPropertyGeneric,
) => {
  // 先按默认的规则找
  let found = properties.find((property) => property.id === id);

  // 找不到同时指定了key则再根据key再找一次
  if (!found && key) {
    found = properties.find((property) => property[key] === id);
  }

  return found;
};

export const getColumnName = (property: ModelPropertyColumn, options?: { showUnit: boolean }) => {
  const { showUnit = true } = options || {};
  const { name, unit } = property;
  return `${name}${showUnit && unit ? `（${unit}）` : ''}`;
};

export const getColumnMinWidth = (
  property: Partial<ModelPropertyColumn>,
  options?: { fontSize?: number; hasSort?: boolean; offset?: number; min?: number },
) => {
  const { fontSize = 12, hasSort = false, offset = 42, min = 0 } = options ?? {};

  const content = typeof property?.name === 'string' ? property.name : '';

  // 字母/数字/空白按 0.7 倍字号宽度，其余（CJK 等宽字符）按 1 倍字号宽度
  const letterCount = (content.match(/[\w\s\\(\\)]/g) ?? []).length;
  const totalCount = content.length;
  const contentWidth = (totalCount - letterCount) * fontSize + letterCount * fontSize * 0.7;

  const finalWidth = contentWidth + (hasSort ? 22 : 0) + offset;

  return Math.ceil(Math.max(finalWidth, min));
};
