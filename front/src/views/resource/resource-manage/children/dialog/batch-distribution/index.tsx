import { Button, Dialog, Message } from 'bkui-vue';
import { PropType, defineComponent, ref } from 'vue';
import BusinessSelector from '@/components/business-selector/index.vue';
import './index.scss';
import { Senarios, useWhereAmI } from '@/hooks/useWhereAmI';
import { useResourceStore } from '@/store';

export enum DResourceType {
  cvms = 'cvms',
  disks = 'disks',
  eips = 'eips',
  firewall = 'vendors/gcp/firewalls/rules',
  network_interfaces = 'network_interfaces',
  route_tables = 'route_table_ids',
  security_groups = 'security_groups',
  subnets = 'subnets',
  vpcs = 'vpcs',
  templates = 'argument_templates',
  load_balancers = 'load_balancers',
  certs = 'certs',
}

export const DResourceTypeMap = {
  [DResourceType.cvms]: {
    key: 'cvm_ids',
    name: '主机',
  },
  [DResourceType.disks]: {
    key: 'disk_ids',
    name: '云硬盘',
  },
  [DResourceType.eips]: {
    key: 'eip_ids',
    name: '弹性IP',
  },
  [DResourceType.firewall]: {
    key: 'firewall_rule_ids',
    name: '防火墙',
  },
  [DResourceType.network_interfaces]: {
    key: 'network_interface_ids	',
    name: '网络接口',
  },
  [DResourceType.route_tables]: {
    key: 'route_table_ids',
    name: '路由表',
  },
  [DResourceType.security_groups]: {
    key: 'security_group_ids',
    name: '安全组',
  },
  [DResourceType.subnets]: {
    key: 'subnet_ids',
    name: '子网',
  },
  [DResourceType.vpcs]: {
    key: 'vpc_ids',
    name: 'VPC',
  },
  [DResourceType.templates]: {
    key: 'template_ids',
    name: '参数模板',
  },
  [DResourceType.load_balancers]: {
    key: 'lb_ids',
    name: '负载均衡',
  },
  [DResourceType.certs]: {
    key: 'cert_ids',
    name: '证书',
  },
};

export const BatchDistribution = defineComponent({
  props: {
    selections: {
      type: Array as PropType<Array<any>>,
      required: true,
    },
    type: {
      type: String as PropType<DResourceType>,
      required: true,
    },
    getData: {
      type: Function as PropType<() => void>,
      required: true,
    },
  },
  setup(props) {
    const { whereAmI } = useWhereAmI();
    const selectedBizId = ref('');
    const isShow = ref(false);
    const isLoading = ref(false);
    const resourceStore = useResourceStore();
    const handleConfirm = async () => {
      isLoading.value = true;
      try {
        await resourceStore.assignBusiness(props.type, {
          [DResourceTypeMap[props.type].key]: props.selections?.map((v) => v.id) || [],
          bk_biz_id: selectedBizId.value,
        });
        Message({
          theme: 'success',
          message: '批量分配成功！',
        });
        props.getData?.();
      } catch (error) {
        Message({
          theme: 'error',
          message: '批量分配失败！',
        });
      } finally {
        isLoading.value = false;
        isShow.value = false;
      }
    };
    return () => (
      <>
        {whereAmI.value === Senarios.resource ? (
          <Button
            class={'mw88 ml8 mr8'}
            onClick={() => {
              isShow.value = true;
            }}
            disabled={!props.selections.length}>
            批量分配
          </Button>
        ) : null}
        <Dialog
          class={'batch-dialog'}
          isShow={isShow.value}
          title={`批量分配/${DResourceTypeMap[props.type].name}分配`}
          theme={'primary'}
          quickClose
          onClosed={() => (isShow.value = false)}
          onConfirm={handleConfirm}
          isLoading={isLoading.value}>
          <p class='selected-host-count-tip'>
            已选择
            <span class='selected-host-count'>{props.selections.length}</span>个{DResourceTypeMap[props.type].name}
            ，可选择所需分配的目标业务
          </p>
          <p class='mb6'>目标业务</p>
          <BusinessSelector
            v-model={selectedBizId.value}
            authed={true}
            class='mb32'
            auto-select={true}></BusinessSelector>
        </Dialog>
      </>
    );
  },
});
