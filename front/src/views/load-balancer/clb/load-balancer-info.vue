<script setup lang="ts">
import { computed, ref, watch, h, withDirectives } from 'vue';
import { Button, Loading, Message, Switcher, Table, Tag, bkTooltips } from 'bkui-vue';
import DetailInfo from '@/views/resource/resource-manage/common/info/detail-info';
import CorsConfigDialog from '@/views/business/load-balancer/clb-view/specific-clb-manager/clb-detail/CorsConfigDialog';
import AddSnatIpDialog from '@/views/business/load-balancer/clb-view/specific-clb-manager/clb-detail/AddSnatIpDialog';
import Confirm from '@/components/confirm';
import { useRouteLinkBtn, TypeEnum, IDetail } from '@/hooks/useRouteLinkBtn';
import StatusNormal from '@/assets/image/Status-normal.png';
import StatusUnknown from '@/assets/image/Status-unknown.png';
import { timeFormatter, formatTags } from '@/common/util';
import { CHARGE_TYPE, CLB_SPECS, LB_ISP, LB_TYPE_MAP } from '@/common/constant';
import { useBusinessStore } from '@/store';
import { useRegionsStore } from '@/store/useRegionsStore';
import { IP_VERSION_MAP } from '@/constants';
import { QueryRuleOPEnum } from '@/typings';
import { useI18n } from 'vue-i18n';
import { getInstVip } from '@/utils';
import '@/views/business/load-balancer/clb-view/specific-clb-manager/clb-detail/index.scss';
import { FieldList } from '@/views/resource/resource-manage/common/info-list/types';
import { ILoadBalancerDetails } from '@/store/load-balancer/clb';

const props = defineProps<{ lbId: string; details: ILoadBalancerDetails }>();
const detailsInfo = computed(() => props.details ?? ({} as IDetail));

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
    name: t('名称'),
    prop: 'name',
    edit: true,
  },
  {
    name: t('所属网络'),
    prop: 'cloud_vpc_id',
    render() {
      return useRouteLinkBtn(detailsInfo.value, {
        id: 'vpc_id',
        name: 'cloud_vpc_id',
        type: TypeEnum.VPC,
      });
    },
    copyContent: detailsInfo.value.cloud_vpc_id || '--',
  },
  {
    name: t('ID'),
    prop: 'cloud_id',
  },
  {
    name: t('删除保护'),
    render() {
      return h('div', {}, [
        h(Switcher, {
          theme: 'primary',
          class: 'mr5',
          modelValue: isProtected.value,
          disabled: isLoading.value,
          onChange: async (val) => {
            isLoading.value = true;
            isProtected.value = val;
            try {
              await updateLb({
                delete_protect: val,
              });
            } catch (_e) {
              isProtected.value = !val;
            } finally {
              isLoading.value = false;
            }
          },
        }),
        h(Tag, { theme: isProtected.value ? 'success' : '' }, isProtected.value ? t('已开启') : t('未开启')),

        withDirectives(
          h('i', {
            class: 'hcm-icon bkhcm-icon-info-line ml10',
          }),
          [
            [
              bkTooltips,
              {
                content: t('开启删除保护后，在云控制台或调用 API 均无法删除该实例'),
                placement: 'top-end',
              },
            ],
          ],
        ),
      ]);
    },
    copy: false,
  },
  {
    name: t('状态'),
    render() {
      return h('div', { class: 'status-wrapper' }, [
        h('img', {
          src: !detailsInfo.value.status ? StatusUnknown : StatusNormal,
          class: 'mr6',
          width: 14,
          height: 14,
        }),
        h('span', !detailsInfo.value.status ? t('创建中') : t('正常运行')),
      ]);
    },
    copyContent: !detailsInfo.value.status ? t('创建中') : t('正常运行'),
  },
  {
    name: t('IP版本'),
    prop: 'ip_version',
    render() {
      return IP_VERSION_MAP[detailsInfo.value.ip_version];
    },
  },
  {
    name: t('网络类型'),
    prop: 'lb_type',
    render() {
      return LB_TYPE_MAP[detailsInfo.value.lb_type];
    },
  },
  {
    name: t('创建时间'),
    prop: 'created_at',
    render() {
      return timeFormatter(detailsInfo.value.created_at);
    },
  },
  {
    name: t('地域'),
    prop: 'region',
    render() {
      return regionStore.getRegionName(detailsInfo.value.vendor, detailsInfo.value.region);
    },
  },
  {
    name: t('可用区域'),
    prop: 'zones',
    render() {
      const mains = detailsInfo.value.zones;
      const backups = detailsInfo.value.backup_zones;
      const mainsStr = mains
        ?.map((zone: string) => `${regionStore.getZoneName(zone, detailsInfo.value.vendor)}-(主)`)
        .join(',');
      const backupsStr = backups
        ?.map((zone: string) => `${regionStore.getZoneName(zone, detailsInfo.value.vendor)}-(备)`)
        .join(',');
      return `${mainsStr}${backupsStr?.length ? `,${backupsStr}` : ''}`;
    },
  },
  {
    name: t('标签'),
    prop: 'tags',
    render(val) {
      return formatTags(val);
    },
  },
];

