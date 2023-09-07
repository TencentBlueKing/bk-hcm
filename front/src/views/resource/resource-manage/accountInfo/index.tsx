import { RESOURCE_DETAIL_TABS } from '@/common/constant';
import { Tab } from 'bkui-vue';
import { BkTabPanel } from 'bkui-vue/lib/tab';
import { defineComponent, ref, watch } from 'vue';
import { RouterView, useRoute, useRouter } from 'vue-router';

export default defineComponent({
  setup() {
    const activeTab = ref(RESOURCE_DETAIL_TABS[0].key);
    const router = useRouter();
    const route = useRoute();
    watch(
      () => activeTab.value,
      (val) => {
        router.push({
          path: val,
          query: route.query,
        });
      },
      {
        immediate: true,
      },
    );
    return () => (
      <>
        <div class={'account-info-container'}>
          <Tab
            v-model:active={activeTab}
            type='card-grid'>
               {
                RESOURCE_DETAIL_TABS.map(({ key, label }) => (
                  <BkTabPanel
                    key={key}
                    label={label}
                    name={key}
                  >
                    <RouterView/>
                  </BkTabPanel>
                ))
               }
          </Tab>
        </div>

      </>
    );
  },
});
