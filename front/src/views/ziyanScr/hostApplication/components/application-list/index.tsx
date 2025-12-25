// 服务请求下的主机申领，业务下单据管理的主机申请已迁移至views/ticket下
import { defineComponent, onMounted, ref, watch, reactive } from 'vue';
import './index.scss';
import { useBusinessMapStore } from '@/store/useBusinessMap';
import { Button, Message, Table, Sideslider } from 'bkui-vue';
import { useTable } from '@/hooks/useTable/useTable';
import useColumns from '@/views/resource/resource-manage/hooks/use-scr-columns';
import { useRoute, useRouter } from 'vue-router';
import routerAction from '@/router/utils/action';
import moment from 'moment';
import WName from '@/components/w-name';
import { Copy, DataShape, HelpDocumentFill } from 'bkui-vue/lib/icon';
import { useApplyStages } from '@/views/ziyanScr/hooks/use-apply-stages';
import CommonSideslider from '@/components/common-sideslider';
import GridContainer from '@/components/layout/grid-container/grid-container.vue';
import GridItemFormElement from '@/components/layout/grid-container/grid-item-form-element.vue';
import GridItem from '@/components/layout/grid-container/grid-item.vue';
import { timeFormatter } from '@/common/util';
import http from '@/http';
import { useZiyanScrStore } from '@/store';
import SuborderDetail from '../suborder-detail';
import CommonDialog from '@/components/common-dialog';
import throttle from 'lodash/throttle';
import MatchPanel from '../match-panel';
import { getRegionCn, getZoneCn } from '@/views/ziyanScr/cvm-web/transform';
import { getResourceTypeName } from '../transform';
import { getTypeCn } from '@/views/ziyanScr/cvm-produce/transform';
import useTimeoutPoll from '@/hooks/use-timeout-poll';
import useSearchQs from '@/hooks/use-search-qs';
import { useBusinessGlobalStore } from '@/store/business-global';
import { getDateRange, transformFlatCondition } from '@/utils/search';
import type { ModelProperty } from '@/model/typings';
import { getModel } from '@/model/manager';
import HocSearch from '@/model/hoc-search.vue';
import { HostApplySearchNonBusiness } from '@/model/order/host-apply-search';
import { serviceShareBizSelectedKey } from '@/constants/storage-symbols';
import { SCR_RESOURCE_TYPE_NAME, ScrResourceType } from '@/constants';
import { RES_ASSIGN_TYPE } from '@/components/device-type-selector/constants';
import type { ICvmDeviceTypeFormData } from '@/components/device-type-selector/typings';
import UserValue from '@/components/display-value/user-value.vue';

