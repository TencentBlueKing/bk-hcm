import { computed, defineComponent, type PropType, ref, watch } from 'vue';
import { useResourcePlanStore } from '@/store';
import CommonSideslider from '@/components/common-sideslider';
import Basic from './basic';
import CVM from './cvm';
import CBS from './cbs';
import cssModule from './index.module.scss';
import { useI18n } from 'vue-i18n';
import type { IPlanTicket, IPlanTicketDemand } from '@/typings/resourcePlan';
import Type from './type';
import { AdjustType } from '@/typings/plan';
import { isEqual, omit } from 'lodash';

export default defineComponent({
  props: {
    isShow: {
      type: Boolean,
    },
    modelValue: {
      type: Object as PropType<IPlanTicket>,
    },
    initDemand: {
      type: Object as PropType<IPlanTicketDemand>,
    },
    // 接口原始数据，用于判断某条预测单最终是否变更，不计编辑次数
    originDemand: {
      type: Object as PropType<IPlanTicketDemand>,
    },
    isEdit: {
      type: Boolean,
    },
    // 添加时自动填充的参数数据，通常来源于页面跳转自动打开添加页面时传入
    initAddParams: {
      type: Object as PropType<Partial<IPlanTicketDemand>>,
    },
    currentGlobalBusinessId: Number,
  },

  emits: ['update:isShow', 'update:modelValue', 'updateDemand', 'hidden'],

  setup(props, { emit }) {
    const { t } = useI18n();
    const resourcePlanStore = useResourcePlanStore();

    const basicRef = ref(null);
    const cvmRef = ref(null);
    const cbsRef = ref(null);
    const resourceType = ref('cvm');
    const isSubmitDisabled = ref(false);
    const submitTooltips = ref({
      content: '',
      disabled: true,
    });
    const planTicketDemand = ref<IPlanTicketDemand>();
    const adjustType = ref();

    // 记录当前次编辑操作的原始数据，用于当前次操作场景的判断
    let currEditOriginPlanTicketDemand: IPlanTicketDemand;
    const is931Business = computed(() => props.currentGlobalBusinessId === 931);
    const initData = async () => {
      resourceType.value =
        props.initDemand?.demand_res_types.length < 2
          ? props.initDemand.demand_res_types?.[0]?.toLocaleLowerCase()
          : 'cvm';
      adjustType.value =
        props.initDemand && props.initDemand.adjustType === AdjustType.time ? AdjustType.time : AdjustType.config;

      currEditOriginPlanTicketDemand = {
        obs_project: props.initAddParams?.obs_project || '',
        expect_time: '',
        region_id: props.initAddParams?.region_id || '',
        zone_id: props.initAddParams?.zone_id || '',
        demand_source: '指标变化',
        remark: '',
        demand_res_types: ['CVM', 'CBS'],
        cvm: { res_mode: '按机型', device_class: '', device_type: '', os: 0, cpu_core: 0, memory: 0 },
        cbs: { disk_type: '', disk_type_name: '', disk_io: 15, disk_size: 0, disk_num: 0, disk_per_size: 0 },
        ...props.initDemand,
      };

      planTicketDemand.value = { ...currEditOriginPlanTicketDemand };

      // 回填初始的device_type和device_class数据
      if (resourceType.value === 'cvm' && props.initAddParams?.cvm?.device_type) {
        const result = await resourcePlanStore.getDeviceTypes(undefined, [props.initAddParams?.cvm?.device_type]);
        if (result?.data?.details?.length) {
          planTicketDemand.value.cvm.device_type = result.data.details[0].device_type;
          planTicketDemand.value.cvm.device_class = result.data.details[0].device_class;
        }
      }
    };

    const disableType = ref<AdjustType>(AdjustType.none);
    const handleDisableTypeChange = (type: AdjustType) => {
      disableType.value = type;
    };

    const handleClose = () => {
      emit('update:isShow', false);
    };

    const handleSubmit = async () => {
      await validate();
      if (props.initDemand) {
        const demandIndex = props.modelValue.demands.findIndex((demand) => demand === props.initDemand);
        emit('update:modelValue', {
          ...props.modelValue,
          demands: [
            ...props.modelValue.demands.slice(0, demandIndex),
            planTicketDemand.value,
            ...props.modelValue.demands.slice(demandIndex + 1),
          ],
        });
      } else {
        emit('update:modelValue', {
          ...props.modelValue,
          demands: [...props.modelValue.demands, planTicketDemand.value],
        });
      }
      handleClose();
    };

    const handleUpdate = async () => {
      await validate();
      const ignoreFields = ['adjustType', 'remark'];
      const isChanged = !isEqual(omit(props.originDemand, ignoreFields), omit(planTicketDemand.value, ignoreFields));
      emit('updateDemand', { ...planTicketDemand.value, adjustType: isChanged ? adjustType.value : AdjustType.none });
      handleClose();
    };

    const validate = () => {
      return Promise.all([basicRef.value.validate(), cvmRef.value.validate(), cbsRef.value.validate()]);
    };

    const clearValidate = () => {
      Promise.all([basicRef.value?.clearValidate(), cvmRef.value?.clearValidate(), cbsRef.value?.clearValidate()]);
    };

    const handleShown = () => {
      clearValidate();
    };

    // 单独监听 props.initDemand 和 props.initAddParams 的变化，用于初始化数据
    watch(() => [props.initDemand, props.initAddParams], initData, { immediate: true, deep: true });
    watch(
      () => props.isShow,
      (isShow) => {
        if (isShow) {
          initData();
        }
      },
      { immediate: true },
    );

    watch(() => resourceType.value, clearValidate);

    return () => (
      <CommonSideslider
        width='960'
        class={cssModule.home}
        isSubmitDisabled={isSubmitDisabled.value}
        submitTooltips={submitTooltips.value}
        isShow={props.isShow}
        title={props.initDemand ? t('修改预测需求') : t('增加预测需求')}
        handleClose={handleClose}
        onUpdate:isShow={handleClose}
        onHandleSubmit={props.isEdit ? handleUpdate : handleSubmit}
        onHandleShown={handleShown}
        onHidden={() => emit('hidden')}>
        {props.initDemand && props.isEdit && (
          <Type v-model={adjustType.value} type={props.initDemand.adjustType} disableType={disableType.value} />
        )}
        <Basic
          ref={basicRef}
          v-model:isSubmitDisabled={isSubmitDisabled.value}
          v-model:submitTooltips={submitTooltips.value}
          v-model:planTicketDemand={planTicketDemand.value}
          v-model:resourceType={resourceType.value}
          type={props.isEdit ? adjustType.value : AdjustType.none}
          originPlanTicketDemand={currEditOriginPlanTicketDemand}
          is931Business={is931Business.value}
          onDisableTypeChange={handleDisableTypeChange}
        />
        <CVM
          ref={cvmRef}
          v-model:planTicketDemand={planTicketDemand.value}
          resourceType={resourceType.value}
          class={cssModule.mt16}
          type={props.isEdit ? adjustType.value : AdjustType.none}
          originPlanTicketDemand={currEditOriginPlanTicketDemand}
          onDisableTypeChange={handleDisableTypeChange}
        />
        <CBS
          ref={cbsRef}
          v-model:planTicketDemand={planTicketDemand.value}
          resourceType={resourceType.value}
          class={cssModule.mt16}
          type={props.isEdit ? adjustType.value : AdjustType.none}
          originPlanTicketDemand={currEditOriginPlanTicketDemand}
          onDisableTypeChange={handleDisableTypeChange}
        />
      </CommonSideslider>
    );
  },
});
