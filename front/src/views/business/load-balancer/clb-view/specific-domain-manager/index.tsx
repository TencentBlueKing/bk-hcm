import { defineComponent, onMounted, onUnmounted, reactive, ref, watch, nextTick } from 'vue';
// import components
import { Button, Form, Input, Link, Message, Select } from 'bkui-vue';
import { Done, Error, Plus } from 'bkui-vue/lib/icon';
import BatchOperationDialog from '@/components/batch-operation-dialog';
import CommonSideslider from '@/components/common-sideslider';
import AddOrUpdateDomainSideslider from '../components/AddOrUpdateDomainSideslider';
import TargetGroupSelector from '../components/TargetGroupSelector';
// use stores
import { useLoadBalancerStore, useBusinessStore, useAccountStore } from '@/store';
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
import { getTableNewRowClass } from '@/common/util';
import bus from '@/common/bus';
import { QueryRuleOPEnum } from '@/typings';
import { SCHEDULER_MAP } from '@/constants';

const { FormItem } = Form;
const { Option } = Select;

export default defineComponent({
  // 导航完成前, 预加载域名所对应监听器以及负载均衡的详情数据, 并存入store中
  async beforeRouteEnter(to, _, next) {
    const businessStore = useBusinessStore();
    const loadBalancerStore = useLoadBalancerStore();
    // 监听器详情
    const { data: listenerDetail } = await businessStore.detail('listeners', to.query.listener_id as string);
    // 负载均衡详情
    const { data: lbDetail } = await businessStore.detail('load_balancers', listenerDetail.lb_id);
    // 当前节点为：监听器
    loadBalancerStore.setCurrentSelectedTreeNode({ ...listenerDetail, lb: lbDetail });
    next();
  },
  // eslint-disable-next-line vue/prop-name-casing
  props: { id: String, listener_id: String },
  setup(props) {
    // use hooks
    const { t } = useI18n();
    // use stores
    const loadBalancerStore = useLoadBalancerStore();
    const formInstance = ref(null);
    const businessStore = useBusinessStore();
    const accountStore = useAccountStore();
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
        field: 'target_group_name',
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
                  <Link
                    class='target-group-name-btn'
                    theme='primary'
                    href={`/#/business/loadbalancer/group-view/${data.target_group_id}?bizs=${accountStore.bizs}&type=detail&vendor=${loadBalancerStore.currentSelectedTreeNode.vendor}`}
                    onClick={() => loadBalancerStore.setTgSearchTarget(cell)}>
                    {cell || '--'}
                  </Link>
                  {/* <span class={'target-group-name-btn'}></span> */}
                  {/* 第一期不支持更新绑定的目标组 */}
                  {/* <EditLine class={'target-group-edit-icon'} onClick={() => (editingID.value = data.id)} /> */}
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
        // sort: true,
        filter: {
          list: Object.keys(CLB_BINDING_STATUS).map((bindingStatus) => ({
            value: bindingStatus,
            text: CLB_BINDING_STATUS[bindingStatus],
          })),
        },
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
                  bus.$emit('resetLbTree');
                });
              }}>
              {t('删除')}
            </Button>
          </div>
        ),
      },
    ];

    const deleteRulesBatch = async (ids: string[]) => {
      isSubmitLoading.value = true;
      try {
        await businessStore.deleteRules(loadBalancerStore.currentSelectedTreeNode.vendor, props.listener_id, {
          lbl_id: props.listener_id,
          rule_ids: ids,
        });
        isBatchDeleteDialogShow.value = false;
        Message({ message: '删除成功', theme: 'success' });
        await getListData();
      } finally {
        isSubmitLoading.value = false;
      }
    };

    const tableSettings = generateColumnsSettings(tableColumns);
    const isCurRowSelectEnable = (row: any) => {
      if (whereAmI.value === Senarios.business) return true;
      if (row.id) {
        return row.bk_biz_id === -1;
      }
    };
    const { CommonTable, getListData, clearFilter } = useTable({
      searchOptions: {
        searchData: [
          { name: 'URL路径', id: 'url' },
          {
            name: '轮询方式',
            id: 'scheduler',
            children: Object.keys(SCHEDULER_MAP).map((scheduler) => ({
              id: scheduler,
              name: SCHEDULER_MAP[scheduler],
            })),
          },
          // todo: 待后端支持
          // { name: '目标组', id: 'target_group_id' },
          // {
          //   name: '同步状态',
          //   id: 'binding_status',
          //   children: Object.keys(CLB_BINDING_STATUS).map((bindingStatus) => ({
          //     id: bindingStatus,
          //     name: CLB_BINDING_STATUS[bindingStatus],
          //   })),
          // },
        ],
      },
      tableOptions: {
        columns: tableColumns,
        extra: {
          settings: tableSettings.value,
          onSelectionChange: (selections: any) => handleSelectionChange(selections, isCurRowSelectEnable),
          onSelectAll: (selections: any) => handleSelectionChange(selections, isCurRowSelectEnable, true),
          rowClass: getTableNewRowClass(),
        },
      },
      requestOption: {
        type: `vendors/${loadBalancerStore.currentSelectedTreeNode.vendor}/listeners/${props.listener_id}/rules`,
        sortOption: { sort: 'created_at', order: 'DESC' },
        filterOption: {
          rules: [{ field: 'domain', op: QueryRuleOPEnum.EQ, value: props.id }],
        },
        async resolveDataListCb(dataList: any) {
          // 如果数据列表为空，直接返回
          if (!dataList?.length) return [];

          // 提取目标组ID列表和规则ID列表
          const tgIds = new Set<string>();
          const ruleIds = new Set<string>();
          dataList.forEach(({ target_group_id, id }: { target_group_id: string; id: string }) => {
            tgIds.add(target_group_id);
            ruleIds.add(id);
          });

          // 并发查询目标组名称和同步状态
          const [targetGroupList, bindingStatusList] = await Promise.all([
            // 查询目标组名称
            businessStore.getTargetGroupList({
              page: { count: false, start: 0, limit: 500 },
              filter: {
                op: QueryRuleOPEnum.AND,
                rules: [{ field: 'id', op: QueryRuleOPEnum.IN, value: Array.from(tgIds) }],
              },
              fields: ['id', 'name'],
            }),
            // 查询同步状态
            loadBalancerStore.queryRulesBindingStatusList(
              loadBalancerStore.currentSelectedTreeNode.vendor,
              loadBalancerStore.currentSelectedTreeNode.id,
              { rule_ids: Array.from(ruleIds) },
            ),
          ]);

          // 构建目标组ID到名称的映射
          const targetGroupMap = new Map<string, string>(
            targetGroupList.data.details.map(({ id, name }: { id: string; name: string }) => [id, name]),
          );

          // 构建规则ID到同步状态的映射
          const bindingStatusMap = new Map<string, string>(
            bindingStatusList.map(({ rule_id, binding_status }: { rule_id: string; binding_status: string }) => [
              rule_id,
              binding_status,
            ]),
          );

          // 返回增强后的数据列表
          return dataList.map((data: any) => ({
            ...data,
            target_group_name: targetGroupMap.get(data.target_group_id) || '--',
            binding_status: bindingStatusMap.get(data.id) || '--',
          }));
        },
      },
    });

    const { isSubmitLoading, isBatchDeleteDialogShow, tableProps, handleBatchDeleteListener } = useBatchDeleteListener(
      tableColumns,
      selections,
      resetSelections,
      getListData,
      true,
    );

    watch(
      () => [props.listener_id, props.id],
      ([id, domain]) => {
        // 清空选中项, 避免切换域名后, 选中项不变
        resetSelections();
        clearFilter();
        id &&
          getListData(
            [
              {
                field: 'domain',
                op: QueryRuleOPEnum.EQ,
                value: domain,
              },
            ],
            `vendors/${loadBalancerStore.currentSelectedTreeNode.vendor}/listeners/${id}/rules`,
          );
      },
    );

    const handleSubmit = async () => {
      await formInstance.value.validate();
      isSubmitLoading.value = true;
      const lbl_id = props.listener_id;
      const { rule_id, url, scheduler, target_group_id } = formData;
      const { vendor } = loadBalancerStore.currentSelectedTreeNode;
      const promise = isEdit.value
        ? businessStore.updateUrl({ lbl_id, rule_id, url, scheduler, target_group_id, vendor })
        : businessStore.createRules({ lbl_id, url, scheduler, domains: [props.id], target_group_id, vendor });
      try {
        await promise;
        Message({
          message: isEdit.value ? '编辑成功' : '创建成功',
          theme: 'success',
        });
        isDomainSidesliderShow.value = false;
        await getListData();
        bus.$emit('resetLbTree');
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

    const targetGroupSelectorRef = ref();
    // 当侧边栏显示时, 刷新目标组select-option-list
    watch(isDomainSidesliderShow, (val) => {
      if (!val) return;
      nextTick(() => {
        targetGroupSelectorRef.value.handleRefresh();
      });
    });

    onMounted(() => {
      bus.$on('showAddUrlSideslider', handleAddUrlSidesliderShow);
    });

    onUnmounted(() => {
      bus.$off('showAddUrlSideslider');
    });

    return () => (
      <div class={'url-list-container has-breadcrumb'}>
        <CommonTable>
          {{
            operation: () => (
              <>
                <Button theme={'primary'} onClick={handleAddUrlSidesliderShow} class='mr8'>
                  <Plus class={'f20'} />
                  {t('新增 URL 路径')}
                </Button>
                <Button onClick={handleBatchDeleteListener} disabled={!selections.value.length}>
                  {t('批量删除')}
                </Button>
              </>
            ),
          }}
        </CommonTable>
        <BatchOperationDialog
          v-model:isShow={isBatchDeleteDialogShow.value}
          title='批量删除 URL 路径'
          theme='danger'
          confirmText='删除'
          tableProps={tableProps}
          list={selections.value}
          isSubmitLoading={isSubmitLoading.value}
          onHandleConfirm={() => deleteRulesBatch(tableProps.data.map(({ id }) => id))}>
          {{
            tips: () => (
              <>
                已选择<span class='blue'>{selections.value.length}</span>个URL路径，可以直接删除。
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
            <span class={'create-url-text-item-value'}>{loadBalancerStore.currentSelectedTreeNode.lb?.name}</span>
          </p>
          <p class={'create-url-text-item'}>
            <span class={'create-url-text-item-label'}>{t('监听器名称')}：</span>
            <span class={'create-url-text-item-value'}>{loadBalancerStore.currentSelectedTreeNode.name}</span>
          </p>
          <p class={'create-url-text-item'}>
            <span class={'create-url-text-item-label'}>{t('域名')}：</span>
            <span class={'create-url-text-item-value'}>{props.id}</span>
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
            {!isEdit.value && (
              <FormItem label={t('目标组')} required>
                <TargetGroupSelector
                  ref={targetGroupSelectorRef}
                  v-model={formData.target_group_id}
                  accountId={loadBalancerStore.currentSelectedTreeNode.account_id}
                  cloudVpcId={loadBalancerStore.currentSelectedTreeNode.lb.cloud_vpc_id}
                  region={loadBalancerStore.currentSelectedTreeNode.lb.region}
                  protocol={loadBalancerStore.currentSelectedTreeNode.protocol}
                />
              </FormItem>
            )}
          </Form>
        </CommonSideslider>
        {/* 编辑域名 */}
        <AddOrUpdateDomainSideslider originPage='domain' />
      </div>
    );
  },
});
