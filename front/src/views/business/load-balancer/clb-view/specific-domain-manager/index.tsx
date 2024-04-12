import { defineComponent, onMounted, onUnmounted, reactive, ref, watch } from 'vue';
// import components
import { Button, Form, Input, Message, Select } from 'bkui-vue';
import { Done, EditLine, Error, Plus, Spinner } from 'bkui-vue/lib/icon';
import BatchOperationDialog from '@/components/batch-operation-dialog';
import CommonSideslider from '@/components/common-sideslider';
// use stores
import { useLoadBalancerStore } from '@/store/loadbalancer';
// import custom hooks
import { useTable } from '@/hooks/useTable/useTable';
import useColumns from '@/views/resource/resource-manage/hooks/use-columns';
import { useI18n } from 'vue-i18n';
// import constants
import { SYNC_STAUS_MAP } from '@/common/constant';
// import static resources
import StatusSuccess from '@/assets/image/success-account.png';
import StatusFailure from '@/assets/image/failed-account.png';
import StatusPartialSuccess from '@/assets/image/result-waiting.png';
import './index.scss';
import { RuleModeList } from '../specific-listener-manager/domain-list/useAddOrUpdateDomain';
import { useBusinessStore } from '@/store';
import Confirm from '@/components/confirm';
import useSelection from '@/views/resource/resource-manage/hooks/use-selection';
import { Senarios, useWhereAmI } from '@/hooks/useWhereAmI';
import useBatchDeleteListener from '../specific-clb-manager/listener-list/useBatchDeleteListener';
import bus from '@/common/bus';

const { FormItem } = Form;
const { Option } = Select;

export default defineComponent({
  setup() {
    // use hooks
    const { t } = useI18n();
    // use stores
    const loadBalancerStore = useLoadBalancerStore();
    const formInstance = ref(null);
    const businessStore = useBusinessStore();
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
    });
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
                  <Select />
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
        field: 'syncStatus',
        isDefaultShow: true,
        render: ({ cell }: any) => {
          let icon = StatusFailure;
          switch (cell) {
            case 'b': {
              icon = StatusSuccess;
              break;
            }
            case 'c': {
              icon = StatusFailure;
              break;
            }
            case 'd': {
              icon = StatusPartialSuccess;
              break;
            }
          }
          return (
            <div class={'url-status-container'}>
              {cell === 'a' ? (
                <Spinner fill='#3A84FF' width={13} height={13} class={'mr6'} />
              ) : (
                <img src={icon} class='mr6' width={13} height={13} />
              )}
              <span
                class={`${cell === 'd' ? 'url-sync-partial-success-status' : ''}`}
                v-bk-tooltips={{
                  content: '成功 89 个，失败 105 个',
                  disabled: cell !== 'd',
                }}>
                {SYNC_STAUS_MAP[cell]}
              </span>
            </div>
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
      await businessStore.deleteRules(loadBalancerStore.currentSelectedTreeNode.listener.id, {
        lbl_id: loadBalancerStore.currentSelectedTreeNode.listener.id,
        rule_ids: ids,
      });
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
            id: 'status',
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
          'row-class': ({ syncStatus }: { syncStatus: string }) => {
            if (syncStatus === 'a') {
              return 'binding-row';
            }
          },
          onSelectionChange: (selections: any) => handleSelectionChange(selections, isCurRowSelectEnable),
          onSelectAll: (selections: any) => handleSelectionChange(selections, isCurRowSelectEnable, true),
        },
      },
      requestOption: {
        type: `vendors/tcloud/listeners/${loadBalancerStore.currentSelectedTreeNode.listener_id}/rules`,
      },
    });

    const { isSubmitLoading, isBatchDeleteDialogShow, tableProps, handleBatchDeleteListener } = useBatchDeleteListener(
      tableColumns,
      selections,
      resetSelections,
      getListData,
    );

    watch(
      () => loadBalancerStore.currentSelectedTreeNode,
      (val) => {
        const { listener_id, type } = val;
        if (type !== 'domain') return;
        // 只有 type='domain' 时, 才去请求对应 listener+domain 下的 url 列表
        getListData([], `vendors/tcloud/listeners/${listener_id}/rules`);
      },
    );
    const handleSubmit = async () => {
      await formInstance.value.validate();
      isSubmitLoading.value = true;
      const promise = isEdit.value
        ? businessStore.updateUrl({
            lbl_id: loadBalancerStore.currentSelectedTreeNode.listener.id,
            rule_id: formData.rule_id,
            url: formData.url,
            scheduler: formData.scheduler,
          })
        : businessStore.createRules({
            lbl_id: loadBalancerStore.currentSelectedTreeNode.listener.id,
            rules: [
              {
                url: formData.url,
                scheduler: formData.scheduler,
                domains: [loadBalancerStore.currentSelectedTreeNode.domain],
              },
            ],
          });
      await promise;
      Message({
        message: isEdit.value ? '编辑成功' : '创建成功',
        theme: 'success',
      });
      isDomainSidesliderShow.value = false;
      isSubmitLoading.value = false;
      await getListData();
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

    onMounted(() => {
      bus.$on('showAddUrlSideslider', handleAddUrlSidesliderShow);
    });

    onUnmounted(() => {
      bus.$off('showAddUrlSideslider');
    });

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
            <span class={'create-url-text-item-label'}>{t('监听器名称')}：</span>
            <span class={'create-url-text-item-value'}>{loadBalancerStore.currentSelectedTreeNode.listener.name}</span>
          </p>
          <p class={'create-url-text-item'}>
            <span class={'create-url-text-item-label'}>{t('域名')}：</span>
            <span class={'create-url-text-item-value'}>{loadBalancerStore.currentSelectedTreeNode.domain}</span>
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
              <Select
                disabled
                v-bk-tooltips={{
                  content: '暂不支持 todo',
                }}
              />
            </FormItem>
          </Form>
        </CommonSideslider>
      </div>
    );
  },
});
