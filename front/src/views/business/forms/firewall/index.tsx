import { VendorEnum } from '@/common/constant';
import { defineComponent, reactive, ref, watch } from 'vue';
import FormSelect from '@/views/business/components/form-select.vue';
import VpcSelector from '@/components/vpc-selector/index.vue';
import './index.scss';
import { useWhereAmI } from '@/hooks/useWhereAmI';
import http from '@/http';
import { useAccountStore } from '@/store';

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

export default defineComponent({
  setup() {
    const formModel = reactive({
      account_id: 0, // 云账号
      vendor: VendorEnum.GCP, // 云厂商
      name: '', // 名称
      cloud_vpc_id: 0, // 所属的VPC
      type: DirectionType.out, // 流量方向
      priority: 0, // 优先级
      source_tags: [], // 来源标记
      target_tags: [], // 目标标记
      source_ranges: [], // 来源
      destination_ranges: [], // 目标
      allowed: [], // 允许的协议和端口
      denied: [], // 拒绝的协议和端口
      disabled: true, // 创建是否不要立即应用到目标
      memo: '', // 备注
    });

    const ip_type = ref(IpType.ipv4);
    const is_source_marked = ref(false);
    const is_destination_marked = ref(false);
    const is_rule_allowed = ref(false);
    const protocolAndPorts = reactive({
      protocol: '',
      ports: '',
    });
    const isPortsDisabled = ref(false);

    const { isResourcePage } = useWhereAmI();
    const accountStore = useAccountStore();

    const { BK_HCM_AJAX_URL_PREFIX } = window.PROJECT_CONFIG;

    const handleSubmit = async () => {
      console.log(formModel);
      // @ts-ignore
      if (formModel.allowed.ports) formModel.allowed.ports = [formModel.allowed.ports];
      // @ts-ignore
      if (formModel.denied.ports) formModel.allowed.ports = [formModel.allowed.ports];
      await http.post(
        isResourcePage
          ? `${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/vendors/gcp/firewalls/rules/create`
          : `${BK_HCM_AJAX_URL_PREFIX}/api/v1/cloud/bizs/${accountStore.bizs}/vendors/gcp/firewalls/rules/create`,
        formModel,
      );
    };

    watch(
      () => is_rule_allowed.value,
      (isAllowed) => {
        if (isAllowed) {
          formModel.allowed.push(protocolAndPorts);
          formModel.denied = [];
        } else {
          formModel.denied.push(protocolAndPorts);
          formModel.allowed = [];
        }
      },
      {
        immediate: true,
      },
    );

    watch(
      () => protocolAndPorts.protocol,
      (val) => {
        protocolAndPorts.ports = '';
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
      () => ip_type.value,
      () => {
        protocolAndPorts.ports = '';
        protocolAndPorts.protocol = '';
      },
      {
        immediate: true,
      },
    );

    return () => (
      <div class={'firewall-form-container'}>
        <FormSelect
          hidden={['region']}
          onChange={(val: any) => {
            console.log(val.account_id, val.vendor);
            formModel.account_id = val.account_id;
            formModel.vendor = val.vendor;
          }}></FormSelect>
        <bk-form class={'pr20'}>
          <bk-form-item label={'名称'} property={'name'}>
            <bk-input v-model={formModel.name}></bk-input>
          </bk-form-item>
          <bk-form-item label={'所属的vpc'} property={'name'}>
            <VpcSelector
              vendor={formModel.vendor}
              v-model={formModel.cloud_vpc_id}
            />
          </bk-form-item>
          <bk-form-item label={'流量方向'} property={'type'}>
            <bk-radio-group v-model={formModel.type}>
              <bk-radio label={DirectionType.in}>入站流量</bk-radio>
              <bk-radio label={DirectionType.out}>出站流量</bk-radio>
            </bk-radio-group>
          </bk-form-item>
          <bk-form-item
            label={'优先级'}
            property={'priority'}
            description={'优先级范围从 0 到 65535'}>
            <bk-input
              vendor={formModel.priority}
              v-model={formModel.priority}
              type='number'
            />
          </bk-form-item>
          <bk-form-item label={'IP类型'}>
            <bk-select v-model={ip_type.value} type='number' clearable={false}>
              {[IpType.ipv4, IpType.ipv6].map(v => (
                <bk-option key={v} value={v} label={v}></bk-option>
              ))}
            </bk-select>
          </bk-form-item>
          <bk-form-item label={'来源'} property={'source_ranges'}>
            <bk-tag-input
              v-model={formModel.source_ranges}
              allowCreate
              hasDeleteIcon
            />
          </bk-form-item>
          <bk-form-item label={'目标'} property={'destination_ranges'}>
            <bk-tag-input
              v-model={formModel.destination_ranges}
              allowCreate
              hasDeleteIcon
            />
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
            <bk-form-item property={'source_tags'}>
              <bk-tag-input
                v-model={formModel.source_tags}
                allowCreate
                hasDeleteIcon
                placeholder='输入来源标记'
              />
            </bk-form-item>
          ) : null}
          <bk-form-item label={'目标标记'}>
            <bk-radio-group v-model={is_destination_marked.value}>
              <bk-radio label={true}>启用</bk-radio>
              <bk-radio label={false}>禁用</bk-radio>
            </bk-radio-group>
          </bk-form-item>
          {is_destination_marked.value ? (
            <bk-form-item property={'target_tags'}>
              <bk-tag-input
                v-model={formModel.target_tags}
                allowCreate
                hasDeleteIcon
                placeholder='输入目标标记'
              />
            </bk-form-item>
          ) : null}
          <bk-form-item label={'协议端口'}>
            <bk-input
              class={'firewall-input-select-warp'}
              disabled={isPortsDisabled.value}
              v-model={protocolAndPorts.ports}>
              {{
                prefix: () => (
                  <bk-select v-model={protocolAndPorts.protocol}>
                    {ip_type.value === IpType.ipv6
                      ? Object.entries({
                        ...Protocols,
                        ...IPV6_Special_Protocols,
                      }).map(([key, val]) => (
                          <bk-option
                            label={key}
                            value={val}
                            key={key}></bk-option>
                      ))
                      : Object.entries({
                        ...Protocols,
                        ...IPV4_Special_Protocols,
                      }).map(([key, val]) => (
                          <bk-option
                            label={key}
                            value={val}
                            key={key}></bk-option>
                      ))}
                  </bk-select>
                ),
              }}
            </bk-input>
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
              提交创建
            </bk-button>
            <bk-button class='ml10'>取消</bk-button>
          </bk-form-item>
        </bk-form>
      </div>
    );
  },
});
