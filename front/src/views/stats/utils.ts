import dayjs from 'dayjs';

export function getRecentMonths(months = 6, includeCurrent = true) {
  const end = dayjs().endOf('month');
  const startMonthOffset = includeCurrent ? months - 1 : months;
  const start = dayjs().subtract(startMonthOffset, 'month').startOf('month');

  return {
    startDate: start.toDate(),
    endDate: end.toDate(),
  };
}

export function getMonthRange(date: Date, format = 'YYYY-MM-DD') {
  return {
    startTime: dayjs(date).startOf('month').format(format),
    endTime: dayjs(date).endOf('month').format(format),
  };
}

/**
 * 基于bk_biz_id合并生成表格数据
 * @param currentData 当前数据
 * @param compareData 对比数据
 * @param fields 字段
 * @returns 表格数据
 */
export function mergeTableData(currentData: any[], compareData: any[], fields: string[]) {
  // 1. 提取所有唯一的bk_biz_id
  const allBizIds = [
    ...new Set([...currentData.map((item) => item.bk_biz_id), ...compareData.map((item) => item.bk_biz_id)]),
  ];

  // 2. 创建数据映射便于快速查找
  const currentMap = new Map(currentData.map((item) => [item.bk_biz_id, item]));
  const compareMap = new Map(compareData.map((item) => [item.bk_biz_id, item]));

  // 3. 使用Object.keys读取字段配置并生成表格数据
  const tableData = allBizIds.map((bizId) => {
    const row: any = { bk_biz_id: bizId };
    const currentItem = currentMap.get(bizId) || {};
    const compareItem = compareMap.get(bizId) || {};

    // 动态处理每个字段
    fields.forEach((field) => {
      row[`current_${field}`] = currentItem[field] !== undefined ? currentItem[field] : undefined;
      row[`compare_${field}`] = compareItem[field] !== undefined ? compareItem[field] : undefined;
    });

    return row;
  });

  return tableData;
}

/**
 * 计算同比或环比增长率
 * @param {number|string} currentValue - 当前值（本期数）
 * @param {number|string} previousValue - 对比值（上期数）
 * @param {number} decimals - 保留小数位数，默认为2
 * @returns {object} 格式化后的同比或环比增长率，如 { text: '11.11%', class: 'up' }
 */
export function calculateGrowthRate(currentValue: number | string, previousValue: number | string, decimals = 2) {
  const current = parseFloat(currentValue as string);
  const previous = parseFloat(previousValue as string);
  const result = { text: '--', class: 'none', value: 0 };
  if (isNaN(current) || isNaN(previous)) {
    return result;
  }

  // 处理除数为零的情况
  if (previous === 0) {
    return result;
  }

  // 计算增长率
  const growthRate = ((current - previous) / previous) * 100;

  result.text = growthRate === 0 ? '0%' : `${Math.abs(growthRate).toFixed(decimals)}%`;
  result.value = growthRate;
  // eslint-disable-next-line no-nested-ternary
  result.class = growthRate === 0 ? 'flat' : growthRate > 0 ? 'up' : 'down';
  return result;
}

/**
 * 计算给定时间段的同比（上年同期）或环比（上月同期）时间段
 * @param {string} startDate - 起始日期，格式 'YYYY-MM'
 * @param {string} endDate - 结束日期，格式 'YYYY-MM'
 * @param {string} period - 周期类型，'yoy' 或 'mom'
 * @param {string} format - 日期格式，默认为 'YYYY-MM'
 * @returns {object} 包含当前期和对比期的对象
 */
export function getPeriodRange(startDate: Date, endDate: Date, period: 'yoy' | 'mom', format = 'YYYY-MM') {
  // 解析当前期日期
  const currentStart = dayjs(startDate);
  const currentEnd = dayjs(endDate);

  // 计算同比期：减一年或一个月
  const comparisonStart = currentStart.subtract(1, period === 'yoy' ? 'year' : 'month').format(format);
  const comparisonEnd = currentEnd.subtract(1, period === 'yoy' ? 'year' : 'month').format(format);

  return {
    currentRange: {
      start: currentStart.format(format),
      end: currentEnd.format(format),
    },
    comparisonRange: {
      start: comparisonStart,
      end: comparisonEnd,
    },
  };
}
