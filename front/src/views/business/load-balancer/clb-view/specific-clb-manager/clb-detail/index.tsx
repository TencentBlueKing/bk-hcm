import { PropType, defineComponent, ref, watch } from 'vue';
import './index.scss';
import DetailInfo from '@/views/resource/resource-manage/common/info/detail-info';
import { Switcher, Tag } from 'bkui-vue';
import StatusNormal from '@/assets/image/Status-normal.png';
import StatusUnknown from '@/assets/image/Status-unknown.png';
import { timeFormatter } from '@/common/util';
import { useRouteLinkBtn, TypeEnum, IDetail } from '@/hooks/useRouteLinkBtn';

import { CHARGE_TYPE, CLB_SPECS, LB_ISP, LB_TYPE_MAP } from '@/common/constant';
import { useRegionsStore } from '@/store/useRegionsStore';
import { IP_VERSION_MAP } from '@/constants';

export default defineComponent({
  props: {
    detail: Object as PropType<IDetail>,
    getDetails: Function,
    updateLb: Function,
    id: String,
  },
  setup(props) {
    const regionStore = useRegionsStore();
    const isProtected = ref(false);
    const isLoading = ref(false);
    const resourceFields = [
      {
        name: '名称',
        prop: 'name',
        edit: true,
      },
      {
        name: '所属网络',
        prop: 'cloud_vpc_id',
        render() {
          return useRouteLinkBtn(props.detail, {
            id: 'vpc_id',
            name: 'cloud_vpc_id',
            type: TypeEnum.VPC,
          });
        },
      },
      {
        name: 'ID',
        prop: 'cloud_id',
      },
      {
        name: '删除保护',
        render: () => (
          <div>
            <Switcher
              theme='primary'
              class={'mr5'}
              modelValue={isProtected.value}
              disabled={isLoading.value}
              onChange={async (val) => {
                isLoading.value = true;
                isProtected.value = val;
                try {
                  await props.updateLb({
                    delete_protect: val,
                  });
                } catch (_e) {
                  isProtected.value = !val;
                } finally {
                  isLoading.value = false;
                }
              }}
            />
            <Tag theme={isProtected.value ? 'success' : ''}> {isProtected.value ? '已开启' : '未开启'} </Tag>
            <i
              class='hcm-icon bkhcm-icon-info-line ml10'
              v-bk-tooltips={{
                content: '开启删除保护后，在云控制台或调用 API 均无法删除该实例',
                placement: 'top-end',
              }}></i>
          </div>
        ),
      },
      {
        name: '状态',
        render() {
          return (
            <div class={'status-wrapper'}>
              <img src={!props.detail.status ? StatusUnknown : StatusNormal} class={'mr6'} width={14} height={14}></img>
              <span>{!props.detail.status ? '创建中' : '正常运行'}</span>
            </div>
          );
        },
      },
      {
        name: 'IP版本',
        prop: 'ip_version',
        render() {
          return IP_VERSION_MAP[props.detail.ip_version];
        },
      },
      {
        name: '网络类型',
        prop: 'lb_type',
        render() {
          return LB_TYPE_MAP[props.detail.lb_type];
        },
      },
      {
        name: '创建时间',
        prop: 'created_at',
        render() {
          return timeFormatter(props.detail.created_at);
        },
      },
      {
        name: '地域',
        prop: 'region',
        render() {
          return regionStore.getRegionName(props.detail.vendor, props.detail.region);
        },
      },
      {
        name: '可用区域',
        prop: 'zones',
        render() {
          const mains = props.detail.zones;
          const backups = props.detail.backup_zones;
          const mainsStr = mains
            ?.map((zone: string) => `${regionStore.getRegionName(props.detail.vendor, zone)}(主)`)
            .join(',');
          const backupsStr = backups
            ?.map((zone: string) => `${regionStore.getRegionName(props.detail.vendor, zone)}(备)`)
            .join(',');
          return `${mainsStr}${backupsStr?.length ? `,${backupsStr}` : ''}`;
        },
      },
    ];

    const configFields = [
      {
        name: '负载均衡域名',
        prop: 'domain',
      },
      {
        name: '实例计费模式',
        render() {
          return CHARGE_TYPE[props.detail?.extension?.charge_type] || '--';
        },
      },
      {
        name: '负载均衡VIP',
        render: () => {
          return props.detail?.public_ipv4_addresses?.concat(props.detail?.public_ipv6_addresses) || '--';
        },
      },
      {
        name: '带宽计费模式',
        render: () => {
          return props.detail?.extension?.internet_charge_type || '--';
        },
      },
      {
        name: '规格类型',
        render: () => {
          return CLB_SPECS[props.detail?.extension?.sla_type] || '--';
        },
      },
      {
        name: '带宽上限',
        render: () => {
          return props.detail?.extension?.internet_max_bandwidth_out || '--';
        },
      },
      {
        name: '运营商',
        render: () => {
          return LB_ISP[props.detail?.extension?.vip_isp] || '--';
        },
      },
    ];

    watch(
      () => props.detail.extension,
      (extension) => {
        isProtected.value = extension.delete_protect || false;
      },
      {
        deep: true,
        immediate: true,
      },
    );

    return () => (
      <div class={'clb-detail-continer'}>
        <div class='mb32'>
          <p class={'clb-detail-info-title'}>资源信息</p>
          <DetailInfo
            fields={resourceFields}
            detail={props.detail}
            onChange={async (payload) => {
              await props.updateLb(payload);
              await props.getDetails(props.id);
            }}
          />
        </div>
        <div>
          <p class={'clb-detail-info-title'}>配置信息</p>
          <DetailInfo fields={configFields} detail={props.detail} />
        </div>
      </div>
    );
  },
});
