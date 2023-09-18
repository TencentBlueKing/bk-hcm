import { Card } from 'bkui-vue';
import { PropType, defineComponent } from 'vue';
import './index.scss';

export default defineComponent({
  props: {
    title: {
      type: Function as PropType<() => string | HTMLElement>,
      required: true,
    },
    layout: {
      type: String as unknown as PropType<'flow' | 'grid'>,
      default: 'flow',
    },
  },
  setup(props, { slots }) {
    return () => (
      <Card
        class={'common-card'}
        showHeader={false}
        showFooter={false}
      >
        <p class={'common-card-title'}>
          {
            props.title?.()
          }
        </p>
        <div class={`account-form-card-content ${true ? 'common-card-content-grid-layout' : ''}`}>
          {
            slots.default()
          }
          <div>999999696969</div>
        </div>
      </Card>
    );
  },
});
