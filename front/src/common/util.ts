import dayjs, { OpUnitType, QUnitType } from 'dayjs';
import utc from 'dayjs/plugin/utc';

dayjs.extend(utc);

// 获取 cookie object
export function getCookies(strCookie = document.cookie): any {
  if (!strCookie) {
    return {};
  }
  const arrCookie = strCookie.split('; '); // 分割
  const cookiesObj = {};
  arrCookie.forEach((cookieStr) => {
    const arr = cookieStr.split('=');
    const [key, value] = arr;
    if (key) {
      cookiesObj[key] = value;
    }
  });
  return cookiesObj;
}

/**
 * 检查是不是 object 类型
 * @param item
 * @returns {boolean}
 */
export function isObject(item: any) {
  return item && Object.prototype.toString.apply(item) === '[object Object]';
}

/**
 * 深度合并多个对象
 * @param objectArray 待合并列表
 * @returns {object} 合并后的对象
 */
export function deepMerge(...objectArray: any) {
  return objectArray.reduce((acc: any, obj: any) => {
    Object.keys(obj || {}).forEach((key) => {
      const pVal = acc[key];
      const oVal = obj[key];

      if (isObject(pVal) && isObject(oVal)) {
        acc[key] = deepMerge(pVal, oVal);
      } else {
        acc[key] = oVal;
      }
    });

    return acc;
  }, {});
}

/**
 * 时间格式化，自动转换成本地时区
 * @param val 待格式化时间
 * @param format 格式
 * @returns 格式化后的时间
 */
export function timeFormatter(val: any, format = 'YYYY-MM-DD HH:mm:ss', defaultVal = '--') {
  return val ? dayjs(val).format(format) : defaultVal;
}

/**
 * 格式化为UTC时间，忽略本地时区
 * @param val 待格式化时间
 * @param format 格式
 * @returns 格式化后的时间
 */
export function timeUTCFormatter(val: string, format = 'YYYY-MM-DD HH:mm:ss', defaultVal = '--') {
  return val ? dayjs.utc(val).format(format) : defaultVal;
}

/**
 * 相对当前的时间
 * @param val 待比较的时间
 * @returns 相对的时间字符串
 */
export function timeFromNow(val: string, unit: QUnitType | OpUnitType = 'minute') {
  return dayjs().diff(val, unit);
}

/**
 * 格式化当前时间与传入时间差值
 *  @param val 待比较的时间
 * @returns 几秒前，几分钟前，几小时前，几天前
 */
export function parseTimeFromNow(val: string) {
  const value = timeFromNow(val, 'second');
  if (value < 60) return `${value}秒前`;
  if (value < 3600) return `${Math.floor(value / 60)}分钟前`;
  if (value < 86400) return `${Math.floor(value / 3600)}小时前`;
  return `${Math.floor(value / 86400)}天前`;
}

/**
 * 为表格设置new标识(配合useTable使用)
 * @returns 'row-class': ({ created_at }: { created_at: string }) => string
 */
export function getTableNewRowClass() {
  return ({ created_at, updated_at }: { created_at?: string; updated_at?: string }) => {
    if ((created_at && timeFromNow(created_at) <= 5) || (updated_at && timeFromNow(updated_at) <= 5)) {
      return 'table-new-row';
    }
  };
}

export function classes(dynamicCls: object, constCls = ''): string {
  return Object.entries(dynamicCls)
    .filter((entry) => entry[1])
    .map((entry) => entry[0])
    .join(' ')
    .concat(constCls ? ` ${constCls}` : '');
}

/**
 * 获取Cookie
 * @param {String} name
 */
export const getCookie = (name: string) => {
  const reg = new RegExp(`(^|)${name}=([^;]*)(;|$)`);
  const data = document.cookie.match(reg);
  if (data) {
    return unescape(data[2]);
  }
  return null;
};

/**
 * 删除Cookie
 * @param {String} name
 */
export const deleteCookie = (name: string) => {
  document.cookie = `${name}=; expires=Thu, 01 Jan 1970 00:00:00 UTC; path=/`;
};

/**
 * 对象转为 url query 字符串
 *
 * @param {*} param 要转的参数
 * @param {string} key key
 *
 * @return {string} url query 字符串
 */
export function json2Query(param: any, key?: any) {
  const mappingOperator = '=';
  const separator = '&';
  let paramStr = '';
  if (
    param instanceof String ||
    typeof param === 'string' ||
    param instanceof Number ||
    typeof param === 'number' ||
    param instanceof Boolean ||
    typeof param === 'boolean'
  ) {
    // @ts-ignore
    paramStr += separator + key + mappingOperator + encodeURIComponent(param);
  } else {
    if (param) {
      Object.keys(param).forEach((p) => {
        const value = param[p];
        const k =
          key === null || key === '' || key === undefined ? p : key + (param instanceof Array ? `[${p}]` : `.${p}`);
        paramStr += separator + json2Query(value, k);
      });
    }
  }
  return paramStr.substr(1);
}

