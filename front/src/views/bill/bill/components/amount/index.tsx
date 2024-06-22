import { PropType, defineComponent } from 'vue';
import cssModule from './index.module.scss';
import { useI18n } from 'vue-i18n';

export default defineComponent({
  props: {
    isAdjust: Boolean,
    showType: {
      type: String as PropType<'vertical' | 'horizontal'>,
      default: 'horizontal',
    },
  },
  setup(props) {
    const { t } = useI18n();

    return () => (
      <div
        class={{
          [cssModule['amount-wrapper']]: true,
          [cssModule.vertical]: props.showType === 'vertical',
        }}>
        <span>
          {t('共计')}
          {props.isAdjust ? t('增加') : t('人民币')}：<span class={cssModule.money}>xxx</span>
          {props.isAdjust && (
            <>
              &nbsp;|&nbsp;<span class={cssModule.money}>xxx</span>
            </>
          )}
        </span>
        <span>
          {t('共计')}
          {props.isAdjust ? t('减少') : t('美金')}：<span class={cssModule.money}>xxx</span>
          {props.isAdjust && (
            <>
              &nbsp;|&nbsp;<span class={cssModule.money}>xxx</span>
            </>
          )}
        </span>
      </div>
    );
  },
});
