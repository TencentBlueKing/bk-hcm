import { defineComponent, ref, reactive, computed, onMounted, watch } from 'vue';
import './index.scss';
import SchemePreview from '../components/scheme-preview';
import SchemeBlankPage from './components/scheme-blank-page';
import SchemeUserProportionShowDialog from './components/scheme-user-proportion-show-dialog';
import { IPageQuery } from '@/typings';
import { IBizTypeList, IBizType, IGenerateSchemesReqParams, IGenerateSchemesResData } from '@/typings/scheme';
import { useSchemeStore } from '@/store';

export default defineComponent({
  name: 'SchemeRecommendationPage',
  setup() {
    const toggleClose = ref(false);
    const schemeStore = useSchemeStore();

    const initLoading = ref(false);
    const countriesList = ref<Array<string>>([]);
    const selectedCountriesList = ref<Array<string>>([]);
    const bizTypeList = ref<IBizTypeList>([]);
    const selectedBizType = computed<IBizType>(() =>
      bizTypeList.value.find((item) => item.biz_type === formData.biz_type),
    );

    const generateSchemesLoading = ref(false);
    const formData = reactive<IGenerateSchemesReqParams>({
      biz_type: '',
      cover_ping: null,
      deployment_architecture: [],
      user_distribution: [],
      user_distribution_mode: 'default',
    });

    const countryChangeLoading = ref(false);
    const clearLastData = () => {
      formData.user_distribution = [];
      schemeStore.setUserDistribution([]);
      schemeStore.setRecommendationSchemes([]);
      scene.value = "blank";
    }
    const handleChangeCountry = async () => {
      clearLastData();
      countryChangeLoading.value = true;
      const res = await schemeStore.queryUserDistributions(
        selectedCountriesList.value.reduce((prev, item) => {
          prev.push({ name: item });
          return prev;
        }, [])
      );
      countryChangeLoading.value = false;
      formData.user_distribution = res.data;
      schemeStore.setUserDistribution(res.data);
    };
    const isUserProportionDetailDialogShow = ref(false);

    const generateSchemes = async () => {
      generateSchemesLoading.value = true;
      const res: IGenerateSchemesResData = await schemeStore.generateSchemes(formData);
      generateSchemesLoading.value = false;
      schemeStore.setRecommendationSchemes(res.data);
      scene.value = "preview";
    };
    const formItemOptions = computed(() => [
      {
        label: '用户分布地区',
        required: true,
        content: () => (
          <bk-select
            loading={initLoading.value}
            v-model={selectedCountriesList.value}
            multiple
            show-select-all
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
        required: true,
        content: () => (
          <bk-select loading={initLoading.value} v-model={formData.biz_type}>
            {bizTypeList.value.map((bizType) => (
              <bk-option key={bizType.id} value={bizType.biz_type} label={bizType.biz_type} />
            ))}
          </bk-select>
        ),
      },
      {
        label: '用户网络容忍',
        extClass: 'prompt-icon-wrap',
        content: [
          {
            label: '网络延迟',
            content: () => (
              <bk-input type='number' v-model={formData.cover_ping}></bk-input>
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
        content: () => (
          <div class='flex-row'>
            <bk-select class='flex-1' v-model={formData.user_distribution_mode}>
              <bk-option label='默认分布占比' value='default' />
            </bk-select>
            <div class={`user-proportion-detail-btn-wrap${formData.user_distribution.length ? '' : ' disabled'}`} 
              onClick={() => {
                formData.user_distribution.length &&
                  (isUserProportionDetailDialogShow.value = true);
              }}>
              <i class='hcm-icon bkhcm-icon-file'></i>
              <span class={'btn-text'}>占比详情</span>
            </div>
          </div>
        ),
      },
      {
        label: '部署架构',
        extClass: 'prompt-icon-wrap',
        content: () => (
          <bk-checkbox-group v-model={formData.deployment_architecture}>
            <bk-checkbox label='distributed'>分布式部署</bk-checkbox>
            <bk-checkbox label='centralized' disabled>集中式部署</bk-checkbox>
          </bk-checkbox-group>
        ),
      },
      {
        label: '',
        content: () => (
          <>
            <bk-button class='mr8' theme='primary' onClick={generateSchemes} loading={countryChangeLoading.value}>
              选型推荐
            </bk-button>
            <bk-button>清空</bk-button>
          </>
        ),
      },
    ]);
    const scene = ref<'blank' | 'preview'>('blank');

    const getInitData = async () => {
      initLoading.value = true;
      const pageQuery: IPageQuery = {
        count: false,
        start: 0,
        limit: 500
      };
      const [res1, res2] = await Promise.all([
        schemeStore.listCountries(),
        schemeStore.listBizTypes(pageQuery),
      ]);
      initLoading.value = false;
      countriesList.value = res1.data.details;
      bizTypeList.value = res2.data.details;
    };

    onMounted(() => {
      getInitData();
    });

    watch(
      () => formData.biz_type,
      (val) => {
        Object.assign(formData, {
          cover_ping: val ? selectedBizType.value.cover_ping : null,
          deployment_architecture: val ? selectedBizType.value.deployment_architecture : '',
        })
      },
    );

    return () => (
      <bk-loading loading={generateSchemesLoading.value} opacity='1' style="height: 100%">
        <div class='scheme-recommendation-page'>
          <div class={`business-attributes-container${toggleClose.value ? ' close' : ''}`}>
            <div class='title-wrap'>
              <div class='title-text'>业务属性</div>
              <i class='hcm-icon bkhcm-icon-shouqi' onClick={() => (toggleClose.value = !toggleClose.value)}></i>
            </div>
            <div class='content-wrap'>
              <bk-form form-type='vertical'>
                {formItemOptions.value.map(
                  ({ label, required, content, extClass }) => (
                    <bk-form-item label={label} required={required} class={extClass}>
                      {Array.isArray(content) ? (
                        <div class='sub-form-item-wrap'>
                          {content.map((sub) => (
                            <bk-form-item label={sub.label}>
                              {sub.content()}
                            </bk-form-item>
                          ))}
                        </div>
                      ) : (
                        content()
                      )}
                    </bk-form-item>
                  ),
                )}
              </bk-form>
            </div>
          </div>
          <div class='scheme-recommendation-container'>
            <div class='content-container'>
              {scene.value === 'blank' ? (
                <SchemeBlankPage />
              ) : (
                <SchemePreview />
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
