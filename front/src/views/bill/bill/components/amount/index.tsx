import { PropType, defineComponent, onMounted, ref } from 'vue';
import cssModule from './index.module.scss';
import { useI18n } from 'vue-i18n';
import { BillsSummarySum, BillsSummarySumResData } from '@/typings/bill';
import { Loading } from 'bkui-vue';

export default defineComponent({
  props: {
    isAdjust: Boolean,
    showType: {
      type: String as PropType<'vertical' | 'horizontal'>,
      default: 'horizontal',
    },
    api: Function as PropType<(...args: any) => Promise<BillsSummarySumResData>>,
    payload: Function as PropType<() => object>,
  },
  setup(props, { expose }) {
    const { t } = useI18n();
    const amountInfo = ref<BillsSummarySum>();
    const isLoading = ref(false);

    const getAmountInfo = async () => {
      isLoading.value = true;
      try {
        const res = await props.api(props.payload());
        amountInfo.value = res.data;
      } finally {
        isLoading.value = false;
      }
    };

    onMounted(() => {
      props.api && props.payload && getAmountInfo();
    });

    expose({ refreshAmountInfo: getAmountInfo });

    return () => (
      <div
        class={{
          [cssModule['amount-wrapper']]: true,
          [cssModule.vertical]: props.showType === 'vertical',
        }}>
        <span class={cssModule.item}>
          {t('共计')}
          {props.isAdjust ? t('增加') : t('人民币')}：
          <Loading loading={isLoading.value} opacity={1} style={{ minWidth: '80px' }} size='small'>
            <span class={cssModule.money}>￥{amountInfo.value?.cost_map?.USD?.RMBCost || 0}</span>
            {props.isAdjust && (
              <>
                &nbsp;|&nbsp;<span class={cssModule.money}>xxx</span>
              </>
            )}
          </Loading>
        </span>
        <span class={cssModule.item}>
          {t('共计')}
          {props.isAdjust ? t('减少') : t('美金')}：
          <Loading loading={isLoading.value} opacity={1} style={{ minWidth: '80px' }} size='small'>
            <span class={cssModule.money}>＄{amountInfo.value?.cost_map?.USD?.Cost || 0}</span>
            {props.isAdjust && (
              <>
                &nbsp;|&nbsp;<span class={cssModule.money}>xxx</span>
              </>
            )}
          </Loading>
        </span>
      </div>
    );
  },
});
