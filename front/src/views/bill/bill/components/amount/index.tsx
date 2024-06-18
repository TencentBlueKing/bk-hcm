import { defineComponent } from 'vue';
import cssModule from './index.module.scss';

export default defineComponent({
  setup() {
    return () => (
      <div class={cssModule['amount-wrapper']}>
        <span>
          总金额（人民币）：<span class={cssModule['money-text']}>￥xxx</span>
        </span>
        <span>
          总金额（美元）：<span class={cssModule['money-text']}>＄xxx</span>
        </span>
      </div>
    );
  },
});
