/* eslint-disable @typescript-eslint/member-ordering */
import type { RouteLocationRaw, RouteRecordName } from 'vue-router';
import * as menuSymbols from '@/constants/menu-symbol';

const symbolRegistry = new Map<string, symbol>();
for (const value of Object.values(menuSymbols)) {
  if (typeof value === 'symbol') {
    symbolRegistry.set(value.description!, value);
  }
}

export class HistoryStorage {
  private static key = 'history';

  get history() {
    return HistoryStorage.get();
  }

  private static serialize(data: RouteLocationRaw): string {
    if (typeof data === 'string') return btoa(JSON.stringify(data));
    const plain = { ...data } as Record<string, any>;
    if (typeof plain.name === 'symbol') {
      plain.name = plain.name.description;
    }
    return btoa(JSON.stringify(plain));
  }

  private static deserialize(encoded: string): RouteLocationRaw {
    const data = JSON.parse(atob(encoded));
    if (typeof data === 'string') return data;
    if (typeof data.name === 'string' && symbolRegistry.has(data.name)) {
      data.name = symbolRegistry.get(data.name);
    }
    return data;
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
    const historyList = this.get();
    historyList.push(this.serialize(data));
    window.sessionStorage.setItem(this.key, JSON.stringify(historyList));
  }

  static remove(name: RouteRecordName) {
    const historyList = this.get();
    const index = historyList.findIndex((item) => {
      const history = this.deserialize(item) as Record<string, any>;
      return history?.name === name;
    });
    if (index !== -1) {
      historyList.splice(index, 1);
      window.sessionStorage.setItem(this.key, JSON.stringify(historyList));
    }
  }

  static pop(): RouteLocationRaw {
    const historyList = this.get();
    const record = historyList.pop();
    return this.deserialize(record);
  }

  static clear() {
    window.sessionStorage.setItem(this.key, JSON.stringify([]));
  }
}
