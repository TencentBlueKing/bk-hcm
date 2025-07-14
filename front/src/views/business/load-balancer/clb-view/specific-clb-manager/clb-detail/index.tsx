import { PropType, computed, defineComponent, ref, watch } from 'vue';
import { Button, Loading, Message, Switcher, Table, Tag } from 'bkui-vue';
import DetailInfo from '@/views/resource/resource-manage/common/info/detail-info';
import CorsConfigDialog from './CorsConfigDialog';
import AddSnatIpDialog from './AddSnatIpDialog';
import Confirm from '@/components/confirm';
import { useRouteLinkBtn, TypeEnum, IDetail } from '@/hooks/useRouteLinkBtn';
import StatusNormal from '@/assets/image/Status-normal.png';
import StatusUnknown from '@/assets/image/Status-unknown.png';
import { timeFormatter, formatTags, parseTimeFromNow } from '@/common/util';
import { CHARGE_TYPE, CLB_SPECS, LB_ISP, LB_TYPE_MAP } from '@/common/constant';
import { useBusinessStore } from '@/store';
import { useRegionsStore } from '@/store/useRegionsStore';
import { IP_VERSION_MAP } from '@/constants';
import { QueryRuleOPEnum } from '@/typings';
import { useI18n } from 'vue-i18n';
import { formatBandwidth, getInstVip } from '@/utils';
import './index.scss';
import { FieldList } from '@/views/resource/resource-manage/common/info-list/types';

