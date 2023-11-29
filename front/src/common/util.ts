import dayjs from 'dayjs';
// 获取 cookie object
export function getCookies(strCookie = document.cookie): any {
  if (!strCookie) {
    return {};
  }
  const arrCookie = strCookie.split('; ');// 分割
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
  return (item && Object.prototype.toString.apply(item) === '[object Object]');
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
 * 时间格式化
 * @param val 待格式化时间
 * @param format 格式
 * @returns 格式化后的时间
 */
export function timeFormatter(val: any, format = 'YYYY-MM-DD HH:mm:ss') {
  return val ? dayjs(val).format(format) : '--';
}

export function classes(dynamicCls: object, constCls = ''): string {
  return Object.entries(dynamicCls).filter(entry => entry[1])
    .map(entry => entry[0])
    .join(' ')
    .concat(constCls ? ` ${constCls}` : '');
};

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
    param instanceof String
      || typeof param === 'string'
      || param instanceof Number
      || typeof param === 'number'
      || param instanceof Boolean
      || typeof param === 'boolean'
  ) {
    // @ts-ignore
    paramStr += separator + key + mappingOperator + encodeURIComponent(param);
  } else {
    if (param) {
      Object.keys(param).forEach((p) => {
        const value = param[p];
        const k = key === null || key === '' || key === undefined
          ? p
          : key + (param instanceof Array ? `[${p}]` : `.${p}`);
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
  const windowHeight = document.compatMode === 'CSS1Compat'
    ? document.documentElement.clientHeight
    : document.body.clientHeight;

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
  const size = value / (1024 ** index);
  return `${size.toFixed(digits)}${uints[index]}`;
}

/**
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