/**
 * 浏览器视口的高度
 *
 * @return {number} 浏览器视口的高度
 */
export function getWindowHeight() {
  const windowHeight =
    document.compatMode === 'CSS1Compat' ? document.documentElement.clientHeight : document.body.clientHeight;

  return windowHeight;
}

/**
 * 将字节大小转换为更易读的的MB/GB等单位大小
 * @param value 原大小，byte
 * @param digits 保留小数位数
 * @returns 转换后的大小，如 4026531840 -> 4GB
 */
export function formatStorageSize(value: number, digits = 0) {
  const uints = ['Bytes', 'KB', 'MB', 'GB', 'TB', 'PB', 'EB', 'ZB', 'YB'];
  const index = Math.floor(Math.log(value) / Math.log(1024));
  const size = value / 1024 ** index;
  return `${size.toFixed(digits)}${uints[index]}`;
}

/**
 * 获取当前网络得分对应的颜色
 * @param score 得分数值
 */
export function getScoreColor(score: number) {
  if (score > 0 && score < 180) {
    return '#00A62B';
  }
  if (score >= 180 && score <= 360) {
    return '#FF9D00';
  }
  if (score > 360) {
    return '#EA3636';
  }
  return '#63656E';
}
/*
 * 将 Map 类型的数据 key-value 互换，并输出一个对象
 * @param map 需要转换的 Map 对象
 * @returns 转换后的普通 js 对象
 */
export function swapMapKeysAndValuesToObj(map: Map<string, string>) {
  const _obj = {};
  for (const [key, value] of map) {
    _obj[value] = key;
  }
  return _obj;
}

// 求两个数组的差集
export function getDifferenceSet(origin: Array<string>, compare: Array<string>) {
  const set = new Set(origin);
  compare.forEach((item) => {
    if (set.has(item)) {
      set.delete(item);
    }
  });
  return Array.from(set);
}

// localStorage 操作类
export const localStorageActions = {
  set(key: string, value: any) {
    if (typeof value === 'object') {
      value = JSON.stringify(value);
    }
    localStorage.setItem(key, value);
  },
  get(key: string, parseFn?: (value: string) => any) {
    const value = localStorage.getItem(key);
    if (value) {
      if (parseFn) {
        return parseFn(value);
      }
      return JSON.parse(value);
    }
    return null;
  },
  remove(key: string) {
    localStorage.removeItem(key);
  },
  clear() {
    localStorage.clear();
  },
};

/**
 * 获取指定的 URL 查询参数值
 * @param {string} param 要获取的查询参数名
 * @param {string} url 可选，指定的 URL，默认为当前浏览器地址
 * @returns {string | null} 查询参数值，如果不存在则返回 null
 */
export const getQueryStringParams = (param: string, url = window.location.href) => {
  let queryParams;

  if (url.includes('#')) {
    // 如果 URL 包含 #，假定是 hash 路由，需要从 hash 中解析查询参数
    const hash = url.split('#')[1]; // 获取 hash 部分
    if (hash.includes('?')) {
      const search = hash.split('?')[1]; // 从 hash 中分离查询字符串
      queryParams = new URLSearchParams(search);
    } else {
      // 如果 hash 中没有查询字符串，提前返回 null
      return null;
    }
  } else {
    // 如果是常规路由，直接从 URL 对象解析查询字符串
    const urlObj = new URL(url);
    queryParams = new URLSearchParams(urlObj.search);
  }

  return queryParams.get(param);
};

/**
 * 判断值是否为空
 * @param { string | array | object | null | undefined } value 值
 * @returns boolean
 */
export const isEmpty = (value: unknown) => {
  if (value === '' || value === null || value === undefined) {
    return true;
  }
  if (Array.isArray(value) && value.length < 1) {
    return true;
  }
  if (Object.prototype.toString.call(value) === '[object Object]' && Object.keys(value).length < 1) {
    return true;
  }
  return false;
};

// 标签解析
export const formatTags = (data: { [k: string]: any }) => {
  return (
    Object.entries(data ?? {})
      .map((item) => item.join(':'))
      .join(';') || '--'
  );
};

export const resolveApiPathByBusinessId = (prefix: string, suffix: string, businessId?: number) => {
  return businessId ? `${prefix}/bizs/${businessId}/${suffix}` : `${prefix}/${suffix}`;
};
