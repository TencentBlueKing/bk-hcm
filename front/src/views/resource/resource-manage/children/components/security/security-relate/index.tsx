import { defineComponent, ref } from 'vue';
import { BkRadioButton, BkRadioGroup } from 'bkui-vue/lib/radio';
import SecurityCvmTable from './SecurityCvmTable';
import SecurityLbTable from './SecurityLbTable';
import './index.scss';

export default defineComponent({
  name: 'SecurityRelate',
  setup() {
    const types = ref([
      { label: '云主机', value: 'cvm' },
      { label: '负载均衡', value: 'lb' },
    ]);
    const selectedType = ref<'cvm' | 'lb'>('cvm');

    return () => (
      <div class='security-relate-page'>
        <section class='top-bar'>
          <BkRadioGroup v-model={selectedType.value} class='tabs-wrap'>
            {types.value.map(({ label, value }) => (
              <BkRadioButton key={value} label={value} class='mw88'>
                {label}
              </BkRadioButton>
            ))}
          </BkRadioGroup>
        </section>
        <section class='table-wrap'>
          {(function () {
            switch (selectedType.value) {
              case 'cvm':
                return <SecurityCvmTable />;
              case 'lb':
                return <SecurityLbTable />;
            }
          })()}
        </section>
      </div>
    );
  },
});
