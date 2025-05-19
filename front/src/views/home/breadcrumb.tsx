import { computed, defineComponent } from 'vue';
import { useRoute } from 'vue-router';
import useBreadcrumb from '@/hooks/use-breadcrumb';
import type { RouteMetaConfig } from '@/router/meta';
import routerAction from '@/router/utils/action';
import { HistoryStorage } from '@/router/utils/history-storage';

import './breadcrumb.scss';

export default defineComponent({
  setup() {
    const { breadcrumb } = useBreadcrumb();
    const route = useRoute();

    const currentTitle = computed(() => {
      const routeMeta = route.meta as RouteMetaConfig;
      return breadcrumb.title ?? routeMeta.title ?? routeMeta?.menu?.i18n;
    });

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

    const handleBack = () => {
      routerAction.redirect(from.value, { back: true });
    };

    return {
      from,
      breadcrumb,
      currentTitle,
      handleBack,
    };
  },

  render() {
    return (
      this.breadcrumb.display && (
        <div class='navigation-breadcrumb'>
          <div class='breadcrumb-content'>
            {this.from && (
              <i onClick={this.handleBack} class={'icon hcm-icon bkhcm-icon-arrows--left-line pr10 back-icon'} />
            )}
            <span class='breadcrumb-name'>{this.currentTitle}</span>
          </div>
          <div id='breadcrumbExtra' class='breadcrumb-extra'></div>
        </div>
      )
    );
  },
});
