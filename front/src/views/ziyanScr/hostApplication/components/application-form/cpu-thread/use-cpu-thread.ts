import { computed, Ref } from 'vue';

// 支持超线程开关的机型族前缀列表
export const CPU_THREAD_SUPPORTED_DEVICE_FAMILIES = [
  // 标准型
  'SA9',
  'SA9e',
  'S9',
  'S9e',
  'S9pro',
  'SA5',
  'S8',
  'SA4',
  'S6',
  'SA3',
  'SA2',
  'S5',
  'S5se',
  // 内存型
  'MA9',
  'MA9e',
  'M9',
  'M9e',
  'M9pro',
  'MA5',
  'M8',
  'M6ce',
  'M6',
  'MA3',
  'MA2',
  'M5',
  // 计算型
  'C6',
  'C5',
  'C4',
  'CN3',
  'C3',
  // 高IO型
  'IT5',
  'IT3',
  'IA3se',
  // 大数据型
  'D3',
  'D2',
];

// 超线程功能业务白名单
export const CPU_THREAD_BIZ_WHITELIST = [100948];

// CPU超线程开关：1表示关闭超线程，2表示开启超线程
export const CpuThreadSwitch = {
  DISABLED: 1, // 关闭超线程
  ENABLED: 2, // 开启超线程（默认）
} as const;

/**
 * 判断机型是否支持超线程开关
 * @param deviceType 机型名称，如 "SA9.LARGE16"
 * @returns 是否支持超线程开关
 */
export const isDeviceTypeSupportCpuThread = (deviceType: string): boolean => {
  if (!deviceType) return false;
  // 机型格式为 "SA9.LARGE16"，需要匹配前缀
  const deviceFamily = deviceType.split('.')[0];
  return CPU_THREAD_SUPPORTED_DEVICE_FAMILIES.some((family) => deviceFamily === family);
};

/**
 * 判断业务是否在超线程功能白名单内
 * @param bizId 业务ID
 * @returns 是否在白名单内
 */
export const isBizInCpuThreadWhitelist = (bizId: number): boolean => {
  return CPU_THREAD_BIZ_WHITELIST.includes(bizId);
};

/**
 * 判断CPU超线程开关
 * @param bizId 业务ID
 * @param formData 表单数据
 * @returns 是否开启超线程
 */
export const useCpuThread = (bizId: Ref<number>, formData: Ref<any>) => {
  const enabledCpuThread = computed(() => {
    return (
      isDeviceTypeSupportCpuThread(formData.value?.spec?.device_type) && isBizInCpuThreadWhitelist(Number(bizId.value))
    );
  });

  return {
    enabledCpuThread,
  };
};
