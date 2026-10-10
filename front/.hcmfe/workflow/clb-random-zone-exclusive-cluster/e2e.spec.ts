/**
 * CLB 随机可用区与独占集群校验的本地 mock 回归。
 * 对应同目录 test.md 的 P0-01～P0-06、P1-01。
 */
import { expect, test, type Page } from '@playwright/test';
import { mockApis, type MockApisOptions } from '../../../e2e/mocks/api';
import { mockData } from '../../../e2e/mocks/data';

const PAGE_URL = '/#/business/load-balancer/apply?bizs=1';
const item = (page: Page, label: string) => page.getByText(label, { exact: true }).last().locator('../..');
const select = (page: Page, label: string) => item(page, label).locator('.bk-select').first();
const tagRequest = (page: Page) =>
  page.waitForRequest((request) => request.url().endsWith('/exclusive_clusters/tags/list'));

async function openPurchaseForm(page: Page, options?: MockApisOptions) {
  await mockApis(page, options);
  await page.goto(PAGE_URL);
  await page.getByRole('button', { name: '添加' }).click();
}

async function selectRegion(page: Page) {
  await select(page, '云地域').click();
  await page.getByText('测试地域', { exact: true }).last().click();
}

async function selectExclusive(page: Page) {
  await select(page, '负载均衡规格类型').click();
  await page.getByText('独占型', { exact: true }).last().click();
}

const clusterRow = (page: Page, label: string) =>
  page.getByText(label, { exact: true }).last().locator('..').locator('..');

