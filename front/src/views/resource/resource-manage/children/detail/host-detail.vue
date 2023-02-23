<script lang="ts" setup>
import DetailHeader from '../../common/header/detail-header';
import DetailTab from '../../common/tab/detail-tab';
import HostInfo from '../components/host/host-info.vue';
import HostNetwork from '../components/host/host-network.vue'
import HostIp from '../components/host/host-ip.vue'
import HostDrive from '../components/host/host-drive.vue';
import HostSecurity from '../components/host/host-security.vue';

import {
  useI18n,
} from 'vue-i18n';
import useShutdown from '../../hooks/use-shutdown';
import useReboot from '../../hooks/use-reboot';
import usePassword from '../../hooks/use-password';
import useRefund from '../../hooks/use-refund';
import useBootUp from '../../hooks/use-boot-up';

const {
  t,
} = useI18n();

const {
  isShowShutdown,
  handleShutdown,
  HostShutdown,
} = useShutdown();

const {
  isShowReboot,
  handleReboot,
  HostReboot,
} = useReboot();

const {
  isShowPassword,
  handlePassword,
  HostPassword,
} = usePassword();

const {
  isShowRefund,
  handleRefund,
  HostRefund,
} = useRefund();

const {
  isShowBootUp,
  handleBootUp,
  HostBootUp,
} = useBootUp();

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

const componentMap = {
  detail: HostInfo,
  network: HostNetwork,
  ip: HostIp,
  drive: HostDrive,
  security: HostSecurity
}
</script>

<template>
  <detail-header>
    主机：（xxx）
    <template #right>
      <bk-button
        class="w100 ml10"
        theme="primary"
        @click="handleBootUp"
      >
        {{ t('开机') }}
      </bk-button>
      <bk-button
        class="w100 ml10"
        theme="primary"
        @click="handleShutdown"
      >
        {{ t('关机') }}
      </bk-button>
      <bk-button
        class="w100 ml10"
        theme="primary"
        @click="handleReboot"
      >
        {{ t('重启') }}
      </bk-button>
      <bk-button
        class="w100 ml10"
        theme="primary"
        @click="handlePassword"
      >
        {{ t('重置密码') }}
      </bk-button>
      <bk-button
        class="w100 ml10"
        theme="primary"
        @click="handleRefund"
      >
        {{ t('退回') }}
      </bk-button>
    </template>
  </detail-header>

  <detail-tab
    :tabs="hostTabs"
  >
    <template #default="type">
      <component :is="componentMap[type]"></component>
    </template>
  </detail-tab>

  <host-shutdown
    v-model:isShow="isShowShutdown"
    :title="t('关机')"
  />

  <host-reboot
    v-model:isShow="isShowReboot"
    :title="t('重启')"
  />

  <host-password
    v-model:isShow="isShowPassword"
    :title="t('修改密码')"
  />

  <host-refund
    v-model:isShow="isShowRefund"
    :title="t('主机回收')"
  />

  <host-boot-up
    v-model:isShow="isShowBootUp"
    :title="t('开机')"
  />
</template>

<style lang="scss" scoped>
.w100 {
  width: 100px;
}
.w60 {
  width: 60px;
}
</style>
