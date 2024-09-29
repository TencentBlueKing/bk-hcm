import { defineComponent } from 'vue';

import { ArrowsLeft } from 'bkui-vue/lib/icon';

import { useRoute, useRouter } from 'vue-router';

import './detail-header.scss';
import { Senarios, useWhereAmI } from '@/hooks/useWhereAmI';
import { useCalcTopWithNotice } from '@/views/home/hooks/useCalcTopWithNotice';

export default defineComponent({
  components: { ArrowsLeft },
  props: { backRouteName: String },
  setup(props) {
    const router = useRouter();
    const route = useRoute();
    const { whereAmI } = useWhereAmI();

    const [calcTop] = useCalcTopWithNotice(52);

    const goBack = () => {
      if (props.backRouteName) {
        router.replace({ name: props.backRouteName, query: { ...route.query } });
        return;
      }
      if (window.history.state.back) {
        router.back();
      } else {
        router.replace({
          path: '/resource/resource',
          query: {
            type: 'subnet',
          },
        });
      }
    };

    return {
      goBack,
      whereAmI,
      calcTop,
    };
  },

  render() {
    return (
      <section
        class={`detail-header-main ${this.whereAmI === Senarios.resource ? 'ml-24' : ''}`}
        style={{ width: this.whereAmI === Senarios.resource ? '85%' : 'calc(100% - 240px)', top: this.calcTop }}>
        <div class='title-content'>
          <arrows-left class='detail-header-arrows-left' onClick={this.goBack} />
          {this.$slots.default?.()}
        </div>
        <div>{this.$slots.right?.()}</div>
      </section>
    );
  },
});
