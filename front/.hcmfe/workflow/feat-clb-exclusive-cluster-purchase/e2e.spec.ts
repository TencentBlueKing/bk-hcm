/**
 * feat-clb-exclusive-cluster-purchase（CLB 独占集群购买、单据与详情适配）E2E 用例
 *
 * 对应手测清单：同目录 test.md（P0 主流程 / P1 异常分支）
 * 运行方式（全部接口走 mock，不依赖真实环境与真实账号）：
 *   npm run e2e -- .hcmfe/workflow/feat-clb-exclusive-cluster-purchase/e2e.spec.ts
 */
import { expect, test, type Locator, type Page } from '@playwright/test';

import { mockApis, type MockApiOptions } from '../../../e2e/mocks/api';
import * as mockData from '../../../e2e/mocks/data';

// ---------- 页面入口 ----------
// 只写业务路径，域名由 playwright.config.ts 的 baseURL 提供
const ticketDetailUrl = (id: string) => `/#/business/ticket/detail?bizs=${mockData.TEST_BIZS}&id=${id}`;
const lbDetailUrl = (id: string) =>
  `/#/business/load-balancer/resource/clb/details/${id}?bizs=${mockData.TEST_BIZS}&details_active=info`;
const applyClbUrl = `/#/business/load-balancer/apply?bizs=${mockData.TEST_BIZS}`;

// ---------- 选择器 ----------
// 单据详情：grid-item(label + item-content)；CLB 详情：li.info-list-item(label + item-value)
const gridValue = (page: Page, label: string): Locator =>
  page
    .locator('.grid-item')
    .filter({ has: page.getByText(label, { exact: true }) })
    .locator('.item-content');

const infoValue = (page: Page, label: string): Locator =>
  page
    .locator('li.info-list-item')
    .filter({ has: page.getByText(label, { exact: true }) })
    .locator('.item-value');

// 购买页：表单行按可见 label 定位，下拉项用 .bk-select-option（浏览器内保持稳定）
const formItem = (page: Page, label: string): Locator =>
  page.locator('.bk-form-item').filter({ has: page.getByText(label, { exact: true }) });

const configCard = (page: Page): Locator => page.locator('.configure-card-container');

const pickOption = async (page: Page, text: string | RegExp): Promise<void> => {
  await page.locator('.bk-select-option').filter({ hasText: text }).first().click();
};

const visibleOptions = (page: Page): Promise<string[]> => page.locator('.bk-select-option:visible').allInnerTexts();

// ============================================================
// 单据 1069995598138114062：单据展示独占集群参数
// ============================================================
test.describe('负载均衡申请单详情 - 独占集群参数', () => {
  test('P0-07 独占型申请单在规格类型之后展示申请时的四层与七层参数', async ({ page }) => {
    await mockApis(page);
    await page.goto(ticketDetailUrl(mockData.TEST_APP_ID));

    await expect(page.getByText('负载均衡申请单详情')).toBeVisible();
    await expect(gridValue(page, '负载均衡规格类型')).toHaveText('独占型');
    await expect(gridValue(page, '四层集群标签')).toHaveText(mockData.L4_TAG);
    await expect(gridValue(page, '四层集群名称')).toHaveText(mockData.L4_CLUSTER_NAME);
    await expect(gridValue(page, '四层集群 IP')).toHaveText(mockData.TEST_LB_VIP);
    await expect(gridValue(page, '七层集群标签')).toHaveText(mockData.L7_TAG_WITH_EGRESS);

    // 阅读顺序：独占区块排在「负载均衡规格类型」之后（两列网格同一行时比较横坐标）
    const specBox = await gridValue(page, '负载均衡规格类型').boundingBox();
    const l4Box = await gridValue(page, '四层集群标签').boundingBox();
    const sameRow = Math.abs(l4Box.y - specBox.y) < 1;
    expect(l4Box.y > specBox.y || (sameRow && l4Box.x > specBox.x)).toBe(true);
  });

  test('P1-11 申请单 content 解析失败时不白屏，其它参数正常展示', async ({ page }) => {
    await mockApis(page, { applicationDetail: mockData.mockApplicationDetails.broken });
    await page.goto(ticketDetailUrl(mockData.TEST_APP_ID));

    await expect(page.getByText('负载均衡申请单详情')).toBeVisible();
    await expect(page.getByText('参数信息')).toBeVisible();
    await expect(gridValue(page, '云厂商')).toBeVisible();
    await expect(gridValue(page, '四层集群标签')).toHaveCount(0);
  });

  test('P1-12 随机分配时四层集群名称与 IP 展示空值', async ({ page }) => {
    await mockApis(page, { applicationDetail: mockData.mockApplicationDetails.random });
    await page.goto(ticketDetailUrl(mockData.TEST_APP_ID));

    await expect(gridValue(page, '四层集群名称')).toHaveText('--');
    await expect(gridValue(page, '四层集群 IP')).toHaveText('--');
    await expect(gridValue(page, '四层集群标签')).toHaveText(mockData.L4_TAG);
  });
});

