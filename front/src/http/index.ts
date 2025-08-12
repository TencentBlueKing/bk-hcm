/**
 * @file axios 封装
 * @author
 */

import axios, { AxiosInstance, AxiosRequestConfig } from 'axios';
import Cookies from 'js-cookie';
import { Message } from 'bkui-vue';
import { defaults } from 'lodash';
import { v4 as uuidv4 } from 'uuid';
import { showLoginModal } from '@/utils/login-helper';
import bus from '@/common/bus';
import CachedPromise from './cached-promise';
import RequestQueue from './request-queue';

interface HttpApi {
  [key: string]: any;
}

type HttpMethodType = 'delete' | 'get' | 'head' | 'options' | 'post' | 'put' | 'patch';

// axios 实例
const axiosInstance: AxiosInstance = axios.create({
  baseURL: window.PROJECT_CONFIG.BK_HCM_AJAX_URL_PREFIX,
  withCredentials: true,
  headers: { 'X-REQUESTED-WITH': 'XMLHttpRequest' },
});

/**
 * request interceptor
 */
axiosInstance.interceptors.request.use(
  (config: any) => {
    if (config.globalHeaders) {
      config.headers['X-Bkapi-Request-Id'] = uuidv4();
    }
    // 在发起请求前，注入CSRFToken，解决跨域
    injectCSRFTokenToHeaders();
    return config;
  },
  (error) => Promise.reject(error),
);

/**
 * response interceptor
 */
axiosInstance.interceptors.response.use(
  (response) => response,
  (error) => Promise.reject(error),
);

const http: HttpApi = {
  queue: new RequestQueue(),
  cache: new CachedPromise(),
  cancelRequest: (requestId: string) => {
    return http.queue.cancel(requestId);
  },
  cancelCache: (requestId: string) => http.cache.delete(requestId),
  cancel: (requestId: string) => Promise.all([http.cancelRequest(requestId), http.cancelCache(requestId)]),
  download: async (config: CombinedRequestConfig) => {
    defaults(config, { method: 'post', responseType: 'blob', originalResponse: true, globalError: true });
    // 设置请求配置默认值
    try {
      const { data, headers } = await axiosInstance(config);
      if (headers['content-type'] === 'application/octet-stream') {
        const downloadUrl = URL.createObjectURL(data);
        const a = document.createElement('a');
        a.href = downloadUrl;
        // 设置默认文件名
        [, a.download] = headers['content-disposition'].match(/filename="(.+)"/);
        // 下载
        document.body.appendChild(a);
        a.click();
        // 释放资源
        document.body.removeChild(a);
        URL.revokeObjectURL(downloadUrl);
      } else if (headers['content-type'] === 'application/json') {
        // 下载业务报错
        const reader = new FileReader();
        reader.onload = () => {
          const error = JSON.parse(reader.result as string);
          Message({ theme: 'error', message: error.message });
        };
        reader.readAsText(data);
      } else {
        throw new Error(`unknown Content-Type: ${headers['content-type']}`);
      }
    } catch (error) {
      if (config.globalError) {
        Message({ theme: 'error', message: (error as Error).message });
      }
      return Promise.reject(error);
    }
  },
  setHeader: (key: string, value: string) => {
    axiosInstance.defaults.headers[key] = value;
  },
  deleteHeader: (key: string) => {
    delete axiosInstance.defaults.headers[key];
  },
};

const methodsWithoutData: HttpMethodType[] = ['get', 'head', 'options'];
const methodsWithData: HttpMethodType[] = ['post', 'put', 'patch', 'delete'];
const allMethods = [...methodsWithoutData, ...methodsWithData];

const TokenInvalidCode = 2000009;

// 在自定义对象 http 上添加各请求方法
allMethods.forEach((method) => {
  Object.defineProperty(http, method, {
    get() {
      return getRequest(method);
    },
  });
});

/**
 * 获取 http 不同请求方式对应的函数
 *
 * @param {string} http method 与 axios 实例中的 method 保持一致
 *
 * @return {Function} 实际调用的请求函数
 */
function getRequest(method: HttpMethodType) {
  if (methodsWithData.includes(method)) {
    return (url: string, data: object, config: object) => getPromise(method, url, data, config);
  }
  return (url: string, config: object) => getPromise(method, url, null, config);
}

/**
 * 实际发起 http 请求的函数，根据配置调用缓存的 promise 或者发起新的请求
 *
 * @param {method} http method 与 axios 实例中的 method 保持一致
 * @param {string} 请求地址
 * @param {Object} 需要传递的数据, 仅 post/put/patch 三种请求方式可用
 * @param {Object} 用户配置，包含 axios 的配置与本系统自定义配置
 *
 * @return {Promise} 本次http请求的Promise
 */
async function getPromise(method: HttpMethodType, url: string, data: object | null, userConfig = {}) {
  const config = initConfig(method, url, userConfig);
  let promise;
  if (config.cancelPrevious) {
    await http.cancel(config.requestId);
  }

  if (config.clearCache) {
    http.cache.delete(config.requestId);
  } else {
    promise = http.cache.get(config.requestId);
  }
  if (config.fromCache && promise) {
    return promise;
  }

  promise = new Promise(async (resolve, reject) => {
    const axiosRequest = methodsWithData.includes(method)
      ? axiosInstance[method](url, data, config)
      : axiosInstance[method](url, config);

    try {
      const response = await axiosRequest;
      Object.assign(config, response.config || {});
      handleResponse({ config, response, resolve, reject });
    } catch (error: any) {
      Object.assign(config, error.config);
      reject(error);
    }
  })
    .catch((error) => {
      return handleReject(error, config);
    })
    .finally(() => {
      http.queue.delete(config.requestId);
    });

  // 添加请求队列
  http.queue.set(config);
  // 添加请求缓存
  http.cache.set(config.requestId, promise);

  return promise;
}

