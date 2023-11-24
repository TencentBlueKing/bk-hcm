import { defineComponent, ref, /* reactive, */ computed } from 'vue';
import './index.scss';
import SchemeBlankPage from './components/scheme-blank-page';
// import { GenerateSchemesReqParams } from './types/index';
import SchemeUserProportionShowDialog from './components/scheme-user-proportion-show-dialog';
import SchemePreview from '../components/scheme-preview';

export default defineComponent({
  name: 'SchemeRecommendationPage',
  setup() {
    const toggleClose = ref(false);

    // const countriesList = ref<Array<string>>([]);
    // const formData = reactive<GenerateSchemesReqParams>({
    //   biz_type: '',
    //   cover_ping: null,
    //   deployment_architecture: [],
    //   user_distribution: [],
    // });
    const formItemOptions = computed(() => [
      {
        label: '用户分布地区',
        required: true,
        content: () => {
          return (
            <bk-select>
              <bk-option value='1' label='本科以下' />
              <bk-option value='2' label='本科以上' />
            </bk-select>
          );
        },
      },
      {
        label: '业务类型',
        required: true,
        content: () => {
          return (
            <bk-select>
              <bk-option value='1' label='本科以下' />
              <bk-option value='2' label='本科以上' />
            </bk-select>
          );
        },
      },
      {
        label: '用户网络容忍',
        extClass: 'prompt-icon-wrap',
        content: [
          {
            label: '网络延迟',
            content: () => {
              return <bk-input type='number'></bk-input>;
            },
          },
          {
            label: 'ping抖动',
            content: () => {
              return <bk-input type='number'></bk-input>;
            },
          },
          {
            label: '丢包率',
            content: () => {
              return <bk-input type='number'></bk-input>;
            },
          },
        ],
      },
      {
        label: '用户分布占比',
        content: () => {
          return (
            <div class='flex-row'>
              <bk-select class='flex-1'>
                <bk-option value='1' label='默认分布占比' />
                <bk-option value='2' label='本科以上' />
              </bk-select>
              <div class='user-proportion-detail-btn-wrap'>
                <i class='hcm-icon bkhcm-icon-file'></i>
                <span
                  class='btn-text'
                  onClick={() => {
                    isUserProportionDetailDialogShow.value = true;
                  }}>
                  占比详情
                </span>
              </div>
            </div>
          );
        },
      },
      {
        label: '部署架构',
        extClass: 'prompt-icon-wrap',
        content: () => {
          return (
            <bk-checkbox-group>
              <bk-checkbox label='分布式部署' />
              <bk-checkbox label='集中式部署' />
            </bk-checkbox-group>
          );
        },
      },
      {
        label: '部署架构',
        extClass: 'prompt-icon-wrap',
        content: () => {
          return (
            <>
              <bk-button class='mr8' theme='primary'>
                选型推荐
              </bk-button>
              <bk-button>清空</bk-button>
            </>
          );
        },
      },
    ]);
    const isUserProportionDetailDialogShow = ref(false);
    const scene = ref<'blank' | 'preview'>('blank');

    return () => (
      <>
        <div class='scheme-recommendation-page'>
          <div
            class={`business-attributes-container${
              toggleClose.value ? ' close' : ''
            }`}>
            <div class='title-wrap'>
              <div class='title-text'>业务属性</div>
              <i
                class='hcm-icon bkhcm-icon-shouqi'
                onClick={() => (toggleClose.value = !toggleClose.value)}></i>
            </div>
            <div class='content-wrap'>
              <bk-form form-type='vertical'>
                {formItemOptions.value.map(({ label, required, content, extClass }) => {
                  return (
                      <bk-form-item
                        label={label}
                        required={required}
                        class={extClass}>
                        {Array.isArray(content) ? (
                          <div class='sub-form-item-wrap'>
                            {content.map((sub) => {
                              return (
                                <bk-form-item label={sub.label}>
                                  {sub.content()}
                                </bk-form-item>
                              );
                            })}
                          </div>
                        ) : (
                          content()
                        )}
                      </bk-form-item>
                  );
                })}
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
        />
      </>
    );
  },
});
