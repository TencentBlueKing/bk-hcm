import { VendorEnum } from '@/common/constant';
import { TargetGroupOperationScene } from '@/constants';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import http from '@/http';
import { IListResData } from '@/typings';
import { defineStore } from 'pinia';
import { reactive, Ref, ref } from 'vue';

export interface ITreeNode {
  [key: string]: any;
  lb: Record<string, any>; // 当前域名节点所属的负载均衡信息，非域名节点时不生效
  listener: Record<string, any>; // 当前域名节点所属的监听器信息，非域名节点时不生效
}

export interface ITargetGroupDetail {
  id: string;
  cloud_id: string;
  name: string;
  vendor: string;
  account_id: string;
  bk_biz_id: number;
  target_group_type: string;
  vpc_id: string;
  cloud_vpc_id: string;
  region: string;
  protocol: string;
  port: number;
  weight: number;
  health_check: {
    health_switch: number;
    time_out: number;
    interval_time: number;
    health_num: number;
    un_health_num: number;
    http_code: number;
    check_type: string;
    http_check_path: string;
    http_check_domain: string;
    http_check_method: string;
    source_ip_type: number;
    extended_code: string;
  };
  memo: string;
  creator: string;
  reviser: string;
  created_at: string;
  updated_at: string;
  target_list: any[];
}

interface IRulesBindingStatus {
  rule_id: string;
  binding_status: string;
}

export const useLoadBalancerStore = defineStore('load-balancer', () => {
  const { getBusinessApiPath } = useWhereAmI();

  // state - 目标组id
  const targetGroupId = ref('');
  const setTargetGroupId = (v: string) => {
    targetGroupId.value = v;
  };

  const listenerDetailWithTargetGroup = ref({} as any);
  const setListenerDetailWithTargetGroup = (v: any) => {
    listenerDetailWithTargetGroup.value = v;
  };

  // state - lb-tree - 当前选中的资源
  // eslint-disable-next-line @typescript-eslint/consistent-type-assertions
  const currentSelectedTreeNode: Ref<ITreeNode> = ref({} as ITreeNode);
  const setCurrentSelectedTreeNode = (node: ITreeNode) => {
    // 其中, node 可能为 lb, listener, domain 节点
    currentSelectedTreeNode.value = node;
  };

  // state - 目标组操作场景
  const updateCount = ref(0); // 记录修改次数: 当值为2时, 重新对场景进行判断(判断第一次回显数据的影响)
  const setUpdateCount = (v: number) => {
    updateCount.value = v;
  };
  const currentScene = ref<TargetGroupOperationScene>(); // 记录当前操作类型
  const setCurrentScene = (v: TargetGroupOperationScene) => {
    currentScene.value = v;
  };
  const defaultTargetGroupOperateLockState = { singleUpdateRsId: '' };
  const targetGroupOperateLockState = reactive(defaultTargetGroupOperateLockState);
  const setTargetGroupOperateLockState = (v: typeof defaultTargetGroupOperateLockState) => {
    Object.assign(targetGroupOperateLockState, v);
  };
  const resetTargetGroupOperateLockState = () => {
    Object.assign(targetGroupOperateLockState, defaultTargetGroupOperateLockState);
  };

  // state - lb-tree的搜索条件, 用于链接跳转
  const lbTreeSearchTarget = ref();
  const setLbTreeSearchTarget = (v: any) => {
    lbTreeSearchTarget.value = v;
  };

  // state - 目标组视角下的搜索条件, 用于链接跳转
  const tgSearchTarget = ref();
  const setTgSearchTarget = (v: any) => {
    tgSearchTarget.value = v;
  };

  // 查询规则绑定目标组状态接口
  const queryRulesBindingStatusListLoading = ref(false);
  const queryRulesBindingStatusList = async (vendor: VendorEnum, lblId: string, payload: { rule_ids: string[] }) => {
    queryRulesBindingStatusListLoading.value = true;
    try {
      const res: IListResData<IRulesBindingStatus[]> = await http.post(
        `/api/v1/cloud/${getBusinessApiPath()}vendors/${vendor}/listeners/${lblId}/rules/binding_status/list`,
        payload,
      );
      return res.data?.details || [];
    } catch (error) {
      console.error(error);
      return Promise.reject(error);
    } finally {
      queryRulesBindingStatusListLoading.value = false;
    }
  };

  return {
    targetGroupId,
    setTargetGroupId,
    currentSelectedTreeNode,
    setCurrentSelectedTreeNode,
    updateCount,
    setUpdateCount,
    currentScene,
    setCurrentScene,
    targetGroupOperateLockState,
    setTargetGroupOperateLockState,
    resetTargetGroupOperateLockState,
    lbTreeSearchTarget,
    setLbTreeSearchTarget,
    tgSearchTarget,
    setTgSearchTarget,
    listenerDetailWithTargetGroup,
    setListenerDetailWithTargetGroup,
    queryRulesBindingStatusListLoading,
    queryRulesBindingStatusList,
  };
});
