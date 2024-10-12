/* eslint-disable @typescript-eslint/member-ordering */
import type { RouteLocationRaw, RouteRecordNameGeneric } from 'vue-router';

export class HistoryStorage {
  private static key = 'history';

  get history() {
    return HistoryStorage.get();
  }

  static get() {
    let historyList = [];
    try {
      historyList = JSON.parse(window.sessionStorage.getItem(this.key)) || [];
      if (!Array.isArray(historyList)) {
        historyList = [historyList];
      }
    } catch (e) {
      historyList = [];
    }
    return historyList;
  }

  static append(data: RouteLocationRaw) {
    const base64 = btoa(JSON.stringify(data));
    const historyList = this.get();
    historyList.push(base64);
    window.sessionStorage.setItem(this.key, JSON.stringify(historyList));
  }

  static remove(name: RouteRecordNameGeneric) {
    const historyList = this.get();
    const index = historyList.findIndex((item) => {
      const history = JSON.parse(atob(item));
      return history.name === name;
    });
    if (index !== -1) {
      historyList.splice(index, 1);
      window.sessionStorage.setItem(this.key, JSON.stringify(historyList));
    }
  }

  static pop(): RouteLocationRaw {
    const historyList = this.get();
    const record = historyList.pop();
    const route = JSON.parse(atob(record));
    return route;
  }

  static clear() {
    window.sessionStorage.setItem(this.key, JSON.stringify([]));
  }
}
