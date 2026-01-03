import { defineComponent, ref, watch, onMounted, reactive } from 'vue';
import { RouteLocationRaw, useRoute, useRouter } from 'vue-router';
import routerAction from '@/router/utils/action';
import cssModule from './index.module.scss';

import { Button, Message } from 'bkui-vue';
import { Copy, DataShape, HelpDocumentFill } from 'bkui-vue/lib/icon';
import GridContainer from '@/components/layout/grid-container/grid-container.vue';
import GridItemFormElement from '@/components/layout/grid-container/grid-item-form-element.vue';
import GridItem from '@/components/layout/grid-container/grid-item.vue';
import WName from '@/components/w-name';
import StageDetailSideslider from './stage-detail';

import moment from 'moment';
import { useI18n } from 'vue-i18n';
import throttle from 'lodash/throttle';
import { useUserStore, useZiyanScrStore } from '@/store';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import useScrColumns from '@/views/resource/resource-manage/hooks/use-scr-columns';
import { useTable } from '@/hooks/useTable/useTable';
import useSearchQs from '@/hooks/use-search-qs';
import { useRequireTypes } from '@/views/ziyanScr/hooks/use-require-types';
import { useApplyStages } from '@/views/ziyanScr/hooks/use-apply-stages';
import { getResourceTypeName } from '@/views/ziyanScr/hostApplication/components/transform';
import { getRegionCn, getZoneCn } from '@/views/ziyanScr/cvm-web/transform';
import http from '@/http';
import useTimeoutPoll from '@/hooks/use-timeout-poll';
import { getDateRange, transformFlatCondition } from '@/utils/search';
import type { ModelProperty } from '@/model/typings';
import { getModel } from '@/model/manager';
import HocSearch from '@/model/hoc-search.vue';
import { HostApplySearch } from '@/model/order/host-apply-search';
import { GLOBAL_BIZS_KEY } from '@/common/constant';
import { SCR_RESOURCE_TYPE_NAME, ScrResourceType } from '@/constants';
import { RES_ASSIGN_TYPE } from '@/components/device-type-selector/constants';
import type { ICvmDeviceTypeFormData } from '@/components/device-type-selector/typings';
import UserValue from '@/components/display-value/user-value.vue';

const { BK_HCM_AJAX_URL_PREFIX } = window.PROJECT_CONFIG;