test.describe('CLB - 独占集群查询与校验', () => {
  test('P0-01 随机可用区查询不传 zones 且可选独占型', async ({ page }) => {
    await openPurchaseForm(page);
    const requestPromise = tagRequest(page);
    await selectRegion(page);
    const request = await requestPromise;
    expect(request.postDataJSON()).not.toHaveProperty('zones');
    await select(page, '负载均衡规格类型').click();
    await expect(page.getByText('独占型', { exact: true }).last()).toBeVisible();
  });

  test('P0-02 指定可用区仍按 zones 数组查询', async ({ page }) => {
    await openPurchaseForm(page);
    await selectRegion(page);
    await item(page, '可用区').locator('.bk-select').last().click();
    const requestPromise = tagRequest(page);
    await page.locator('.bk-select-option-item:visible').filter({ hasText: '测试一区' }).click();
    expect((await requestPromise).postDataJSON()).toMatchObject({ zones: ['ap-test-1a'] });
  });

  test('P0-03 勾选集群后预选首个标签并保留随机分配', async ({ page }) => {
    await openPurchaseForm(page);
    await selectRegion(page);
    await selectExclusive(page);
    await page.getByText('四层集群', { exact: true }).last().click();
    await page.getByText('七层标签', { exact: true }).last().click();
    const l4Selects = clusterRow(page, '四层集群').locator('.bk-select input');
    await expect(l4Selects.nth(0)).toHaveValue('l4-test');
    await expect(l4Selects.nth(1)).toHaveValue('随机分配');
    await expect(l4Selects.nth(2)).toHaveValue('随机分配');
    await expect(clusterRow(page, '七层标签').locator('.bk-select input')).toHaveValue('l7-test');
  });

  test('P0-04 四层配置缺失时在规格类型下提示四层错误', async ({ page }) => {
    await openPurchaseForm(page, { clusterTags: mockData.clusterTags.filter((tag) => tag.cluster_type === 'STGW') });
    await selectRegion(page);
    await selectExclusive(page);
    await page.getByText('四层集群', { exact: true }).last().click();
    await expect(item(page, '负载均衡规格类型').getByText('请完整配置四层集群')).toBeVisible();
    await expect(item(page, '独占集群').getByText('请完整配置四层集群')).toHaveCount(0);
  });

  test('P0-04/P0-05 取消未配置的七层后及时清除七层错误', async ({ page }) => {
    await openPurchaseForm(page, { clusterTags: mockData.clusterTags.filter((tag) => tag.cluster_type === 'TGW') });
    await selectRegion(page);
    await selectExclusive(page);
    await page.getByText('四层集群', { exact: true }).last().click();
    await page.getByText('七层标签', { exact: true }).last().click();
    await expect(item(page, '负载均衡规格类型').getByText('请选择七层集群标签')).toBeVisible();
    await page.getByText('七层标签', { exact: true }).last().click();
    await expect(item(page, '负载均衡规格类型').getByText('请选择七层集群标签')).toHaveCount(0);
  });

  test('P0-06 切离独占型再返回时不保留复选框和标签', async ({ page }) => {
    await openPurchaseForm(page);
    await selectRegion(page);
    await selectExclusive(page);
    await page.getByText('四层集群', { exact: true }).last().click();
    await page.getByText('七层标签', { exact: true }).last().click();
    await select(page, '负载均衡规格类型').click();
    await page.getByText('共享型', { exact: true }).last().click();
    await selectExclusive(page);
    await expect(clusterRow(page, '四层集群').getByRole('checkbox')).not.toBeChecked();
    await expect(clusterRow(page, '七层标签').getByRole('checkbox')).not.toBeChecked();
  });

  test('P1-01 没有可用标签时独占型不可选', async ({ page }) => {
    await openPurchaseForm(page, { clusterTags: [] });
    await selectRegion(page);
    await select(page, '负载均衡规格类型').click();
    await expect(page.getByText('独占型', { exact: true })).toHaveCount(0);
  });

  test('P1-01 标签查询失败时独占型不可选', async ({ page }) => {
    await openPurchaseForm(page, { clusterTagsFail: true });
    await selectRegion(page);
    await select(page, '负载均衡规格类型').click();
    await expect(page.getByText('独占型', { exact: true })).toHaveCount(0);
  });

  test('P1-02 快速切换可用区后旧响应不回填', async ({ page }) => {
    await openPurchaseForm(page, { staleClusterTags: { stale: mockData.staleClusterTags, delayMs: 1500 } });
    await selectRegion(page);
    // 第一次查询（随机可用区）被 mock 扣住，立即切换到具体可用区触发第二次查询
    await item(page, '可用区').locator('.bk-select').last().click();
    await page.locator('.bk-select-option-item:visible').filter({ hasText: '测试一区' }).click();
    await selectExclusive(page);
    await page.getByText('四层集群', { exact: true }).last().click();
    const l4TagInput = clusterRow(page, '四层集群').locator('.bk-select input').nth(0);
    await expect(l4TagInput).toHaveValue('l4-test');
    // 等旧响应窗口过去后再核验终态：旧数据不回填、勾选不复活
    await page.waitForTimeout(1800);
    await expect(l4TagInput).toHaveValue('l4-test');
    await expect(clusterRow(page, '四层集群').getByRole('checkbox')).toBeChecked();
    await l4TagInput.click();
    await expect(page.locator('.bk-select-option-item:visible').filter({ hasText: 'l4-stale' })).toHaveCount(0);
  });

  async function selectConcreteL4Cluster(page: Page) {
    await clusterRow(page, '四层集群').locator('.bk-select').nth(1).click();
    await page.locator('.bk-select-option-item:visible').filter({ hasText: '测试四层集群' }).click();
  }

  test('P1-03 IP 延迟加载：加载中不误报，完成后可选', async ({ page }) => {
    await openPurchaseForm(page, { idleVips: ['10.0.0.1'], idleVipsDelayMs: 1200 });
    await selectRegion(page);
    await selectExclusive(page);
    await page.getByText('四层集群', { exact: true }).last().click();
    const vipsLoaded = page
      .waitForResponse((response) => response.url().endsWith('/exclusive_clusters/idle_vips/list'))
      .then(() => undefined);
    await selectConcreteL4Cluster(page);
    // 加载窗口内（慢响应未返回）：无失败/配置错误提示
    await expect(item(page, '负载均衡规格类型').getByText('独占集群信息获取失败，请重新选择或稍后重试')).toHaveCount(0);
    await expect(item(page, '负载均衡规格类型').getByText('请完整配置四层集群')).toHaveCount(0);
    // 加载完成后：IP 下拉出现具体 IP
    await vipsLoaded;
    await clusterRow(page, '四层集群').locator('.bk-select').nth(2).click();
    await expect(page.locator('.bk-select-option-item:visible').filter({ hasText: '10.0.0.1' })).toBeVisible();
  });

  test('P1-03 集群无空闲 IP 时提示并清空选择', async ({ page }) => {
    await openPurchaseForm(page, { idleVips: [] });
    await selectRegion(page);
    await selectExclusive(page);
    await page.getByText('四层集群', { exact: true }).last().click();
    await selectConcreteL4Cluster(page);
    await expect(page.getByText('该集群暂无空闲 IP，请选择其它集群')).toBeVisible();
    await expect(clusterRow(page, '四层集群').locator('.bk-select input').nth(1)).toHaveValue('');
  });

  test('P1-03 IP 请求失败时在规格类型下提示获取失败', async ({ page }) => {
    await openPurchaseForm(page, { idleVipsFail: true });
    await selectRegion(page);
    await selectExclusive(page);
    await page.getByText('四层集群', { exact: true }).last().click();
    await selectConcreteL4Cluster(page);
    await expect(item(page, '负载均衡规格类型').getByText('独占集群信息获取失败，请重新选择或稍后重试')).toBeVisible();
  });
});
