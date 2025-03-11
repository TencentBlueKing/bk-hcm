import { ref } from 'vue';
import { ArrowsRight } from 'bkui-vue/lib/icon';
import { formatBillSymbol } from '@/utils';
import { CURRENCY_ALIAS_MAP } from '@/constants';

export default () => {
  const changeCurrencyChecked = ref(false);

  const getColElement = (money: string, converted: string, currency: string) => {
    const normalData = formatBillSymbol(money, currency);
    if (!changeCurrencyChecked.value || currency !== CURRENCY_ALIAS_MAP.USD) {
      return normalData;
    }
    return (
      <div class={'current-currency'}>
        <span class={'dollar'}>{normalData}</span>
        <ArrowsRight class={'arrow-right'} />
        <span> {formatBillSymbol(converted, CURRENCY_ALIAS_MAP.CNY)} </span>
      </div>
    );
  };

  const handleChangeCurrencyChecked = (val: boolean) => {
    changeCurrencyChecked.value = val;
  };

  return {
    handleChangeCurrencyChecked,
    getColElement,
  };
};
