import { RESOURCE_DETAIL_TABS } from '@/common/constant';
import { Tab } from 'bkui-vue';
import { BkTabPanel } from 'bkui-vue/lib/tab';
import { defineComponent, ref, watch } from 'vue';
import { RouterView, useRoute, useRouter } from 'vue-router';
import './index.scss';
import { useResourceAccountStore } from '@/store/useResourceAccountStore';

export default defineComponent({
  setup() {
    const activeTab = ref(RESOURCE_DETAIL_TABS[0].key);
    const router = useRouter();
    const route = useRoute();
    const resourceAccountStore = useResourceAccountStore();
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
          <Tab v-model:active={activeTab.value} type='card-grid'>
            {RESOURCE_DETAIL_TABS.filter(({ key }) => resourceAccountStore.resourceAccount?.id || key === '/resource/resource/account/manage').map(({ key, label }) => (
              <BkTabPanel key={key} label={label} name={key}>
                <RouterView />
              </BkTabPanel>
            ))}
          </Tab>
        </div>
      </>
    );
  },
});
