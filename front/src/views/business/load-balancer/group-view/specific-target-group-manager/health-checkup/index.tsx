import { computed, defineComponent, ref } from 'vue';
import { Button, Switcher } from 'bkui-vue';
import './index.scss';

export default defineComponent({
  name: 'HealthCheckupPage',
  setup() {
    const isOpen = ref(true);
    const healthDetailInfo = computed(() => [
      {
        label: '是否启用',
        value: () => <Switcher v-model={isOpen.value} theme='primary' />,
      },
      {
        label: '健康探测源IP',
        value: '100.64.0.0/10网段',
      },
      {
        label: '检查方式',
        value: 'TCP',
      },
      {
        label: '检查端口',
        value: '4600',
      },
      {
        label: '检查端口',
        value: ['响应超时(2秒)', '检查间隔(5秒)', '不健康阈值(3次)', '健康阈值(3次),'],
      },
    ]);
    return () => (
      <div class='health-checkup-page'>
        <Button class='fixed-operate-btn' outline theme='primary'>
          配置
        </Button>
        <div class='detail-info-container'>
          {healthDetailInfo.value.map(({ label, value }) => {
            let _valueNode = null;
            if (typeof value === 'function') {
              _valueNode = value();
            } else {
              if (isOpen.value) {
                if (Array.isArray(value)) {
                  _valueNode = value.map(v => (
                    <>
                      {' '}
                      {v};<br />{' '}
                    </>
                  ));
                } else {
                  _valueNode = value;
                }
              } else {
                _valueNode = '-';
              }
            }
            return (
              <div class='info-item'>
                <span class='info-item-label'>{label}</span>
                {typeof value === 'function' ? null : ':'}
                <span class={`info-item-value${Array.isArray(value) ? ' multiline' : ''}`}>
                  {_valueNode}
                </span>
              </div>
            );
          })}
        </div>
      </div>
    );
  },
});
