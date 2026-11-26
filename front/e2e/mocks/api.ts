import type { Page } from '@playwright/test';
import { mockData } from './data';

function ok(data: unknown): string {
  return JSON.stringify({ code: 0, message: 'ok', result: true, data });
}

const sleep = (ms: number) => new Promise((resolve) => setTimeout(resolve, ms));

export interface MockApisOptions {
  clusterTags?: typeof mockData.clusterTags;
  /** tags/list 直接返回 HTTP 500，模拟请求失败 */
  clusterTagsFail?: boolean;
  /** 首次 tags 请求延迟 delayMs 后才返回 stale 数据，用于竞态回归 */
  staleClusterTags?: { stale: typeof mockData.clusterTags; delayMs: number };
  /** idle_vips/list 返回的空闲 IP 列表（默认空 = 无空闲 IP） */
  idleVips?: string[];
  /** idle_vips/list 直接返回 HTTP 500 */
  idleVipsFail?: boolean;
  /** idle_vips/list 响应延迟，模拟慢接口 */
  idleVipsDelayMs?: number;
  /** 证书列表数据，默认使用共享 mock */
  certificates?: typeof mockData.certificates;
}

const matchCertificateRule = (cert: (typeof mockData.certificates)[number], rule: any): boolean => {
  if (rule?.op === 'or') return (rule.rules ?? []).some((item: any) => matchCertificateRule(cert, item));
  const value = cert[rule.field as keyof typeof cert];
  if (rule.op === 'eq') return value === rule.value;
  if (rule.op === 'in') return Array.isArray(rule.value) && rule.value.includes(value);
  if (rule.op === 'cs' || rule.op === 'cis') {
    const source = String(value ?? '');
    const keyword = String(rule.value ?? '');
    return rule.op === 'cis' ? source.toLowerCase().includes(keyword.toLowerCase()) : source.includes(keyword);
  }
  return true;
};

const filterCertificates = (certificates: typeof mockData.certificates, filter: any) =>
  (filter?.rules ?? []).reduce(
    (result: typeof mockData.certificates, rule: any) => result.filter((cert) => matchCertificateRule(cert, rule)),
    certificates,
  );

export async function mockApis(page: Page, options: MockApisOptions = {}): Promise<void> {
  await page.addInitScript(() => localStorage.setItem('bk-hcm-release-note-version', 'e2e'));
  let tagsCallCount = 0;
  await page.route('**/api/**', async (route) => {
    const { pathname } = new URL(route.request().url());
    const requestBody = route.request().postDataJSON();
    let data: unknown = { details: [], count: 0 };
    if (pathname === '/api/v1/web/auth/verify') {
      const resources = (route.request().postDataJSON() as { resources?: unknown[] })?.resources ?? [];
      data = { results: resources.map(() => ({ authorized: true })) };
    } else if (pathname === '/api/v1/web/users') {
      data = { username: 'test-user' };
    } else if (pathname === '/api/v1/web/bk_bizs/list' || pathname === '/api/v1/web/authorized/bizs/list') {
      data = mockData.businesses;
    } else if (pathname === '/api/v1/web/changelogs') {
      data = [{ version: 'e2e', time: '2026-01-01', is_current: true }];
    } else if (/\/api\/v1\/cloud\/accounts\/bizs\/\d+$/.test(pathname)) {
      data = mockData.accounts;
    } else if (/\/api\/v1\/cloud\/bizs\/\d+\/collections\/bizs$/.test(pathname)) {
      data = [];
    } else if (pathname.endsWith('/network_type')) {
      data = { NetworkAccountType: 'STANDARD' };
    } else if (pathname.endsWith('/regions/list')) {
      data = pathname.includes('/tcloud/') ? { details: [mockData.region], count: 1 } : { details: [], count: 0 };
    } else if (pathname.endsWith('/zones/list')) {
      data = { details: [mockData.zone], count: 1 };
    } else if (pathname.endsWith('/resources/describe')) {
      data = mockData.resource;
    } else if (pathname.endsWith('/load_balancers/quotas')) {
      data = [];
    } else if (pathname.endsWith('/exclusive_clusters/tags/list')) {
      const callIndex = tagsCallCount;
      tagsCallCount += 1;
      if (options.clusterTagsFail) {
        return route.fulfill({ status: 500, contentType: 'application/json', body: 'mock tags failure' });
      }
      if (options.staleClusterTags && callIndex === 0) {
        await sleep(options.staleClusterTags.delayMs);
        return route.fulfill({
          contentType: 'application/json',
          body: ok({ details: options.staleClusterTags.stale }),
        });
      }
      data = { details: options.clusterTags ?? mockData.clusterTags };
    } else if (pathname.endsWith('/exclusive_clusters/idle_vips/list')) {
      if (options.idleVipsDelayMs) await sleep(options.idleVipsDelayMs);
      if (options.idleVipsFail) {
        return route.fulfill({ status: 500, contentType: 'application/json', body: 'mock idle vips failure' });
      }
      const vips = options.idleVips ?? [];
      data = { details: vips, count: vips.length };
    } else if (pathname.endsWith('/bandwidth_packages/query')) {
      data = { packages: [], total_count: 0 };
    } else if (pathname.endsWith('/certs/list')) {
      const certificates = filterCertificates(options.certificates ?? mockData.certificates, requestBody?.filter);
      const page = requestBody?.page ?? {};
      if (page.count === true) {
        data = { details: [], count: certificates.length };
      } else {
        const start = Number(page.start) || 0;
        const limit = Number(page.limit) || certificates.length;
        data = { details: certificates.slice(start, start + limit), count: certificates.length };
      }
    } else if (/\/load_balancers\/clb-e2e$/.test(pathname)) {
      data = mockData.clbDetails;
    }
    return route.fulfill({
      contentType: 'application/json',
      body: ok(data),
    });
  });
}

export async function mockBootApis(page: Page): Promise<void> {
  await mockApis(page);
}
