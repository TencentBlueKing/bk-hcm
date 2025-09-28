import { computed, defineComponent, PropType } from 'vue';

import { ArrowsLeft } from 'bkui-vue/lib/icon';

import { RouteLocationRaw, useRoute, useRouter } from 'vue-router';
import routerAction from '@/router/utils/action';

import './detail-header.scss';
import { Senarios, useWhereAmI } from '@/hooks/useWhereAmI';
import { useCalcTopWithNotice } from '@/views/home/hooks/useCalcTopWithNotice';
import { HistoryStorage } from '@/router/utils/history-storage';
import { RouteMetaConfig } from '@/router/meta';
import { GLOBAL_BIZS_KEY } from '@/common/constant';

export default defineComponent({
  components: { ArrowsLeft },
  props: { to: Object as PropType<RouteLocationRaw>, useRouterAction: Boolean },
  setup(props) {
    const router = useRouter();
    const route = useRoute();
    const { whereAmI } = useWhereAmI();

    const [calcTop] = useCalcTopWithNotice(52);

    const defaultFrom = computed(() => {
      const routeMeta = route.meta as RouteMetaConfig;
      const menu = routeMeta.menu || {};
      if (menu.relative) {
        return { name: Array.isArray(menu.relative) ? menu.relative[0] : menu.relative };
      }
      return null;
    });

    const from = computed(() => {
      if (Object.hasOwn(route.query, '_f')) {
        try {
          return HistoryStorage.pop();
        } catch (error) {
          return defaultFrom.value;
        }
      }
      return defaultFrom.value;
    });

    const goBack = () => {
      const { to } = props;
      if (to) {
        router.replace(to);
        return;
      }
      if (props.useRouterAction) {
        routerAction.redirect(from.value, { back: true });
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
