// 将输入的字符串形式的数字转换并格式化为指定精度的字符串表示
export function formatBillCost(value: string, fixed = 3): string {
  if (!value?.trim()) {
    return '0';
  }

  const num = parseFloat(value);
  if (isNaN(num)) {
    return '0';
  }

  return num % 1 === 0 ? num.toString() : num.toFixed(fixed);
}
