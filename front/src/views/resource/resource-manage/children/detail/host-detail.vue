<script lang="ts" setup>
import DetailHeader from '../../common/header/detail-header';
import DetailInfo from '../../common/info/detail-info';
import DetailTab from '../../common/tab/detail-tab';
import HostInfo from '../components/host/host-info.vue';
import HostSubnet from '../components/host/host-subnet.vue';
import HostDrive from '../components/host/host-drive.vue';
import HostSecurity from '../components/host/host-security.vue';
import {
  AngleRight,
} from 'bkui-vue/lib/icon';

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

// 更多
const moreOperations = [
  {
    name: t('重置密码'),
    handler: handlePassword,
  },
  {
    name: t('退回'),
    handler: handleRefund,
  },
];

const hostFields = [
  {
    name: 'ID',
    value: '1234223',
  },
  {
    name: '账号ID',
    value: '1234223',
    link: 'http://www.baidu.com',
  },
  {
    name: '账号名称',
    value: '1234223',
  },
  {
    name: '备注',
    value: '1234223',
    edit: true,
  },
  {
    name: 'VPC',
    value: '1234223',
    copy: '1234223',
  },
];
const hostTabs = [
  {
    name: '详细信息',
    value: 'detail',
  },
  {
    name: '网络',
    value: 'network',
  },
  {
    name: '云硬盘',
    value: 'drive',
  },
  {
    name: '安全',
    value: 'security',
  },
];
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
      <bk-dropdown
        class="ml10"
        placement="right-start"
      >
        <bk-button>
          <span class="w60">
            {{ t('更多') }}
          </span>
          <angle-right
            width="16"
            height="16"
          />
        </bk-button>
        <template #content>
          <bk-dropdown-menu>
            <bk-dropdown-item
              v-for="operation in moreOperations"
              :key="operation.name"
              @click="operation.handler"
            >
              {{ operation.name }}
            </bk-dropdown-item>
          </bk-dropdown-menu>
        </template>
      </bk-dropdown>
    </template>
  </detail-header>

  <detail-info
    :fields="hostFields"
  />

  <detail-tab
    :tabs="hostTabs"
  >
    <template #default="type">
      <host-info v-if="type === 'detail'"></host-info>
      <host-subnet v-if="type === 'network'"></host-subnet>
      <host-drive v-if="type === 'drive'"></host-drive>
      <host-security v-if="type === 'security'"></host-security>
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
