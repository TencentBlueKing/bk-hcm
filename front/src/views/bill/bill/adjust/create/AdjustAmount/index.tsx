import { PropType, computed, defineComponent } from 'vue';
import './index.scss';
import { formatBillCost } from '@/utils';

export default defineComponent({
  props: {
    adjustData: Object as PropType<{
      increaseSum: number;
      decreaseSum: number;
    }>,
    currency: String as PropType<'RMB' | 'USD'>,
  },
  setup(props) {
    const moneySign = computed(() => {
      return props.currency === 'RMB' ? '￥' : '$';
    });
    return () => (
      <div class={'amount-wrapper'}>
        <span class={'item'}>
          增加：
          <span class={'money'}>{`${moneySign.value} ${formatBillCost(`${props.adjustData.increaseSum}`)}`}</span>
        </span>
        <span class={'item'}>
          减少：
          <span class={'money'}>{`${moneySign.value} ${formatBillCost(`${props.adjustData.decreaseSum}`)}`}</span>
        </span>
      </div>
    );
  },
});
