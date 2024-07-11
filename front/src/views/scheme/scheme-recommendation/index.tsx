import { defineComponent, ref, reactive, computed, onMounted, watch, onBeforeUnmount } from 'vue';
import './index.scss';
import SchemePreview from '../components/scheme-preview';
import SchemeBlankPage from './components/scheme-blank-page';
import SchemeUserProportionShowDialog from './components/scheme-user-proportion-show-dialog';
import { IPageQuery } from '@/typings';
import { IBizTypeList, IBizType, IGenerateSchemesReqParams, IGenerateSchemesResData } from '@/typings/scheme';
import { useSchemeStore } from '@/store';
import SchemeRecommendDetail from '../components/scheme-recommend-detail';
import { onBeforeRouteLeave } from 'vue-router';
import { InfoBox } from 'bkui-vue';
import { useVerify } from '@/hooks';
import ErrorPage from '@/views/error-pages/403';

export default defineComponent({
  name: 'SchemeRecommendationPage',
  setup() {
    const toggleClose = ref(false);
    const schemeStore = useSchemeStore();

    const countryInitLoading = ref(false);
    const bizTypesInitLoading = ref(false);
    const countriesList = ref<Array<string>>([]);
    const bizTypeList = ref<IBizTypeList>([]);
    const selectedBizType = computed<IBizType>(() => {
      return bizTypeList.value.find((item) => item.biz_type === formData.biz_type);
    });

    const generateSchemesLoading = ref(false);
    const formData = reactive<IGenerateSchemesReqParams>({
      selected_countries: [],
      biz_type: '',
      cover_ping: null,
      deployment_architecture: [],
      user_distribution: [],
      user_distribution_mode: 'default',
    });

    const isAllSchemeSaved = computed(() => {
      return schemeStore.recommendationSchemes.reduce((acc, cur) => {
        acc &&= cur.isSaved;
        return acc;
      }, true);
    });

    const countryChangeLoading = ref(false);
    const clearLastData = () => {
      formData.user_distribution = [];
      schemeStore.setUserDistribution([]);
    };

    const { authVerifyData } = useVerify();
    if (!authVerifyData.value.permissionAction.cloud_selection_recommend) return () => <ErrorPage />;

    const handleChangeCountry = async () => {
      clearLastData();
      countryChangeLoading.value = true;
      const res = await schemeStore.queryUserDistributions(
        formData.selected_countries.reduce((prev, item) => {
          prev.push({ name: item });
          return prev;
        }, []),
      );
      countryChangeLoading.value = false;
      formData.user_distribution = res.data;
      schemeStore.setUserDistribution(res.data);
    };
    const isUserProportionDetailDialogShow = ref(false);

    const generateSchemes = async () => {
      await formRef.value.validate();
      generateSchemesLoading.value = true;
      const res: IGenerateSchemesResData = await schemeStore.generateSchemes(formData);
      schemeStore.setSchemeConfig(formData.cover_ping, formData.biz_type, formData.deployment_architecture);
      generateSchemesLoading.value = false;
      schemeStore.setRecommendationSchemes(
        res.data
          .filter(({ cover_rate }) => cover_rate >= 0.65)
          .map((item, idx) => ({
            ...item,
            id: `${idx}`,
            name: `方案${idx + 1}`,
            isSaved: false,
          })),
      );
      scene.value = 'preview';
    };

    const viewDetail = (idx: number) => {
      scene.value = 'detail';
      schemeStore.setSelectedSchemeIdx(idx);
    };

    const formItemOptions = computed(() => [
      {
        label: '用户分布地区',
        required: true,
        property: 'selected_countries',
        content: () => (
          <bk-select
            loading={countryInitLoading.value || countryChangeLoading.value}
            v-model={formData.selected_countries}
            multiple
            show-select-all
            filterable
            input-search={false}
            onBlur={handleChangeCountry}
            onClear={clearLastData}>
            {countriesList.value.map((country, index) => (
              <bk-option key={index} value={country} label={country} />
            ))}
          </bk-select>
        ),
      },
      {
        label: '业务类型',
        property: 'biz_type',
        content: () => (
          <bk-select loading={bizTypesInitLoading.value} v-model={formData.biz_type}>
            {bizTypeList.value.map((bizType) => (
              <bk-option key={bizType.id} value={bizType.biz_type} label={bizType.biz_type} />
            ))}
          </bk-select>
        ),
      },
      {
        label: '用户网络容忍',
        required: true,
        extClass: 'prompt-icon-wrap',
        tips: '用户到 IDC 的网络质量容忍',
        left: '96px',
        content: [
          {
            label: '网络延迟',
            property: 'cover_ping',
            required: true,
            content: () => (
              <bk-input class='with-suffix' type='number' v-model={formData.cover_ping} min={1} suffix='ms'></bk-input>
            ),
          },
          /* {
            label: 'ping抖动',
            content: () => <bk-input type='number' disabled></bk-input>,
          },
          {
            label: '丢包率',
            content: () => <bk-input type='number' disabled></bk-input>,
          }, */
        ],
      },
      {
        label: '用户分布占比',
        required: true,
        content: () => (
          <div class='flex-row' style={{ overflow: 'hidden' }}>
            <bk-select class='flex-1' v-model={formData.user_distribution_mode} clearable={false}>
              <bk-option label='默认分布占比' value='default' />
            </bk-select>
            <div
              class={`user-proportion-detail-btn-wrap${formData.user_distribution.length ? '' : ' disabled'}`}
              onClick={() => {
                formData.user_distribution.length && (isUserProportionDetailDialogShow.value = true);
              }}>
              <i class='hcm-icon bkhcm-icon-file'></i>
              <span class={'btn-text'}>占比详情</span>
            </div>
          </div>
        ),
      },
      {
        label: '部署架构',
        property: 'deployment_architecture',
        required: true,
        extClass: 'prompt-icon-wrap',
        tips: '分布式部署：全局模块集中部署，功能模块分区域部署。\n集中式部署：适用于同一套服务器覆盖所有用户的场景。',
        left: '68px',
        content: () => (
          <bk-checkbox-group v-model={formData.deployment_architecture}>
            <bk-checkbox label='distributed'>分布式部署</bk-checkbox>
            <bk-checkbox label='centralized' disabled>
              集中式部署
            </bk-checkbox>
          </bk-checkbox-group>
        ),
      },
      {
        label: '',
        content: () => (
          <>
            <bk-button class='mr8' theme='primary' disabled={countryChangeLoading.value} onClick={generateSchemes}>
              选型推荐
            </bk-button>
            <bk-button onClick={clearFormData}>清空</bk-button>
          </>
        ),
      },
    ]);
    const scene = ref<'blank' | 'preview' | 'detail'>('blank');
    const formRef = ref();
    const formRules = {};

    const clearFormData = () => {
      clearLastData();
      Object.assign(formData, {
        selected_countries: [],
        biz_type: '',
        cover_ping: null,
        deployment_architecture: [],
        user_distribution: [],
        user_distribution_mode: 'default',
      });
    };

    const getInitCountryList = async () => {
      countryInitLoading.value = true;
      const res = await schemeStore.listCountries();
      countryInitLoading.value = false;
      countriesList.value = res.data.details.sort((prev, next) => prev.localeCompare(next, 'zh'));
    };

    const getInitBizTypeList = async () => {
      bizTypesInitLoading.value = true;
      const pageQuery: IPageQuery = {
        count: false,
        start: 0,
        limit: 500,
      };
      const res = await schemeStore.listBizTypes(pageQuery);
      bizTypesInitLoading.value = false;
      bizTypeList.value = res.data.details;
      formData.biz_type = bizTypeList.value?.[0].biz_type;
    };

    onMounted(() => {
      getInitCountryList();
      getInitBizTypeList();
    });

    watch(
      () => formData.biz_type,
      (val) => {
        Object.assign(formData, {
          cover_ping: val ? selectedBizType.value.cover_ping : null,
          deployment_architecture: val ? selectedBizType.value.deployment_architecture : '',
        });
      },
    );

    function confirmLeave(event: any) {
      if (!schemeStore.recommendationSchemes.length || !formData.selected_countries.length || isAllSchemeSaved.value)
        return;
      // 生成选型方案后再让离开的用户二次确认
      (event || window.event).returnValue = '关闭提示';
      return '关闭提示';
    }

    onBeforeUnmount(() => {
      window.removeEventListener('beforeunload', confirmLeave);
    });

    onMounted(() => {
      window.addEventListener('beforeunload', confirmLeave);
    });

    onBeforeRouteLeave((to, from, next) => {
      if (!schemeStore.recommendationSchemes.length || !formData.selected_countries.length || isAllSchemeSaved.value) {
        next();
      } else {
        InfoBox({
          title: '确定离开当前页面?',
          subTitle: '离开会导致未保存方案丢失',
          onConfirm: () => next(),
        });
      }
    });

    return () => (
      <bk-loading loading={generateSchemesLoading.value} opacity='1' style='height: 100%'>
        <div class='scheme-recommendation-page'>
          <div
            style={{
              display: scene.value !== 'detail' ? 'block' : 'none',
            }}
            class={`business-attributes-container${toggleClose.value ? ' close' : ''}`}>
            <div class='title-wrap'>
              <div class='title-text'>业务属性</div>
              {<i class='hcm-icon bkhcm-icon-shouqi' onClick={() => (toggleClose.value = !toggleClose.value)}></i>}
            </div>
            <div class='content-wrap'>
              <bk-form form-type='vertical' ref={formRef} model={formData} rules={formRules}>
                {formItemOptions.value.map(({ label, required, content, extClass, property, tips, left }) => (
                  <bk-form-item label={label} required={required} class={extClass} property={property}>
                    {extClass && tips && (
                      <i
                        v-bk-tooltips={{ content: tips, placement: 'right' }}
                        class='hcm-icon bkhcm-icon-prompt'
                        style={{ left }}></i>
                    )}
                    {Array.isArray(content) ? (
                      <div class='sub-form-item-wrap'>
                        {content.map((sub) => (
                          <bk-form-item label={sub.label} required={sub.required} property={sub.property}>
                            {sub.content()}
                          </bk-form-item>
                        ))}
                      </div>
                    ) : (
                      content()
                    )}
                  </bk-form-item>
                ))}
              </bk-form>
            </div>
            {!toggleClose.value && (
              <div class='right-handle' onClick={() => (toggleClose.value = !toggleClose.value)}></div>
            )}
          </div>
          <div class='scheme-recommendation-container'>
            <div class='content-container' style={{ padding: toggleClose.value ? '0 26px' : '0' }}>
              {scene.value === 'blank' ? (
                <SchemeBlankPage />
              ) : (
                <>
                  <SchemePreview
                    style={{
                      display: scene.value === 'preview' ? 'block' : 'none',
                    }}
                    onViewDetail={viewDetail}
                  />
                  {scene.value === 'detail' ? <SchemeRecommendDetail onBack={() => (scene.value = 'preview')} /> : null}
                </>
              )}
            </div>
          </div>
        </div>
        <SchemeUserProportionShowDialog
          v-model:isShow={isUserProportionDetailDialogShow.value}
          treeData={formData.user_distribution}
        />
      </bk-loading>
    );
  },
});
