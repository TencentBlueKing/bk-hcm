import { computed, defineComponent, nextTick, onBeforeMount, PropType, ref } from 'vue';
import Panel from '@/components/panel';
import { Button, DatePicker, Select, Checkbox } from 'bkui-vue';
import { Info as InfoIcon } from 'bkui-vue/lib/icon';
import cssModule from './index.module.scss';
import BusinessSelector from '@/components/business-selector/index.vue';
import WName from '@/components/w-name';
import { useI18n } from 'vue-i18n';
import { timeFormatter } from '@/common/util';
import {
  IDeviceType,
  IListResourcesDemandsParam,
  IOpProductsResult,
  IPlanProducts,
  IRegion,
  IZone,
} from '@/typings/resourcePlan';
import { useResourcePlanStore } from '@/store';
import dayjs from 'dayjs';
import isoWeek from 'dayjs/plugin/isoWeek';
import { useRoute, useRouter } from 'vue-router';
import { RESOURCE_DEMANDS_STATUS_NAME } from '@/components/resource-plan/constants';
import ObsProjectSelector from '@/views/business/resource-plan/children/obs-project-selector.vue';
import { useWhereAmI } from '@/hooks/useWhereAmI';

dayjs.extend(isoWeek);

export default defineComponent({
  props: {
    isBiz: {
      type: Boolean,
      required: true,
    },
    expectTimeRange: Object as PropType<{ start: string; end: string }>,
  },
  emits: ['search', 'update:expectTimeRange'],
  setup(props, { emit }) {
    const { Option } = Select;
    const router = useRouter();
    const route = useRoute();
    const { t } = useI18n();
    const { getBizsId } = useWhereAmI();
    const resourcePlanStore = useResourcePlanStore();
    const initialSearchModel: Partial<IListResourcesDemandsParam> = {
      bk_biz_ids: [], // 业务
      op_product_ids: [], // 产品
      plan_product_ids: [], // 规划产品
      obs_projects: [], // OBS项目类型
      demand_classes: [], // 预测需求
      device_classes: [], // 机型分类
      device_types: [], // 机型规格
      region_ids: [], // 地区城市
      zone_ids: [], // 可用区
      plan_types: [], // 预测类型
      expiring_only: false, // 过期状态
      expect_time_range: props.expectTimeRange, // 期望交付时间范围
      statuses: [], // 状态
    };

    const opProductList = ref<{ op_product_id: number; op_product_name: string }[]>([]);
    const planProductsList = ref<IPlanProducts[]>([]);
    const demandClassList = ref<string[]>([]);
    const deviceClassList = ref<string[]>([]);
    const deviceTypeList = ref<IDeviceType[]>([]);
    const regionList = ref<IRegion[]>([]);
    const zoneList = ref<IZone[]>([]);
    const planClassList = ref<string[]>([]);

    const isLoadingOpProducts = ref(false);
    const isLoadingPlanProducts = ref(false);
    const isLoadingDemandClass = ref(false);
    const isLoadingDeviceClass = ref(false);
    const isLoadingDeviceType = ref(false);
    const isLoadingRegion = ref(false);
    const isLoadingZone = ref(false);
    const isLoadingPlanClass = ref(false);

    const searchModel = ref(JSON.parse(JSON.stringify(initialSearchModel)));
    const showRollingServerProject = computed(() => (props.isBiz ? 931 === getBizsId() : true));

    const handleSearch = () => {
      storeSearchModelInQuery(JSON.stringify(searchModel.value));
      emit('search', searchModel.value);
    };

    const handleReset = () => {
      storeSearchModelInQuery('');
      searchModel.value = JSON.parse(JSON.stringify(initialSearchModel));
      emit('search', searchModel.value);
    };

    const storeSearchModelInQuery = (searchModelStr: string) => {
      router.replace({
        query: { ...route.query, searchModel: searchModelStr },
      });
    };

    const handleChangeDate = (key: string, val: string[]) => {
      if (val[0] && val[1]) {
        const range = {
          start: timeFormatter(val[0], 'YYYY-MM-DD'),
          end: timeFormatter(val[1], 'YYYY-MM-DD'),
        };
        searchModel.value[key] = range;
        emit('update:expectTimeRange', range);
      } else {
        searchModel.value[key] = undefined;
      }
    };

    const getOpProductsList = () => {
      isLoadingOpProducts.value = true;
      resourcePlanStore
        .getOpProductsList()
        .then((data: IOpProductsResult) => {
          opProductList.value = data?.data?.details || [];
        })
        .finally(() => {
          isLoadingOpProducts.value = false;
        });
    };

    const getPlanProductsList = () => {
      isLoadingPlanProducts.value = true;
      resourcePlanStore
        .getPlanProductsList()
        .then((data: { data: { details: IPlanProducts[] } }) => {
          planProductsList.value = data?.data?.details || [];
        })
        .finally(() => {
          isLoadingPlanProducts.value = false;
        });
    };

    const getDemandClassList = () => {
      isLoadingDemandClass.value = true;
      resourcePlanStore
        .getDemandClasses()
        .then((data: { data: { details: string[] } }) => {
          demandClassList.value = data?.data?.details || [];
        })
        .finally(() => {
          isLoadingDemandClass.value = false;
        });
    };

    const getDeviceClassList = () => {
      isLoadingDeviceClass.value = true;
      resourcePlanStore
        .getDeviceClasses()
        .then((data: { data: { details: string[] } }) => {
          deviceClassList.value = data?.data?.details || [];
        })
        .finally(() => {
          isLoadingDeviceClass.value = false;
        });
    };

    const getDeviceTypeList = () => {
      isLoadingDeviceType.value = true;
      resourcePlanStore
        .getDeviceTypes()
        .then((data: { data: { details: IDeviceType[] } }) => {
          deviceTypeList.value = data?.data?.details || [];
        })
        .finally(() => {
          isLoadingDeviceType.value = false;
        });
    };

    const getPlanClassList = () => {
      isLoadingPlanClass.value = true;
      resourcePlanStore
        .getPlanTypes()
        .then((data: { data: { details: string[] } }) => {
          planClassList.value = data?.data?.details || [];
        })
        .finally(() => {
          isLoadingPlanClass.value = false;
        });
    };

    const getRegionList = () => {
      isLoadingRegion.value = true;
      resourcePlanStore
        .getRegions()
        .then((data: { data: { details: IRegion[] } }) => {
          regionList.value = data?.data?.details || [];
        })
        .finally(() => {
          isLoadingRegion.value = false;
        });
    };

    const getZoneList = () => {
      isLoadingZone.value = true;
      resourcePlanStore
        .getZones(searchModel.value.region_ids)
        .then((data: { data: { details: IZone[] } }) => {
          zoneList.value = data?.data?.details || [];
        })
        .finally(() => {
          isLoadingZone.value = false;
        });
    };

    const onChangeRegion = () => {
      searchModel.value.zone_ids = [];
      getZoneList();
    };

    onBeforeMount(() => {
      if (!props.isBiz) {
        getOpProductsList();
        getPlanProductsList();
      }
      getDemandClassList();
      getDeviceClassList();
      getDeviceTypeList();
      getPlanClassList();
      getRegionList();
      getZoneList();
      nextTick(() => {
        if (route.query.searchModel) {
          const querySearchModel = JSON.parse(route.query.searchModel as string);
          if (!showRollingServerProject.value) {
            const idx = querySearchModel.obs_projects.indexOf('滚服项目');
            idx !== -1 && querySearchModel.obs_projects.splice(idx, 1);
          }
          searchModel.value = querySearchModel;
          emit('update:expectTimeRange', searchModel.value.expect_time_range);
        }
        handleSearch();
      });
    });

    return () => (
      <Panel class={cssModule['mb-16']}>
        {{
          title: () =>
            props.isBiz && (
              <div class={cssModule.infoItem}>
                <InfoIcon class={cssModule.icon} />
                <span>{t('限自研云CVM资源预测，IDC物理主机资源申请，请联系')}</span>
                <WName name={'ICR'} alias={t('ICR(IEG资源服务助手)')}></WName>
                <span>{t('确认')}</span>
              </div>
            ),
          default: () => (
            <>
              <div class={cssModule['search-grid']}>
                {!props.isBiz && (
                  <>
                    <div>
                      <div class={cssModule['search-label']}>{t('业务')}</div>
                      <BusinessSelector
                        v-model={searchModel.value.bk_biz_ids}
                        multiple={true}
                        authed={true}
                        autoSelect={true}
                        isShowAll={true}
                        clearable={true}
                      />
                    </div>
                    <div>
                      <div class={cssModule['search-label']}>{t('运营产品')}</div>
                      <Select multiple v-model={searchModel.value.op_product_ids} loading={isLoadingOpProducts.value}>
                        {opProductList.value.map((item) => (
                          <Option name={item.op_product_name} id={item.op_product_id} />
                        ))}
                      </Select>
                    </div>
                    <div>
                      <div class={cssModule['search-label']}>{t('规划产品')}</div>
                      <Select
                        multiple
                        v-model={searchModel.value.plan_product_ids}
                        loading={isLoadingPlanProducts.value}>
                        {planProductsList.value.map((item) => (
                          <Option name={item.plan_product_name} id={item.plan_product_id} />
                        ))}
                      </Select>
                    </div>
                  </>
                )}
                <div>
                  <div class={cssModule['search-label']}>{t('项目类型')}</div>
                  <ObsProjectSelector
                    v-model={searchModel.value.obs_projects}
                    multiple
                    showRollingServerProject={showRollingServerProject.value}
                  />
                </div>
                <div>
                  <div class={cssModule['search-label']}>{t('预测类型')}</div>
                  <Select multiple v-model={searchModel.value.demand_classes} loading={isLoadingDemandClass.value}>
                    {demandClassList.value.map((item) => (
                      <Option name={item} id={item} />
                    ))}
                  </Select>
                </div>
                <div>
                  <div class={cssModule['search-label']}>{t('机型类型')}</div>
                  <Select multiple v-model={searchModel.value.device_classes} loading={isLoadingDeviceClass.value}>
                    {deviceClassList.value.map((item) => (
                      <Option name={item} id={item} />
                    ))}
                  </Select>
                </div>
                <div>
                  <div class={cssModule['search-label']}>{t('机型规格')}</div>
                  <Select multiple v-model={searchModel.value.device_types} loading={isLoadingDeviceType.value}>
                    {deviceTypeList.value.map((item) => (
                      <Option id={item.device_type} name={item.device_type} />
                    ))}
                  </Select>
                </div>
                <div>
                  <div class={cssModule['search-label']}>{t('期望到货时间')}</div>
                  <DatePicker
                    modelValue={[searchModel.value.expect_time_range?.start, searchModel.value.expect_time_range?.end]}
                    onChange={(val: string[]) => handleChangeDate('expect_time_range', val)}
                    type='daterange'
                    clearable={false}
                  />
                </div>
                <div>
                  <div class={cssModule['search-label']}>{t('城市')}</div>
                  <Select
                    multiple
                    v-model={searchModel.value.region_ids}
                    loading={isLoadingRegion.value}
                    onChange={onChangeRegion}>
                    {regionList.value.map((item) => (
                      <Option id={item.region_id} name={item.region_name} />
                    ))}
                  </Select>
                </div>
                <div>
                  <div class={cssModule['search-label']}>{t('可用区')}</div>
                  <Select multiple v-model={searchModel.value.zone_ids} loading={isLoadingZone.value}>
                    {zoneList.value.map((item) => (
                      <Option name={item.zone_name} id={item.zone_id} />
                    ))}
                  </Select>
                </div>
                <div>
                  <div class={cssModule['search-label']}>{t('预测类型')}</div>
                  <Select multiple v-model={searchModel.value.plan_types} loading={isLoadingPlanClass.value}>
                    {planClassList.value.map((item) => (
                      <Option name={item} id={item} />
                    ))}
                  </Select>
                </div>
                <div>
                  <div class={cssModule['search-label']}>{t('状态')}</div>
                  <Select multiple v-model={searchModel.value.statuses}>
                    {Object.entries(RESOURCE_DEMANDS_STATUS_NAME).map(([id, name]) => (
                      <Option name={name} id={id} />
                    ))}
                  </Select>
                </div>
              </div>
              <div class={cssModule['search-checkbox-wrapper']}>
                <Checkbox v-model={searchModel.value.expiring_only}>{t('本月即将过期')}</Checkbox>
              </div>
              <Button theme='primary' class={cssModule['search-button']} onClick={handleSearch}>
                {t('查询')}
              </Button>
              <Button onClick={handleReset} class={cssModule['search-button']}>
                {t('重置')}
              </Button>
            </>
          ),
        }}
      </Panel>
    );
  },
});
