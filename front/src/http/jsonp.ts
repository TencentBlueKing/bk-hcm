export const jsonp = <T>(
  url: string,
  data: object,
  options?: { callbackName?: string; version?: string },
): Promise<T> => {
  if (!url) throw new Error('invalid URL');

  const { callbackName = '', version = '' } = options || {};

  const callback = (callbackName || `CALLBACK${Math.random().toString().slice(9, 18)}`) as keyof Window;

  const JSONP = document.createElement('script');
  JSONP.setAttribute('type', 'text/javascript');

  const headEle = document.getElementsByTagName('head')[0];

  const t = version || Date.now();
  let query = `&_t=${t}`;
  if (data) {
    if (typeof data === 'string') {
      query = `&${data}`;
    } else if (typeof data === 'object') {
      for (const [key, value] of Object.entries(data)) {
        query += `&${key}=${encodeURIComponent(value)}`;
      }
    }
  }

  let promiseRejecter: (event: Event | string) => void;

  const rejectFn = (event: Event | string) => {
    promiseRejecter?.(event);
    window.removeEventListener('error', rejectFn);
  };

  JSONP.src = `${url}?callback=${callback}${query}`;
  JSONP.onerror = rejectFn;

  window.addEventListener('error', rejectFn);

  return new Promise((resolve, reject) => {
    promiseRejecter = reject;
    try {
      (window[callback] as unknown) = (result: T) => {
        resolve(result);
        headEle.removeChild(JSONP);
        delete window[callback];
      };
      headEle.appendChild(JSONP);
    } catch (err) {
      reject(err);
    }
  });
};