// ============================================================
// 单据 1069995598138114066：CLB 详情独占集群字段
// ============================================================
test.describe('CLB 资源详情 - 配置信息独占字段', () => {
  test('P0-08 独占型 CLB 展示实例规格与四层/七层集群字段，四层 IP 与负载均衡 VIP 一致', async ({ page }) => {
    await mockApis(page);
    await page.goto(lbDetailUrl(mockData.TEST_LB_ID));

    await expect(infoValue(page, '实例规格')).toHaveText('独占型');
    await expect(infoValue(page, '四层集群标签')).toHaveText(mockData.L4_TAG);
    await expect(infoValue(page, '四层集群名称')).toHaveText(mockData.L4_CLUSTER_NAME);
    await expect(infoValue(page, '七层集群标签')).toHaveText(mockData.L7_TAG_WITH_EGRESS);

    const vip = await infoValue(page, '负载均衡VIP').innerText();
    const l4Ip = await infoValue(page, '四层集群 IP').innerText();
    expect(l4Ip).toBe(vip);
    expect(l4Ip).toBe(mockData.TEST_LB_VIP);
  });

  test('P1-12 独占型但集群字段缺失时展示空值，且不阻断其它配置信息', async ({ page }) => {
    await mockApis(page, { lbDetails: mockData.mockLbDetails.exclusiveMissingFields });
    await page.goto(lbDetailUrl(mockData.TEST_LB_ID));

    await expect(infoValue(page, '实例规格')).toHaveText('独占型');
    await expect(infoValue(page, '四层集群标签')).toHaveText('--');
    await expect(infoValue(page, '四层集群名称')).toHaveText('--');
    await expect(infoValue(page, '四层集群 IP')).toHaveText('--');
    await expect(infoValue(page, '带宽计费模式')).toBeVisible();
  });

  test('P0-09 共享型不渲染独占集群字段且规格显示共享型', async ({ page }) => {
    await mockApis(page, { lbDetails: mockData.mockLbDetails.shared });
    await page.goto(lbDetailUrl(mockData.TEST_LB_ID));

    await expect(infoValue(page, '实例规格')).toHaveText('共享型');
    await expect(page.getByText('四层集群标签')).toHaveCount(0);
    await expect(page.getByText('七层集群标签')).toHaveCount(0);
  });

  test('P0-09 性能容量型展示档位且不渲染独占集群字段', async ({ page }) => {
    await mockApis(page, { lbDetails: mockData.mockLbDetails.performance });
    await page.goto(lbDetailUrl(mockData.TEST_LB_ID));

    await expect(infoValue(page, '实例规格')).toHaveText('超强型2规格');
    await expect(page.getByText('四层集群标签')).toHaveCount(0);
  });
});

