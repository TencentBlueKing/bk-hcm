import { VendorEnum } from '@/common/constant';
import { PropType, defineComponent, reactive, ref, watch } from 'vue';
import FormSelect from '@/views/business/components/form-select.vue';
import VpcSelector from '@/components/vpc-selector/index.vue';
import './index.scss';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import http from '@/http';
import { useAccountStore } from '@/store';
import { validateIpCidr } from '@/views/resource/resource-manage/children/dialog/security-rule/security-rule-validators';
import { cloneDeep } from 'lodash-es';

enum IpType {
  ipv4 = 'IPv4',
  ipv6 = 'IPv6',
}

enum DirectionType {
  out = 'EGRESS',
  in = 'INGRESS',
}

const Protocols = {
  ALL: 'all',
  TCP: 'tcp',
  UDP: 'udp',
};

const IPV4_Special_Protocols = {
  ICMP: 'icmp',
};

const IPV6_Special_Protocols = {
  ICMPV6: '58',
};

type ProtocolAndPorts = {
  protocol: string;
  port: Array<string>;
};

const _formModel = {
  account_id: 0, // 云账号
  vendor: VendorEnum.GCP, // 云厂商
  name: '', // 名称
  cloud_vpc_id: '', // 所属的VPC
  type: DirectionType.out, // 流量方向
  priority: 0, // 优先级
  source_tags: [] as Array<string>, // 来源标记
  target_tags: [] as Array<string>, // 目标标记
  source_ranges: [] as Array<string>, // 来源
  destination_ranges: [] as Array<string>, // 目标
  allowed: [] as Array<ProtocolAndPorts>, // 允许的协议和端口
  denied: [] as Array<ProtocolAndPorts>, // 拒绝的协议和端口
  disabled: true, // 创建是否不要立即应用到目标
  memo: '', // 备注
  id: '', // 编辑时用的唯一标志
};

