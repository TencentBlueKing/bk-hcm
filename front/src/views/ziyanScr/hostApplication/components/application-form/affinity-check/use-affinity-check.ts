import { ref } from 'vue';
import http from '@/http';

export interface IAffinityCheckResultItem {
  zone: string;
  device_type: string;
  replicas: number;
  status: 1 | 2; // 匹配状态（1:CRP预检有数据 2:CRP预检无数据）
  max_cut_num: number;
  ips: string[];
}

export const useAffinityCheck = () => {
  const isLoading = ref(false);
  const isResultDialogShow = ref(false);
  const affinityCheckResult = ref<IAffinityCheckResultItem[]>([]);

  const affinityCheck = async (params: {
    bk_biz_id: number;
    specs: Array<{ zones: string[]; device_type: string; replicas: number }>;
  }) => {
    isLoading.value = true;
    try {
      const { data } = await http.post(`/api/v1/woa/bizs/${params.bk_biz_id}/task/apply/match/check`, params);
      affinityCheckResult.value = data?.details || [];
      isResultDialogShow.value = true;
    } catch (error) {
      console.error(error);
    } finally {
      isLoading.value = false;
    }
  };

  return {
    isLoading,
    isResultDialogShow,
    affinityCheckResult,
    affinityCheck,
  };
};
