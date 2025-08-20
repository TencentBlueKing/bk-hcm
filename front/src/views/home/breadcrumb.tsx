import { computed, defineComponent } from 'vue';
import { useRoute } from 'vue-router';
import useBreadcrumb from '@/hooks/use-breadcrumb';
import { useBack } from '@/router/hooks/use-back';
import type { RouteMetaConfig } from '@/router/meta';

import './breadcrumb.scss';

export default defineComponent({
  setup() {
    const { breadcrumb } = useBreadcrumb();
    const { from, handleBack } = useBack();
    const route = useRoute();

    const currentTitle = computed(() => {
      const routeMeta = route.meta as RouteMetaConfig;
      return breadcrumb.title ?? routeMeta.title ?? routeMeta?.menu?.i18n;
    });

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
              <i
                onClick={() => this.handleBack()}
                class={'icon hcm-icon bkhcm-icon-arrows--left-line pr10 back-icon'}
              />
            )}
            <span class='breadcrumb-name'>{this.currentTitle}</span>
          </div>
          <div id='breadcrumbExtra' class='breadcrumb-extra'></div>
        </div>
      )
    );
  },
});
