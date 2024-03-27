import { defineComponent, ref, watch } from 'vue';
import './index.scss';
import DetailInfo from '@/views/resource/resource-manage/common/info/detail-info';
import { Switcher, Tag } from 'bkui-vue';
import StatusNormal from '@/assets/image/Status-normal.png';
import StatusUnknown from '@/assets/image/Status-unknown.png';
import { useLoadBalancerStore } from '@/store/loadbalancer';
import { useBusinessStore } from '@/store';

export default defineComponent({
  setup() {
    const detail: { [key: string]: any } = ref({});
    const loadBalancerStore = useLoadBalancerStore();
    const businessStore = useBusinessStore();
    const resourceFields = [
      {
        name: '名称',
        prop: 'name',
        edit: true,
      },
      {
        name: '所属网络',
        prop: 'vpc_id',
      },
      {
        name: 'ID',
        prop: 'cloud_id',
      },
      {
        name: '删除保护',
        render: () => (
          <div>
            <Switcher class={'mr5'} />
            <Tag>未开启</Tag>
          </div>
        ),
      },
      {
        name: '状态',
        render() {
          return (
            <div class={'status-wrapper'}>
              <img src={!detail.value.status ? StatusUnknown : StatusNormal} class={'mr6'} width={14} height={14}></img>
              <span>{!detail.value.status ? '创建中' : '正常运行'}</span>
            </div>
          );
        },
      },
      {
        name: 'IP版本',
        prop: 'ip_type',
      },
      {
        name: '网络类型',
        prop: 'ip_version',
      },
      {
        name: '创建时间',
        prop: 'created_at',
      },
      {
        name: '地域',
        prop: 'region',
      },
      {
        name: '可用区域',
        prop: 'zones',
      },
    ];

    const configFields = [
      {
        name: '负载均衡域名',
        prop: 'domain',
      },
      {
        name: '实例计费模式',
        prop: '',
      },
      {
        name: '负载均衡VIP',
        render: () => {
          return detail.value?.extension?.vip_isp;
        },
      },
      {
        name: '带宽计费模式',
        render: () => {
          return detail.value?.extension?.internet_charge_type;
        },
      },
      {
        name: '规格类型',
        render: () => {
          return detail.value?.extension?.sla_type;
        },
      },
      {
        name: '带宽上限',
        render: () => {
          return detail.value?.extension?.internet_max_bandwidth_out;
        },
      },
      {
        name: '运营商',
        render: () => {
          return detail.value?.extension?.vip_isp;
        },
      },
    ];

    const getDetails = async (id: string) => {
      const res = await businessStore.getLbDetail(id);
      detail.value = res.data;
    };

    watch(
      () => loadBalancerStore.currentSelectedTreeNode.id,
      (id) => {
        if (id) getDetails(id);
      },
      {
        immediate: true,
      },
    );

    return () => (
      <div class={'clb-detail-continer'}>
        <div>
          <p class={'clb-detail-info-title'}>资源信息</p>
          <DetailInfo fields={resourceFields} detail={detail.value} class={'ml60'} />
        </div>
        <div>
          <p class={'clb-detail-info-title'}>配置信息</p>
          <DetailInfo fields={configFields} detail={detail.value} class={'ml60'} />
        </div>
      </div>
    );
  },
});