export default defineComponent({
  setup() {
    const router = useRouter();
    const { t } = useI18n();
    const userStore = useUserStore();
    const scrStore = useZiyanScrStore();
    const { getBusinessApiPath } = useWhereAmI();
    const route = useRoute();

    const { transformApplyStages } = useApplyStages();
    const { transformRequireTypes } = useRequireTypes();

    const stageDetailSidesliderRef = ref();

    const stageDetailSlideState = reactive({
      suborderId: undefined,
    });

    const orderClipboard = ref<any>({});
    const machineDetails = ref([]);

    const { columns } = useScrColumns('applicationList');
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

    const getOrderRoute = (row: any) => {
      const { order_id, bk_biz_id, bk_username, resource_type, stage } = row;
      let routeParams: RouteLocationRaw = {
        name: 'HostApplicationsDetail',
        params: { id: order_id },
        query: { [GLOBAL_BIZS_KEY]: bk_biz_id, creator: bk_username, bkBizId: bk_biz_id, resource_type },
      };
      if (stage === 'UNCOMMIT') {
        routeParams = { name: 'applyCvm', query: { ...routeParams.query, order_id, unsubmitted: 1 } };
      }
      routerAction.redirect(routeParams, { history: true });
    };

    const modify = (data: any) => {
      router.push({ name: 'HostApplicationsModify', query: { ...route.query, ...data } });
    };

    const reapply = (data: any) => {
      routerAction.redirect(
        { name: 'applyCvm', query: { ...route.query, order_id: data.order_id, unsubmitted: 0 } },
        { history: true },
      );
    };

    const throttleInfo = ref(null);
    // 已交付设备
    const getDeliveredDevices = (params: any) => {
      return http.post(
        `${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/${getBusinessApiPath()}task/findmany/apply/device`,
        params,
      );
    };
    // 查询交付IP和固号IP
    const getDeliveredHostField = (row: any, fieldKey: any) => {
      const params = {
        filter: {
          condition: 'AND',
          rules: [
            { field: 'suborder_id', operator: 'equal', value: row.suborder_id },
            { field: 'bk_biz_id', operator: 'in', value: [row.bk_biz_id] },
          ],
        },
      };
      return getDeliveredDevices(params).then((res: any) => {
        const value = res?.data?.info?.map((item: any) => item[fieldKey]) || [];
        return value;
      });
    };
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
    const handleCellMouseEnter = (row: any) => {
      if (row.success_num > 0) {
        throttleInfo.value(row);
      }
    };
    // 获取匹配详情
    const getMatchDetails = async (subOrderId: number) => {
      return http.post(`${BK_HCM_AJAX_URL_PREFIX}/api/v1/woa/${getBusinessApiPath()}task/find/apply/detail`, {
        suborder_id: subOrderId,
      });
    };

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

    const handleChangeStageSlideShow = (isShow: boolean) => {
      if (isShow) {
        stageDetailPolling.resume();
      } else {
        stageDetailPolling.reset();
      }
    };

    const searchFields = getModel(HostApplySearch).getProperties();
    const searchQs = useSearchQs({ key: 'filter', properties: searchFields });

    const { CommonTable, getListData, pagination } = useTable({
      tableOptions: {
        columns: [
          {
            label: t('单号/子单号'),
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
                </div>
              );
            },
          },
          {
            label: t('单据状态'),
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
                  if (isUpgradeCvm) return t('备货状态异常');
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
                      {t('备货状态异常')} <HelpDocumentFill fill='#ffbb00' width={12} height={12} class={'ml4'} />
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
                    {t('修改需求重试')}
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
                      const { data: list } = await getMatchDetails(data.suborder_id);
                      stageDetailSlideState.suborderId = data.suborder_id;
                      machineDetails.value = list.info;
                      stageDetailSidesliderRef.value.triggerShow(true);
                    }}>
                    {t('查看详情')}
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
            label: t('需求类型'),
            field: 'require_type',
            width: 100,
            render: ({ cell }: any) => transformRequireTypes(cell),
          },
          {
            label: t('需求摘要'),
            width: 250,
            render: ({ data }: any) => {
              const isUpgradeCvm = ScrResourceType.UPGRADECVM === data.resource_type;
              return (
                <div>
                  <div>
                    {t('资源类型')}：{data?.resource_type ? getResourceTypeName(data?.resource_type) : '--'}
                  </div>
                  {/* 机型配置调整不展示机型、园区 */}
                  {!isUpgradeCvm && (
                    <>
                      <div>
                        {t('机型')}：{data.spec?.device_type || '--'}
                      </div>
                      <div style={{ display: 'grid', gridTemplateColumns: 'auto 1fr' }}>
                        <div>{t('园区')}：</div>
                        <div>
                          {data.spec?.zones?.[0] === 'all'
                            ? `${getRegionCn(data.spec.region)}(全部可用区)`
                            : data.spec?.zones?.map((zone: string) => {
                                return <div>{getZoneCn(zone)}</div>;
                              }) || '--'}
                        </div>
                      </div>
                      <div>
                        {t('分布')}：
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
            label: t('申请人'),
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
            label: t('需求数'),
            field: 'total_num',
            width: 120,
            render: ({ row, cell }: any) => {
              if (row.modify_time > 0) {
                return `${row.total_num}(原需求数${row.origin_num})`;
              }
              return cell;
            },
          },
          {
            label: t('待交付数'),
            field: 'pending_num',
            width: 90,
          },
          {
            label: t('已交付数'),
            field: 'success_num',
            width: 120,
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
                      v-bk-tooltips={{ content: t('复制 IP') }}>
                      <Copy />
                    </Button>
                    <Button
                      text
                      theme={'primary'}
                      class='mr8'
                      v-clipboard:copy={assetIds.join('\n')}
                      v-bk-tooltips={{ content: t('复制固资号') }}>
                      <Copy />
                    </Button>
                    <Button
                      text
                      theme={'primary'}
                      onClick={() => goToCmdb(ips)}
                      v-bk-tooltips={{ content: t('去蓝鲸配置平台管理资源') }}>
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
            label: t('操作'),
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
                    {t('再次申请')}
                  </Button>
                  <Button
                    size='small'
                    text
                    theme={'primary'}
                    class='mr8'
                    disabled={retryBtnDisabled(row) || isUpgradeCvm}
                    onClick={async () => {
                      await scrStore.retryOrder({ suborder_id: [row.suborder_id] });
                      Message({ theme: 'success', message: t('重试成功') });
                      getListData();
                    }}>
                    {t('重试')}
                  </Button>
                  <Button
                    size='small'
                    text
                    theme={'primary'}
                    class='mr8'
                    disabled={opBtnDisabled(row)}
                    onClick={async () => {
                      await scrStore.stopOrder({ suborder_id: [row.suborder_id] });
                      Message({ theme: 'success', message: t('终止成功') });
                      getListData();
                    }}>
                    {t('终止')}
                  </Button>
                </div>
              );
            },
          },
        ],
        extra: {
          onRowMouseEnter: (e: any, row: any) => {
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
        return {
          url: `/api/v1/woa/${getBusinessApiPath()}task/findmany/apply`,
          payload: transformFlatCondition(condition.value, searchFields),
        };
      },
    });

    const condition = ref<Record<string, any>>({});
    const searchValues = ref<Record<string, any>>({});

    const getSearchCompProps = (field: ModelProperty) => {
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
      const { order_id: orderId, ...rest } = searchValues.value;
      const orderIds = orderId.filter((item: string) => /^\d+$/.test(item));
      const suborderIds = orderId.filter((item: string) => /^\d+-\d+$/.test(item));

      searchQs.set({ ...rest, order_id: orderIds, suborder_id: suborderIds });
    };

    const handleReset = () => {
      searchQs.clear();
    };

    watch(
      () => route.query,
      async (query) => {
        condition.value = searchQs.get(query, {
          create_at: getDateRange('last30d', true),
          bk_username: [userStore.username],
        });

        // 将子单号合并到主单号条件中
        const { order_id: orderId, suborder_id: suborderId, ...rest } = condition.value;
        searchValues.value = { ...rest, order_id: [...(orderId || []), ...(suborderId || [])] };

        getListData();
      },
      { immediate: true },
    );

    onMounted(() => {
      throttleDeliveredHostField();
    });

    return () => (
      <>
        <div style={{ padding: '24px 24px 0 24px' }}>
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
        <section class={cssModule['table-wrapper']}>
          <CommonTable />
        </section>
        <StageDetailSideslider
          ref={stageDetailSidesliderRef}
          details={machineDetails.value}
          onChangeSlideShow={handleChangeStageSlideShow}
        />
      </>
    );
  },
});
