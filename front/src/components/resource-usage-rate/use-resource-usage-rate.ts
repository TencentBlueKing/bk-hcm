import { ref } from 'vue';
import dayjs from 'dayjs';
import http from '@/http';

export interface IDeviceLoadUsage {
  threshold: number;
  cpu_usage: number;
  achieved_kpi: boolean;
  empty_load_cpu_core: number;
  empty_load_os: number;
  low_load_cpu_core: number;
  low_load_os: number;
}

export interface IDeviceCpuUsageTrend {
  date: string;
  cpu_usage: number;
}

export const useResourceUsageRate = () => {
  const deviceLoadUsageLoading = ref(true);
  const deviceCpuUsageTrendLoading = ref(false);

  const getDeviceLoadUsage = async (params: { bk_biz_id: number }): Promise<IDeviceLoadUsage> => {
    deviceLoadUsageLoading.value = true;
    try {
      const { data } = await http.post(`/api/v1/woa/bizs/${params.bk_biz_id}/device/load_usage`, {
        // T-2 的日期
        date: dayjs().subtract(2, 'day').format('YYYY-MM-DD'),
      });
      return data;
    } catch (error) {
      console.error(error);
    } finally {
      deviceLoadUsageLoading.value = false;
    }
  };

  const getDeviceCpuUsageTrend = async (params: { bk_biz_id: number }): Promise<IDeviceCpuUsageTrend[]> => {
    deviceCpuUsageTrendLoading.value = true;
    try {
      const { data } = await http.post(`/api/v1/woa/bizs/${params.bk_biz_id}/device/cpu_usage/trend`, {
        time_granularity: 'month',
        // 最近 6 个月，不包括最近一个月
        date_range: {
          start: dayjs().subtract(6, 'month').format('YYYY-MM-DD'),
          end: dayjs().subtract(1, 'month').format('YYYY-MM-DD'),
        },
      });
      return data ?? [];
    } catch (error) {
      console.error(error);
    } finally {
      deviceCpuUsageTrendLoading.value = false;
    }
  };

  return {
    deviceLoadUsageLoading,
    getDeviceLoadUsage,
    deviceCpuUsageTrendLoading,
    getDeviceCpuUsageTrend,
  };
};
