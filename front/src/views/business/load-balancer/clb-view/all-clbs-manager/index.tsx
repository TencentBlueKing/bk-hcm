import { defineComponent } from 'vue';
import { useRouter, useRoute } from 'vue-router';
// import components
import { Button, Message } from 'bkui-vue';
import { BkRadioButton, BkRadioGroup } from 'bkui-vue/lib/radio';
import BatchOperationDialog from '@/components/batch-operation-dialog';
// import hooks
import useColumns from '@/views/resource/resource-manage/hooks/use-columns';
import useSelection from '@/views/resource/resource-manage/hooks/use-selection';
import { useTable } from '@/hooks/useTable/useTable';
import { useI18n } from 'vue-i18n';
import { useWhereAmI, Senarios } from '@/hooks/useWhereAmI';
import useBatchDeleteLB from './useBatchDeleteLB';
import { useBusinessStore, useResourceStore } from '@/store';
// import utils
import { getTableNewRowClass } from '@/common/util';
import { asyncGetListenerCount } from '@/utils';
// import types
import { CLB_STATUS_MAP, LB_NETWORK_TYPE_MAP } from '@/constants';
import { DoublePlainObject } from '@/typings';
import './index.scss';
import Confirm from '@/components/confirm';
import { useVerify } from '@/hooks';
import { useGlobalPermissionDialog } from '@/store/useGlobalPermissionDialog';
export default defineComponent({
  name: 'AllClbsManager',
  setup() {
    // use hooks
    const router = useRouter();
    const route = useRoute();
    const { t } = useI18n();
    const businessStore = useBusinessStore();
    const { whereAmI } = useWhereAmI();
    const { selections, handleSelectionChange, resetSelections } = useSelection();
    const { authVerifyData, handleAuth } = useVerify();
    const globalPermissionDialogStore = useGlobalPermissionDialog();

    const isRowSelectEnable = ({ row, isCheckAll }: DoublePlainObject) => {
      if (isCheckAll) return true;
      return isCurRowSelectEnable(row);
    };
    const isCurRowSelectEnable = (row: any) => {
      if (whereAmI.value === Senarios.business) return true;
      if (row.id) {
        return row.bk_biz_id === -1;
      }
    };
    const resourceStore = useResourceStore();
    const handleDelete = (data: any) => {
      Confirm('请确定删除负载均衡', `将删除负载均衡【${data.name}】`, async () => {
        await resourceStore
          .deleteBatch('load_balancers', {
            ids: [data.id],
          })
          .then(() => {
            Message({
              message: '删除成功',
              theme: 'success',
            });
            getListData();
          });
      });
    };
    const { columns, settings } = useColumns('lb');
    const { CommonTable, getListData } = useTable({
      searchOptions: {
        searchData: [
          {
            id: 'name',
            name: '负载均衡名称',
          },
          {
            id: 'domain',
            name: '负载均衡域名',
          },
          // {
          //   id: 'public_ipv4_addresses',
          //   name: '负载均衡VIP',
          // },
          {
            id: 'lb_type',
            name: '网络类型',
          },
          // {
          //   id: 'listener_num',
          //   name: '监听器数量',
          // },
          {
            id: 'ip_version',
            name: 'IP版本',
          },
          {
            id: 'vendor',
            name: '云厂商',
          },
          {
            id: 'region',
            name: '地域',
          },
          {
            id: 'zones',
            name: '可用区域',
          },
          {
            id: 'cloud_vpc_id',
            name: '所属VPC',
          },
        ],
      },
      tableOptions: {
        columns: [
          ...columns,
          {
            label: '操作',
            width: 120,
            render: ({ data }: { data: any }) => {
              return (
                <Button
                  text
                  theme='primary'
                  class={`${
                    !authVerifyData?.value?.permissionAction?.load_balancer_delete ? 'hcm-no-permision-text-btn' : ''
                  }`}
                  disabled={
                    authVerifyData?.value?.permissionAction?.load_balancer_delete &&
                    (data.listenerNum > 0 || data.delete_protect)
                  }
                  v-bk-tooltips={
                    data.listenerNum > 0
                      ? { content: '该负载均衡已绑定监听器, 不可删除', disabled: !(data.listenerNum > 0) }
                      : { content: t('该负载均衡已开启删除保护, 不可删除'), disabled: !data.delete_protect }
                  }
                  onClick={() => {
                    if (!authVerifyData?.value?.permissionAction?.load_balancer_delete) {
                      handleAuth('clb_resource_delete');
                      globalPermissionDialogStore.setShow(true);
                    } else handleDelete(data);
                  }}>
                  删除
                </Button>
              );
            },
          },
        ],
        extra: {
          settings: settings.value,
          isRowSelectEnable,
          onSelectionChange: (selections: any) => handleSelectionChange(selections, isCurRowSelectEnable),
          onSelectAll: (selections: any) => handleSelectionChange(selections, isCurRowSelectEnable, true),
          rowClass: getTableNewRowClass(),
        },
      },
      requestOption: {
        type: 'load_balancers/with/delete_protection',
        sortOption: { sort: 'created_at', order: 'DESC' },
        async resolveDataListCb(dataList: any[]) {
          return asyncGetListenerCount(
            businessStore.asyncGetListenerCount,
            dataList.map((item) => {
              item.lb_type = LB_NETWORK_TYPE_MAP[item.lb_type];
              item.status = CLB_STATUS_MAP[item.status];
              return item;
            }),
          );
        },
      },
    });

    const handleApply = () => {
      router.push({
        path: '/business/service/service-apply/clb',
        query: { ...route.query },
      });
    };

    // 批量删除负载均衡
    const {
      isBatchDeleteDialogShow,
      isSubmitLoading,
      isSubmitDisabled,
      radioGroupValue,
      tableProps,
      handleRemoveSelection,
      handleClickBatchDelete,
      handleBatchDeleteSubmit,
      computedListenersList,
    } = useBatchDeleteLB(
      [
        ...columns.slice(1, 7),
        {
          label: '',
          width: 50,
          minWidth: 50,
          render: ({ data }: any) => (
            <Button text onClick={() => handleRemoveSelection(data.id)}>
              <i class='hcm-icon bkhcm-icon-minus-circle-shape'></i>
            </Button>
          ),
        },
      ],
      selections,
      resetSelections,
      getListData,
    );

    return () => (
      <div class='common-card-wrap'>
        {/* 负载均衡list */}
        <CommonTable>
          {{
            operation: () => (
              <>
                <Button
                  class={`mw64 ${
                    !authVerifyData?.value?.permissionAction?.load_balancer_create ? 'hcm-no-permision-btn' : ''
                  }`}
                  theme='primary'
                  onClick={() => {
                    if (!authVerifyData?.value?.permissionAction?.load_balancer_create) {
                      handleAuth('clb_resource_create');
                      globalPermissionDialogStore.setShow(true);
                    } else handleApply();
                  }}>
                  购买
                </Button>
                <Button class='mw88' onClick={handleClickBatchDelete} disabled={selections.value.length === 0}>
                  批量删除
                </Button>
              </>
            ),
          }}
        </CommonTable>
        {/* 批量删除负载均衡 */}
        <BatchOperationDialog
          class='batch-delete-lb-dialog'
          v-model:isShow={isBatchDeleteDialogShow.value}
          title={t('批量删除负载均衡')}
          theme='danger'
          confirmText='删除'
          isSubmitLoading={isSubmitLoading.value}
          isSubmitDisabled={isSubmitDisabled.value}
          tableProps={tableProps}
          list={computedListenersList.value}
          onHandleConfirm={handleBatchDeleteSubmit}>
          {{
            tips: () => (
              <>
                已选择<span class='blue'>{tableProps.data.length}</span>个负载均衡，其中
                <span class='red'>{tableProps.data.filter(({ listenerNum }) => listenerNum > 0).length}</span>
                个存在监听器、
                <span class='red'>{tableProps.data.filter(({ delete_protect }) => delete_protect).length}</span>
                个负载均衡开启了删除保护，不可删除。
              </>
            ),
            tab: () => (
              <BkRadioGroup v-model={radioGroupValue.value}>
                <BkRadioButton label={true}>{t('可删除')}</BkRadioButton>
                <BkRadioButton label={false}>{t('不可删除')}</BkRadioButton>
              </BkRadioGroup>
            ),
          }}
        </BatchOperationDialog>
      </div>
    );
  },
});