export default defineComponent({
  setup() {
    const businessMapStore = useBusinessMapStore();
    const { transformApplyStages } = useApplyStages();
    const machineDetails = ref([]);
    const isMatchPanelShow = ref(false);
    const isDialogShow = ref(false);
    const curRow = ref({});
    const curSuborder = ref({
      step_name: '',
      step_id: 1,
      suborder_id: 0,
    });

    const stageDetailSlideState = reactive({
      show: false,
      suborderId: undefined,
    });

    const scrStore = useZiyanScrStore();
    const businessGlobalStore = useBusinessGlobalStore();

    const reapply = (data: any) => {
      routerAction.redirect(
        {
          path: '/service/hostApplication/apply',
          query: { order_id: data.order_id, unsubmitted: 0 },
        },
        { history: true },
      );
    };
    const modify = (data: any) => {
      router.push({
        path: '/service/hostApplication/modify',
        query: { ...data },
      });
    };

    const { columns } = useColumns('applicationList');
    const router = useRouter();
    const route = useRoute();
    const orderClipboard = ref({});
    columns.splice(3, 0);

    const opBtnDisabled = (row: any) => {
      if (row.stage === 'RUNNING' && row.status === 'MATCHING') {
        return true;
      }
      if (!row.suborder_id) {
        return true;
      }
      if (
        ['wait', 'MATCHED_SOME', 'MATCHING'].includes(row.status) ||
        (row.stage === 'SUSPEND' && row.status === 'TERMINATE')
      ) {
        return false;
      }
      if (['UNCOMMIT', 'PAUSED'].includes(row.status)) {
        return true;
      }
      if (['AUDIT'].includes(row.stage) && !row.status) {
        return true;
      }
      if (['TERMINATE'].includes(row.stage)) {
        return true;
      }
      if (row.stage === 'DONE' && row.status === 'DONE') {
        return true;
      }
      return false;
    };

    // 原opBtnDisabled方法重试与终止操作共用了，可能是一个错误的实现，这里将其拆开
    const retryBtnDisabled = (row: any) => {
      if (row.stage === 'RUNNING' && row.status === 'MATCHING') {
        return true;
      }
      if (!row.suborder_id) {
        return true;
      }
      if (
        ['wait', 'MATCHED_SOME', 'MATCHING'].includes(row.status) ||
        (row.stage === 'SUSPEND' && row.status === 'TERMINATE')
      ) {
        return false;
      }
      if (['UNCOMMIT', 'PAUSED'].includes(row.status)) {
        return true;
      }
      if (['AUDIT'].includes(row.stage) && !row.status) {
        return true;
      }
      if (['TERMINATE', 'CONFIRMING'].includes(row.stage)) {
        return true;
      }
      if (row.stage === 'DONE' && row.status === 'DONE') {
        return true;
      }
      return false;
    };

    const searchFields = getModel(HostApplySearchNonBusiness).getProperties();
    const searchQs = useSearchQs({ key: 'filter', properties: searchFields });

    const { CommonTable, getListData, pagination } = useTable({
      tableOptions: {
        columns: [
          {
            label: '单号/子单号',
            width: 100,
            render: ({ data }: any) => {
              return (
                <div>
                  <div>
                    <Button theme='primary' text onClick={() => getOrderRoute(data)}>
                      {data.order_id}
                    </Button>
                  </div>
                  <div>
                    <p>{data.suborder_id || '无'}</p>
                  </div>
                  {data.source === 'purchase_to_resource_pool' && (
                    <div>
                      <bk-tag theme='info'>资源池</bk-tag>
                    </div>
                  )}
                </div>
              );
            },
          },
          {
            label: '业务',
            render: ({ data }: any) =>
              businessMapStore.getNameFromBusinessMap(data.bk_biz_id) || data.bk_biz_id || '--',
          },
          {
            label: '单据状态',
            field: 'stage',
            width: 170,
            render: ({ data }: any) => {
              const { create_at, stage, resource_type, spec } = data;
              const diffHours = moment(new Date()).diff(moment(create_at), 'hours');
              const isAbnormal = diffHours >= 2 && stage === 'RUNNING';
              const resourceTypeName = SCR_RESOURCE_TYPE_NAME[resource_type as keyof typeof ScrResourceType];

              const isIdcpm = resource_type === ScrResourceType.IDCPM;
              const isUpgradeCvm = ScrResourceType.UPGRADECVM === resource_type;

              const stageClass = (stage: string) => {
                if (stage === 'UNCOMMIT') return 'c-text-3';
                if (stage === 'AUDIT') return 'c-text-2';
                if (stage === 'DONE') return 'c-success';
                if (isAbnormal) return 'c-warning';
                if (stage === 'RUNNING') return 'c-text-1';
                if (stage === 'TERMINATE') return 'c-danger';
                if (stage === 'SUSPEND') return 'c-danger';
              };

              const abnormalStatus = () => {
                if (stage === 'SUSPEND') {
                  if (isUpgradeCvm) return '备货状态异常';
                  return (
                    <div
                      class={'flex-row align-item-center'}
                      v-bk-tooltips={{
                        content: (
                          <>
                            {spec?.failed_zone_ids?.length > 0 && (
                              <div>
                                可用区：
                                {[...new Set(spec.failed_zone_ids)].map((zone) => getZoneCn(zone)).join('，')}
                                。已尝试生产主机失败
                              </div>
                            )}
                            <span>建议：修改需求重试</span>
                          </>
                        ),
                      }}>
                      备货状态异常 <HelpDocumentFill fill='#ffbb00' width={12} height={12} class={'ml4'} />
                    </div>
                  );
                }
                return null;
              };

              const modifyButton = () => {
                const isDisabled = isIdcpm || isUpgradeCvm;
                const tooltipsOption = {
                  // eslint-disable-next-line no-nested-ternary
                  disabled: isDisabled ? (isIdcpm ? !isIdcpm : !isUpgradeCvm) : true,
                  content: `${resourceTypeName}不支持修改,请联系ICR(IEG资源服务助手)`,
                };

                return (
                  <Button
                    class='mr8'
                    theme='primary'
                    text
                    size='small'
                    disabled={isDisabled}
                    v-bk-tooltips={tooltipsOption}
                    onClick={() => modify(data)}>
                    修改需求重试
                  </Button>
                );
              };

              const progressButton = () => {
                return (
                  <Button
                    theme='primary'
                    text
                    size='small'
                    onClick={async () => {
                      stageDetailSlideState.show = true;
                      stageDetailSlideState.suborderId = data.suborder_id;
                      const { data: list } = await getMatchDetails(data.suborder_id);
                      machineDetails.value = list.info;
                    }}>
                    查看详情
                  </Button>
                );
              };

              return (
                <div>
                  <p class={stageClass(stage)}>
                    {stage !== 'SUSPEND' && transformApplyStages(stage)}
                    {abnormalStatus()}
                  </p>
                  <p>
                    {stage === 'SUSPEND' ? modifyButton() : null}
                    {['RUNNING', 'DONE', 'SUSPEND'].includes(stage) ? progressButton() : null}
                  </p>
                </div>
              );
            },
          },
          {
            label: '需求类型',
            field: 'require_type',
            width: 100,
            render: ({ data }: any) => getTypeCn(data.require_type),
          },
          {
            label: '需求摘要',
            width: 250,
            render: ({ data }: any) => {
              const isUpgradeCvm = ScrResourceType.UPGRADECVM === data.resource_type;
              return (
                <div>
                  <div>资源类型：{data?.resource_type ? getResourceTypeName(data?.resource_type) : '--'}</div>
                  {/* 机型配置调整不展示机型、园区 */}
                  {!isUpgradeCvm && (
                    <>
                      <div>机型：{data.spec?.device_type || '--'}</div>
                      <div style={{ display: 'grid', gridTemplateColumns: 'auto 1fr' }}>
                        <div>园区：</div>
                        <div>
                          {data.spec?.zones?.[0] === 'all'
                            ? `${getRegionCn(data.spec.region)}(全部可用区)`
                            : data.spec?.zones?.map((zone: string) => {
                                return <div>{getZoneCn(zone)}</div>;
                              }) || '--'}
                        </div>
                      </div>
                      <div>
                        分布：
                        {RES_ASSIGN_TYPE[data.spec?.res_assign as ICvmDeviceTypeFormData['resAssignType']]?.label ??
                          '--'}
                      </div>
                    </>
                  )}
                </div>
              );
            },
          },
          {
            label: '申请人',
            width: 150,
            render: ({ data }: any) => {
              return (
                <WName name={data.bk_username}>
                  <UserValue value={data.bk_username} />
                </WName>
              );
            },
          },
          {
            label: `需求数`,
            width: 120,
            field: 'total_num',
            render: ({ row, cell }: any) => {
              if (row.modify_time > 0) {
                return `${row.total_num}(原需求数${row.origin_num})`;
              }
              return cell;
            },
          },
          {
            label: '待交付数',
            width: 90,
            field: 'pending_num',
            render({ cell, data }: any) {
              const isUpgradeCvm = ScrResourceType.UPGRADECVM === data.resource_type;
              return !isUpgradeCvm && cell > 0 ? (
                <Button
                  theme='primary'
                  text
                  onClick={() => {
                    curRow.value = data;
                    isMatchPanelShow.value = true;
                  }}>
                  {cell}
                </Button>
              ) : (
                cell
              );
            },
          },
          {
            label: '已交付数',
            field: 'success_num',
            width: 150,
            render: ({ data }: any) => {
              if (data.success_num > 0) {
                const ips = orderClipboard.value?.[data.suborder_id]?.ips || [];
                const assetIds = orderClipboard.value?.[data.suborder_id]?.assetIds || [];
                const goToCmdb = (ips: string[]) => {
                  window.open(`http://bkcc.oa.com/#/business/${data.bk_biz_id}/index?ip=text=${ips.join(',')}`);
                };

                return (
                  <div class={'flex-row align-item-center'}>
                    {data.success_num}
                    <Button
                      text
                      theme={'primary'}
                      class='ml8 mr8'
                      v-clipboard:copy={ips.join('\n')}
                      v-bk-tooltips={{
                        content: '复制 IP',
                      }}>
                      <Copy />
                    </Button>
                    <Button
                      text
                      theme={'primary'}
                      class='mr8'
                      v-clipboard:copy={assetIds.join('\n')}
                      v-bk-tooltips={{
                        content: '复制固资号',
                      }}>
                      <Copy />
                    </Button>
                    <Button
                      text
                      theme={'primary'}
                      onClick={() => goToCmdb(ips)}
                      v-bk-tooltips={{
                        content: '去蓝鲸配置平台管理资源',
                      }}>
                      <DataShape />
                    </Button>
                  </div>
                );
              }

              return <span>{data.success_num}</span>;
            },
          },
          ...columns,
          {
            label: '操作',
            fixed: 'right',
            width: 200,
            render: ({ row }: any) => {
              const isUpgradeCvm = ScrResourceType.UPGRADECVM === row.resource_type;
              return (
                <div>
                  <Button
                    // 滚服项目暂不支持再次申请
                    disabled={row.status === 'UNCOMMIT' || row.require_type === 6 || isUpgradeCvm}
                    size='small'
                    onClick={() => reapply(row)}
                    text
                    theme={'primary'}
                    class='mr8'>
                    再次申请
                  </Button>
                  <Button
                    size='small'
                    text
                    theme={'primary'}
                    class='mr8'
                    disabled={retryBtnDisabled(row) || isUpgradeCvm}
                    onClick={async () => {
                      await scrStore.retryOrder({ suborder_id: [row.suborder_id] });
                      Message({ theme: 'success', message: '重试成功' });
                      getListData();
                    }}>
                    重试
                  </Button>
                  <Button
                    size='small'
                    text
                    theme={'primary'}
                    class='mr8'
                    disabled={opBtnDisabled(row)}
                    onClick={async () => {
                      await scrStore.stopOrder({ suborder_id: [row.suborder_id] });
                      Message({ theme: 'success', message: '终止成功' });
                      getListData();
                    }}>
                    终止
                  </Button>
                </div>
              );
            },
          },
        ],
        extra: {
          onRowMouseEnter: (e, row) => {
            handleCellMouseEnter(row);
          },
          rowHeight: 24,
        },
      },
      requestOption: {
        dataPath: 'data.info',
        immediate: false,
      },
      scrConfig: () => {
        const payload = transformFlatCondition(condition.value, searchFields);
        if (payload.bk_biz_id?.[0] === 0) {
          payload.bk_biz_id = businessGlobalStore.businessAuthorizedList.map((item: any) => item.id);
        }
        return {
          url: '/api/v1/woa/task/findmany/apply',
          payload,
        };
      },
    });

    const condition = ref<Record<string, any>>({});
    const searchValues = ref<Record<string, any>>({});

    const getSearchCompProps = (field: ModelProperty) => {
      if (field.type === 'business') {
        return {
          scope: 'auth',
          showAll: true,
          emptySelectAll: true,
          cacheKey: serviceShareBizSelectedKey,
        };
      }
      if (field.type === 'enum') {
        return {
          option: field.option,
        };
      }
      if (field.id === 'create_at') {
        return {
          type: 'daterange',
          format: 'yyyy-MM-dd',
          clearable: false,
        };
      }
      if (field.id === 'order_id') {
        return {
          collapseTags: true,
          pasteFn: (value: string) =>
            value
              .split(/\r\n|\n|\r/)
              .filter((tag) => /^\d+(-\d+)?$/.test(tag)) // 匹配纯数字或数字-数字格式
              .map((tag) => ({ id: tag, name: tag })),
          placeholder: '请输入主单号/子单号',
        };
      }
      return {};
    };

    const handleSearch = () => {
      // TODO: 实际无效
      pagination.start = 0;

      // 将子单号从主单号条件中分离
      const { order_id: orderId, bk_biz_id: bkBizId, ...rest } = searchValues.value;
      const orderIds = orderId.filter((item: string) => /^\d+$/.test(item));
      const suborderIds = orderId.filter((item: string) => /^\d+-\d+$/.test(item));

      searchQs.set({
        ...rest,
        order_id: orderIds,
        suborder_id: suborderIds,
        bk_biz_id: bkBizId?.length ? bkBizId : [0],
      });
    };

    const handleReset = () => {
      searchQs.clear();
    };

    watch(
      () => route.query,
      async (query) => {
        const defaultCondition = {
          create_at: getDateRange('last30d', true),
          bk_biz_id: businessGlobalStore.getCacheSelected(serviceShareBizSelectedKey) ?? [0],
        };
        condition.value = searchQs.get(query, defaultCondition);

        // 将子单号合并到主单号条件中
        const { order_id: orderId, suborder_id: suborderId, ...rest } = condition.value;
        searchValues.value = { ...rest, order_id: [...(orderId || []), ...(suborderId || [])] };

        getListData();
      },
      { immediate: true },
    );

    watch(
      () => stageDetailSlideState.show,
      (val) => {
        if (val) {
          stageDetailPolling.resume();
        } else {
          stageDetailPolling.reset();
        }
      },
    );

    const stageDetailPolling = useTimeoutPoll(
      async () => {
        const { data: list } = await getMatchDetails(stageDetailSlideState.suborderId);
        machineDetails.value = list.info;
      },
      30000,
      {
        max: 60,
      },
    );

    const getOrderRoute = (row: any) => {
      const { order_id, bk_biz_id, bk_username, resource_type, stage } = row;
      let routeParams: any = {
        name: 'host-application-detail',
        params: { id: order_id },
        query: { creator: bk_username, bkBizId: bk_biz_id, resource_type },
      };
      if (stage === 'UNCOMMIT') {
        routeParams = { path: '/service/hostApplication/apply', query: { order_id, unsubmitted: 1 } };
      }
      routerAction.redirect(routeParams, { history: true });
    };
    // 获取匹配详情
    const getMatchDetails = async (subOrderId: number) => {
      return http.post('/api/v1/woa/task/find/apply/detail', {
        suborder_id: subOrderId,
      });
    };
    // 已交付设备
    const getDeliveredDevices = (params) => {
      return http.post('/api/v1/woa/task/findmany/apply/device', params);
    };
    // 查询交付IP和固号IP
    const getDeliveredHostField = (row, fieldKey) => {
      const params = {
        filter: {
          condition: 'AND',
          rules: [
            {
              field: 'suborder_id',
              operator: 'equal',
              value: row.suborder_id,
            },
            {
              field: 'bk_biz_id',
              operator: 'in',
              value: [row.bk_biz_id],
            },
          ],
        },
      };
      return getDeliveredDevices(params).then((res) => {
        const value = res?.data?.info?.map((item) => item[fieldKey]) || [];
        return value;
      });
    };
    const throttleInfo = ref(null);
    const throttleDeliveredHostField = () => {
      throttleInfo.value = throttle(async (row) => {
        const [ips, assetIds] = await Promise.all([
          getDeliveredHostField(row, 'ip'),
          getDeliveredHostField(row, 'asset_id'),
        ]);
        orderClipboard.value[row.suborder_id] = {
          ips,
          assetIds,
        };
      }, 200);
    };
    const handleCellMouseEnter = (row) => {
      if (row.success_num > 0) {
        throttleInfo.value(row);
      }
    };

    onMounted(() => {
      throttleDeliveredHostField();
    });
    return () => (
      <div class={'apply-list-container scr-application-list'}>
        <div class={'filter-container'} style={{ margin: '0 24px 20px 24px' }}>
          <GridContainer layout='vertical' column={4} content-min-width={'1fr'} gap={[16, 60]}>
            {searchFields
              // 子单号不单独作为一个搜索框，而是集成到主单号框内
              .filter((field) => field.id !== 'suborder_id')
              .map((field) => (
                <GridItemFormElement key={field.id} label={field.name}>
                  <HocSearch
                    is={field.type}
                    display={field.meta?.display}
                    v-model={searchValues.value[field.id]}
                    {...getSearchCompProps(field)}
                  />
                </GridItemFormElement>
              ))}
            <GridItem span={4}>
              <div style={{ display: 'flex', gap: '8px' }}>
                <bk-button theme='primary' style={{ minWidth: '86px' }} onClick={handleSearch}>
                  查询
                </bk-button>
                <bk-button style={{ minWidth: '86px' }} onClick={handleReset}>
                  重置
                </bk-button>
              </div>
            </GridItem>
          </GridContainer>
        </div>
        <div class='btn-container oper-btn-pad'>
          <Button
            theme='primary'
            onClick={() => {
              router.push({
                path: '/service/hostApplication/apply',
                query: route.query,
              });
            }}>
            新增申请
          </Button>
        </div>
        <CommonTable />
        <CommonSideslider v-model:isShow={stageDetailSlideState.show} title='资源匹配详情' width={1100} noFooter>
          <Table
            showOverflowTooltip
            border={['outer', 'col', 'row']}
            data={machineDetails.value}
            columns={[
              {
                field: 'step_id',
                label: 'ID',
                width: 40,
              },
              {
                field: 'step_name',
                label: '步骤名称',
                width: '100',
              },
              {
                field: 'status',
                label: '状态',
                width: 80,
                render({ data }: any) {
                  if (data.status === -1) return <span>未执行</span>;
                  if (data.status === 0) return <span>成功</span>;
                  if (data.status === 1) return <span>执行中</span>;
                  return <span>失败</span>;
                },
              },
              {
                field: 'message',
                label: '状态说明',
                width: 100,
              },
              {
                label: '概要',
                width: '250',
                render({ data }: any) {
                  return (
                    <div>
                      <span>
                        <span class='c-text-2 fz-12'>总数：</span>
                        <span>{data.total_num || '-'}</span>
                      </span>
                      <span class='ml-10'>
                        <span class='c-text-2 fz-12'>成功：</span>
                        <span class='c-success'>{data.success_num || '-'}</span>
                      </span>
                      <span class='ml-10'>
                        <span class='c-text-2 fz-12'>进行中：</span>
                        <span>{data.running_num || '-'}</span>
                      </span>
                      <span class='ml-10'>
                        <span class='c-text-2 fz-12'>失败：</span>
                        <span class='c-danger'>{data.fail_num || '-'}</span>
                      </span>
                    </div>
                  );
                },
              },
              {
                field: 'start_at',
                label: '开始时间',
                width: 160,
                render: ({ data }: any) => (data.status === -1 ? '-' : timeFormatter(data.start_at)),
              },
              {
                field: 'end_at',
                label: '结束时间',
                width: 160,
                render: ({ data }: any) => (![0, 2].includes(data.status) ? '-' : timeFormatter(data.end_at)),
              },
              {
                field: 'operation',
                label: '操作',
                render: ({ data }: any) => (
                  <div>
                    {data.step_id > 1 ? (
                      <Button
                        text
                        theme='primary'
                        onClick={() => {
                          isDialogShow.value = true;
                          curSuborder.value = data;
                        }}>
                        查看详情
                      </Button>
                    ) : (
                      '--'
                    )}
                  </div>
                ),
              },
            ]}></Table>
        </CommonSideslider>

        <CommonDialog v-model:isShow={isDialogShow.value} title={`资源${curSuborder.value.step_name}详情`} width={1200}>
          <SuborderDetail
            suborderId={curSuborder.value.suborder_id}
            stepId={curSuborder.value.step_id}
            isShow={isDialogShow.value}
          />
        </CommonDialog>

        <Sideslider v-model:isShow={isMatchPanelShow.value} title='待匹配' width={1600} renderDirective='if'>
          <MatchPanel data={curRow.value} handleClose={() => (isMatchPanelShow.value = false)} />
        </Sideslider>
      </div>
    );
  },
});