const configFields: FieldList = [
  {
    name: t('负载均衡域名'),
    prop: 'domain',
  },
  {
    name: t('实例计费模式'),
    render() {
      return CHARGE_TYPE[detailsInfo.value.extension?.charge_type] || '--';
    },
  },
  {
    name: t('负载均衡VIP'),
    render: () => {
      return getInstVip(detailsInfo.value);
    },
  },
  {
    name: t('带宽计费模式'),
    render: () => {
      return detailsInfo.value.extension?.internet_charge_type || '--';
    },
  },
  {
    name: t('规格类型'),
    render: () => {
      return CLB_SPECS[detailsInfo.value.extension?.sla_type] || '--';
    },
  },
  {
    name: t('带宽上限'),
    render: () => {
      return detailsInfo.value.extension?.internet_max_bandwidth_out || '--';
    },
  },
  {
    name: t('运营商'),
    render: () => {
      return LB_ISP[detailsInfo.value.extension?.vip_isp] || '--';
    },
  },
];

watch(
  () => detailsInfo.value,
  async () => {
    // 当 lbInfo 信息变更时, 重新获取监听器数量, vpc 详情
    const listenerNumRes = await businessStore.asyncGetListenerCount({ lb_ids: [detailsInfo.value.id] });
    const vpcDetailRes = await businessStore.detail('vpcs', detailsInfo.value.vpc_id);
    if (isCorsV1.value && detailsInfo.value.extension?.target_vpc) {
      // 通过cloud_id查list接口, 从而获取到target_vpc_name
      const res = await businessStore.list(
        {
          filter: {
            op: QueryRuleOPEnum.AND,
            rules: [{ field: 'cloud_id', op: QueryRuleOPEnum.EQ, value: detailsInfo.value.extension?.target_vpc }],
          },
          page: { count: false, start: 0, limit: 1 },
        },
        `vendors/${detailsInfo.value.vendor}/vpcs`,
      );
      [targetVpcDetail.value] = res.data.details;
    }
    listenerNum.value = listenerNumRes.data.details[0]?.num;
    vpcDetail.value = vpcDetailRes.data;
  },
  {
    deep: true,
  },
);

const updateLb = async (payload: Record<string, any>) => {
  await businessStore.updateLbDetail(detailsInfo.value.vendor, {
    id: detailsInfo.value.id,
    ...payload,
  });
  Message({
    message: t('更新成功'),
    theme: 'success',
  });
};

// 重新加载负载均衡详情
const handleReloadLbDetail = async () => {
  try {
    isReloadLoading.value = true;
    await props.getDetails(props.lbId);
  } finally {
    isReloadLoading.value = false;
  }
};

// 删除Snat IP
const handleDeleteSnatIp = (data: { ip: string; subnet_id: string }) => {
  Confirm('请确定删除SNAT IP', `将删除SNAT IP【${data.ip}】`, async () => {
    await businessStore.deleteSnatIps(props.lbId, detailsInfo.value.vendor, { delete_ips: [data.ip] });
    Message({ theme: 'success', message: t('删除成功') });
    await handleReloadLbDetail();
  });
};

const corsColumns = [
  {
    label: t('内网IP'),
    field: 'ip',
  },
  {
    label: t('所属子网'),
    field: 'subnet_id',
  },
  {
    label: t('操作'),
    render({ data }: any) {
      return h(Button, { text: true, theme: 'primary', onClick: () => handleDeleteSnatIp(data) }, t('删除'));
    },
  },
];

const isShowCorsConfig = ref(false);
const isShowAddSnatIp = ref(false);

