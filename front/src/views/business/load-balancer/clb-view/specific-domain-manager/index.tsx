import { defineComponent, onMounted, onUnmounted, reactive, ref, watch } from 'vue';
// import components
import { Button, Form, Input, Message, Select } from 'bkui-vue';
import { Done, EditLine, Error, Plus } from 'bkui-vue/lib/icon';
import BatchOperationDialog from '@/components/batch-operation-dialog';
import CommonSideslider from '@/components/common-sideslider';
import AddOrUpdateDomainSideslider from '../components/AddOrUpdateDomainSideslider';
// use stores
import { useLoadBalancerStore, useBusinessStore, useResourceStore } from '@/store';
// import custom hooks
import { useTable } from '@/hooks/useTable/useTable';
import useColumns from '@/views/resource/resource-manage/hooks/use-columns';
import { useI18n } from 'vue-i18n';
// import constants
import { CLB_BINDING_STATUS } from '@/common/constant';
// import static resources
import StatusSuccess from '@/assets/image/success-account.png';
import StatusLoading from '@/assets/image/status_loading.png';
import './index.scss';
import { RuleModeList } from '../components/AddOrUpdateDomainSideslider/useAddOrUpdateDomain';
import Confirm from '@/components/confirm';
import useSelection from '@/views/resource/resource-manage/hooks/use-selection';
import { Senarios, useWhereAmI } from '@/hooks/useWhereAmI';
import useBatchDeleteListener from '../specific-clb-manager/listener-list/useBatchDeleteListener';
import { getTableRowClassOption } from '@/common/util';
import bus from '@/common/bus';

const { FormItem } = Form;
const { Option } = Select;

