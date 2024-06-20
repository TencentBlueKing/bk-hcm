import { defineComponent } from 'vue';
import cssModule from './index.module.scss';

export default defineComponent({
  props: {
    title: {
      type: String,
    },
  },

  setup(props, { slots }) {
    return () => (
      <section class={cssModule.home}>
        {props.title ? <span class={cssModule.title}>{props.title}</span> : ''}
        {slots.default()}
      </section>
    );
  },
});
