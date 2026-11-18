/* eslint-disable @typescript-eslint/member-ordering */
import type { RouteLocationRaw } from 'vue-router';
import * as menuSymbols from '@/constants/menu-symbol';

const symbolRegistry = new Map<string, symbol>();
for (const value of Object.values(menuSymbols)) {
  if (typeof value === 'symbol' && value.description) {
    symbolRegistry.set(value.description, value);
  }
}

export class HistoryStorage {
  private static key = 'history';

  get history() {
    return HistoryStorage.get();
  }

  private static serialize(data: RouteLocationRaw): string {
    const serializable = typeof data === 'string' ? data : ({ ...data } as Record<string, any>);
    if (typeof serializable !== 'string' && typeof serializable.name === 'symbol') {
      serializable.name = serializable.name.description ?? '';
    }
    const bytes = new TextEncoder().encode(JSON.stringify(serializable));
    let binary = '';
    bytes.forEach((byte) => {
      binary += String.fromCharCode(byte);
    });
    return btoa(binary);
  }

  private static deserialize(encoded: string): RouteLocationRaw {
    const binary = atob(encoded);
    const bytes = Uint8Array.from(binary, (char) => char.charCodeAt(0));
    const data = JSON.parse(new TextDecoder().decode(bytes));
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

  // 删除最后一条记录（即后退的目标页）
  static removeLast() {
    const historyList = this.get();
    if (!historyList.length) return;
    historyList.pop();
    window.sessionStorage.setItem(this.key, JSON.stringify(historyList));
  }

  // 读取并删除最后一条记录。
  static pop(): RouteLocationRaw {
    const historyList = this.get();
    const record = historyList.pop();
    if (record === undefined) {
      throw new Error('history stack is empty');
    }
    const route = this.deserialize(record);
    window.sessionStorage.setItem(this.key, JSON.stringify(historyList));
    return route;
  }

  // 只读取最后一条记录，不会从栈中删除。
  static peek(): RouteLocationRaw | null {
    const historyList = this.get();
    const record = historyList[historyList.length - 1];
    if (!record) return null;
    try {
      return this.deserialize(record);
    } catch {
      return null;
    }
  }

  static clear() {
    window.sessionStorage.setItem(this.key, JSON.stringify([]));
  }
}
