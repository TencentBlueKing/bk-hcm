import { ModelProperty } from '@/model/typings';

export const findProperty = (id: ModelProperty['id'], properties: ModelProperty[], key?: keyof ModelProperty) => {
  // 先按默认的规则找
  let found = properties.find((property) => property.id === id);

  // 找不到同时指定了key则再根据key再找一次
  if (!found && key) {
    found = properties.find((property) => property[key] === id);
  }

  return found;
};
