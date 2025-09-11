import { computed, ref } from 'vue';
// import components
import { Checkbox, Message } from 'bkui-vue';
// import stores
import { useBusinessStore } from '@/store';
// import types
import { QueryRuleOPEnum } from '@/typings';

export default () => {
  // use stores
  const businessStore = useBusinessStore();

  let account_id = ''; // 同一个负载均衡, 肯定同属于一个账号

  // dialog相关
  const isBatchDeleteRsShow = ref(false);
  const isBatchDeleteRsSubmitLoading = ref(false);
  // submit-handler
  const batchDeleteRs = async () => {
    const target_ids: string[] = [];
    for (const [, target_ids] of reqDataMap) {
      target_ids.push(...target_ids);
    }
    try {
      isBatchDeleteRsSubmitLoading.value = true;
      await businessStore.batchDeleteTargets({ account_id, target_ids });
      Message({ theme: 'success', message: '批量移除RS成功' });
      isBatchDeleteRsShow.value = false;
    } finally {
      isBatchDeleteRsSubmitLoading.value = false;
    }
  };

  // table相关
  // 切换单个rs的选中状态
  const handleChangeChecked = (tgId: string, rs: any, isChecked: boolean, isBreak = false) => {
    rs.isChecked = isChecked;
    if (isChecked) {
      // 如果tg下已有rs被选中, 则添加到最后; 否则, 将rs作为数组第一项添加进去
      reqDataMap.has(tgId) ? reqDataMap.get(tgId).push(rs.id) : reqDataMap.set(tgId, [rs.id]);
    } else {
      // 如果设置了isBreak=true, 则提前结束函数.(全不选时可以设置为true)
      if (isBreak) return;
      const rsList = reqDataMap.get(tgId);
      // 取消选中时, rsList一定有值. 找出对应rs的索引, 移除就行
      const idx = rsList.findIndex((id) => id === rs.id);
      rsList.splice(idx, 1);

      // 如果rsList为空, 则移除map中对应的映射
      if (rsList.length === 0) reqDataMap.delete(tgId);
    }
  };
  // 是否全选
  const isCheckedAll = computed({
    get() {
      return batchDeleteRsTableData.value
        .reduce((prev, curr) => [...prev, ...curr.ipList], [])
        .every((item: any) => item.isChecked === true);
    },
    set(isChecked: boolean) {
      // 清空reqDataMap, 避免全选之前已经选中过rs带来的影响
      reqDataMap.clear();
      batchDeleteRsTableData.value.forEach((item) => {
        // 遍历ipList, 修改checked状态
        item.ipList.forEach((rs: any) => {
          handleChangeChecked(item.tgId, rs, isChecked, true);
        });
      });
    },
  });
  const batchDeleteRsTableColumn = [
    {
      type: 'expand',
      width: 32,
      minWidth: 32,
      colspan: 7,
      resizable: false,
      label: () => <Checkbox v-model={isCheckedAll.value} />,
    },
    {
      label: '内网IP',
      field: 'private_ip_address',
      resizable: false,
    },
    {
      label: '公网IP',
      field: 'public_ip_address',
      resizable: false,
    },
    {
      label: '名称',
      field: 'name',
      resizable: false,
    },
    {
      label: '资源类型',
      field: 'inst_type',
      resizable: false,
    },
    {
      label: '端口',
      field: 'port',
      resizable: false,
    },
    {
      label: '权重',
      field: 'weight',
      resizable: false,
    },
  ];
  const isBatchDeleteRsTableLoading = ref(false);
  const batchDeleteRsTableData = ref([]);
  const tgPaginationMap = new Map<string, { start: number; hasNext: boolean }>(); // 映射: target_group_id 和 pagination
  const reqDataMap = new Map<string, string[]>([]); // 映射: target_group_id 和 target_ids

  /**
   * 初始化映射关系
   * @param tgIds 目标组id列表
   */
  const initMap = (tgIds: string[], accountId: string) => {
    // 更新账号id
    account_id = accountId;

    // target_group_id 和 pagination
    tgPaginationMap.clear();
    tgIds.forEach((id) => {
      tgPaginationMap.set(id, { start: 0, hasNext: true });
    });

    // target_group_id 和 target_ids
    reqDataMap.clear();
  };

  /**
   * 获取对应目标组下的rs列表
   * @param tgId 目标组id
   */
  const getRsListOfTargetGroup = async (tgId: string) => {
    const [detailRes, countRes] = await Promise.all(
      [false, true].map((isCount) =>
        businessStore.getRsList(tgId, {
          filter: {
            op: QueryRuleOPEnum.AND,
            rules: [],
          },
          page: {
            count: isCount,
            start: isCount ? 0 : tgPaginationMap.get(tgId).start,
            limit: isCount ? 0 : 500,
          },
        }),
      ),
    );
    return {
      ipCount: countRes.data.count,
      ipList: detailRes.data.details.map(
        ({ id, private_ip_address, public_ip_address, inst_name, inst_type, port, weight }: any) => ({
          id,
          isChecked: false,
          private_ip_address,
          public_ip_address,
          inst_name,
          inst_type,
          port,
          weight,
        }),
      ),
    };
  };

  /**
   * 获取目标组信息以及其所属rs列表
   * @param tgIds 目标组id列表
   */
  const getRsListOfTargetGroups = async (targetGroups: any[]) => {
    batchDeleteRsTableData.value = [];
    const promises = targetGroups.map(({ id, name, account_id }) =>
      getRsListOfTargetGroup(id).then(({ ipCount, ipList }) => ({
        tgId: id,
        tgName: name,
        account_id,
        ipCount,
        ipList,
      })),
    );
    try {
      isBatchDeleteRsTableLoading.value = true;
      batchDeleteRsTableData.value = await Promise.all(promises);
    } finally {
      isBatchDeleteRsTableLoading.value = false;
    }
  };

  // search相关
  const batchDeleteRsSearchData = [] as any[];

  return {
    isBatchDeleteRsShow,
    isBatchDeleteRsSubmitLoading,
    isBatchDeleteRsTableLoading,
    batchDeleteRsTableColumn,
    batchDeleteRsTableData,
    initMap,
    getRsListOfTargetGroups,
    handleChangeChecked,
    batchDeleteRsSearchData,
    batchDeleteRs,
  };
};
