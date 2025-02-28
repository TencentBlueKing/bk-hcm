import { ref } from 'vue';
import { defineStore } from 'pinia';
import type { IAccountItem, IListResData } from '@/typings';
import { getPrivateIPs, getPublicIPs } from '@/utils';
import http from '@/http';

export interface ICvmItem {
  id: string;
  cloud_id: string;
  name: string;
  vendor: string;
  bk_biz_id: number;
  bk_cloud_id: number;
  account_id: string;
  region: string;
  zone: string;
  cloud_vpc_ids: string[];
  vpc_ids: string[];
  cloud_subnet_ids: string[];
  subnet_ids: string[];
  cloud_image_id: string;
  image_id: string;
  os_name: string;
  memo: string;
  status: string;
  private_ipv4_addresses: string[];
  private_ipv6_addresses: string[];
  public_ipv4_addresses: string[];
  public_ipv6_addresses: string[];
  machine_type: string;
  cloud_created_time: string;
  cloud_launched_time: string;
  cloud_expired_time: string;
  creator: string;
  reviser: string;
  created_at: string;
  updated_at: string;
}

export interface ICvmsAssignBizsPreviewItem {
  cvm_id: string;
  match_type: 'auto' | 'manual' | 'no_match';
  bk_cloud_id: number;
  bk_biz_id: number;
}

export type CvmsAssignPreviewItem = ICvmItem &
  Omit<ICvmsAssignBizsPreviewItem, 'cvm_id'> & { account_name: string; bk_biz_ids: number[] };

export type CvmBatchAssignOpItem = {
  account_name: string;
  cloud_vpc_id: string;
  onlyShowUnConfirmed: boolean;
  tableData: CvmsAssignPreviewItem[];
  unConfirmedCount?: number;
  hostCount?: number;
  bkCloudCount?: number;
};

export interface IMatchHostsItem {
  bk_host_id: number;
  private_ipv4_addresses: string[];
  public_ipv4_addresses: string[];
  bk_cloud_id: number;
  bk_biz_id: number;
  region: string;
  bk_host_name: string;
  bk_os_name: string;
  create_time: string;
}

export const useHostStore = defineStore('host', () => {
  // 云地域
  const regionList = ref([]);

  const fetchData = async <T>(url: string, data: object): Promise<T[]> => {
    const res: IListResData<T[]> = await http.post(url, data);
    return res.data.details;
  };

  const isAssignPreviewLoading = ref(false);
  const getAssignPreviewList = async (cvms: ICvmItem[]) => {
    isAssignPreviewLoading.value = true;
    try {
      // 并行获取预览数据和账户列表数据
      const [cvmsAssignBizsPreviewList, accountList] = await Promise.all([
        fetchData<ICvmsAssignBizsPreviewItem>('/api/v1/cloud/cvms/assign/bizs/preview', {
          cvm_ids: cvms.map((cvm) => cvm.id),
        }),
        fetchData<IAccountItem>('/api/v1/cloud/accounts/list', {
          page: { start: 0, limit: 500, count: false },
          filter: {
            op: 'and',
            rules: [{ field: 'id', value: [...new Set(cvms.map((cvm) => cvm.account_id))], op: 'in' }],
          },
        }),
      ]);

      return cvms.flatMap<CvmsAssignPreviewItem>((item) => {
        const previewItem = cvmsAssignBizsPreviewList.find((i) => i.cvm_id === item.id);
        const accountItem = accountList.find((i) => i.id === item.account_id);
        // 将预览数据和账户数据合并到选择的项目中
        return {
          ...item,
          match_type: previewItem?.match_type,
          bk_biz_id: previewItem?.bk_biz_id,
          bk_cloud_id: previewItem?.bk_cloud_id,
          account_name: accountItem?.name || '',
          bk_biz_ids: accountItem?.bk_biz_ids || [],
          private_ip_address: getPrivateIPs(item),
          public_ip_address: getPublicIPs(item),
        };
      });
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    } finally {
      isAssignPreviewLoading.value = false;
    }
  };

  const isAssignHostsMatchLoading = ref(false);
  const getAssignHostsMatchList = async (account_id: string, private_ipv4_addresses: string[]) => {
    isAssignHostsMatchLoading.value = true;
    try {
      const list = await fetchData<IMatchHostsItem>('/api/v1/cloud/cvms/assign/hosts/match/list', {
        account_id,
        private_ipv4_addresses,
      });
      return list;
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    } finally {
      isAssignHostsMatchLoading.value = false;
    }
  };

  const isAssignCvmsToBizsLoading = ref(false);
  const assignCvmsToBiz = async (cvms: { cvm_id: string; bk_biz_id: number; bk_cloud_id: number }[]) => {
    isAssignCvmsToBizsLoading.value = true;
    try {
      await http.post('/api/v1/cloud/cvms/assign/bizs', { cvms });
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    } finally {
      isAssignCvmsToBizsLoading.value = false;
    }
  };

  return {
    regionList,
    isAssignPreviewLoading,
    getAssignPreviewList,
    isAssignHostsMatchLoading,
    getAssignHostsMatchList,
    isAssignCvmsToBizsLoading,
    assignCvmsToBiz,
  };
});
