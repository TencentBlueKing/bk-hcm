import { computed, defineComponent, reactive, ref } from 'vue';
import { Container, Button, Switcher, Form, Tag, Input } from 'bkui-vue';
import { BkRadio, BkRadioGroup } from 'bkui-vue/lib/radio';
import './index.scss';
import CommonSideslider from '@/components/common-sideslider';

const { Row, Col } = Container;
const { FormItem } = Form;

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
    const isHealthCheckupConfigShow = ref(false);
    const formData = reactive({
      sourceIp: '1',
      healthCheckType: 'TCP',
      port: '',
      responseTimeout: '',
      healthCheckInterval: '',
      unhealthyThreshold: '',
      healthyThreshold: '',
    });
    const formItemOptions = computed(() => [
      {
        label: '探测来源IP',
        property: 'sourceIp',
        required: true,
        content: () => (
          <BkRadioGroup v-model={formData.sourceIp} class='radio-groups'>
            <BkRadio label='1'>
              <div class='radio-item-wrap'>
                <div class='item-label'>
                  云专用探测 IP 段
                  <Tag class='ml12' theme='success'>
                    推荐
                  </Tag>
                </div>
                <div class='item-desc'>
                  腾讯云内网专用探测网段是100.64.0.0/10，非固定IP，安全组默认放通该网段。
                  <br />
                  后端服务器有iptables等其他安全策略时，需放通此网段
                </div>
              </div>
            </BkRadio>
            <BkRadio label='2'>
              <div class='radio-item-wrap'>
                <div class='item-label'>负载均衡 VIP</div>
                <div class='item-desc'>需同时在后端服务器安全组和iptables放通VIP地址</div>
              </div>
            </BkRadio>
          </BkRadioGroup>
        ),
      },
      {
        label: '检查方式',
        property: 'healthCheckType',
        required: true,
        content: () => (
          <BkRadioGroup v-model={formData.healthCheckType}>
            <BkRadio label='TCP'>TCP</BkRadio>
            <BkRadio label='HTTP'>HTTP</BkRadio>
            <BkRadio label='custom'>自定义</BkRadio>
          </BkRadioGroup>
        ),
      },
      {
        label: '检查端口',
        property: 'port',
        required: true,
        content: () => <Input v-model={formData.port} />,
      },
      [
        {
          label: '响应超时',
          property: 'responseTimeout',
          required: true,
          span: 12,
          content: () => <Input v-model={formData.responseTimeout} placeholder='0' type='number' suffix='秒' />,
        },
        {
          label: '检查间隔',
          property: 'healthCheckInterval',
          required: true,
          span: 12,
          content: () => <Input v-model={formData.healthCheckInterval} placeholder='0' type='number' suffix='秒' />,
        },
      ],
      [
        {
          label: '不健康阈值',
          property: 'unhealthyThreshold',
          required: true,
          span: 12,
          content: () => <Input v-model={formData.unhealthyThreshold} placeholder='0' type='number' suffix='秒' />,
        },
        {
          label: '健康阈值',
          property: 'healthyThreshold',
          required: true,
          span: 12,
          content: () => <Input v-model={formData.healthyThreshold} placeholder='0' type='number' suffix='秒' />,
        },
      ],
    ]);
    return () => (
      <div class='health-checkup-page'>
        <Button
          class='fixed-operate-btn'
          outline
          theme='primary'
          onClick={() => (isHealthCheckupConfigShow.value = true)}>
          配置
        </Button>
        <div class='detail-info-container'>
          {healthDetailInfo.value.map(({ label, value }) => {
            let valueVNode = null;
            if (typeof value === 'function') {
              valueVNode = value();
            } else {
              if (isOpen.value) {
                if (Array.isArray(value)) {
                  valueVNode = value.map((v) => (
                    <>
                      {' '}
                      {v};<br />{' '}
                    </>
                  ));
                } else {
                  valueVNode = value;
                }
              } else {
                valueVNode = '-';
              }
            }
            return (
              <div class='info-item'>
                <span class='info-item-label'>{label}</span>
                {typeof value === 'function' ? null : ':'}
                <span class={`info-item-value${Array.isArray(value) ? ' multiline' : ''}`}>{valueVNode}</span>
              </div>
            );
          })}
        </div>
        <CommonSideslider
          class='health-checkup-config-sideslider'
          v-model:isShow={isHealthCheckupConfigShow.value}
          title='健康检查配置'
          width='640'>
          <Form formType='vertical'>
            <Container margin={0}>
              {formItemOptions.value.map((item) => (
                <Row>
                  {Array.isArray(item) ? (
                    item.map(({ label, property, required, span, content }) => (
                      <Col span={span}>
                        <FormItem label={label} property={property} required={required}>
                          {content()}
                        </FormItem>
                      </Col>
                    ))
                  ) : (
                    <Col span={24}>
                      <FormItem label={item.label} property={item.property} required={item.required}>
                        {item.content()}
                      </FormItem>
                    </Col>
                  )}
                </Row>
              ))}
            </Container>
          </Form>
        </CommonSideslider>
      </div>
    );
  },
});