export default defineComponent({
  props: {
    detail: Object as PropType<IDetail>,
    getDetails: Function,
    updateLb: Function,
    id: String,
  },
  setup(props) {
    const { t } = useI18n();
    const businessStore = useBusinessStore();

    const isReloadLoading = ref(false);
    const regionStore = useRegionsStore();
    const isProtected = ref(false);
    const isLoading = ref(false);
    const listenerNum = ref(0);
    const vpcDetail = ref(null);
    const targetVpcDetail = ref(null);
    const resourceFields: FieldList = [
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
        copyContent: props.detail?.cloud_vpc_id || '--',
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
            <Tag theme={isProtected.value ? 'success' : ''}> {isProtected.value ? t('已开启') : t('未开启')} </Tag>
            <i
              class='hcm-icon bkhcm-icon-info-line ml10'
              v-bk-tooltips={{
                content: '开启删除保护后，在云控制台或调用 API 均无法删除该实例',
                placement: 'top-end',
              }}></i>
          </div>
        ),
        copy: false,
      },
      {
        name: '状态',
        render() {
          return (
            <div class={'status-wrapper'}>
              <img src={!props.detail.status ? StatusUnknown : StatusNormal} class={'mr6'} width={14} height={14}></img>
              <span>{!props.detail.status ? t('创建中') : t('正常运行')}</span>
            </div>
          );
        },
        copyContent: !props.detail.status ? t('创建中') : t('正常运行'),
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
            ?.map((zone: string) => `${regionStore.getZoneName(zone, props.detail.vendor)}-(主)`)
            .join(',');
          const backupsStr = backups
            ?.map((zone: string) => `${regionStore.getZoneName(zone, props.detail.vendor)}-(备)`)
            .join(',');
          return `${mainsStr}${backupsStr?.length ? `,${backupsStr}` : ''}`;
        },
      },
      {
        name: '标签',
        prop: 'tags',
        render(val) {
          return formatTags(val);
        },
      },
      {
        name: '同步时间',
        prop: 'sync_time',
        render: () => `${timeFormatter(props.detail.sync_time)}(${parseTimeFromNow(props.detail.sync_time)})`,
      },
    ];

    const configFields: FieldList = [
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
          return getInstVip(props.detail);
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
        render: () => formatBandwidth(props.detail?.bandwidth),
      },
      {
        name: '运营商',
        render: () => {
          const displayValue = props.detail?.isp ? LB_ISP[props.detail.isp] ?? props.detail.isp : '--';
          return displayValue;
        },
      },
    ];

    watch(
      () => props.detail,
      async () => {
        // 当 lbInfo 信息变更时, 重新获取监听器数量, vpc 详情
        const listenerNumRes = await businessStore.asyncGetListenerCount({ lb_ids: [props.detail?.id] });
        const vpcDetailRes = await businessStore.detail('vpcs', props.detail?.vpc_id);
        if (isCorsV1.value && props.detail.extension?.target_vpc) {
          // 通过cloud_id查list接口, 从而获取到target_vpc_name
          const res = await businessStore.list(
            {
              filter: {
                op: QueryRuleOPEnum.AND,
                rules: [{ field: 'cloud_id', op: QueryRuleOPEnum.EQ, value: props.detail.extension?.target_vpc }],
              },
              page: { count: false, start: 0, limit: 1 },
            },
            `vendors/${props.detail?.vendor}/vpcs`,
          );
          [targetVpcDetail.value] = res.data.details;
        }
        listenerNum.value = listenerNumRes.data.details[0]?.num;
        vpcDetail.value = vpcDetailRes.data;
      },
      {
        deep: true,
        immediate: true,
      },
    );

    // 重新加载负载均衡详情
    const handleReloadLbDetail = async () => {
      try {
        isReloadLoading.value = true;
        await props.getDetails(props.id);
      } finally {
        isReloadLoading.value = false;
      }
    };

    // 删除Snat IP
    const handleDeleteSnatIp = (data: { ip: string; subnet_id: string }) => {
      Confirm('请确定删除SNAT IP', `将删除SNAT IP【${data.ip}】`, async () => {
        await businessStore.deleteSnatIps(props.id, props.detail.vendor, { delete_ips: [data.ip] });
        Message({ theme: 'success', message: '删除成功' });
        await handleReloadLbDetail();
      });
    };

    const corsColumns = [
      {
        label: '内网IP',
        field: 'ip',
      },
      {
        label: '所属子网',
        field: 'subnet_id',
      },
      {
        label: '操作',
        render({ data }: any) {
          return (
            <Button text theme='primary' onClick={() => handleDeleteSnatIp(data)}>
              {t('删除')}
            </Button>
          );
        },
      },
    ];

    const isShowCorsConfig = ref(false);
    const isShowAddSnatIp = ref(false);

    // 是否开启跨域
    const isCorsOpen = computed(() => {
      return (
        props.detail.extension?.target_region || props.detail.extension?.target_vpc || props.detail.extension?.snat_pro
      );
    });
    // 是否为跨域1.0
    const isCorsV1 = computed(() => {
      return props.detail.extension?.target_region || props.detail.extension?.target_vpc;
    });
    // 是否为跨域2.0
    const isCorsV2 = computed(() => {
      return props.detail.extension?.snat_pro;
    });

    // 开启/关闭跨域2.0
    const isSnatproChange = ref(false);
    const isSnatproOpen = ref(false);
    const handleChangeSnatPro = async (snat_pro: boolean) => {
      isSnatproChange.value = true;
      isSnatproOpen.value = snat_pro;
      try {
        await businessStore.updateLbDetail(props.detail.vendor, { id: props.id, snat_pro });
        Message({ theme: 'success', message: '修改成功' });
        await props.getDetails(props.id);
      } catch (error) {
        isSnatproOpen.value = false;
      } finally {
        isSnatproChange.value = false;
      }
    };

    watch(
      () => props.detail.extension,
      (extension) => {
        isProtected.value = extension.delete_protect || false;
        isSnatproOpen.value = extension.snat_pro || false;
      },
      {
        deep: true,
        immediate: true,
      },
    );

    return () => (
      <Loading class={'clb-detail-continer'} loading={isReloadLoading.value} opacity={1}>
        <div class='mb32'>
          <p class={'clb-detail-info-title'}>{t('资源信息')}</p>
          <DetailInfo
            fields={resourceFields}
            detail={props.detail}
            onChange={async (payload) => {
              await props.updateLb(payload);
              await props.getDetails(props.id);
            }}
            globalCopyable
          />
        </div>
        <div>
          <p class={'clb-detail-info-title'}>{t('配置信息')}</p>
          <DetailInfo fields={configFields} detail={props.detail} globalCopyable />
        </div>
        <div>
          <p class={'clb-detail-info-title'}>{t('跨域配置')}</p>
          <div class='cors-config-container'>
            {/* 跨域1.0 */}
            {isCorsV1.value && (
              <>
                <div class='cors-config-item'>
                  <div class='cors-config-item-title'>{t('跨地域绑定1.0')}</div>
                  <div class='cors-config-item-content'>
                    {t('跨地域绑定某一VPC内的云服务器')}
                    <Button text theme='primary' class='ml10' onClick={() => (isShowCorsConfig.value = true)}>
                      {t('预览')}
                    </Button>
                  </div>
                </div>
                <CorsConfigDialog
                  v-model:isShow={isShowCorsConfig.value}
                  lbInfo={props.detail}
                  listenerNum={listenerNum.value}
                  vpcDetail={vpcDetail.value}
                  targetVpcDetail={targetVpcDetail.value}
                />
              </>
            )}
            {/* 跨域2.0 */}
            {(isCorsV2.value || !isCorsOpen.value) && (
              <>
                <div class='cors-config-item'>
                  <div class='cors-config-item-title'>{t('跨地域绑定2.0')}</div>
                  <div class='cors-config-item-content'>
                    <div>
                      <Switcher
                        class='mr10'
                        modelValue={isSnatproOpen.value}
                        theme='primary'
                        onChange={handleChangeSnatPro}
                        disabled={props.detail?.extension?.snat_ips?.length > 0 || isSnatproChange.value}
                        v-bk-tooltips={{
                          content: '当前负载均衡已绑定SNAT IP，不可关闭跨域',
                          disabled: props.detail?.extension?.snat_ips?.length === 0,
                        }}
                      />
                      {t('跨多个地域，绑定多个非本VPC内的IP，以及云下IDC内部的IP')}
                    </div>
                    <div class='snat-ip-container'>
                      <div class='top-bar'>
                        <Button text theme='primary' onClick={() => (isShowAddSnatIp.value = true)}>
                          <i class='hcm-icon bkhcm-icon-plus-circle-shape mr5'></i>
                          {t('新增 SNAT 的 IP')}
                        </Button>
                        <span class='desc'>{t('绑定IDC内部的IP，则需要添加SNAT IP。云上IP，则无需增加。')}</span>
                      </div>
                      <Table columns={corsColumns} data={props.detail?.extension?.snat_ips}></Table>
                    </div>
                  </div>
                </div>
                <AddSnatIpDialog
                  v-model:isShow={isShowAddSnatIp.value}
                  lbInfo={props.detail}
                  vpcDetail={vpcDetail.value}
                  reloadLbDetail={handleReloadLbDetail}
                />
              </>
            )}
          </div>
        </div>
      </Loading>
    );
  },
});