export default defineComponent({
  // eslint-disable-next-line vue/prop-name-casing
  props: { domain: String, listener_id: String },
  setup(props) {
    // use hooks
    const { t } = useI18n();
    // use stores
    const loadBalancerStore = useLoadBalancerStore();
    const formInstance = ref(null);
    const businessStore = useBusinessStore();
    const resourceStore = useResourceStore();
    const isEdit = ref(false);
    const { selections, handleSelectionChange, resetSelections } = useSelection();
    const { whereAmI } = useWhereAmI();
    const isDomainSidesliderShow = ref(false);
    const { columns, generateColumnsSettings } = useColumns('url');
    const editingID = ref('-1');
    const formData = reactive({
      scheduler: '',
      certificate: '',
      url: '',
      rule_id: '',
      target_group_id: '',
    });
    const targetList = ref([]);
    const tableColumns = [
      ...columns,
      {
        label: t('目标组'),
        field: 'target_group_id',
        isDefaultShow: true,
        render: ({ cell, data }: any) => {
          return (
            <div class={'flex-row align-item-center target-group-name'}>
              {editingID.value === data.id ? (
                <div class={'flex-row align-item-center'}>
                  <Select>
                    {targetList.value.map(({ id, name }) => (
                      <Option name={name} id={id} key={id}></Option>
                    ))}
                  </Select>
                  <Done width={20} height={20} class={'submit-edition-icon'} onClick={() => (editingID.value = '-1')} />
                  <Error
                    width={13}
                    height={13}
                    class={'submit-edition-icon'}
                    onClick={() => (editingID.value = '-1')}
                  />
                </div>
              ) : (
                <span class={'flex-row align-item-center'}>
                  <span class={'target-group-name-btn'}>{cell || '--'}</span>
                  <EditLine class={'target-group-edit-icon'} onClick={() => (editingID.value = data.id)} />
                </span>
              )}
            </div>
          );
        },
      },
      {
        label: t('同步状态'),
        field: 'binding_status',
        isDefaultShow: true,
        render: ({ cell }: { cell: string }) => {
          let icon = StatusSuccess;
          switch (cell) {
            case 'binding':
              icon = StatusLoading;
              break;
            case 'success':
              icon = StatusSuccess;
              break;
          }
          return cell ? (
            <div class='status-column-cell'>
              <img class={`status-icon${cell === 'binding' ? ' spin-icon' : ''}`} src={icon} alt='' />
              <span>{CLB_BINDING_STATUS[cell]}</span>
            </div>
          ) : (
            '--'
          );
        },
      },
      {
        label: t('操作'),
        field: 'actions',
        isDefaultShow: true,
        render: ({ data }: any) => (
          <div class='operate-groups'>
            <Button
              text
              theme='primary'
              onClick={() => {
                isEdit.value = true;
                isDomainSidesliderShow.value = true;
                formData.url = data.url;
                formData.scheduler = data.scheduler;
                formData.rule_id = data.id;
              }}>
              {t('编辑')}
            </Button>
            <Button
              text
              theme='primary'
              onClick={() => {
                Confirm('请确定删除URL', `将删除URL【${data.url}】`, async () => {
                  await deleteRulesBatch([data.id]);
                });
              }}>
              {t('删除')}
            </Button>
          </div>
        ),
      },
    ];

    const deleteRulesBatch = async (ids: string[]) => {
      await businessStore.deleteRules(props.listener_id, { lbl_id: props.listener_id, rule_ids: ids });
      Message({
        message: '删除成功',
        theme: 'success',
      });
      await getListData();
      isBatchDeleteDialogShow.value = false;
    };

    const tableSettings = generateColumnsSettings(tableColumns);
    const isCurRowSelectEnable = (row: any) => {
      if (whereAmI.value === Senarios.business) return true;
      if (row.id) {
        return row.bk_biz_id === -1;
      }
    };
    const { CommonTable, getListData } = useTable({
      searchOptions: {
        searchData: [
          {
            name: 'URL路径',
            id: 'url',
          },
          {
            name: '轮询方式',
            id: 'scheduler',
          },
          {
            name: '目标组',
            id: 'target_group_id',
          },
          {
            name: '同步状态',
            id: 'binding_status',
          },
          {
            name: '操作',
            id: 'actions',
          },
        ],
      },
      tableOptions: {
        columns: tableColumns,
        extra: {
          settings: tableSettings.value,
          onSelectionChange: (selections: any) => handleSelectionChange(selections, isCurRowSelectEnable),
          onSelectAll: (selections: any) => handleSelectionChange(selections, isCurRowSelectEnable, true),
          ...getTableRowClassOption(),
        },
      },
      requestOption: {
        type: `vendors/tcloud/listeners/${props.listener_id}/rules`,
        sortOption: { sort: 'created_at', order: 'DESC' },
      },
    });

    const { isSubmitLoading, isBatchDeleteDialogShow, tableProps, handleBatchDeleteListener } = useBatchDeleteListener(
      tableColumns,
      selections,
      resetSelections,
      getListData,
    );

    watch(
      () => props.listener_id,
      (id) => {
        id && getListData([], `vendors/tcloud/listeners/${id}/rules`);
      },
    );

    const handleSubmit = async () => {
      await formInstance.value.validate();
      isSubmitLoading.value = true;
      const promise = isEdit.value
        ? businessStore.updateUrl({
            lbl_id: props.listener_id,
            rule_id: formData.rule_id,
            url: formData.url,
            scheduler: formData.scheduler,
            target_group_id: formData.target_group_id,
          })
        : businessStore.createRules({
            lbl_id: props.listener_id,
            url: formData.url,
            scheduler: formData.scheduler,
            domains: [props.domain],
            target_group_id: formData.target_group_id,
          });
      try {
        await promise;
        Message({
          message: isEdit.value ? '编辑成功' : '创建成功',
          theme: 'success',
        });
        isDomainSidesliderShow.value = false;
        await getListData();
      } finally {
        isSubmitLoading.value = false;
      }
    };

    const resetFormData = () => {
      formData.url = '';
      formData.certificate = '';
      formData.scheduler = '';
    };

    // click-handler - 新增url路径
    const handleAddUrlSidesliderShow = () => {
      isDomainSidesliderShow.value = true;
      isEdit.value = false;
      resetFormData();
    };

    // 获取监听器详情
    const getListenerDetail = async (id: string) => {
      const { data } = await resourceStore.detail('listeners', id);
      loadBalancerStore.setCurrentSelectedTreeNode(data);
    };

    onMounted(() => {
      bus.$on('showAddUrlSideslider', handleAddUrlSidesliderShow);
      getTargetGroupsList();
      !loadBalancerStore.currentSelectedTreeNode?.id && getListenerDetail(props.listener_id);
    });

    onUnmounted(() => {
      bus.$off('showAddUrlSideslider');
    });

    const getTargetGroupsList = async () => {
      const res = await businessStore.getResourceGroupList({
        filter: { op: 'and', rules: [] },
        page: {
          start: 0 * 100,
          limit: 500,
        },
      });
      targetList.value = res.data.details;
    };

    return () => (
      <div class={'url-list-container has-selection has-breadcrumb'}>
        <CommonTable>
          {{
            operation: () => (
              <div class={'flex-row align-item-center'}>
                <Button theme={'primary'} onClick={handleAddUrlSidesliderShow}>
                  <Plus class={'f20'} />
                  {t('新增 URL 路径')}
                </Button>
                <Button onClick={handleBatchDeleteListener} disabled={!selections.value.length}>
                  {t('批量删除')}
                </Button>
              </div>
            ),
          }}
        </CommonTable>
        <BatchOperationDialog
          v-model:isShow={isBatchDeleteDialogShow.value}
          title='批量删除 URL 路径'
          theme='danger'
          confirmText='删除'
          tableProps={tableProps}
          isSubmitLoading={isSubmitLoading.value}
          onHandleConfirm={() => deleteRulesBatch(tableProps.data.map(({ id }) => id))}>
          {{
            tips: () => (
              <>
                已选择 <span class='blue'> {selections.value.length} </span> 个URL路径
              </>
            ),
          }}
        </BatchOperationDialog>
        <CommonSideslider
          title={isEdit.value ? '编辑 URL 路径' : '新增 URL 路径'}
          width={640}
          v-model:isShow={isDomainSidesliderShow.value}
          isSubmitLoading={isSubmitLoading.value}
          onHandleSubmit={handleSubmit}>
          <p class={'create-url-text-item'}>
            <span class={'create-url-text-item-label'}>{t('负载均衡名称')}：</span>
            <span class={'create-url-text-item-value'}>{loadBalancerStore.currentSelectedTreeNode.lb.name}</span>
          </p>
          <p class={'create-url-text-item'}>
            <span class={'create-url-text-item-label'}>{t('监听器名称')}：</span>
            <span class={'create-url-text-item-value'}>{loadBalancerStore.currentSelectedTreeNode.listener.name}</span>
          </p>
          <p class={'create-url-text-item'}>
            <span class={'create-url-text-item-label'}>{t('域名')}：</span>
            <span class={'create-url-text-item-value'}>{props.domain}</span>
          </p>
          <Form formType='vertical' model={formData} ref={formInstance}>
            <FormItem label={t('URL路径')} required property='url'>
              <Input v-model={formData.url} />
            </FormItem>
            <FormItem label={t('均衡方式')} required property='scheduler'>
              <Select v-model={formData.scheduler}>
                {RuleModeList.map(({ id, name }) => (
                  <Option name={name} id={id} />
                ))}
              </Select>
            </FormItem>
            <FormItem label={t('目标组')} required>
              <Select v-model={formData.target_group_id}>
                {targetList.value.map(({ id, name }) => (
                  <Option name={name} id={id} key={id}></Option>
                ))}
              </Select>
            </FormItem>
          </Form>
        </CommonSideslider>
        {/* 编辑域名 */}
        <AddOrUpdateDomainSideslider originPage='domain' />
      </div>
    );
  },
});
