import { ComputedRef, defineComponent, inject, reactive, useTemplateRef, watch } from 'vue';
import { useRouter, useRoute } from 'vue-router';
// import components
import { Button, Message } from 'bkui-vue';
import { BkRadioButton, BkRadioGroup } from 'bkui-vue/lib/radio';
import BatchOperationDialog from '@/components/batch-operation-dialog';
import BatchImportComp from './batch-import-comp/index.vue';
import SyncAccountResource from '@/components/sync-account-resource/index.vue';
import BatchCopy from './batch-copy.vue';
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
import { asyncGetListenerCount, buildMultipleValueRulesItem, parseIP } from '@/utils';
// import types
import { CLB_STATUS_MAP, LB_NETWORK_TYPE_MAP } from '@/constants';
import { DoublePlainObject } from '@/typings';
import './index.scss';
import Confirm from '@/components/confirm';
import { useVerify } from '@/hooks';
import { useGlobalPermissionDialog } from '@/store/useGlobalPermissionDialog';
import { ResourceTypeEnum, VendorEnum, VendorMap } from '@/common/constant';
import { ValidateValuesFunc } from 'bkui-vue/lib/search-select/utils';

export default defineComponent({
  name: 'AllClbsManager',
  setup() {
    // use hooks
    const router = useRouter();
    const route = useRoute();
    const { t } = useI18n();
    const businessStore = useBusinessStore();
    const { whereAmI, getBizsId } = useWhereAmI();
    const { selections, handleSelectionChange, resetSelections } = useSelection();
    const { authVerifyData, handleAuth } = useVerify();
    const globalPermissionDialogStore = useGlobalPermissionDialog();
    const createClbActionName: ComputedRef<'clb_resource_create' | 'biz_clb_resource_create'> =
      inject('createClbActionName');
    const deleteClbActionName: ComputedRef<'clb_resource_delete' | 'biz_clb_resource_delete'> =
      inject('deleteClbActionName');

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
        await resourceStore.deleteBatch('load_balancers', { ids: [data.id] });
        Message({ message: '删除成功', theme: 'success' });
        getListData();
      });
    };
    const { columns, settings } = useColumns('lb');
    const validateValues: ValidateValuesFunc = async (item, values) => {
      if (!item) return '请选择条件';
      if ('lb_vip' === item.id) {
        const { IPv4List, IPv6List } = parseIP(values[0].id);
        return Boolean(IPv4List.length || IPv6List.length) ? true : 'IP格式有误';
      }
      return true;
    };
    const { CommonTable, getListData, dataList } = useTable({
      searchOptions: {
        searchData: [
          { id: 'name', name: '负载均衡名称' },
          { id: 'cloud_id', name: '负载均衡ID' },
          { id: 'domain', name: '负载均衡域名' },
          { id: 'lb_vip', name: '负载均衡VIP' },
          {
            id: 'lb_type',
            name: '网络类型',
            children: Object.keys(LB_NETWORK_TYPE_MAP).map((lbType) => ({
              id: lbType,
              name: LB_NETWORK_TYPE_MAP[lbType as keyof typeof LB_NETWORK_TYPE_MAP],
            })),
          },
          {
            id: 'ip_version',
            name: t('IP版本'),
            children: [
              { id: 'ipv4', name: 'IPv4' },
              { id: 'ipv6', name: 'IPv6' },
              { id: 'ipv6_dual_stack', name: 'IPv6DualStack' },
              { id: 'ipv6_nat64', name: 'IPv6Nat64' },
            ],
          },
          {
            id: 'vendor',
            name: t('云厂商'),
            children: [{ id: VendorEnum.TCLOUD, name: VendorMap[VendorEnum.TCLOUD] }],
          },
          { id: 'zones', name: '可用区域' },
          {
            id: 'status',
            name: '状态',
            children: Object.keys(CLB_STATUS_MAP).map((key) => ({ id: key, name: CLB_STATUS_MAP[key] })),
          },
          { id: 'cloud_vpc_id', name: '所属VPC' },
        ],
        extra: {
          validateValues,
        },
        conditionFormatterMapper: {
          cloud_id: (value: string) => {
            return buildMultipleValueRulesItem('cloud_id', value);
          },
        },
      },
      tableOptions: {
        columns: [
          ...columns,
          {
            label: '操作',
            width: 120,
            fixed: 'right',
            render: ({ data }: { data: any }) => {
              return (
                <Button
                  text
                  theme='primary'
                  class={{
                    'hcm-no-permision-text-btn': !authVerifyData?.value?.permissionAction?.[deleteClbActionName.value],
                  }}
                  disabled={
                    authVerifyData?.value?.permissionAction?.[deleteClbActionName.value] &&
                    (data.listenerNum > 0 || data.delete_protect)
                  }
                  v-bk-tooltips={
                    data.listenerNum > 0
                      ? { content: '该负载均衡已绑定监听器, 不可删除', disabled: !(data.listenerNum > 0) }
                      : { content: t('该负载均衡已开启删除保护, 不可删除'), disabled: !data.delete_protect }
                  }
                  onClick={() => {
                    if (!authVerifyData?.value?.permissionAction?.[deleteClbActionName.value]) {
                      handleAuth(deleteClbActionName.value);
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
          return asyncGetListenerCount(businessStore.asyncGetListenerCount, dataList);
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
      getListData,
    );

    const tableRef = useTemplateRef<typeof CommonTable>('table-comp');
    const clearSelection = () => {
      resetSelections();
      tableRef.value?.clearSelection();
    };
    watch(
      () => dataList.value,
      () => {
        clearSelection();
      },
    );

    const syncDialogState = reactive({ isShow: false, isHidden: true, businessId: undefined });
    const handleSync = () => {
      syncDialogState.isShow = true;
      syncDialogState.isHidden = false;
      syncDialogState.businessId = getBizsId();
    };

    return () => (
      <div class='common-card-wrap'>
        {/* 负载均衡list */}
        <CommonTable ref='table-comp'>
          {{
            operation: () => (
              <>
                <Button
                  class={`mw64 ${
                    !authVerifyData?.value?.permissionAction?.[createClbActionName.value] ? 'hcm-no-permision-btn' : ''
                  }`}
                  theme='primary'
                  onClick={() => {
                    if (!authVerifyData?.value?.permissionAction?.[createClbActionName.value]) {
                      handleAuth(createClbActionName.value);
                      globalPermissionDialogStore.setShow(true);
                    } else handleApply();
                  }}>
                  购买
                </Button>
                <Button
                  class={[
                    'mw88',
                    { 'hcm-no-permision-btn': !authVerifyData?.value?.permissionAction?.[deleteClbActionName.value] },
                  ]}
                  onClick={() => {
                    if (!authVerifyData?.value?.permissionAction?.[deleteClbActionName.value]) {
                      handleAuth(deleteClbActionName.value);
                      globalPermissionDialogStore.setShow(true);
                      return;
                    }
                    handleClickBatchDelete();
                  }}
                  disabled={selections.value.length === 0}>
                  批量删除
                </Button>
                {/* 批量导入 */}
                <BatchImportComp />
                <bk-button disabled={selections.value.length > 0} onClick={handleSync}>
                  同步负载均衡
                </bk-button>
                <BatchCopy selections={selections.value} />
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
        {!syncDialogState.isHidden && (
          <SyncAccountResource
            v-model={syncDialogState.isShow}
            title='同步负载均衡'
            desc='从云上同步该业务的所有负载均衡数据，包括负载均衡，监听器等'
            resourceType={ResourceTypeEnum.CLB}
            businessId={syncDialogState.businessId}
            resourceName='load_balancer'
            onHidden={() => {
              syncDialogState.isHidden = true;
              syncDialogState.businessId = undefined;
            }}
          />
        )}
      </div>
    );
  },
});
