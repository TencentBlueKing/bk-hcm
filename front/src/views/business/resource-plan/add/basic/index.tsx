import { defineComponent, type PropType, ref, onBeforeMount } from 'vue';
import { useI18n } from 'vue-i18n';
import { useResourcePlanStore } from '@/store/resourcePlan';
import Panel from '@/components/panel';
import cssModule from './index.module.scss';
import WName from '@/components/w-name';
import type { IPlanTicket } from '@/typings/resourcePlan';

export default defineComponent({
  props: {
    modelValue: Object as PropType<IPlanTicket>,
  },

  emits: ['update:modelValue'],

  setup(props, { emit, expose }) {
    const { t } = useI18n();
    const resourcePlanStore = useResourcePlanStore();

    const isLoadingDemandClasses = ref(false);
    const opRelationLoading = ref(false);
    const isShowNoOpRelation = ref(false);
    const bizListRelationLoading = ref(false);
    const demandClasses = ref([]);
    const productName = ref('');
    const productId = ref();
    const formRef = ref();
    const bizListLength = ref(0);
    const bizNameList = ref();

    const handleUpdateModelValue = (key: string, value: unknown) => {
      emit('update:modelValue', {
        ...props.modelValue,
        [key]: value,
      });
    };

    const handleInitDemandClassList = () => {
      isLoadingDemandClasses.value = true;
      resourcePlanStore
        .getDemandClasses()
        .then((data: { data: { details: string[] } }) => {
          demandClasses.value = data?.data?.details || [];
        })
        .finally(() => {
          isLoadingDemandClasses.value = false;
        });
    };

    // 业务所属运营产品
    const getBizOprelation = async () => {
      try {
        opRelationLoading.value = true;
        const res = await resourcePlanStore.getBizOrgRelation(props.modelValue.bk_biz_id);

        if (res.code === 0) {
          productName.value = res.data?.op_product_name;
          productId.value = res.data.op_product_id;
          isShowNoOpRelation.value = false;
          getBizListWithOperation();
        } else {
          isShowNoOpRelation.value = true;
        }
      } catch (error) {
        console.error(error, 'error');
      } finally {
        opRelationLoading.value = false;
      }
    };

    // 查询运营产品对应业务列表
    const getBizListWithOperation = async () => {
      try {
        bizListRelationLoading.value = true;

        const res = await resourcePlanStore.getBizsByOpProductList({
          op_product_id: productId.value,
        });

        bizListLength.value = res.data?.details?.length || 0;
        bizNameList.value = res.data?.details?.map((item) => item?.bk_biz_name).join(',');
      } catch (error) {
        console.error(error, 'error');
      } finally {
        bizListRelationLoading.value = false;
      }
    };

    const validate = () => {
      return formRef.value.validate();
    };

    onBeforeMount(() => {
      getBizOprelation();
      handleInitDemandClassList();
    });

    expose({
      validate,
    });

    return () => (
      <Panel title={t('基本信息')}>
        <bk-form form-type='vertical' ref={formRef} model={props.modelValue} class={cssModule.home}>
          <bk-form-item label={t('运营产品')}>
            <bk-input disabled={true} loading={opRelationLoading.value} modelValue={productName.value} />
            {isShowNoOpRelation.value && (
              <div class={cssModule['op-relation']}>
                <span class={cssModule['warning-text']}> {t('当前业务无运营产品，')}</span>
                {t('请联系')}
                <WName name={'ICR'} alias={t('ICR(IEG资源服务助手)')}></WName>
                {t('确认')}
              </div>
            )}
          </bk-form-item>
          <bk-form-item label={t('运营产品关联业务')}>
            <bk-input disabled={true} loading={bizListRelationLoading.value} modelValue={bizNameList.value}></bk-input>
          </bk-form-item>
          <bk-alert theme='warning' class={cssModule['biz-list']}>
            {t(`注意：当前运营产品有${bizListLength.value}个业务，资源预测额度在这${bizListLength.value}个业务中共用`)}
          </bk-alert>
          <bk-form-item label={t('预测用途')} property='demand_class' required class={cssModule['forecast-type']}>
            <bk-select
              clearable
              loading={isLoadingDemandClasses.value}
              modelValue={props.modelValue.demand_class}
              onChange={(val: string) => handleUpdateModelValue('demand_class', val)}>
              {demandClasses.value.map((demandClass) => (
                <bk-option id={demandClass} name={demandClass}></bk-option>
              ))}
            </bk-select>
          </bk-form-item>
        </bk-form>
      </Panel>
    );
  },
});
