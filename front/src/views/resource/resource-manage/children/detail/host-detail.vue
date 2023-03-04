<script lang="ts" setup>
import DetailHeader from '../../common/header/detail-header';
import DetailTab from '../../common/tab/detail-tab';
import HostInfo from '../components/host/host-info/index.vue';
import HostNetwork from '../components/host/host-network/index.vue';
import HostIp from '../components/host/host-ip.vue';
import HostDrive from '../components/host/host-drive.vue';
import HostSecurity from '../components/host/host-security.vue';
import {
  useResourceStore,
} from '@/store/resource';

import {
  useI18n,
} from 'vue-i18n';
import {
  InfoBox,
  Message,
} from 'bkui-vue';
// import useShutdown from '../../hooks/use-shutdown';
// import useReboot from '../../hooks/use-reboot';
// import usePassword from '../../hooks/use-password';
import useRefund from '../../hooks/use-refund';
// import useBootUp from '../../hooks/use-boot-up';
import useDetail from '@/views/resource/resource-manage/hooks/use-detail';

import {
  ref,
} from 'vue';


import {
  useRoute,
} from 'vue-router';

const {
  t,
} = useI18n();

const route = useRoute();

const resourceStore = useResourceStore();

const hostId = ref<any>(route.query?.id);
const cloudType = ref<any>(route.query?.type);
// 搜索过滤相关数据
const filter = ref({ op: 'and', rules: [] });
const cvmStatus = ref({
  start: ['RUNNING'],
  stop: ['STOPPED', 'SHUTOFF', 'STOPPING', 'shutting-down', 'PowerState', 'stopped'],
});

const {
  loading,
  detail,
} = useDetail(
  'cvms',
  hostId.value,
);

console.log('extension', detail);

// const {
//   isShowShutdown,
//   handleShutdown,
//   HostShutdown,
// } = useShutdown();

// const {
//   isShowReboot,
//   handleReboot,
//   HostReboot,
// } = useReboot();

// const {
//   isShowPassword,
//   handlePassword,
//   HostPassword,
// } = usePassword();

const {
  isShowRefund,
  handleRefund,
  HostRefund,
} = useRefund();

// const {
//   isShowBootUp,
//   handleBootUp,
//   HostBootUp,
// } = useBootUp();

const hostTabs = [
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
if (cloudType.value === 'tcloud' || cloudType.value === 'aws') {    // 腾讯云和Aws没有网络接口
  hostTabs.splice(1, 1);
}

const componentMap = {
  detail: HostInfo,
  network: HostNetwork,
  ip: HostIp,
  drive: HostDrive,
  security: HostSecurity,
};

const handleCvmOperate = (type: string) => {
  let title = '开机';
  if (type === 'stop') {
    title = '关机';
  } else if (type === 'reboot') {
    title = '重启';
  }
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
    await resourceStore.cvmOperate(type, { ids: [hostId.value] });
    Message({
      message: t('操作成功'),
      theme: 'success',
    });
  } catch (error) {
    console.log(error);
  } finally {
  }
};

</script>

<template>
  <detail-header>
    主机：ID（{{`${hostId}`}}）
    <template #right>
      <bk-button
        class="w100 ml10"
        theme="primary"
        :disabled="cvmStatus.start.includes(detail.status)"
        @click="() => {
          handleCvmOperate('start')
        }"
      >
        {{ t('开机') }}
      </bk-button>
      <bk-button
        class="w100 ml10"
        theme="primary"
        :disabled="cvmStatus.stop.includes(detail.status)"
        @click="() => {
          handleCvmOperate('stop')
        }"
      >
        {{ t('关机') }}
      </bk-button>
      <bk-button
        class="w100 ml10"
        theme="primary"
        @click="() => {
          handleCvmOperate('reboot')
        }"
      >
        {{ t('重启') }}
      </bk-button>
      <!-- <bk-button
        class="w100 ml10"
        theme="primary"
        @click="handlePassword"
      >
        {{ t('重置密码') }}
      </bk-button> -->
      <bk-button
        class="w100 ml10"
        theme="primary"
        @click="handleRefund"
      >
        {{ t('回收') }}
      </bk-button>
    </template>
  </detail-header>

  <div class="host-detail">
    <detail-tab
      :tabs="hostTabs"
    >
      <template #default="type">
        <bk-loading
          :loading="loading"
        >
          <component :is="componentMap[type]" :data="detail" :type="cloudType" :filter="filter"></component>
        </bk-loading>
      </template>
    </detail-tab>
  </div>

  <!-- <host-shutdown
    v-model:isShow="isShowShutdown"
    :title="t('关机')"
  />

  <host-reboot
    v-model:isShow="isShowReboot"
    :title="t('重启')"
  /> -->

  <!-- <host-password
    v-model:isShow="isShowPassword"
    :title="t('修改密码')"
  /> -->

  <host-refund
    v-model:isShow="isShowRefund"
    :title="t('主机回收')"
  />

  <!-- <host-boot-up
    v-model:isShow="isShowBootUp"
    :title="t('开机')" -->
  <!-- /> -->
</template>

<style lang="scss" scoped>
.w100 {
  width: 100px;
}
.w60 {
  width: 60px;
}
:deep(.detail-tab-main) .bk-tab-content {
  height: calc(100vh - 300px);
}
</style>
