import { timeFormatter } from '@/common/util';

export const formatDiskBoolean = (value?: boolean | null) => {
  if (typeof value !== 'boolean') return '--';
  return value ? '是' : '否';
};

export const formatDiskExpiry = (value?: string) => (value?.startsWith('0000-00-00') ? '--' : timeFormatter(value));
