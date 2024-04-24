import { computed, defineComponent, reactive, ref, watch } from 'vue';
import { Container, Button, Switcher, Form, Tag, Input } from 'bkui-vue';
import { BkRadio, BkRadioGroup } from 'bkui-vue/lib/radio';
import './index.scss';
import CommonSideslider from '@/components/common-sideslider';
import { useBusinessStore, useLoadBalancerStore } from '@/store';

const { Row, Col } = Container;
const { FormItem } = Form;

export default defineComponent({
  name: 'HealthCheckupPage',
  props: {
    detail: {
      required: true,
      type: Object,
    },
    getTargetGroupDetail: Function,
  },
  setup(props) {
    const isOpen = ref(false);
    const loadbalancerStore = useLoadBalancerStore();
    const businessStore = useBusinessStore();
    const isSubmitLoading = ref(false);
    const healthDetailInfo = computed(() => [
      {
        label: '是否启用',
        value: () => (isOpen.value ? '已启用' : '未启用'),
      },
      {
        label: '健康探测源IP',
        value: props.detail.health_check?.health_switch || '-',
      },
      {
        label: '检查方式',
        value: props.detail.health_check?.check_type || '-',
      },
      {
        label: '检查端口',
        value: props.detail.health_check?.check_port || '-',
      },
      {
        label: '检查选型',
        value: [
          `响应超时(${props.detail.health_check?.time_out}秒)`,
          `检查间隔(${props.detail.health_check?.interval_time}秒)`,
          `不健康阈值(${props.detail.health_check?.un_health_num}次)`,
          `健康阈值(${props.detail.health_check?.health_num}次),`,
        ],
      },
    ]);
    const isHealthCheckupConfigShow = ref(false);
    const formData = reactive({
      health_switch: false,
      check_type: 'TCP',
      check_port: '',
      time_out: '',
      interval_time: '',
      un_health_num: '',
      health_num: '',
    });
    function resetFormData() {
      for (const key in formData) {
        if (Object.hasOwnProperty.call(formData, key)) {
          switch (key) {
            case 'health_switch':
              formData[key] = false;
              break;
            case 'check_type':
              formData[key] = 'TCP';
              break;
            default:
              formData[key] = '';
              break;
          }
        }
      }
    }
    const formItemOptions = computed(() => [
      {
        label: '是否启用',
        content: () => <Switcher v-model={formData.health_switch} theme='primary' />,
      },
      {
        label: '探测来源IP',
        property: 'health_switch',
        required: true,
        content: () => (
          <BkRadioGroup v-model={formData.health_switch} class='radio-groups'>
            <BkRadio label={1}>
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
            <BkRadio label={2}>
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
        property: 'check_type',
        required: true,
        content: () => (
          <BkRadioGroup v-model={formData.check_type}>
            <BkRadio label='TCP'>TCP</BkRadio>
            <BkRadio label='HTTP'>HTTP</BkRadio>
            <BkRadio label='CUSTOM'>自定义</BkRadio>
          </BkRadioGroup>
        ),
      },
      {
        label: '检查端口',
        property: 'check_port',
        required: true,
        content: () => <Input v-model={formData.check_port} />,
      },
      [
        {
          label: '响应超时',
          property: 'time_out',
          required: true,
          span: 12,
          content: () => <Input v-model={formData.time_out} placeholder='0' type='number' suffix='秒' />,
        },
        {
          label: '检查间隔',
          property: 'interval_time',
          required: true,
          span: 12,
          content: () => <Input v-model={formData.interval_time} placeholder='0' type='number' suffix='秒' />,
        },
      ],
      [
        {
          label: '不健康阈值',
          property: 'un_health_num',
          required: true,
          span: 12,
          content: () => <Input v-model={formData.un_health_num} placeholder='0' type='number' suffix='秒' />,
        },
        {
          label: '健康阈值',
          property: 'health_num',
          required: true,
          span: 12,
          content: () => <Input v-model={formData.health_num} placeholder='0' type='number' suffix='秒' />,
        },
      ],
    ]);

    watch(
      () => props.detail,
      (detail) => {
        isOpen.value = !!detail.health_check?.health_switch;
      },
      {
        immediate: true,
        deep: true,
      },
    );

    const handleSubmit = async () => {
      isSubmitLoading.value = true;
      try {
        await businessStore.updateHealthCheck({
          id: loadbalancerStore.targetGroupId,
          health_check: {
            ...formData,
            check_port: +formData.check_port,
            time_out: +formData.time_out,
            interval_time: +formData.interval_time,
            un_health_num: +formData.un_health_num,
            health_num: +formData.health_num,
          },
        });
        isHealthCheckupConfigShow.value = false;
        resetFormData();
        props.getTargetGroupDetail?.(loadbalancerStore.targetGroupId);
      } finally {
        isSubmitLoading.value = false;
      }
    };

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
          onHandleSubmit={handleSubmit}
          isSubmitLoading={isSubmitLoading.value}
          width='640'>
          <Form formType='vertical'>
            <Container margin={0}>
              {formItemOptions.value.map((item) => (
                <Row>
                  {Array.isArray(item) ? (
                    item.map(({ label, property, required, span, content }) => (
                      <Col span={span}>
                        <FormItem label={label} property={property} required={required} labelPosition='top'>
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
