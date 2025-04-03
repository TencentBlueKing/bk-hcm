import { defineComponent, PropType, ref } from 'vue';

import './detail-tab.scss';

type Tab = {
  name: string;
  value: string;
};

export default defineComponent({
  props: {
    tabs: Array as PropType<Tab[]>,
    active: String as PropType<any>,
    onChange: Function as PropType<(val: string) => void>,
  },

  setup(props) {
    const activeTab = ref(props.active || props.tabs[0].value);

    return {
      activeTab,
    };
  },

  render() {
    return (
      <>
        <bk-tab
          v-model:active={this.activeTab}
          type='card-grid'
          class={`detail-tab-main ${this.$attrs?.class}`}
          onChange={this.onChange}>
          {this.tabs.map((tab) => {
            return (
              <>
                <bk-tab-panel name={tab.value} label={tab.name} key={tab.name}>
                  {tab.value === this.activeTab ? this.$slots.default(this.activeTab) : ''}
                </bk-tab-panel>
              </>
            );
          })}
        </bk-tab>
      </>
    );
  },
});
