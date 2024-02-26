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
      type: String as PropType<'flow' | 'grid'>,
      default: 'flow',
    },
  },
  setup(props, { slots }) {
    return () => (
      <Card
        class={'common-card'}
        border={false}
        showHeader={false}
        showFooter={false}
      >
        <p class={'common-card-title'}>
          {
            props.title?.()
          }
        </p>
        <div class={`common-card-content ${props.layout === 'grid' ? 'common-card-content-grid-layout' : ''}`}>
          {
            slots.default()
          }
        </div>
      </Card>
    );
  },
});