// ============================================================
// 单据 1069995598138114049：购买支持独占型规格
// ============================================================
test.describe('购买负载均衡 - 独占型规格', () => {
  const capturedApplies: Array<Record<string, any>> = [];

  test.beforeEach(() => {
    capturedApplies.length = 0;
  });

  /** 打开购买页 -> 侧滑表单 -> 填好地域 / VPC / 可用区 */
  async function openApplyForm(page: Page, opt: MockApiOptions = {}): Promise<void> {
    await mockApis(page, {
      ...opt,
      override: (ctx) => {
        if (ctx.pathname.endsWith('/applications/types/create_load_balancer')) {
          capturedApplies.push(ctx.body);
          return {};
        }
        return opt.override?.(ctx);
      },
    });
    await page.goto(applyClbUrl);
    await expect(page.locator('.account-selector')).toBeVisible();
    await page.getByRole('button', { name: '添加' }).click();
    await expect(page.getByText('添加负载均衡')).toBeVisible();

    await formItem(page, '云地域').locator('.bk-select').first().click();
    await pickOption(page, /^广州$/);
    await expect(formItem(page, 'VPC').locator('.bk-select').first()).toBeEnabled();
    await page.waitForTimeout(600);
    await formItem(page, 'VPC').locator('.bk-select').first().click();
    await pickOption(page, mockData.TEST_VPC_ID);
    await page.waitForTimeout(600);
    await formItem(page, '可用区').locator('.bk-select').nth(1).click();
    await pickOption(page, mockData.TEST_ZONE_NAME);
    await expect(formItem(page, '运营商类型').getByText('BGP')).toBeVisible();
  }

  /** 选择独占型规格 */
  async function chooseExclusiveSpec(page: Page): Promise<void> {
    await formItem(page, '负载均衡规格类型').locator('.bk-select').first().click();
    await page.locator('.bk-select-option').filter({ hasText: /^独占型$/ }).first().click();
    await expect(page.getByText('独占集群', { exact: true })).toBeVisible();
  }

  /** 勾选四层并选标签，集群保持默认「随机分配」 */
  async function enableL4Random(page: Page): Promise<void> {
    await page.getByText('四层集群', { exact: true }).click();
    await formItem(page, '独占集群').locator('.bk-select').first().click();
    await pickOption(page, /^L4-TAG-A$/);
    await expect(formItem(page, '独占集群').locator('input').nth(2)).toHaveValue('随机分配');
  }

  /** 勾选七层并选标签 */
  async function enableL7(page: Page, tag: RegExp = /^L7-TAG-A$/): Promise<void> {
    await page.getByText('七层标签', { exact: true }).click();
    await formItem(page, '独占集群').locator('.bk-select').nth(3).click();
    await pickOption(page, tag);
  }

  /** 独占集群之外的必填项 */
  async function fillRequiredFields(page: Page): Promise<void> {
    await formItem(page, '安全组放通模式').locator('.bk-select').first().click();
    await pickOption(page, /启用默认放通/);
    await formItem(page, '实例名称').locator('input').first().fill('clb-e2e-01');
    await formItem(page, '带宽上限（Mbps）').locator('input').first().fill('100');
    await page.keyboard.press('Enter');
  }

  /** 保存 -> 配置清单新增一行 -> 提交 -> 拿到请求体 */
  async function saveAndSubmit(page: Page): Promise<Record<string, any>> {
    await page.getByRole('button', { name: '保存' }).click();
    await expect(configCard(page).getByText('暂无数据')).toHaveCount(0);
    await page.getByRole('button', { name: '提交' }).click();
    await expect.poll(() => capturedApplies.length).toBe(1);
    return capturedApplies[0];
  }

  test('P0-01 具备可用独占集群时规格类型可选独占型，并展示四层/七层启用项', async ({ page }) => {
    await openApplyForm(page);

    await formItem(page, '负载均衡规格类型').locator('.bk-select').first().click();
    const exclusiveOption = page.locator('.bk-select-option').filter({ hasText: /^独占型$/ }).first();
    await expect(exclusiveOption).toBeVisible();
    await exclusiveOption.click();

    await expect(page.getByText('独占集群', { exact: true })).toBeVisible();
    await expect(page.getByText('四层集群', { exact: true })).toBeVisible();
    await expect(page.getByText('七层标签', { exact: true })).toBeVisible();
  });

  test('P1-01 两层均未启用时保存被拦截，配置清单不新增', async ({ page }) => {
    await openApplyForm(page);
    await chooseExclusiveSpec(page);

    await page.getByRole('button', { name: '保存' }).click();

    // 校验拦截：给出独占集群的校验文案，侧滑不关闭、配置清单仍为空
    await expect(page.getByText('请至少启用并完整配置一个独占集群')).toBeVisible();
    await expect(page.getByText('添加负载均衡')).toBeVisible();
    await expect(configCard(page).getByText('暂无数据')).toBeVisible();
  });

  test('P1-02 启用层配置不完整时保存被拦截', async ({ page }) => {
    await openApplyForm(page);
    await chooseExclusiveSpec(page);
    // 勾选四层但不选标签 / 集群
    await page.getByText('四层集群', { exact: true }).click();

    await page.getByRole('button', { name: '保存' }).click();

    await expect(page.getByText('请至少启用并完整配置一个独占集群')).toBeVisible();
    await expect(configCard(page).getByText('暂无数据')).toBeVisible();
  });

  test('P0-02 四层随机集群可保存并提交标签下全部候选集群与空 VIP', async ({ page }) => {
    await openApplyForm(page);
    await chooseExclusiveSpec(page);
    await enableL4Random(page);
    await fillRequiredFields(page);

    const payload = await saveAndSubmit(page);
    expect(payload.exclusive).toBe(1);
    expect(payload.sla_type).toBe('');
    expect(payload.cluster_tag).toBeUndefined();
    expect(payload.vip).toBe('');
    expect([...payload.cloud_cluster_ids].sort()).toEqual(
      [mockData.L4_CLUSTER_CLOUD_ID, mockData.L4_CLUSTER_B_CLOUD_ID, mockData.L4_CLUSTER_NO_VIP_ID].sort(),
    );
  });

  test('P0-03 四层指定集群与指定 IP 时可提交该集群与该 IP', async ({ page }) => {
    await openApplyForm(page);
    await chooseExclusiveSpec(page);
    await page.getByText('四层集群', { exact: true }).click();
    await formItem(page, '独占集群').locator('.bk-select').first().click();
    await pickOption(page, /^L4-TAG-A$/);
    await formItem(page, '独占集群').locator('.bk-select').nth(1).click();
    await pickOption(page, new RegExp(`^${mockData.L4_CLUSTER_NAME}$`));
    await formItem(page, '独占集群').locator('.bk-select').nth(2).click();
    await pickOption(page, new RegExp(`^${mockData.L4_IDLE_VIP}$`));
    await fillRequiredFields(page);

    const payload = await saveAndSubmit(page);
    expect(payload.exclusive).toBe(1);
    expect(payload.sla_type).toBe('');
    expect(payload.cloud_cluster_ids).toEqual([mockData.L4_CLUSTER_CLOUD_ID]);
    expect(payload.vip).toBe(mockData.L4_IDLE_VIP);
    expect(payload.cluster_tag).toBeUndefined();
  });

  test('P0-04 仅七层时可提交七层标签且不输出四层字段', async ({ page }) => {
    await openApplyForm(page);
    await chooseExclusiveSpec(page);
    await enableL7(page);
    await fillRequiredFields(page);

    const payload = await saveAndSubmit(page);
    expect(payload.exclusive).toBe(1);
    expect(payload.sla_type).toBe('');
    expect(payload.cluster_tag).toBe(mockData.L7_TAG_WITH_EGRESS);
    expect(payload.cloud_cluster_ids).toBeUndefined();
    expect(payload.vip).toBeUndefined();
  });

  test('P0-05 四层与七层同时启用时两类字段同时输出', async ({ page }) => {
    await openApplyForm(page);
    await chooseExclusiveSpec(page);
    await enableL4Random(page);
    await enableL7(page);
    await fillRequiredFields(page);

    const payload = await saveAndSubmit(page);
    expect(payload.exclusive).toBe(1);
    expect(payload.cluster_tag).toBe(mockData.L7_TAG_WITH_EGRESS);
    expect(payload.cloud_cluster_ids.length).toBe(3);
    expect(payload.vip).toBe('');
  });

  test('P1-08 取消勾选的四层不进入提交参数', async ({ page }) => {
    await openApplyForm(page);
    await chooseExclusiveSpec(page);
    await enableL4Random(page);
    // 取消四层勾选，只保留七层
    await page.getByText('四层集群', { exact: true }).click();
    await enableL7(page);
    await fillRequiredFields(page);

    const payload = await saveAndSubmit(page);
    expect(payload.cluster_tag).toBe(mockData.L7_TAG_WITH_EGRESS);
    expect(payload.cloud_cluster_ids).toBeUndefined();
    expect(payload.vip).toBeUndefined();
  });

  test('P1-09 切回共享型后不携带任何独占参数', async ({ page }) => {
    await openApplyForm(page);
    await chooseExclusiveSpec(page);
    await enableL4Random(page);

    await formItem(page, '负载均衡规格类型').locator('.bk-select').first().click();
    await pickOption(page, /^共享型$/);
    await fillRequiredFields(page);

    const payload = await saveAndSubmit(page);
    expect(payload.exclusive).toBe(0);
    expect(payload.sla_type).toBe('');
    expect(payload.cluster_tag).toBeUndefined();
    expect(payload.cloud_cluster_ids).toBeUndefined();
    expect(payload.vip).toBeUndefined();
  });

  test('P0-06 共享带宽包候选按四层指定集群的出口过滤', async ({ page }) => {
    await openApplyForm(page);
    await chooseExclusiveSpec(page);

    await page.getByText('四层集群', { exact: true }).click();
    await formItem(page, '独占集群').locator('.bk-select').first().click();
    await pickOption(page, /^L4-TAG-A$/);
    // 选具体集群 l4-cluster-a（egress-a），IP 保持随机分配
    await formItem(page, '独占集群').locator('.bk-select').nth(1).click();
    await pickOption(page, new RegExp(`^${mockData.L4_CLUSTER_NAME}$`));

    await page.getByText('共享带宽包', { exact: true }).click();
    await formItem(page, '共享带宽包').locator('.bk-select').first().click();

    const options = await visibleOptions(page);
    expect(options.some((text) => text.includes('bp-egress-a-1'))).toBe(true);
    expect(options.some((text) => text.includes('bp-egress-a-2'))).toBe(true);
    expect(options.some((text) => text.includes('bp-egress-b-1'))).toBe(false);
  });

  test('P1-05 出口集合为空时共享带宽包候选为空且不发请求', async ({ page }) => {
    await openApplyForm(page);
    await chooseExclusiveSpec(page);

    // 只启用七层，且选一个没有同标签 TGW 集群的七层标签 -> 有效出口为空
    await page.getByText('七层标签', { exact: true }).click();
    await formItem(page, '独占集群').locator('.bk-select').nth(3).click();
    await pickOption(page, new RegExp(`^${mockData.L7_TAG_NO_EGRESS}$`));

    await page.getByText('共享带宽包', { exact: true }).click();
    await formItem(page, '共享带宽包').locator('.bk-select').first().click();

    const options = await visibleOptions(page);
    expect(options.some((text) => text.includes('bp-egress'))).toBe(false);
  });

  test('P1-03 具体集群无空闲 IP 时提示不可用并清空该集群选择', async ({ page }) => {
    await openApplyForm(page);
    await chooseExclusiveSpec(page);
    await page.getByText('四层集群', { exact: true }).click();
    await formItem(page, '独占集群').locator('.bk-select').first().click();
    await pickOption(page, /^L4-TAG-A$/);

    // l4-cluster-no-ip 的空闲 IP 接口返回 count=0
    await formItem(page, '独占集群').locator('.bk-select').nth(1).click();
    await pickOption(page, new RegExp(`^${mockData.L4_CLUSTER_NO_VIP_NAME}$`));

    await expect(page.getByText('该集群暂无空闲 IP，请选择其它集群')).toBeVisible();
    await expect(formItem(page, '独占集群').locator('input').nth(2)).toHaveValue('');
  });

  test('P1-04 独占集群标签查询失败时规格类型不提供独占型', async ({ page }) => {
    await openApplyForm(page, { failUrls: ['exclusive_clusters/tags/list'] });

    await formItem(page, '负载均衡规格类型').locator('.bk-select').first().click();

    const options = await visibleOptions(page);
    expect(options).toContain('共享型');
    expect(options).toContain('性能容量型');
    expect(options.some((text) => text.trim() === '独占型')).toBe(false);
  });

  test('P1-06 可用区变化后失效的独占集群保留值被清空', async ({ page }) => {
    await openApplyForm(page);
    await chooseExclusiveSpec(page);
    await enableL4Random(page);
    await expect(formItem(page, '独占集群').locator('input').nth(1)).toHaveValue(mockData.L4_TAG);

    // 切换可用区 -> 上游条件变化（规格类型会一并重置回共享型）
    await formItem(page, '可用区').locator('.bk-select').nth(1).click();
    await pickOption(page, new RegExp(`^${mockData.TEST_ZONE_B_NAME}$`));
    await expect(formItem(page, '运营商类型').getByText('BGP')).toBeVisible();

    // 再次选择独占型：上一轮的标签/集群/IP 保留值必须已被清除
    await chooseExclusiveSpec(page);
    await expect(formItem(page, '独占集群').locator('input').nth(1)).toHaveValue('');
  });

  test('P1-07 取消勾选后保留值不生效，重新勾选后恢复', async ({ page }) => {
    await openApplyForm(page);
    await chooseExclusiveSpec(page);
    await enableL4Random(page);

    await page.getByText('四层集群', { exact: true }).click();
    // 取消勾选：控件置灰、保留值仍在
    await expect(formItem(page, '独占集群').locator('.bk-select').first()).toHaveClass(/is-disabled/);
    await expect(formItem(page, '独占集群').locator('input').nth(1)).toHaveValue(mockData.L4_TAG);

    await page.getByText('四层集群', { exact: true }).click();
    // 重新勾选：保留值恢复生效
    await expect(formItem(page, '独占集群').locator('.bk-select').first()).not.toHaveClass(/is-disabled/);
    await expect(formItem(page, '独占集群').locator('input').nth(1)).toHaveValue(mockData.L4_TAG);
  });

  test('P1-10 切回非独占型后询价请求不携带独占字段', async ({ page }) => {
    const inquiries: Array<Record<string, any>> = [];
    await openApplyForm(page, {
      override: ({ pathname, body }) => {
        if (pathname.endsWith('/load_balancers/prices/inquiry')) {
          inquiries.push(body);
          return { instance_price: null, bandwidth_price: null, lcu_price: null };
        }
        return undefined;
      },
    });

    // 先在独占型下形成前端配置残留
    await chooseExclusiveSpec(page);
    await enableL4Random(page);

    // 切回共享型，并补齐询价前置条件
    await formItem(page, '负载均衡规格类型').locator('.bk-select').first().click();
    await pickOption(page, /^共享型$/);
    await formItem(page, '安全组放通模式').locator('.bk-select').first().click();
    await pickOption(page, /启用默认放通/);
    await formItem(page, '实例名称').locator('input').first().fill('clb-e2e-01');
    await formItem(page, '带宽上限（Mbps）').locator('input').first().fill('100');
    await page.keyboard.press('Enter');

    await expect.poll(() => inquiries.length, { timeout: 20000 }).toBeGreaterThan(0);
    const payload = inquiries[inquiries.length - 1];
    for (const key of ['enable_l4', 'enable_l7', 'l4_cluster_tag', 'l4_cluster_id', 'l4_vip', 'exclusive_cluster_tags']) {
      expect(payload).not.toHaveProperty(key);
    }
    expect(payload.bandwidth_package_id).toBeUndefined();
    expect(payload.exclusive).toBe(0);
  });
});
