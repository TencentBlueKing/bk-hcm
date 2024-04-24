import { computed, defineComponent, reactive, ref, watch } from 'vue';
import { Container, Button, Switcher, Form, Tag, Input, Select } from 'bkui-vue';
import { BkRadio, BkRadioGroup } from 'bkui-vue/lib/radio';
import './index.scss';
import CommonSideslider from '@/components/common-sideslider';
import { useBusinessStore, useLoadBalancerStore } from '@/store';

const { Row, Col } = Container;
const { FormItem } = Form;
const { Option } = Select;

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
      http_check_domain: '', // 域名
      http_check_path: '', // 路径
      http_check_method: 'HEAD', // 请求方式
      http_code: [], // 状态码检测
      http_version: 'HTTP/1.0', // 版本
      context_type: 'TEXT', // 输入格式
      send_context: '', // 检查请求
      recv_context: '', // 检查返回结果
      source_ip_type: 1, // 0（使用LB的VIP作为源IP），1（使用100.64网段IP作为源IP）
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
            case 'http_check_method':
              formData[key] = 'HEAD';
              break;
            case 'http_version':
              formData[key] = 'HTTP/1.0';
            case 'context_type':
              formData[key] = 'TEXT';
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
        isRender: true,
      },
      {
        label: '探测来源IP',
        property: 'source_ip_type',
        required: true,
        content: () => (
          <BkRadioGroup v-model={formData.source_ip_type} class='radio-groups'>
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
            <BkRadio label={0}>
              <div class='radio-item-wrap'>
                <div class='item-label'>负载均衡 VIP</div>
                <div class='item-desc'>需同时在后端服务器安全组和iptables放通VIP地址</div>
              </div>
            </BkRadio>
          </BkRadioGroup>
        ),
        isRender: true,
      },
      {
        label: '检查方式',
        property: 'check_type',
        required: true,
        content: () => (
          <BkRadioGroup v-model={formData.check_type}>
            {!['UDP'].includes(props.detail.protocol) && (
              <>
                <BkRadio label='TCP'>TCP</BkRadio>
                <BkRadio label='HTTP'>HTTP</BkRadio>
              </>
            )}
            {['TCP', 'UDP'].includes(props.detail.protocol) && <BkRadio label='CUSTOM'>自定义</BkRadio>}
            {['UDP'].includes(props.detail.protocol) && <BkRadio label='PING'>PING</BkRadio>}
          </BkRadioGroup>
        ),
        isRender: true,
      },
      {
        label: '检查端口',
        property: 'check_port',
        required: true,
        content: () => <Input v-model={formData.check_port} />,
        isRender: ['TCP', 'UDP'].includes(props.detail.protocol) && !['PING'].includes(formData.check_type),
      },
      [
        {
          label: '检查域名',
          property: 'http_check_domain',
          required: true,
          span: 12,
          content: () => <Input v-model={formData.http_check_domain} />,
          isRender: ['HTTP'].includes(formData.check_type),
        },
        {
          label: '检查路径',
          property: 'http_check_path',
          required: true,
          span: 12,
          content: () => <Input v-model={formData.http_check_path} />,
          isRender: ['HTTP'].includes(formData.check_type),
        },
      ],
      [
        {
          label: 'HTTP请求方式',
          property: 'http_check_domain',
          required: true,
          span: 12,
          content: () => (
            <Select v-model={formData.http_check_method} clearable={false}>
              {['HEAD', 'GET'].map((v) => (
                <Option name={v} id={v} key={v} />
              ))}
            </Select>
          ),
          isRender: ['HTTP'].includes(formData.check_type),
        },
        {
          label: 'HTTP状态码检测',
          property: 'http_code',
          required: true,
          span: 12,
          content: () => (
            <Select v-model={formData.http_code} clearable={false} multiple multiple-mode='tag'>
              {[
                {
                  name: '1xx',
                  id: 1,
                },
                {
                  name: '2xx',
                  id: 2,
                },
                {
                  name: '3xx',
                  id: 4,
                },
                {
                  name: '4xx',
                  id: 8,
                },
                {
                  name: '5xx',
                  id: 16,
                },
              ].map(({ name, id }) => (
                <Option name={name} id={id} key={id} />
              ))}
            </Select>
          ),
          isRender: ['HTTP'].includes(formData.check_type),
        },
      ],
      {
        label: 'HTTP版本',
        property: 'http_version',
        required: true,
        span: 12,
        content: () => (
          <Select v-model={formData.http_version} clearable={false}>
            {[
              {
                name: 'HTTP/1.0',
                id: 'HTTP/1.0',
              },
              {
                name: 'HTTP/1.1',
                id: 'HTTP/1.1',
              },
            ].map(({ name, id }) => (
              <Option name={name} id={id} key={id} />
            ))}
          </Select>
        ),
        isRender: props.detail.protocol === 'TCP' && formData.check_type === 'HTTP',
      },
      {
        label: '输入格式',
        property: 'time_out',
        required: true,
        span: 12,
        content: () => (
          <Select v-model={formData.context_type} clearable={false}>
            {[
              {
                name: '十六进制',
                id: 'HEX',
              },
              {
                name: '文本',
                id: 'TEXT',
              },
            ].map(({ name, id }) => (
              <Option name={name} id={id} key={id} />
            ))}
          </Select>
        ),
        isRender: ['CUSTOM'].includes(formData.check_type),
      },
      {
        label: '检查请求',
        property: 'send_context',
        required: true,
        span: 12,
        content: () => <Input v-model={formData.send_context} type={'textarea'} rows={3} />,
        isRender: ['CUSTOM'].includes(formData.check_type),
      },
      {
        label: '检查返回结果',
        property: 'recv_context',
        required: true,
        span: 12,
        content: () => <Input v-model={formData.recv_context} type={'textarea'} rows={3} />,
        isRender: ['CUSTOM'].includes(formData.check_type),
      },
      [
        {
          label: '响应超时',
          property: 'time_out',
          required: true,
          span: 12,
          content: () => <Input v-model={formData.time_out} placeholder='0' type='number' suffix='秒' />,
          isRender: true,
        },
        {
          label: '检查间隔',
          property: 'interval_time',
          required: true,
          span: 12,
          content: () => <Input v-model={formData.interval_time} placeholder='0' type='number' suffix='秒' />,
          isRender: true,
        },
      ],
      [
        {
          label: '不健康阈值',
          property: 'un_health_num',
          required: true,
          span: 12,
          content: () => <Input v-model={formData.un_health_num} placeholder='0' type='number' suffix='秒' />,
          isRender: true,
        },
        {
          label: '健康阈值',
          property: 'health_num',
          required: true,
          span: 12,
          content: () => <Input v-model={formData.health_num} placeholder='0' type='number' suffix='秒' />,
          isRender: true,
        },
      ],
    ]);

    watch(
      () => props.detail,
      (detail) => {
        isOpen.value = !!detail.health_check?.health_switch;
        const keys = [
          'health_switch',
          'check_type',
          'check_port',
          'http_check_domain',
          'http_check_path',
          'http_check_method',
          'http_code',
          'http_version',
          'context_type',
          'send_context',
          'recv_context',
          'time_out',
          'interval_time',
          'un_health_num',
          'health_num',
        ];

        for (const key of keys) {
          formData[key] = detail.health_check?.[key];
        }
        isSubmitLoading.value = false;
      },
    );

    watch(
      () => formData,
      (formData) => {
        if (formData.health_switch) {
          isOpen.value = true;
        } else {
          isOpen.value = false;
        }
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
            health_switch: formData.health_switch ? 1 : 0,
            check_port: +formData.check_port,
            time_out: +formData.time_out,
            interval_time: +formData.interval_time,
            un_health_num: +formData.un_health_num,
            health_num: +formData.health_num,
            http_code: formData.http_code?.length ? formData.http_code.reduce((acc, cur) => acc + cur, 0) : undefined,
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
              {formData.health_switch ? (
                formItemOptions.value.map((item) => (
                  <Row>
                    {Array.isArray(item)
                      ? item
                          .filter(({ isRender }) => !!isRender)
                          .map(({ label, property, required, span, content }) => (
                            <Col span={span}>
                              <FormItem label={label} property={property} required={required} labelPosition='top'>
                                {content()}
                              </FormItem>
                            </Col>
                          ))
                      : item.isRender && (
                          <Col span={24}>
                            <FormItem label={item.label} property={item.property} required={item.required}>
                              {item.content()}
                            </FormItem>
                          </Col>
                        )}
                  </Row>
                ))
              ) : (
                <FormItem label='是否启用'>{formItemOptions.value[0].content()}</FormItem>
              )}
            </Container>
          </Form>
        </CommonSideslider>
      </div>
    );
  },
});
