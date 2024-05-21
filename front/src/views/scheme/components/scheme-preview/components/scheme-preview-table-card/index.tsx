import { PropType, defineComponent, ref, watch } from 'vue';
import './index.scss';
import { Table, Tag, Loading, Button, PopConfirm } from 'bkui-vue';
import { AngleDown, AngleRight } from 'bkui-vue/lib/icon';
import { useSchemeStore } from '@/store';
import { QueryRuleOPEnum } from '@/typings';
import { IServiceArea } from '@/typings/scheme';
import { VendorEnum, VendorMap } from '@/common/constant';
import { renderVendorIcons } from './renderVendorIcons';
import SaveSchemeButton from '../save-scheme-button';

export default defineComponent({
  props: {
    compositeScore: {
      type: Number,
      required: true,
      default: 0,
    },
    costScore: {
      type: Number,
      required: true,
      default: 0,
    },
    netScore: {
      type: Number,
      required: true,
      default: 0,
    },
    resultIdcIds: {
      type: Array as PropType<Array<string>>,
      required: true,
    },
    idx: {
      type: Number,
      required: true,
      default: -1,
    },
    onViewDetail: {
      required: true,
      type: Function,
    },
    coverRate: {
      required: true,
      type: Number,
      default: 0,
    },
  },
  setup(props) {
    const schemeStore = useSchemeStore();
    const columns = [
      {
        field: 'name',
        label: '部署点名称',
        width: 200,
        render: ({ data }: any) => `${data.region}机房`,
      },
      {
        field: 'vendor',
        label: '云厂商',
        width: 100,
        render: ({ cell }: { cell: VendorEnum }) => {
          return VendorMap[cell];
        },
      },
      {
        field: 'region',
        label: '所在地',
        width: 200,
        render: ({ data }: any) => `${data.country},${data.region}`,
      },
      {
        field: 'service_areas',
        label: '服务区域',
        render: ({ cell, data }: any) => {
          return data.service_area_arr?.length ? (
            <p class={'flex-row align-items-center service-areas-paragraph'}>
              <PopConfirm trigger='click' width={454}>
                {{
                  content: () => (
                    <div class={'service-areas-table-container'}>
                      <div class={'service-areas-table-header'}>
                        <p class={'service-areas-table-header-title'}>服务质量排名</p>
                      </div>
                      <div class={'service-areas-table'}>
                        <Table
                          data={data.service_area_arr}
                          show-overflow-tooltip
                          maxHeight={500}
                          columns={[
                            {
                              field: 'country_name_province_name',
                              label: '地区',
                              align: 'left',
                              render: ({ data }) => (
                                <p class={'index-number-box-container'}>
                                  <div class={`index-number-box bg-color-${data.idx < 3 ? data.idx + 1 : 4}`}>
                                    {`${data.idx + 1} `}
                                  </div>
                                  {`${data.country_name},${data.province_name}`}
                                </p>
                              ),
                            },
                            {
                              field: 'network_latency',
                              label: '网络延迟',
                              width: 150,
                              render: ({ cell }: { cell: number }) => `${Math.floor(cell)} ms`,
                              sort: true,
                            },
                          ]}></Table>
                      </div>
                    </div>
                  ),
                  default: () => (
                    <a class={'scheme-service-areas-icon-box mr4'}>
                      <i class={'icon hcm-icon bkhcm-icon-paiming scheme-service-areas-icon'}></i>
                    </a>
                  ),
                }}
              </PopConfirm>
              <span v-bk-tooltips={{ content: cell, placement: 'top-start' }}>{cell}</span>
            </p>
          ) : (
            '--'
          );
        },
      },
      {
        field: 'ping',
        label: '平均延迟',
        align: 'right',
        render: ({ cell }: { cell: number }) => {
          return `${Math.floor(cell)} ms`;
        },
        width: 100,
      },
      {
        field: 'price',
        align: 'right',
        label: 'IDC 单位成本',
        render: ({ cell }: { cell: number }) => `${cell}`,
        width: 200,
      },
    ];
    const tableData = ref([]);
    const isLoading = ref(false);
    const isExpanded = ref(false);
    const idcServiceAreasMap = ref<
      Map<
        string,
        {
          service_areas: Array<IServiceArea>;
          avg_latency: number;
        }
      >
    >(new Map());
    const schemeVendors = ref([]);
    const isViewDetailBtnLoading = ref(false);

    const handleViewDetail = async () => {
      isViewDetailBtnLoading.value = true;
      await getSchemeDetails();
      isViewDetailBtnLoading.value = false;
      props.onViewDetail(props.idx);
    };

    watch(
      () => isExpanded.value,
      async () => {
        if (isExpanded.value) await getSchemeDetails();
      },
      {
        immediate: true,
      },
    );

    // 部署方案详情页里切换方案时重新拉数据
    watch(
      () => schemeStore.selectedSchemeIdx,
      (idx) => {
        if (+idx === props.idx) getSchemeDetails();
      },
    );

    const getSchemeDetails = async () => {
      // if (!tableData.value.length) {
      isLoading.value = true;
      const listIdcPromise = schemeStore.listIdc(
        {
          op: QueryRuleOPEnum.AND,
          rules: [
            {
              field: 'id',
              op: QueryRuleOPEnum.IN,
              value: props.resultIdcIds,
            },
          ],
        },
        {
          start: 0,
          limit: props.resultIdcIds.length,
        },
      );
      const queryIdcServiceAreaPromise = schemeStore.queryIdcServiceArea(
        'ping',
        props.resultIdcIds,
        schemeStore.userDistribution,
      );
      const [listIdcRes, queryIdcServiceAreaRes] = await Promise.all([listIdcPromise, queryIdcServiceAreaPromise]);
      queryIdcServiceAreaRes.data.forEach((v) => {
        idcServiceAreasMap.value.set(v.idc_id, {
          service_areas: v.service_areas,
          avg_latency: v.avg_latency,
        });
      });
      tableData.value = listIdcRes.data.map((v) => ({
        name: v.name,
        vendor: v.vendor,
        region: v.region,
        price: v.price,
        country: v.country,
        service_areas: idcServiceAreasMap.value.get(v.id)?.service_areas?.reduce((acc, cur) => {
          acc += `${cur.country_name},${cur.province_name}; `;
          return acc;
        }, ''),
        ping: idcServiceAreasMap.value.get(v.id)?.avg_latency,
        id: v.id,
        service_area_arr: idcServiceAreasMap.value
          .get(v.id)
          ?.service_areas?.sort((a, b) => {
            return Math.floor(a.network_latency) - Math.floor(b.network_latency);
          })
          .map((v, idx) => ({
            ...v,
            idx,
          })),
      }));
      schemeVendors.value = Array.from(
        listIdcRes.data.reduce((acc, cur) => {
          acc.add(cur.vendor);
          return acc;
        }, new Set()),
      );
      schemeStore.setSchemeData({
        deployment_architecture: [],
        vendors: schemeVendors.value,
        composite_score: props.compositeScore,
        net_score: props.netScore,
        cost_score: props.costScore,
        name: schemeStore.recommendationSchemes[props.idx].name,
        id: `${props.idx}`,
        idcList: tableData.value.map((item) => ({
          id: item.id,
          name: item.name,
          vendor: item.vendor,
          country: item.country,
          price: item.price,
          region: item.region,
        })),
      });
      isLoading.value = false;
      // }
    };

    watch(
      () => props.idx,
      () => {
        if (props.idx === 0) {
          isExpanded.value = true;
        }
      },
      {
        immediate: true,
      },
    );

    return () => (
      <div class={'scheme-preview-table-card-container'}>
        <div
          onClick={() => (isExpanded.value = !isExpanded.value)}
          class={`scheme-preview-table-card-header ${
            isExpanded.value ? '' : 'scheme-preview-table-card-header-closed'
          }`}>
          {isExpanded.value ? (
            <AngleDown
              width={'40px'}
              height={'30px'}
              fill='#63656E'
              class={'scheme-preview-table-card-header-expand-icon'}
            />
          ) : (
            <AngleRight
              width={'40px'}
              height={'30px'}
              fill='#63656E'
              class={'scheme-preview-table-card-header-expand-icon'}
            />
          )}

          <p class={'scheme-preview-table-card-header-title'}>{schemeStore.recommendationSchemes[props.idx].name}</p>
          <Tag theme='success' radius='11px' class={'scheme-preview-table-card-header-tag'}>
            分布式部署
          </Tag>
          {renderVendorIcons(schemeStore.recommendationSchemes[props.idx].vendors)}
          <div class={'scheme-preview-table-card-header-score'}>
            <div class={'scheme-preview-table-card-header-score-item'}>
              综合评分： <span class={'score-value'}>{props.compositeScore}</span>
            </div>
            <div class={'scheme-preview-table-card-header-score-item'}>
              网络评分： <span class={'score-value'}>{props.netScore}</span>
            </div>
            <div class={'scheme-preview-table-card-header-score-item'}>
              成本评分： <span class={'score-value'}>{props.costScore}</span>
            </div>
          </div>
          <div class={'scheme-preview-table-card-header-operation'} onClick={(e) => e.stopPropagation()}>
            <Button class={'mr8'} onClick={handleViewDetail} loading={isViewDetailBtnLoading.value}>
              查看详情
            </Button>
            <SaveSchemeButton idx={props.idx} />
          </div>
        </div>
        <div
          class={`scheme-preview-table-card-panel ${
            isExpanded.value ? '' : 'scheme-preview-table-card-panel-invisable'
          }`}>
          <Loading loading={isLoading.value}>
            <Table data={tableData.value} columns={columns} show-overflow-tooltip>
              {{
                empty: () => {
                  if (isLoading.value) return null;
                  return '暂无数据';
                },
              }}
            </Table>
          </Loading>
        </div>
      </div>
    );
  },
});
