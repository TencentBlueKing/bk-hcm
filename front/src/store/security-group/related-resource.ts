import { ref } from 'vue';
import { defineStore } from 'pinia';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import http from '@/http';
import { VendorEnum } from '@/common/constant';
import { IQueryResData } from '@/typings';

interface IRelatedResourceSecurityGroupItem {
  id: string;
  vendor: VendorEnum;
  cloud_id: string;
  region: string;
  name: string;
  memo: string;
  cloud_created_time: string;
  cloud_update_time: string;
  tags: Record<string, string>;
  account_id: string;
  bk_biz_id: number;
  mgmt_type: string;
  mgmt_biz_id: number;
  manager: string;
  bak_manager: string;
  usage_biz_ids: number[];
  creator: string;
  reviser: string;
  created_at: string;
  updated_at: string;
  res_id: string;
  res_type: string;
  priority: number;
  rel_creator: string;
  rel_created_at: string;
}

/**
 * 安全组关联资源侧，例如cvm侧、clb侧
 */
export const useSecurityGroupRelatedResourceStore = defineStore('security-group-related-resource', () => {
  const { getBusinessApiPath } = useWhereAmI();

  const querySecurityListWithResInfoLoading = ref(false);
  /**
   * 查询关联资源所绑定的安全组列表
   * @param res_type cvm | load_balancer
   * @param res_id
   */
  const querySecurityGroupListWithResInfo = async (res_type: string, res_id: string) => {
    querySecurityListWithResInfoLoading.value = true;
    try {
      const res: IQueryResData<IRelatedResourceSecurityGroupItem[]> = await http.get(
        `/api/v1/cloud/${getBusinessApiPath()}security_groups/res/${res_type}/${res_id}`,
      );
      return res.data;
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    } finally {
      querySecurityListWithResInfoLoading.value = false;
    }
  };

  return {
    querySecurityListWithResInfoLoading,
    querySecurityGroupListWithResInfo,
  };
});
