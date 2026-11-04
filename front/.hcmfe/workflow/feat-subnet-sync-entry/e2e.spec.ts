/**
 * feat-subnet-sync-entry（子网-增加独立的子网同步入口）E2E 用例
 *
 * 对应手测清单：同目录 test.md
 * 运行方式（本地 mock，不依赖真实环境）：
 *   npm run e2e -- .hcmfe/workflow/feat-subnet-sync-entry/e2e.spec.ts
 */
import { expect, test, type Page } from '@playwright/test';

import { mockApis } from '../../../e2e/mocks/api';
import { mockSubnetSyncAccounts } from '../../../e2e/mocks/data';

const BUSINESS_SUBNET_URL = '/#/business/subnet?bizs=1';
const RESOURCE_SUBNET_URL = '/#/resource/resource?type=subnet&accountId=account-e2e';

const syncButton = (page: Page) => page.getByRole('button', { name: '同步子网' });
const subnetRows = (page: Page) => page.locator('.bk-table-body tr');

test.describe('子网同步入口 - 业务视角', () => {
  test.beforeEach(async ({ page }) => {
    await mockApis(page);
    await page.goto(BUSINESS_SUBNET_URL);
  });

  test('P0-01 业务子网列表在申请按钮旁展示同步入口', async ({ page }) => {
    await expect(page.getByRole('button', { name: '申请' })).toBeVisible();
    await expect(syncButton(page)).toBeVisible();
    await expect(syncButton(page)).toBeEnabled();
  });

  test('P0-03 点击同步入口打开子网同步弹窗', async ({ page }) => {
    await syncButton(page).click();

    await expect(page.getByText('从云上同步该业务的子网数据，不会同步 VPC、负载均衡等其他资源')).toBeVisible();
    await expect(page.getByText('云账号', { exact: true })).toBeVisible();
    await expect(page.getByText('云地域', { exact: true })).toBeVisible();
    await expect(page.getByRole('button', { name: '确定' })).toBeVisible();
  });

  test('P0-04 选择地域后提交子网条件同步', async ({ page }) => {
    await syncButton(page).click();

    const regionItem = page.locator('.bk-form-item').filter({ hasText: '云地域' });
    await regionItem.locator('.bk-select').click();
    await page.locator('.bk-select-option').filter({ hasText: '测试地域' }).first().click();

    const responsePromise = page.waitForResponse(
      (response) => response.request().method() === 'POST' && response.url().includes('/resources/subnet/sync_by_cond'),
    );
    await page.getByRole('button', { name: '确定' }).click();
    const response = await responsePromise;

    expect(response.request().postDataJSON()).toMatchObject({ regions: ['ap-test-1'] });
    await expect(page.getByText('已同步成功')).toBeVisible();
  });

  test('P0-06 勾选子网后条件同步入口不可用', async ({ page }) => {
    await expect(subnetRows(page)).toHaveCount(2);
    await subnetRows(page).first().locator('.bk-checkbox').click();

    await expect(syncButton(page)).toBeDisabled();
  });
});

test.describe('子网同步入口 - 异常分支', () => {
  test('P1-01 后端未注册 subnet 时展示错误信息', async ({ page }) => {
    await page.unrouteAll({ behavior: 'ignoreErrors' });
    await mockApis(page, { subnetSyncFail: true });
    await page.goto(BUSINESS_SUBNET_URL);
    await syncButton(page).click();

    const regionItem = page.locator('.bk-form-item').filter({ hasText: '云地域' });
    await regionItem.locator('.bk-select').click();
    await page.locator('.bk-select-option').filter({ hasText: '测试地域' }).first().click();
    await page.getByRole('button', { name: '确定' }).click();

    await expect(page.getByText('conditional sync resource does not support subnet')).toBeVisible();
  });
});

test.describe('子网同步入口 - 资源账号视角', () => {
  test('P0-02 资源视角未选账号时同步入口仍可用', async ({ page }) => {
    await mockApis(page);
    await page.goto(RESOURCE_SUBNET_URL);

    await expect(page.getByText('子网', { exact: true }).first()).toBeVisible();
    await expect(syncButton(page)).toBeEnabled();
    await syncButton(page).click();

    await expect(page.getByText('从云上同步所选账号的子网数据，不会同步 VPC、负载均衡等其他资源')).toBeVisible();
    await expect(page.getByText('云账号', { exact: true })).toBeVisible();
  });
});

test.describe('外部子网同步 - 云厂商范围', () => {
  for (const { title, url } of [
    { title: 'P1-04 业务视角排除内部厂商并保留外部多云账号', url: BUSINESS_SUBNET_URL },
    { title: 'P1-05 资源全部账号视角排除内部厂商并保留外部多云账号', url: '/#/resource/manage/subnet' },
  ]) {
    test(title, async ({ page }) => {
      await mockApis(page, { accounts: mockSubnetSyncAccounts });
      await page.goto(url);
      await syncButton(page).click();
      const accountItem = page.locator('.bk-form-item').filter({ hasText: '云账号' });
      await accountItem.locator('.bk-select-trigger').click();

      // AccountSelector 的列表项和厂商分组，确保被隐藏的账号也不参与计数。
      const options = page.locator('.account-option');
      await expect(options).toHaveCount(5);
      await expect(page.locator('.vendor-list .vendor-item')).toHaveCount(6);
      await expect(options.filter({ hasText: '内部厂商测试账号' })).toHaveCount(0);
      for (const vendor of ['tcloud', 'aws', 'azure', 'gcp', 'huawei']) {
        await expect(options.filter({ hasText: `外部-${vendor}-账号` })).toHaveCount(1);
      }
    });
  }
});
