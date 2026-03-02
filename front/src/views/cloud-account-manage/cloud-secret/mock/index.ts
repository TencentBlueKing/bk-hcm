/**
 * 云密钥模块 Mock 数据
 * 使用 USE_MOCK 参数控制是否启用 mock 数据
 */

import type { ICloudSecretItem } from '../typings';

// Mock 开关，设置为 true 启用 mock 数据，false 则使用真实接口
export const USE_MOCK = false;

// 模拟延迟时间（毫秒）
const MOCK_DELAY = 300;

// Mock 数据生成工具函数
const generateMockSecretId = (index: number) => `AKID${String(index).padStart(16, '0')}`;

const generateMockAccountId = (index: number) => `100${String(index).padStart(12, '0')}`;

const generateRandomDate = (startYear = 2023, endYear = 2026): string => {
  const start = new Date(startYear, 0, 1).getTime();
  const end = new Date(endYear, 11, 31).getTime();
  const randomTime = start + Math.random() * (end - start);
  return new Date(randomTime).toISOString().replace('T', ' ').substring(0, 19);
};

// Mock 用户列表
const mockUsers = ['zhangsan', 'lisi', 'wangwu', 'zhaoliu', 'qianqi', 'sunba', 'zhoujiu', 'wushi'];

const getRandomUser = () => mockUsers[Math.floor(Math.random() * mockUsers.length)];

// 生成 Mock 数据列表
const generateMockList = (count: number): ICloudSecretItem[] => {
  const list: ICloudSecretItem[] = [];

  for (let i = 1; i <= count; i++) {
    const isEnabled = Math.random() > 0.3;
    const isConsoleLogin = Math.random() > 0.5;
    const mainAccountId = generateMockAccountId(Math.floor(i / 5) + 1);
    const subAccountId = generateMockAccountId(i);
    const createdAt = generateRandomDate(2023, 2025);

    list.push({
      id: `secret_${String(i).padStart(8, '0')}`,
      vendor: 'tcloud',
      status: isEnabled ? 'enabled' : 'disabled',
      account_id: `account_${String(Math.floor(i / 5) + 1).padStart(4, '0')}`,
      sub_account_id: `sub_account_${String(i).padStart(4, '0')}`,
      extension: {
        cloud_secret_id: generateMockSecretId(i),
        cloud_main_account_id: mainAccountId,
        cloud_sub_account_id: subAccountId,
        console_login: isConsoleLogin ? 1 : 0,
      },
      cloud_secret_id: generateMockSecretId(i),
      cloud_main_account_id: mainAccountId,
      cloud_sub_account_id: subAccountId,
      console_login: isConsoleLogin ? 1 : 0,
      tenant_id: `tenant_${String(Math.floor(i / 10) + 1).padStart(4, '0')}`,
      cloud_created_at: createdAt,
      disabled_time: isEnabled ? undefined : generateRandomDate(2025, 2026),
      last_used_time: Math.random() > 0.3 ? generateRandomDate(2025, 2026) : undefined,
      creator: getRandomUser(),
      reviser: getRandomUser(),
      created_at: createdAt,
      updated_at: generateRandomDate(2025, 2026),
      sub_account_manager: getRandomUser(),
      account_manager: getRandomUser(),
    });
  }

  return list;
};

// Mock 数据总数
const MOCK_TOTAL_COUNT = 56;

// 缓存的 Mock 数据
let cachedMockList: ICloudSecretItem[] | null = null;

const getMockList = (): ICloudSecretItem[] => {
  if (!cachedMockList) {
    cachedMockList = generateMockList(MOCK_TOTAL_COUNT);
  }
  return cachedMockList;
};

/**
 * 模拟获取三级账号密钥列表
 * @param params 查询参数
 */
export const mockGetSubAccountSecretList = async (params: {
  page: { start: number; limit: number };
  cloud_secret_ids?: string[];
  status?: string;
  cloud_sub_account_ids?: string[];
  cloud_main_account_ids?: string[];
  sub_account_managers?: string[];
  account_managers?: string[];
}): Promise<{ list: ICloudSecretItem[]; count: number }> => {
  // 模拟网络延迟
  await new Promise((resolve) => setTimeout(resolve, MOCK_DELAY));

  let filteredList = [...getMockList()];

  // 根据条件筛选
  if (params.cloud_secret_ids && params.cloud_secret_ids.length > 0) {
    filteredList = filteredList.filter((item) =>
      params.cloud_secret_ids!.some((id) => item.cloud_secret_id?.toLowerCase().includes(id.toLowerCase())),
    );
  }

  if (params.status) {
    filteredList = filteredList.filter((item) => item.status === params.status);
  }

  if (params.cloud_sub_account_ids && params.cloud_sub_account_ids.length > 0) {
    filteredList = filteredList.filter((item) =>
      params.cloud_sub_account_ids!.some((id) => item.cloud_sub_account_id?.toLowerCase().includes(id.toLowerCase())),
    );
  }

  if (params.cloud_main_account_ids && params.cloud_main_account_ids.length > 0) {
    filteredList = filteredList.filter((item) =>
      params.cloud_main_account_ids!.some((id) => item.cloud_main_account_id?.toLowerCase().includes(id.toLowerCase())),
    );
  }

  if (params.sub_account_managers && params.sub_account_managers.length > 0) {
    filteredList = filteredList.filter((item) =>
      params.sub_account_managers!.some((manager) =>
        item.sub_account_manager?.toLowerCase().includes(manager.toLowerCase()),
      ),
    );
  }

  if (params.account_managers && params.account_managers.length > 0) {
    filteredList = filteredList.filter((item) =>
      params.account_managers!.some((manager) => item.account_manager?.toLowerCase().includes(manager.toLowerCase())),
    );
  }

  // 分页处理
  const { start = 0, limit = 20 } = params.page || {};
  const pagedList = filteredList.slice(start, start + limit);

  return {
    list: pagedList,
    count: filteredList.length,
  };
};

/**
 * 模拟启用/禁用密钥
 * @param params 参数
 */
export const mockUpdateSubAccountSecretStatus = async (
  params: { id: string; status: 'enabled' | 'disabled' }[],
): Promise<{ ids: string[] }> => {
  await new Promise((resolve) => setTimeout(resolve, MOCK_DELAY));

  // 更新缓存中的数据
  const mockList = getMockList();
  params.forEach((param) => {
    const item = mockList.find((i) => i.id === param.id);
    if (item) {
      item.status = param.status;
      if (param.status === 'disabled') {
        item.disabled_time = new Date().toISOString().replace('T', ' ').substring(0, 19);
      } else {
        item.disabled_time = undefined;
      }
    }
  });

  return { ids: params.map((p) => p.id) };
};

/**
 * 模拟删除密钥
 * @param ids 密钥ID列表
 */
export const mockDeleteSubAccountSecret = async (ids: string[]): Promise<{ ids: string[] }> => {
  await new Promise((resolve) => setTimeout(resolve, MOCK_DELAY));

  // 从缓存中移除数据
  if (cachedMockList) {
    cachedMockList = cachedMockList.filter((item) => !ids.includes(item.id));
  }

  return { ids };
};
