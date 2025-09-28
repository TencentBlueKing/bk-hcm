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
import { TARGET_GROUP_PROTOCOLS, VendorEnum, VendorMap } from '@/common/constant';
import { useI18n } from 'vue-i18n';

/**
 * 渲染目标组list
 */
export default () => {
  // use hooks
  const { t } = useI18n();
  const { columns, settings } = useColumns('targetGroup');
  const { selections, handleSelectionChange } = useSelection();
  // use stores
  const businessStore = useBusinessStore();
  const accountStore = useAccountStore();

  const searchData: ISearchItem[] = [
    { id: 'name', name: t('目标组名称') },
    { id: 'protocol', name: t('协议'), children: TARGET_GROUP_PROTOCOLS.map((item) => ({ id: item, name: item })) },
    { id: 'port', name: t('端口') },
    {
      id: 'vendor',
      name: t('云厂商'),
      children: [{ id: VendorEnum.TCLOUD, name: VendorMap[VendorEnum.TCLOUD] }],
    },
    { id: 'cloud_vpc_id', name: t('所属VPC') },
    {
      id: 'health_check.health_switch',
      name: t('健康检查'),
      children: [
        { name: t('已开启'), id: 1 },
        { name: t('未开启'), id: 0 },
      ],
    },
  ];
  const tableColumns = [
    ...columns,
    {
      label: t('操作'),
      width: 120,
      fixed: 'right',
      render: ({ data }: any) => (
        <div>
          <Button text theme={'primary'} onClick={() => handleEditTargetGroup(data)}>
            {t('编辑')}
          </Button>
          <span
            v-bk-tooltips={{
              content: t('已绑定了监听器的目标组不可删除'),
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
              {t('删除')}
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
  const handleEditTargetGroup = async (tgItem: any) => {
    // 获取对应目标组的详情
    const { data } = await businessStore.getTargetGroupDetail(tgItem.id);
    bus.$emit('editTargetGroup', { ...data, rs_list: data.target_list, lb_id: tgItem.lb_id });
  };

  // 删除单个目标组
  const handleDeleteTargetGroup = (id: string, name: string) => {
    Confirm(t('请确定删除目标组'), `${t('将删除目标组【')}${name}${t('】')}`, async () => {
      await businessStore.deleteTargetGroups({
        bk_biz_id: accountStore.bizs,
        ids: [id],
      });
      Message({ message: t('删除成功'), theme: 'success' });
      // 刷新表格数据
      getListData();
      // 刷新左侧目标组列表
      bus.$emit('refreshTargetGroupList');
    });
  };

  return { searchData, selections, CommonTable, getListData };
};
