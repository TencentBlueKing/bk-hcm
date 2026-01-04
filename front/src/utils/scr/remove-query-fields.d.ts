/**
 * 递归移除对象中的空字段（空字符串、空数组、空对象等）
 * 保留 number 和 boolean 类型的值
 * @param data - 需要处理的数据对象
 * @returns 处理后的数据对象
 */
export function removeEmptyFields<T extends Record<string, any>>(data: T): T;
