import { defineComponent, PropType } from 'vue';
import { RouteLocationRaw, useRoute, useRouter } from 'vue-router';
import { Senarios, useWhereAmI } from '@/hooks/useWhereAmI';
import { useCalcTopWithNotice } from '@/views/home/hooks/useCalcTopWithNotice';
import { useBack } from '@/router/hooks/use-back';
import { GLOBAL_BIZS_KEY } from '@/common/constant';

import { ArrowsLeft } from 'bkui-vue/lib/icon';
import './detail-header.scss';

export default defineComponent({
  components: { ArrowsLeft },
  props: { to: Object as PropType<RouteLocationRaw>, fromConfig: Object as PropType<Partial<RouteLocationRaw>> },
  setup(props) {
    const router = useRouter();
    const route = useRoute();
    const { whereAmI } = useWhereAmI();

    const [calcTop] = useCalcTopWithNotice(52);
    const { handleBack } = useBack();

    const goBack = () => {
      const { to, fromConfig } = props;
      if (to) {
        router.replace(to);
        return;
      }
      if (fromConfig) {
        handleBack(fromConfig);
        return;
      }
      if (window.history.state.back) {
        router.back();
        return;
      }
      // TODO：补齐relative之前，先跳转到首页
      router.replace({
        path: '/',
        query: {
          [GLOBAL_BIZS_KEY]: whereAmI.value === Senarios.business ? route.query?.[GLOBAL_BIZS_KEY] : undefined,
        },
      });
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
