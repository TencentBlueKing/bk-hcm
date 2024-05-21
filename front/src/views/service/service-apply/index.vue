<template>
  <div class="template-wrap service-wrap">
    <Tab v-model:active="activeTab" type="card">
      <TabPanel v-for="item in panels" :key="item.name" :name="item.name" :label="item.label">
        <div :style="{ maxHeight: `${cardHeight}px`, overflowY: 'auto', paddingBottom: '50px' }">
          <component v-if="item.name === activeTab" :key="item.name" :is="activeTab" />
        </div>
      </TabPanel>
    </Tab>
  </div>
</template>

<script lang="ts">
import { defineComponent, reactive, toRefs, computed } from 'vue';
import { useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { Tab } from 'bkui-vue';
import { getWindowHeight } from '@/common/util';
import BasicResource from './components/basic-resource/index.vue';
import SeniorResource from './components/senior-resource/index.vue';

const { TabPanel } = Tab;

export default defineComponent({
  name: 'ServiceApply',
  components: {
    Tab,
    TabPanel,
    BasicResource,
    SeniorResource,
  },
  setup() {
    const { t } = useI18n();
    useRouter();
    const tabInfo = reactive({
      currentType: 'card',
      activeTab: '',
      panels: [
        { name: 'BasicResource', label: t('基础'), count: 10 },
        // { name: 'SeniorResource', label: t('高级'), count: 20 },
      ],
    });

    const cardHeight = computed(() => {
      return getWindowHeight() - 260;
    });

    return {
      ...toRefs(tabInfo),
      cardHeight,
    };
  },
});
</script>

<style lang="scss" scoped>
.service-wrap {
  width: 100%;
  min-height: calc(100% - 30px);
  border-radius: 2px;
  background: #ffffff;
  box-shadow: 0px 2px 4px 0px rgb(0 0 0 / 10%);
}
</style>