// 是否开启跨域
const isCorsOpen = computed(() => {
  return (
    detailsInfo.value.extension?.target_region ||
    detailsInfo.value.extension?.target_vpc ||
    detailsInfo.value.extension?.snat_pro
  );
});
// 是否为跨域1.0
const isCorsV1 = computed(() => {
  return detailsInfo.value.extension?.target_region || detailsInfo.value.extension?.target_vpc;
});
// 是否为跨域2.0
const isCorsV2 = computed(() => {
  return detailsInfo.value.extension?.snat_pro;
});

// 开启/关闭跨域2.0
const isSnatproChange = ref(false);
const isSnatproOpen = ref(false);
const handleChangeSnatPro = async (snat_pro: boolean) => {
  isSnatproChange.value = true;
  isSnatproOpen.value = snat_pro;
  try {
    await businessStore.updateLbDetail(detailsInfo.value.vendor, { id: props.lbId, snat_pro });
    Message({ theme: 'success', message: t('修改成功') });
    await props.getDetails(props.lbId);
  } catch (error) {
    isSnatproOpen.value = false;
  } finally {
    isSnatproChange.value = false;
  }
};

watch(
  () => detailsInfo.value.extension,
  (extension) => {
    isProtected.value = extension?.delete_protect || false;
    isSnatproOpen.value = extension?.snat_pro || false;
  },
  {
    deep: true,
    immediate: true,
  },
);
</script>
<template>
  <Loading class="clb-detail-continer" :loading="isReloadLoading" :opacity="1">
    <div class="mb32">
      <p class="clb-detail-info-title">{{ t('资源信息') }}</p>
      <DetailInfo
        :fields="resourceFields"
        :detail="detailsInfo"
        @change="
          async (payload) => {
            await updateLb(payload);
          }
        "
        global-copyable
      />
    </div>
    <div>
      <p class="clb-detail-info-title">{{ t('配置信息') }}</p>
      <DetailInfo :fields="configFields" :detail="detailsInfo" global-copyable />
    </div>
    <div>
      <p class="clb-detail-info-title">{{ t('跨域配置') }}</p>
      <div class="cors-config-container">
        <div class="cors-config-item" v-if="isCorsV1">
          <div class="cors-config-item-title">{{ t('跨地域绑定1.0') }}</div>
          <div class="cors-config-item-content">
            {{ t('跨地域绑定某一VPC内的云服务器') }}
            <Button text theme="primary" class="ml10" @click="() => (isShowCorsConfig = true)">
              {{ t('预览') }}
            </Button>
          </div>
        </div>
        <CorsConfigDialog
          v-model:is-show="isShowCorsConfig"
          :lb-info="detailsInfo"
          :listener-num="listenerNum"
          :vpc-detail="vpcDetail"
          :target-vpc-detail="targetVpcDetail"
        />
        <div class="cors-config-item" v-if="isCorsV2 || !isCorsOpen">
          <div class="cors-config-item-title">{{ t('跨地域绑定2.0') }}</div>
          <div class="cors-config-item-content">
            <div>
              <Switcher
                class="mr10"
                :model-value="isSnatproOpen"
                theme="primary"
                @change="handleChangeSnatPro"
                :disabled="detailsInfo.extension?.snat_ips?.length > 0 || isSnatproChange"
                v-bk-tooltips="{
                  content: '当前负载均衡已绑定SNAT IP，不可关闭跨域',
                  disabled: detailsInfo.extension?.snat_ips?.length === 0,
                }"
              />
              {{ t('跨多个地域，绑定多个非本VPC内的IP，以及云下IDC内部的IP') }}
            </div>
            <div class="snat-ip-container">
              <div class="top-bar">
                <Button text theme="primary" @click="() => (isShowAddSnatIp = true)">
                  <i class="hcm-icon bkhcm-icon-plus-circle-shape mr5"></i>
                  {{ t('新增 SNAT 的 IP') }}
                </Button>
                <span class="desc">{{ t('绑定IDC内部的IP，则需要添加SNAT IP。云上IP，则无需增加。') }}</span>
              </div>
              <Table :columns="corsColumns" :data="detailsInfo.extension?.snat_ips"></Table>
            </div>
          </div>
        </div>
        <AddSnatIpDialog
          v-if="isShowAddSnatIp"
          v-model:is-show="isShowAddSnatIp"
          :lb-info="detailsInfo"
          :vpc-detail="vpcDetail"
          :reload-lb-detail="handleReloadLbDetail"
        />
      </div>
    </div>
  </Loading>
</template>

<style scoped></style>
