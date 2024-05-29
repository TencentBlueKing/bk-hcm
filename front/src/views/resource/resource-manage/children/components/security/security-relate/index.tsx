import { defineComponent, ref } from 'vue';
import { Button } from 'bkui-vue';
import { BkRadioButton, BkRadioGroup } from 'bkui-vue/lib/radio';
import { Plus } from 'bkui-vue/lib/icon';
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

    /**
     * fetch url
     *  POST /api/v1/cloud/security_group/{id}/cvm/list
     *  POST /api/v1/cloud/bizs/{bk_biz_id}/security_group/{id}/cvm/list
     *  POST /api/v1/cloud/security_group/{id}/common/list
     *  POST /api/v1/bizs/{bk_biz_id}/cloud/security_group/{id}/common/list
     */

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
          <div class='operation-btn-wrap'>
            <Button theme='primary'>
              <Plus class='f22' />
              新增绑定
            </Button>
            <Button class='ml12'>批量解绑</Button>
          </div>
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