export default defineComponent({
  props: {
    isEdit: {
      default: false,
      type: Boolean,
    },
    detail: {
      default: {},
      type: Object as PropType<typeof _formModel>,
    },
    isFormDataChanged: Boolean,
    show: Boolean,
  },
  emits: ['cancel', 'success', 'update:isFormDataChanged'],
  setup(props, { emit }) {
    let formModel = reactive({
      ...cloneDeep(_formModel),
      ...(props.isEdit ? props.detail : {}),
    });
    const ip_type = ref(validateIpCidr(formModel?.destination_ranges?.[0]) === 'ipv6' ? IpType.ipv6 : IpType.ipv4);
    const is_source_marked = ref(!!formModel.source_tags?.length);
    const is_destination_marked = ref(!!formModel.destination_ranges?.length);
    const is_rule_allowed = ref(!!formModel.allowed?.length);
    const protocolAndPorts = reactive({
      protocol: props.detail.allowed?.[0]?.protocol || props.detail.denied?.[0]?.protocol || '',
      port: props.detail.allowed?.[0]?.port || [],
    });
    const isPortsDisabled = ref(false);
    const formInstance = ref({});

    const { isResourcePage } = useWhereAmI();
    const accountStore = useAccountStore();

    const { BK_HCM_AJAX_URL_PREFIX } = window.PROJECT_CONFIG;

    const handleSubmit = async () => {
      // @ts-ignore
      await formInstance.value.validate();
      if (!formModel.allowed?.length) delete formModel.allowed;
      if (!formModel.denied?.length) delete formModel.denied;
      if (props.isEdit) {
        await http.put(
          isResourcePage
            ? `${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/vendors/gcp/firewalls/rules/${props.detail.id}`
            : `${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/bizs/${accountStore.bizs}/vendors/gcp/firewalls/rules/${props.detail.id}`,
          formModel,
        );
      } else {
        await http.post(
          isResourcePage
            ? `${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/vendors/gcp/firewalls/rules/create`
            : `${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/bizs/${accountStore.bizs}/vendors/gcp/firewalls/rules/create`,
          formModel,
        );
      }
      emit('success');
      handleCancel();
    };

    const handleCancel = () => {
      emit('cancel');
      formModel = cloneDeep(_formModel);
    };

    watch(
      () => is_rule_allowed.value,
      (isAllowed) => {
        formModel.allowed = [];
        formModel.denied = [];
        if (isAllowed) {
          formModel.allowed?.push(protocolAndPorts);
        } else {
          formModel.denied?.push(protocolAndPorts);
        }
      },
      {
        immediate: true,
      },
    );

    watch(
      () => protocolAndPorts.protocol,
      (val) => {
        switch (val) {
          case Protocols.ALL:
          case IPV4_Special_Protocols.ICMP:
          case IPV6_Special_Protocols.ICMPV6:
            isPortsDisabled.value = true;
            break;
          default:
            isPortsDisabled.value = false;
        }
      },
      {
        immediate: true,
      },
    );

    watch(
      () => protocolAndPorts.protocol,
      () => {
        protocolAndPorts.port = [];
      },
    );

    watch(
      () => ip_type.value,
      () => {
        protocolAndPorts.port = [];
        protocolAndPorts.protocol = '';
      },
    );

    watch(formModel, () => {
      !props.isFormDataChanged && emit('update:isFormDataChanged', true);
    });

    return () => (
      <div class={'firewall-form-container'}>
        {!props.isEdit ? (
          <FormSelect
            type={'security'}
            hidden={['region']}
            show={props.show}
            onChange={(val: any) => {
              formModel.account_id = val.account_id;
              formModel.vendor = val.vendor;
            }}></FormSelect>
        ) : null}
        <bk-form
          class={'pr20'}
          ref={formInstance}
          model={formModel}
          rules={{
            protocol: [
              {
                trigger: 'change',
                message: '协议不能为空',
                validator: () => {
                  return !!protocolAndPorts.protocol;
                },
              },
            ],
          }}>
          <bk-form-item label={'名称'} property={'name'} required>
            <bk-input v-model={formModel.name}></bk-input>
          </bk-form-item>
          <bk-form-item label={'所属的vpc'} property={'cloud_vpc_id'} required>
            <VpcSelector vendor={formModel.vendor} v-model={formModel.cloud_vpc_id} isDisabled={props.isEdit} />
          </bk-form-item>
          <bk-form-item label={'流量方向'} property={'type'} required>
            <bk-radio-group v-model={formModel.type}>
              <bk-radio label={DirectionType.in}>入站流量</bk-radio>
              <bk-radio label={DirectionType.out}>出站流量</bk-radio>
            </bk-radio-group>
          </bk-form-item>
          <bk-form-item label={'优先级'} property={'priority'} required description={'优先级范围从 0 到 65535'}>
            <bk-input
              vendor={formModel.priority}
              v-model_number={formModel.priority}
              min={0}
              max={65535}
              type='number'
            />
          </bk-form-item>
          <bk-form-item label={'IP类型'}>
            <bk-select v-model={ip_type.value} clearable={false}>
              {[IpType.ipv4, IpType.ipv6].map((v) => (
                <bk-option key={v} value={v} label={v}></bk-option>
              ))}
            </bk-select>
          </bk-form-item>
          <bk-form-item label={'来源'} property={'source_ranges'} required>
            <bk-tag-input v-model={formModel.source_ranges} allowCreate allowAutoMatch hasDeleteIcon />
          </bk-form-item>
          <bk-form-item label={'目标'} property={'destination_ranges'} required>
            <bk-tag-input v-model={formModel.destination_ranges} allowCreate allowAutoMatch hasDeleteIcon />
          </bk-form-item>
          {formModel.type === DirectionType.in ? (
            <bk-form-item label={'来源标记'}>
              <bk-radio-group v-model={is_source_marked.value}>
                <bk-radio label={true}>启用</bk-radio>
                <bk-radio label={false}>禁用</bk-radio>
              </bk-radio-group>
            </bk-form-item>
          ) : null}
          {is_source_marked.value ? (
            <bk-form-item property={'source_tags'} required>
              <bk-tag-input
                v-model={formModel.source_tags}
                allowCreate
                allowAutoMatch
                hasDeleteIcon
                placeholder='输入来源标记'
              />
            </bk-form-item>
          ) : null}
          <bk-form-item label={'目标标记'} required>
            <bk-radio-group v-model={is_destination_marked.value}>
              <bk-radio label={true}>启用</bk-radio>
              <bk-radio label={false}>禁用</bk-radio>
            </bk-radio-group>
          </bk-form-item>
          {is_destination_marked.value ? (
            <bk-form-item property={'target_tags'} required>
              <bk-tag-input
                v-model={formModel.target_tags}
                allowCreate
                allowAutoMatch
                hasDeleteIcon
                placeholder='输入目标标记'
              />
            </bk-form-item>
          ) : null}
          <bk-form-item label={'协议'} property={'protocol'}>
            <bk-select v-model={protocolAndPorts.protocol}>
              {ip_type.value === IpType.ipv6
                ? Object.entries({
                    ...Protocols,
                    ...IPV6_Special_Protocols,
                  }).map(([key, val]) => <bk-option label={key} value={val} key={key}></bk-option>)
                : Object.entries({
                    ...Protocols,
                    ...IPV4_Special_Protocols,
                  }).map(([key, val]) => <bk-option label={key} value={val} key={key}></bk-option>)}
            </bk-select>
          </bk-form-item>
          <bk-form-item label={'端口'} property={'port'}>
            <bk-tag-input
              disabled={isPortsDisabled.value}
              v-model={protocolAndPorts.port}
              allowCreate
              allowAutoMatch
              hasDeleteIcon
              placeholder='输入端口'
            />
          </bk-form-item>
          <bk-form-item label={'策略'}>
            <bk-radio-group v-model={is_rule_allowed.value}>
              <bk-radio label={true}>允许</bk-radio>
              <bk-radio label={false}>拒绝</bk-radio>
            </bk-radio-group>
          </bk-form-item>
          <bk-form-item label={'创建后立即应用'}>
            <bk-radio-group v-model={formModel.disabled}>
              <bk-radio label={false}>启用</bk-radio>
              <bk-radio label={true}>禁用</bk-radio>
            </bk-radio-group>
          </bk-form-item>
          <bk-form-item label={'备注'}>
            <bk-input v-model={formModel.memo} type='textarea'></bk-input>
          </bk-form-item>
          <bk-form-item>
            <bk-button theme='primary' class='ml10' onClick={handleSubmit}>
              {props.isEdit ? '确定' : '提交创建'}
            </bk-button>
            <bk-button class='ml10' onClick={handleCancel}>
              取消
            </bk-button>
          </bk-form-item>
        </bk-form>
      </div>
    );
  },
});
