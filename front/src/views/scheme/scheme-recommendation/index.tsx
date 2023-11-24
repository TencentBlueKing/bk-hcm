import { defineComponent, reactive, ref } from 'vue';
import './index.scss';
import blank_1 from '@/assets/image/scheme-blank-1.png';
import blank_2 from '@/assets/image/scheme-blank-2.png';
import { GenerateSchemesParams } from './types/index';
import { Dialog, SearchSelect, Tree } from 'bkui-vue';
// import SchemePreview from '../components/scheme-preview';
// import http from '@/http';
// const { BK_HCM_AJAX_URL_PREFIX } = window.PROJECT_CONFIG;

export default defineComponent({
  name: 'SchemeRecommendationPage',
  setup() {
    const toggleClose = ref(false);

    const countriesList = ref<Array<string>>([]);
    const formData = reactive<GenerateSchemesParams>({
      biz_type: '',
      cover_ping: null,
      deployment_architecture: [],
      user_distribution: [],
    });

    console.log(countriesList, formData);
    const isUserProportionDetailDialogShow = ref(false);
    const treeData = ref([
      {
        name: '方案成熟',
        isOpen: true,
        content:
          '拥有支撑数百款腾讯业务的经验沉淀，兼容各种复杂的系统架构，生于运维 · 精于运维',
        id: '/',
        children: [
          {
            name: 'child-1-方案成熟-拥有支撑数百款腾讯业务的经验沉淀，兼容各种复杂的系统架构，生于运维 · 精于运维',
            content:
              '拥有支撑数百款腾讯业务的经验沉淀，兼容各种复杂的系统架构，生于运维 · 精于运维',
            children: [],
            __uuid: '7dd80a5e-f43e-476a-9b88-0a08f8801162',
          },
          {
            name: 'child-1-覆盖全面',
            content:
              '从配置管理，到作业执行、任务调度和监控自愈，再通过运维大数据分析辅助运营决策，全方位覆盖业务运营的全周期保障管理。',
            children: [],
            __uuid: '9195461a-24cd-46b1-9d7d-08234e7279be',
          },
          {
            name: 'child-1-开放平台',
            content:
              '开放的PaaS，具备强大的开发框架和调度引擎，以及完整的运维开发培训体系，助力运维快速转型升级。',
            children: [
              {
                name: 'child-1-方案成熟',
                content:
                  '拥有支撑数百款腾讯业务的经验沉淀，兼容各种复杂的系统架构，生于运维 · 精于运维',
                children: [],
                __uuid: 'd137d6b3-313e-49df-8e62-9b0549e5906a',
              },
              {
                name: 'child-1-覆盖全面',
                content:
                  '从配置管理，到作业执行、任务调度和监控自愈，再通过运维大数据分析辅助运营决策，全方位覆盖业务运营的全周期保障管理。',
                children: [],
                __uuid: 'ff19cff2-8ab5-4101-a823-63b5fd7c8e73',
              },
              {
                name: 'child-1-开放平台',
                isOpen: true,
                content:
                  '开放的PaaS，具备强大的开发框架和调度引擎，以及完整的运维开发培训体系，助力运维快速转型升级。',
                children: [],
                __uuid: '1b9effd2-d3b4-40be-a5ee-eb30242e3654',
              },
            ],
            __uuid: '69a6edec-7a75-432f-b1a1-2162e8c42293',
          },
        ],
        __uuid: '272197ba-2571-455d-8e89-9edc821f4b8b',
      },
      {
        name: '覆盖全面',
        content:
          '从配置管理，到作业执行、任务调度和监控自愈，再通过运维大数据分析辅助运营决策，全方位覆盖业务运营的全周期保障管理。',
        id: '//',
        children: [
          {
            name: 'child-2-方案成熟',
            content:
              '拥有支撑数百款腾讯业务的经验沉淀，兼容各种复杂的系统架构，生于运维 · 精于运维',
            children: [],
            __uuid: 'c3193bf8-10d5-4b7b-9ddf-5056e79c0024',
          },
          {
            name: 'child-2-覆盖全面',
            content:
              '从配置管理，到作业执行、任务调度和监控自愈，再通过运维大数据分析辅助运营决策，全方位覆盖业务运营的全周期保障管理。',
            children: [],
            __uuid: 'ea53c5da-a150-4164-bafd-06372db9a7d9',
          },
          {
            name: 'child-2-开放平台',
            content:
              '开放的PaaS，具备强大的开发框架和调度引擎，以及完整的运维开发培训体系，助力运维快速转型升级。',
            children: [],
            checked: true,
            __uuid: 'e281b5ca-fdfc-487d-b94b-1d481bd75cb3',
          },
        ],
        __uuid: 'ee578f5a-3928-4077-b4f8-0e714cf8fda3',
      },
      {
        name: '开放平台',
        content:
          '开放的PaaS，具备强大的开发框架和调度引擎，以及完整的运维开发培训体系，助力运维快速转型升级。',
        children: [
          {
            name: 'child-3-方案成熟',
            content:
              '拥有支撑数百款腾讯业务的经验沉淀，兼容各种复杂的系统架构，生于运维 · 精于运维',
            children: [],
            __uuid: 'a9073e51-7702-4f2c-9d5f-f64d84a40680',
          },
          {
            name: 'child-3-覆盖全面',
            content:
              '从配置管理，到作业执行、任务调度和监控自愈，再通过运维大数据分析辅助运营决策，全方位覆盖业务运营的全周期保障管理。',
            children: [],
            __uuid: '4a670b99-9a38-4624-94b7-486b7dac89a0',
          },
          {
            name: 'child-3-开放平台',
            content:
              '开放的PaaS，具备强大的开发框架和调度引擎，以及完整的运维开发培训体系，助力运维快速转型升级。',
            children: [],
            __uuid: '62267a43-bf35-4eaf-8df1-42e463059740',
          },
        ],
        __uuid: 'b334cc34-c5b0-4c25-b926-33aaadb82e7e',
      },
    ]);

    // const reqCountriesData = async () => {
    //   const result = await http.post(
    //     `${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/selections/countries/list`,
    //   );
    //   console.log(result);
    // };

    // onMounted(() => {
    //   reqCountriesData();
    // });

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
                <bk-form-item label='用户分布地区' required>
                  <bk-select>
                    <bk-option value='1' label='本科以下' />
                    <bk-option value='2' label='本科以上' />
                  </bk-select>
                </bk-form-item>
                <bk-form-item label='业务类型' required>
                  <bk-select>
                    <bk-option value='1' label='本科以下' />
                    <bk-option value='2' label='本科以上' />
                  </bk-select>
                </bk-form-item>
                <bk-form-item label='用户网络容忍' class='prompt-icon-wrap'>
                  <div class='sub-form-item-wrap'>
                    <bk-form-item
                      label='网络延迟'
                      class='sub-form-item-content'>
                      <bk-input type='number'></bk-input>
                    </bk-form-item>
                    <bk-form-item
                      label='ping抖动'
                      class='sub-form-item-content'>
                      <bk-input type='number'></bk-input>
                    </bk-form-item>
                    <bk-form-item label='丢包率' class='sub-form-item-content'>
                      <bk-input type='number'></bk-input>
                    </bk-form-item>
                  </div>
                </bk-form-item>
                <bk-form-item label='用户分布占比'>
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
                </bk-form-item>
                <bk-form-item label='部署架构' class='prompt-icon-wrap'>
                  <bk-checkbox-group>
                    <bk-checkbox label='分布式部署' />
                    <bk-checkbox label='集中式部署' />
                  </bk-checkbox-group>
                </bk-form-item>
                <bk-form-item>
                  <bk-button class='mr8' theme='primary'>
                    选型推荐
                  </bk-button>
                  <bk-button>清空</bk-button>
                </bk-form-item>
              </bk-form>
            </div>
          </div>
          <div class='scheme-recommendation-container'>
            <div class='content-container'>
              <div class='item-wrap'>
                <img src={blank_1} alt='' />
                <div class='title-wrap'>
                  <span class='serial-number mr8'>1</span>
                  <span class='title-text'>配置基本信息</span>
                </div>
                <div class='content-wrap'>
                  配置业务基本属性，并查看初步的部署架构、用户分布、部署方案的推荐结果。
                </div>
              </div>
              <i class='hcm-icon bkhcm-icon-arrows-up separator'></i>
              <div class='item-wrap'>
                <img src={blank_2} alt='' />
                <div class='title-wrap'>
                  <span class='serial-number mr8'>2</span>
                  <span class='title-text'>查看方案详情</span>
                </div>
                <div class='content-wrap'>
                  查看方案结果，并基于方案分析与网络、成本数据进一步微调部署方案。
                </div>
              </div>
            </div>
          </div>
        </div>
        <Dialog
          dialogType='show'
          class='user-proportion-detail-dialog'
          isShow={isUserProportionDetailDialogShow.value}
          title='分布权重占比'
          onClosed={() => {
            isUserProportionDetailDialogShow.value = false;
          }}>
          <div class='tips-wrap mb16'>
            <i class='hcm-icon bkhcm-icon-info-line'></i>
            <div class='tips-text'>占比权重说明说明说明</div>
          </div>
          <SearchSelect class="mb16" modelValue={[]} data={[]} placeholder='请输入' />
          <Tree
            data={treeData.value}
            label='name'
            children='children'
            search=''
            show-node-type-icon={false}
            prefixIcon={(params: any, renderType: any) => {
              if (params.children.length === 0) return null;
              console.log(params, renderType);
              return params.isOpen
                ? <i class="hcm-icon bkhcm-icon-minus-circle"></i>
                : <i class="hcm-icon bkhcm-icon-plus-circle"></i>;
            }}
          >
            {{
              nodeAppend: () => (
                <span class="proportion-num">10</span>
              ),
            }}
          </Tree>
        </Dialog>
      </>
    );
  },
});
