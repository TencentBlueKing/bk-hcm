<script lang="ts" setup>
import DetailTab from '../../common/tab/detail-tab';
import HostInfo from '../components/host/host-info/index.vue';
import HostNetwork from '../components/host/host-network/index.vue';
import HostIp from '../components/host/host-ip.vue';
import HostDrive from '../components/host/host-drive.vue';
import HostSecurity from '../components/host/host-security.vue';
import BusinessSelector from '@/components/business-selector/index.vue';
import { useRouter, useRoute } from 'vue-router';
import { useResourceStore } from '@/store/resource';

import { useI18n } from 'vue-i18n';
import { InfoBox, Message } from 'bkui-vue';
import useDetail from '@/views/resource/resource-manage/hooks/use-detail';
import {
  useResourceDetailAuth,
  useResourceDetailRoute,
} from '@/views/resource/resource-manage/hooks/use-resource-detail-route';

import { ref, inject, computed, watchEffect } from 'vue';
import { Senarios, useWhereAmI } from '@/hooks/useWhereAmI';
import { MENU_BUSINESS_RECYCLEBIN, MENU_RESOURCE_RECYCLEBIN } from '@/constants/menu-symbol';
import useBreadcrumb from '@/hooks/use-breadcrumb';
import { VendorEnum, CLOUD_HOST_STATUS } from '@/common/constant';
import { HOST_RUNNING_STATUS, HOST_SHUTDOWN_STATUS } from '../../common/table/HostOperations';

const router = useRouter();
const { t } = useI18n();
const { setTitle } = useBreadcrumb();

const route = useRoute();

const resourceStore = useResourceStore();

const { detailId } = useResourceDetailRoute();
const hostId = ref<any>(detailId.value);
const cloudType = computed<VendorEnum>(() => (route.query?.vendor as VendorEnum) || detail.value?.vendor);
// 搜索过滤相关数据
const filter = ref({ op: 'and', rules: [] });
const isDialogShow = ref(false);
const selectedBizId = ref(0);
const isDialogBtnLoading = ref(false);

const isResourcePage: any = inject('isResourcePage');

// 操作的相关信息
const cvmInfo = ref({
  start: { op: '开机', loading: false, status: HOST_RUNNING_STATUS },
  stop: {
    op: '关机',
    loading: false,
    status: HOST_SHUTDOWN_STATUS,
  },
  reboot: { op: '重启', loading: false },
  destroy: { op: '回收', loading: false },
});

const { loading, detail, getDetail } = useDetail('cvms', hostId.value);
const { authPending, updateSign } = useResourceDetailAuth(computed(() => detail.value.account_id));

watchEffect(() => {
  if (hostId.value) {
    setTitle(`主机详情 - ID ${hostId.value}`);
  }
});

const hostTabs = computed(() => {
  const allTabs = [
    {
      name: '基本信息',
      value: 'detail',
    },
    {
      name: '网络接口',
      value: 'network',
    },
    {
      name: '弹性 IP',
      value: 'ip',
    },
    {
      name: '云硬盘',
      value: 'drive',
    },
    {
      name: '安全组',
      value: 'security',
    },
  ];

  let excludeTabs: string[] = [];
  if (cloudType.value === VendorEnum.TCLOUD || cloudType.value === VendorEnum.AWS) {
    excludeTabs = ['network'];
  } else if (cloudType.value === VendorEnum.GCP) {
    excludeTabs = ['security'];
  } else if (cloudType.value === VendorEnum.OTHER) {
    excludeTabs = ['network', 'ip', 'drive', 'security'];
  }

  return allTabs.filter((tab) => !excludeTabs.includes(tab.value));
});

const componentMap = {
  detail: HostInfo,
  network: HostNetwork,
  ip: HostIp,
  drive: HostDrive,
  security: HostSecurity,
};

const isBindBusiness = computed(() => {
  return detail.value.bk_biz_id !== -1 && isResourcePage.value;
});

const isOtherVendor = computed(() => {
  return detail.value?.vendor === VendorEnum.OTHER;
});

const { whereAmI } = useWhereAmI();

const handleCvmOperate = (type: string) => {
  const title = cvmInfo.value[type].op;
  InfoBox({
    title: `确定${title}`,
    subTitle: `确定将此主机${title}`,
    headerAlign: 'center',
    footerAlign: 'center',
    contentAlign: 'center',
    onConfirm() {
      modifyCvmStatus(type);
    },
  });
};

const modifyCvmStatus = async (type: string) => {
  try {
    Message({
      message: `${cvmInfo.value[type].op}中, 请不要操作`,
      theme: 'warning',
    });
    cvmInfo.value[type].loading = true;
    if (type === 'destroy') {
      await resourceStore.recycledCvmsData({ infos: [{ id: hostId.value, with_disk: true }] });
    } else {
      await resourceStore.cvmOperate(type, { ids: [hostId.value] });
    }
    Message({
      message: t('操作成功'),
      theme: 'success',
    });
    if (type === 'destroy') {
      // 回收成功跳转回收记录
      router.push({
        name: whereAmI.value === Senarios.business ? MENU_BUSINESS_RECYCLEBIN : MENU_RESOURCE_RECYCLEBIN,
        query: { type: 'cvm' },
      });
    } else {
      getDetail();
    }
  } catch (error) {
    console.error(error);
  } finally {
    cvmInfo.value[type].loading = false;
  }
};

