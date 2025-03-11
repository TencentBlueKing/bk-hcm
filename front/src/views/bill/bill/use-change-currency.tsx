import { ref } from 'vue';
import { ArrowsRight } from 'bkui-vue/lib/icon';
import { formatBillSymbol } from '@/utils';
import { CURRENCY_ALIAS_MAP } from '@/constants';

const FIELD_MAP = {
  current_month_cost_synced: ['current_month_cost_synced', 'current_month_rmb_cost_synced'],
  current_month_cost: ['current_month_cost', 'current_month_rmb_cost'],
  last_month_cost_synced: ['last_month_cost_synced', 'last_month_rmb_cost_synced'],
  adjustment_cost: ['adjustment_cost', 'adjustment_rmb_cost'],
};

export default () => {
  const changeCurrencyChecked = ref(false);

  const customRender = (args: any, field: string) => {
    const { data } = args;
    const [money, converted] = FIELD_MAP[field];
    const { currency } = data;
    const normalData = formatBillSymbol(data[money], currency);
    if (!changeCurrencyChecked.value || currency !== CURRENCY_ALIAS_MAP.USD) {
      return normalData;
    }
    return (
      <div class={'current-currency'}>
        <span class={'dollar'}>{normalData}</span>
        <ArrowsRight class={'arrow-right'} />
        <span> {formatBillSymbol(data[converted], CURRENCY_ALIAS_MAP.CNY)} </span>
      </div>
    );
  };

  const handleChangeCurrencyChecked = (val: boolean) => {
    changeCurrencyChecked.value = val;
  };

  return {
    handleChangeCurrencyChecked,
    customRender,
  };
};