/**
 * 处理 http 请求成功结果
 *
 * @param {Object} 请求配置
 * @param {Object} cgi 原始返回数据
 * @param {Function} promise 完成函数
 * @param {Function} promise 拒绝函数
 */
function handleResponse(params: { config: any; response: any; resolve: any; reject: any }) {
  const { config, response, resolve, reject } = params;
  const transformedResponse = response.data;
  const { code, message, data } = transformedResponse;

  if (code !== 0 && config.globalError) {
    reject({ code, message });
    return;
  }
  if (config.originalResponse) {
    resolve(response);
    return;
  }
  if (config.transformData) {
    resolve(data);
    return;
  }

  resolve(transformedResponse);
}

/**
 * 处理 http 请求失败结果
 *
 * @param {Object} Error 对象
 * @param {config} 请求配置
 *
 * @return {Promise} promise 对象
 */
function handleReject(error: any, config: any) {
  if (error.code === TokenInvalidCode) {
    showLoginModal();
    return Promise.reject(error);
  }

  if (axios.isCancel(error)) {
    return Promise.reject(error);
  }
  if (error.code === 2000000 && error.message.includes('do not support create IPv6 full chain loadbalancer')) {
    Message({
      theme: 'error',
      message: '当前账号不支持购买IPv6，请联系云厂商开通, 参考文档https://cloud.tencent.com/document/product/214/39612',
    });
    return Promise.reject(
      '当前账号不支持购买IPv6，请联系云厂商开通, 参考文档https://cloud.tencent.com/document/product/214/39612',
    );
  }
  http.queue.delete(config.requestId);

  // 非2xx请求错误
  if (error.response) {
    const { status, data } = error.response;
    const nextError = { message: error.message, response: error.response };
    if (status === 401) {
      showLoginModal();
    } else if (status === 403) {
      bus.$emit('show-forbidden', error.response.data);
    } else if (status === 404) {
      nextError.message = '不存在';
      Message({ theme: 'error', message: nextError.message });
    } else if (status === 500) {
      nextError.message = '系统出现异常';
      Message({ theme: 'error', message: nextError.message });
    } else if (data?.message && error.code !== 0) {
      nextError.message = data.message;
      Message({ theme: 'error', message: nextError.message });
    } else {
      Message({ theme: 'error', message: error.message });
    }

    return Promise.reject(nextError);
  }

  // 有请求无响应
  if (error.request) {
    Message({ theme: 'error', message: 'Network Error' });
    return Promise.reject(error);
  }

  // 业务错误
  if (config.globalError) {
    handleCustomErrorCode(error);
  }

  return Promise.reject(error);
}

/**
 * 处理自定义错误码
 * @param error 异常
 */
function handleCustomErrorCode(error: any) {
  if (error.code === 2000014) {
    Message({ message: '当前负载均衡正在变更中，云平台限制新的任务同时变更。', theme: 'error' });
    return;
  }
  // zenlayer 账单导入错误码
  if ([2000015, 2000016, 2000017].includes(error.code)) return;

  if (error.code !== 0) Message({ theme: 'error', message: error.message });
}

/**
 * 初始化本系统 http 请求的各项配置
 *
 * @param {string} http method 与 axios 实例中的 method 保持一致
 * @param {string} 请求地址, 结合 method 生成 requestId
 * @param {Object} 用户配置，包含 axios 的配置与本系统自定义配置
 *
 * @return {Promise} 本次 http 请求的 Promise
 */
function initConfig(method: string, url: string, userConfig: object) {
  const defaultConfig = {
    ...getCancelToken(),
    // http 请求默认 id
    requestId: `${method}_${url}`,
    // 是否全局捕获异常
    globalError: true,
    // 是否直接复用缓存的请求
    fromCache: false,
    // 是否在请求发起前清楚缓存
    clearCache: false,
    // 响应结果是否返回原始数据
    originalResponse: false,
    // 转换返回数据，仅返回data对象
    transformData: false,
    // 当路由变更时取消请求
    cancelWhenRouteChange: true,
    // 取消上次请求
    cancelPrevious: false,
    // 是否使用全局headers
    globalHeaders: true,
  };
  return Object.assign(defaultConfig, userConfig);
}
type CombinedRequestConfig = Partial<AxiosRequestConfig & ReturnType<typeof initConfig>>;

/**
 * 生成 http 请求的 cancelToken，用于取消尚未完成的请求
 *
 * @return {Object} {cancelToken: axios 实例使用的 cancelToken, cancelExcutor: 取消http请求的可执行函数}
 */
function getCancelToken() {
  let cancelExcutor;
  const cancelToken = new axios.CancelToken((excutor) => {
    cancelExcutor = excutor;
  });
  return {
    cancelToken,
    cancelExcutor,
  };
}

/**
 * 向 http header 注入 CSRFToken，CSRFToken key 值与后端一起协商制定
 */
export function injectCSRFTokenToHeaders() {
  const CSRFToken = Cookies.get(`${window.PROJECT_CONFIG.BKPAAS_APP_ID}_csrftoken`);
  if (CSRFToken !== undefined) {
    axiosInstance.defaults.headers.common['X-CSRFToken'] = CSRFToken;
  } else {
    console.warn('Can not find csrftoken in document.cookie');
  }
  return CSRFToken;
}

export * from './jsonp';

export default http;