const handleConfirm = async () => {
  isDialogBtnLoading.value = true;
  await resourceStore.assignBusiness('cvms', {
    cvm_ids: [hostId.value],
    bk_biz_id: selectedBizId.value,
  });
  isDialogBtnLoading.value = false;
  isDialogShow.value = false;
};

const disabledOption = computed(() => {
  if (authPending.value) return true;
  // 业务下，判断是否已被回收
  if (!isResourcePage.value) return detail.value?.recycle_status === 'recycling';
  // 资源下，判断是否分配业务，是否已被回收
  return detail.value?.bk_biz_id !== -1 || detail.value?.recycle_status === 'recycling';
});

const bktoolTipsOptions = computed(() => {
  // 资源下，是否分配业务
  if (isResourcePage.value && detail.value?.bk_biz_id !== -1)
    return {
      content: '该主机仅可在业务下操作',
      disabled: detail.value.bk_biz_id === -1,
    };
  // 业务/资源下，是否已被回收
  if (detail.value?.recycle_status === 'recycling')
    return {
      content: '已回收的资源，不支持操作',
      disabled: detail.value.recycle_status !== 'recycling',
    };

  return null;
});
</script>

<template>
  <Teleport defer to="#breadcrumbExtra" v-if="!isOtherVendor">
    <hcm-auth
      v-if="whereAmI === Senarios.resource"
      :sign="updateSign"
      :ignore="authPending"
      tag="span"
      v-slot="{ noPerm }"
    >
      <bk-button
        v-bk-tooltips="(!noPerm && bktoolTipsOptions) || { disabled: true }"
        theme="primary"
        :disabled="noPerm || disabledOption"
        @click="isDialogShow = true"
      >
        {{ t('分配') }}
      </bk-button>
    </hcm-auth>
    <hcm-auth :sign="updateSign" :ignore="authPending" tag="span" v-slot="{ noPerm }">
      <bk-button
        v-bk-tooltips="
          (!noPerm && bktoolTipsOptions) || {
            content: `当前主机处于 ${CLOUD_HOST_STATUS[detail.status]} 状态`,
            disabled: noPerm || !cvmInfo.start.status.includes(detail.status),
          }
        "
        :disabled="noPerm || disabledOption || cvmInfo.start.status.includes(detail.status)"
        :loading="cvmInfo.start.loading"
        @click="handleCvmOperate('start')"
      >
        {{ t('开机') }}
      </bk-button>
    </hcm-auth>
    <hcm-auth :sign="updateSign" :ignore="authPending" tag="span" v-slot="{ noPerm }">
      <bk-button
        v-bk-tooltips="
          (!noPerm && bktoolTipsOptions) || {
            content: `当前主机处于 ${CLOUD_HOST_STATUS[detail.status]} 状态`,
            disabled: noPerm || !cvmInfo.stop.status.includes(detail.status),
          }
        "
        :disabled="noPerm || disabledOption || cvmInfo.stop.status.includes(detail.status)"
        :loading="cvmInfo.stop.loading"
        @click="handleCvmOperate('stop')"
      >
        {{ t('关机') }}
      </bk-button>
    </hcm-auth>
    <hcm-auth :sign="updateSign" :ignore="authPending" tag="span" v-slot="{ noPerm }">
      <bk-dropdown
        trigger="click"
        :disabled="noPerm || disabledOption || cvmInfo.stop.status.includes(detail.status)"
        :popover-options="{ clickContentAutoHide: true }"
      >
        <bk-button
          v-bk-tooltips="(!noPerm && bktoolTipsOptions) || { disabled: true }"
          :disabled="noPerm || disabledOption || cvmInfo.stop.status.includes(detail.status)"
        >
          ⋮
        </bk-button>
        <template #content>
          <bk-dropdown-menu>
            <bk-dropdown-item @click="handleCvmOperate('destroy')">
              {{ t('回收') }}
            </bk-dropdown-item>
            <bk-dropdown-item @click="handleCvmOperate('reboot')">
              {{ t('重启') }}
            </bk-dropdown-item>
          </bk-dropdown-menu>
        </template>
      </bk-dropdown>
    </hcm-auth>
  </Teleport>

  <div class="detail-content-wrap">
    <detail-tab :tabs="hostTabs">
      <template #default="type">
        <bk-loading :loading="loading">
          <component
            v-if="!loading"
            :is="componentMap[type]"
            :data="detail"
            :type="cloudType"
            :filter="filter"
            :is-bind-business="isBindBusiness"
          ></component>
        </bk-loading>
      </template>
    </detail-tab>
  </div>

  <bk-dialog
    :is-show="isDialogShow"
    title="主机分配"
    :theme="'primary'"
    quick-close
    @closed="() => (isDialogShow = false)"
    @confirm="handleConfirm"
    :is-loading="isDialogBtnLoading"
  >
    <p class="mb6-text">目标业务</p>
    <business-selector v-model="selectedBizId" :authed="true" :auto-select="true"></business-selector>
  </bk-dialog>
</template>

<style lang="scss" scoped>
.w100 {
  width: 100px;
}

.w60 {
  width: 60px;
}

.mb6-text {
  margin-bottom: 6px;
  color: #63656e;
}

.btn {
  min-width: 64px;
  margin-right: 8px;
}
</style>
