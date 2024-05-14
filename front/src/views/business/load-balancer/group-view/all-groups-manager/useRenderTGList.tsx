// import components
import { Button, Message } from 'bkui-vue';
import Confirm from '@/components/confirm';
// import stores
import { useAccountStore, useBusinessStore } from '@/store';
// import custom hooks
import useColumns from '@/views/resource/resource-manage/hooks/use-columns';
import useSelection from '@/views/resource/resource-manage/hooks/use-selection';
import { useTable } from '@/hooks/useTable/useTable';
// import types
import { ISearchItem } from 'bkui-vue/lib/search-select/utils';
// import utils
import { getTableNewRowClass } from '@/common/util';
import bus from '@/common/bus';

/**
 * 渲染目标组list
 */
export default () => {
  // use hooks
  const { columns, settings } = useColumns('targetGroup');
  const { selections, handleSelectionChange } = useSelection();
  // use stores
  const businessStore = useBusinessStore();
  const accountStore = useAccountStore();

  const searchData: ISearchItem[] = [
    {
      id: 'name',
      name: '目标组名称',
    },
    // {
    //   id: 'lb_name',
    //   name: '关联的负载均衡',
    // },
    {
      id: 'protocol',
      name: '协议',
    },
    // {
    //   id: 'port',
    //   name: '端口',
    // },
    {
      id: 'vendor',
      name: '云厂商',
    },
    {
      id: 'region',
      name: '地域',
    },
    {
      id: 'cloud_vpc_id',
      name: '所属VPC',
    },
    {
      id: 'health_check.health_switch',
      name: '健康检查',
      children: [
        { name: '已开启', id: 1 },
        { name: '未开启', id: 0 },
      ],
    },
  ];
  const tableColumns = [
    ...columns,
    {
      label: '操作',
      width: 120,
      render: ({ data }: any) => (
        <div>
          <Button text theme={'primary'} onClick={() => handleEditTargetGroup(data.id)}>
            编辑
          </Button>
          <span
            v-bk-tooltips={{
              content: '已绑定了监听器的目标组不可删除',
              disabled: data.listener_num === 0,
            }}>
            <Button
              text
              theme={'primary'}
              disabled={data.listener_num > 0}
              class={'ml16'}
              onClick={() => {
                handleDeleteTargetGroup(data.id, data.name);
              }}>
              删除
            </Button>
          </span>
        </div>
      ),
    },
  ];

  const { CommonTable, getListData } = useTable({
    searchOptions: {
      searchData,
    },
    tableOptions: {
      columns: tableColumns,
      extra: {
        settings: settings.value,
        onSelect: (selections: any) => {
          handleSelectionChange(selections, () => true, false);
        },
        onSelectAll: (selections: any) => {
          handleSelectionChange(selections, () => true, true);
        },
        rowClass: getTableNewRowClass(),
      },
    },
    requestOption: {
      type: 'target_groups',
      sortOption: { sort: 'created_at', order: 'DESC' },
      async resolveDataListCb(dataList: any) {
        if (dataList.length === 0) return;
        return dataList.map((data: any) => {
          const { health_check } = data;
          health_check.health_switch = health_check.health_switch || 0;
          return { ...data, health_check };
        });
      },
    },
  });

  // 编辑单个目标组
  const handleEditTargetGroup = async (id: string) => {
    // 获取对应目标组的详情
    const { data } = await businessStore.getTargetGroupDetail(id);
    bus.$emit('editTargetGroup', { ...data, rs_list: data.target_list });
  };

  // 删除单个目标组
  const handleDeleteTargetGroup = (id: string, name: string) => {
    Confirm('请确定删除目标组', `将删除目标组【${name}】`, () => {
      businessStore
        .deleteTargetGroups({
          bk_biz_id: accountStore.bizs,
          ids: [id],
        })
        .then(() => {
          Message({
            message: '删除成功',
            theme: 'success',
          });
          // 刷新表格数据
          getListData();
          // 刷新左侧目标组列表
          bus.$emit('refreshTargetGroupList');
        });
    });
  };

  return { searchData, selections, CommonTable, getListData };
};
