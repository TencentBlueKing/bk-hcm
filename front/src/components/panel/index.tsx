/* eslint-disable no-nested-ternary */
import { PropType, VNode, defineComponent, computed } from 'vue';
import cssModule from './index.module.scss';

export default defineComponent({
  props: {
    title: {
      type: [Function, String] as PropType<(() => string | HTMLElement | VNode) | String>,
      default: () => '',
    },
    noShadow: Boolean as PropType<boolean>,
  },

  setup(props, { slots }) {
    const renderTitle = computed(() => (typeof props?.title === 'function' ? props.title() : props.title));
    return () => (
      <section class={!props.noShadow ? cssModule.home : undefined}>
        {slots.title ? slots.title() : props.title ? <span class={cssModule.title}>{renderTitle.value}</span> : null}
        {slots.default()}
      </section>
    );
  },
});
