import { defineComponent } from 'vue';

import { ArrowsLeft } from 'bkui-vue/lib/icon';

import { useRouter } from 'vue-router';

import './detail-header.scss';
import { Senarios, useWhereAmI } from '@/hooks/useWhereAmI';

export default defineComponent({
  components: {
    ArrowsLeft,
  },

  setup() {
    const router = useRouter();
    const { whereAmI } = useWhereAmI();

    const goBack = () => {
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
    };
  },

  render() {
    return (
      <>
        <section
          class={`detail-header-main ${this.whereAmI === Senarios.resource ? 'ml-24' : ''}`}
          style={{ width: this.whereAmI === Senarios.resource ? '85%' : 'calc(100% - 240px)' }}>
          <div class='title-content'>
            <arrows-left class='detail-header-arrows-left' onClick={this.goBack} />
            {this.$slots.default?.()}
          </div>
          <div>{this.$slots.right?.()}</div>
        </section>
      </>
    );
  },
});
