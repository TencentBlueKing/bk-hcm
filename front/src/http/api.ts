import fetch from './fetch';

export default {
  getApiDemo(params) {
    return fetch.get('/api/demo', params);
  }
};
