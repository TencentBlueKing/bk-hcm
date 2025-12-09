import { defineComponent, computed, onBeforeMount, ref, PropType } from 'vue';
import { useI18n } from 'vue-i18n';
import Panel from '@/components/panel';
import { timeFormatter } from '@/common/util';
import cssModule from './index.module.scss';
import { useResourcePlanStore } from '@/store';
import { useRoute } from 'vue-router';
import { IPlanDemandResult } from '@/typings/resourcePlan';
import { useBusinessGlobalStore } from '@/store/business-global';

export default defineComponent({
  props: {
    isBiz: {
      type: Boolean,
      default: true,
    },
    updateCurrentBusinessId: Function as PropType<(id: number) => void>,
  },
  setup(props) {
    const { t } = useI18n();
    const route = useRoute();
    const { getPlanDemand, getPlanDemandByOrg } = useResourcePlanStore();
    const { getBusinessNames } = useBusinessGlobalStore();

    const baseInfo = ref<IPlanDemandResult['data']>();
    const isLoading = ref(false);

    const baseList = computed(() => [
      {
        label: t('业务'),
        value: baseInfo.value?.bk_biz_id
          ? getBusinessNames(baseInfo.value.bk_biz_id)?.[0] ?? baseInfo.value.bk_biz_id
          : '--',
      },
      { label: t('资源模式'), value: baseInfo.value?.res_mode },
      { label: t('运营产品'), value: baseInfo.value?.op_product_name },
      { label: t('机型族'), value: baseInfo.value?.device_family },
      { label: t('机型类型'), value: baseInfo.value?.device_class },
      { label: t('机型规格'), value: baseInfo.value?.device_type },
      { label: t('项目类型'), value: baseInfo.value?.obs_project },
      { label: t('实例数'), value: baseInfo.value?.os },
      { label: t('期望到货时间'), value: timeFormatter(baseInfo.value?.expect_time) },
      { label: t('核心类型'), value: baseInfo.value?.core_type },
      { label: t('预测类型'), value: baseInfo.value?.plan_type },
      { label: t('云磁盘类型'), value: baseInfo.value?.disk_type_name },
      { label: t('地域'), value: baseInfo.value?.area_name },
      { label: t('单实例磁盘IO(MB/s)'), value: baseInfo.value?.disk_io },
      { label: t('城市'), value: baseInfo.value?.region_name },
      { label: t('总磁盘大小'), value: baseInfo.value?.disk_size },
    ]);

    const getPlanDemandDetail = async () => {
      const { bizs, demandId } = route.query;
      isLoading.value = true;
      try {
        const result = props.isBiz
          ? await getPlanDemand(+bizs, demandId as string)
          : await getPlanDemandByOrg(demandId as string);
        baseInfo.value = result.data;

        // 设置当前业务
        props.updateCurrentBusinessId(result.data.bk_biz_id);
      } catch (error) {
        console.error('Error fetching plan demand details:', error);
      } finally {
        isLoading.value = false;
      }
    };

    onBeforeMount(() => {
      getPlanDemandDetail();
    });

    return () => (
      <Panel title={t('基本信息')}>
        <bk-loading loading={isLoading.value}>
          <ul class={cssModule.home}>
            {baseList.value.map((item) => (
              <li>
                <span class={cssModule.label}>{item.label}：</span>
                <span class={cssModule.value} title={item.value as string}>
                  {item.value || '--'}
                </span>
              </li>
            ))}
          </ul>
        </bk-loading>
      </Panel>
    );
  },
});
