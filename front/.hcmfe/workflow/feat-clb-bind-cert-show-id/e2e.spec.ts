/**
 * feat-clb-bind-cert-show-id（负载均衡-绑定证书-增加证书ID的展示）E2E 用例
 *
 * 对应手测清单：同目录 test.md（P0 主流程 / P1 异常分支）
 * 运行方式（本地 mock，不依赖真实环境）：
 *   npm run e2e -- .hcmfe/workflow/feat-clb-bind-cert-show-id/e2e.spec.ts
 */
import { expect, test, type Locator, type Page } from '@playwright/test';

import { mockApis } from '../../../e2e/mocks/api';

const LISTENER_PAGE_URL = '/#/business/load-balancer/resource/clb/details/clb-e2e?bizs=1';
const CERT_PAGE_URL = '/#/business/cert?bizs=1';
const LONG_CERT_NAME = 'ieg-shared-certificate-with-a-very-long-remark-for-e2e';
const HIDDEN_CERT_ID = 'CERT-HIDDEN-999';

const certSelector = (page: Page) => page.locator('.cert-selector');
const certOptions = (page: Page) => page.locator('.cert-selector-option');
const certSearchInput = (page: Page) => page.getByPlaceholder('请输入证书ID/名称');

async function openListenerCertificateSelector(page: Page): Promise<Locator> {
  await page.getByRole('button', { name: '新增监听器' }).click();
  const httpsRadio = page.locator('.bk-radio-group').getByText('HTTPS', { exact: true });
  await expect(httpsRadio).toBeVisible();
  await httpsRadio.click();

  const selector = certSelector(page);
  await expect(selector).toBeVisible();
  await selector.locator('.bk-select-trigger').click();
  await expect(certOptions(page).first()).toBeVisible();
  return selector;
}

async function searchCertificates(page: Page, keyword: string): Promise<void> {
  await Promise.all([
    page.waitForResponse(
      (response) => response.request().method() === 'POST' && response.url().includes('/certs/list'),
    ),
    certSearchInput(page).fill(keyword),
  ]);
}

test.describe('负载均衡 - 证书选择器', () => {
  test.beforeEach(async ({ page }) => {
    await mockApis(page);
    await page.goto(LISTENER_PAGE_URL);
  });

  test('P0-01 证书下拉以证书 ID 为主并展示算法、域名和到期时间', async ({ page }) => {
    await openListenerCertificateSelector(page);

    const firstOption = certOptions(page).first();
    await expect(firstOption).toContainText('CERT-VISIBLE-000');
    await expect(firstOption).toContainText('RSA 2048');
    await expect(firstOption).toContainText(LONG_CERT_NAME);
    await expect(firstOption).toContainText('主域名 e2e.example.com');
    await expect(firstOption).toContainText('到期 2027-04-29');
  });

  test('P0-02 按证书 ID 可检索到未加载分页中的证书', async ({ page }) => {
    await openListenerCertificateSelector(page);
    await searchCertificates(page, 'hidden-999');

    const hiddenOption = page.locator('.bk-select-option').filter({ hasText: HIDDEN_CERT_ID });
    await expect(hiddenOption).toBeVisible();
    await expect(hiddenOption).toContainText('ieg 目标证书');
  });

  test('P0-03 按名称检索大小写不敏感', async ({ page }) => {
    await openListenerCertificateSelector(page);
    await searchCertificates(page, 'IEG 目标证书');

    await expect(page.locator('.bk-select-option').filter({ hasText: HIDDEN_CERT_ID })).toBeVisible();
  });

  test('P0-04 选中证书后回显完整证书 ID 和名称', async ({ page }) => {
    const selector = await openListenerCertificateSelector(page);
    await searchCertificates(page, HIDDEN_CERT_ID);

    await page.locator('.bk-select-option').filter({ hasText: HIDDEN_CERT_ID }).click();
    await page.keyboard.press('Escape');

    await expect(selector).toContainText(HIDDEN_CERT_ID);
    await expect(selector).toContainText('ieg 目标证书');
  });

  test('P1-01 已过期证书展示过期标识且不可选', async ({ page }) => {
    await openListenerCertificateSelector(page);

    const expiredOption = page.locator('.bk-select-option').filter({ hasText: 'CERT-VISIBLE-002' });
    await expect(expiredOption).toContainText('已过期');
    await expect(expiredOption).toHaveClass(/is-disabled/);
  });

  test('P1-04 无匹配证书时展示空态，清空后恢复首屏', async ({ page }) => {
    await openListenerCertificateSelector(page);
    await searchCertificates(page, 'NO-SUCH-CERT');

    await expect(page.getByText('无匹配证书').first()).toBeVisible();
    await searchCertificates(page, '');
    await expect(certOptions(page).first()).toContainText('CERT-VISIBLE-000');
  });
});

test.describe('证书管理 - 证书 ID 一致性', () => {
  test.beforeEach(async ({ page }) => {
    await mockApis(page);
    await page.goto(CERT_PAGE_URL);
  });

  test('P0-06 证书列、删除确认使用证书 ID（名称）识别证书', async ({ page }) => {
    await expect(page.getByRole('columnheader', { name: '证书ID' })).toBeVisible();

    const certRow = page.locator('.bk-table-body tr').filter({ hasText: 'CERT-VISIBLE-000' });
    await expect(certRow).toContainText(LONG_CERT_NAME);
    await certRow.getByRole('button', { name: '删除' }).click();

    await expect(page.getByText(/将删除证书【CERT-VISIBLE-000（/)).toBeVisible();
  });
});
