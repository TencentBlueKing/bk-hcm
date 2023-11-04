<script lang="ts" setup>
import DetailHeader from '../../common/header/detail-header';
import DetailTab from '../../common/tab/detail-tab';
import HostInfo from '../components/host/host-info/index.vue';
import HostNetwork from '../components/host/host-network/index.vue';
import HostIp from '../components/host/host-ip.vue';
import HostDrive from '../components/host/host-drive.vue';
import HostSecurity from '../components/host/host-security.vue';
import BusinessSelector from '@/components/business-selector/index.vue';
import bus from '@/common/bus';
import { useRouter,
  useRoute,
} from 'vue-router';
import { useResourceStore } from '@/store/resource';

import {
  useI18n,
} from 'vue-i18n';
import {
  InfoBox,
  Message,
} from 'bkui-vue';
import useDetail from '@/views/resource/resource-manage/hooks/use-detail';

import {
  ref,
  inject,
  computed,
} from 'vue';
import { Senarios, useWhereAmI } from '@/hooks/useWhereAmI';


const router = useRouter();
const {
  t,
} = useI18n();

const route = useRoute();

const resourceStore = useResourceStore();

const hostId = ref<any>(route.query?.id);
const cloudType = ref<any>(route.query?.type);
// 搜索过滤相关数据
const filter = ref({ op: 'and', rules: [] });
const isDialogShow = ref(false);
const selectedBizId = ref(0);
const isDialogBtnLoading = ref(false);

const isResourcePage: any = inject('isResourcePage');
const authVerifyData: any = inject('authVerifyData');

// 操作的相关信息
const cvmInfo = ref({
  start: { op: '开机', loading: false, status: ['RUNNING', 'running'] },
  stop: { op: '关机', loading: false, status: ['STOPPED', 'SHUTOFF', 'STOPPING', 'shutting-down', 'PowerState', 'stopped'] },
  reboot: { op: '重启', loading: false },
  destroy: { op: '回收', loading: false },
});


const actionName = computed(() => {   // 资源下没有业务ID
  console.log('isResourcePage.value', isResourcePage.value);
  return isResourcePage.value ? 'iaas_resource_operate' : 'biz_iaas_resource_operate';
});

const {
  loading,
  detail,
  getDetail,
} = useDetail(
  'cvms',
  hostId.value,
);

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

if (cloudType.value === 'gcp') {    // 腾讯云和Aws没有网络接口
  hostTabs.splice(4, 1);
}

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
    if (type === 'destroy') {  // 回收成功跳转回收记录
      router.push({
        path: '/business/host/recyclebin/cvm',
      });
    } else {
      getDetail();
    }
  } catch (error) {
    console.log(error);
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

// 权限弹窗 bus通知最外层弹出
const showAuthDialog = (authActionName: string) => {
  bus.$emit('auth', authActionName);
};

</script>

<template>
  <detail-header>
    <span class="header-title-prefix">
      主机详情
    </span>
    <span class="header-title-content">
      &nbsp;- ID {{`${hostId}`}}
    </span>
    <!-- <span class="status-stopped" v-if="(detail.bk_biz_id !== -1 && isResourcePage)">
      【已绑定】
    </span> -->
    <template #right>
      <span @click="showAuthDialog(actionName)">
        <bk-button
          class="w100 ml10"
          theme="primary"
          :disabled="(detail.bk_biz_id !== -1 && isResourcePage)
            || !authVerifyData?.permissionAction[actionName]"
          @click="() => isDialogShow = true"
          v-if="whereAmI === Senarios.resource"
        >
          {{ t('分配') }}
        </bk-button>
      </span>
      <span @click="showAuthDialog(actionName)">
        <bk-button
          class="w100 ml10"
          :disabled="cvmInfo.start.status.includes(detail.status) || (detail.bk_biz_id !== -1 && isResourcePage)
            || !authVerifyData?.permissionAction[actionName]"
          :loading="cvmInfo.start.loading"
          @click="() => {
            handleCvmOperate('start')
          }"
        >
          {{ t('开机') }}
        </bk-button>
      </span>
      <span @click="showAuthDialog(actionName)">
        <bk-button
          class="w100 ml10 mr10"
          :disabled="cvmInfo.stop.status.includes(detail.status) || (detail.bk_biz_id !== -1 && isResourcePage)
            || !authVerifyData?.permissionAction[actionName]"
          :loading="cvmInfo.stop.loading"
          @click="() => {
            handleCvmOperate('stop')
          }"
        >
          {{ t('关机') }}
        </bk-button>
      </span>
      <!-- <span @click="showAuthDialog(actionName)">
        <bk-button
          class="w100 ml10"
          theme="primary"
          :disabled="cvmInfo.stop.status.includes(detail.status) || (detail.bk_biz_id !== -1 && isResourcePage)
            || !authVerifyData?.permissionAction[actionName]"
          :loading="cvmInfo.reboot.loading"
          @click="() => {
            handleCvmOperate('reboot')
          }"
        >
          {{ t('重启') }}
        </bk-button>
      </span> -->
      <!-- <bk-button
        class="w100 ml10"
        theme="primary"
        @click="handlePassword"
      >
        {{ t('重置密码') }}
      </bk-button> -->
      <span @click="showAuthDialog(actionName)">
        <bk-dropdown
          trigger="click"
        >
          <bk-button
            :disabled="cvmInfo.stop.status.includes(detail.status) || (detail.bk_biz_id !== -1 && isResourcePage)
              || !authVerifyData?.permissionAction[actionName]">
            ⋮
          </bk-button>
          <template #content>
            <bk-dropdown-menu>
              <bk-dropdown-item
                @click="() => {
                  handleCvmOperate('destroy')
                }"
              >
                {{ t('回收') }}
              </bk-dropdown-item>
              <bk-dropdown-item
                @click="() => {
                  handleCvmOperate('reboot')
                }">
                {{ t('重启') }}
              </bk-dropdown-item>
            </bk-dropdown-menu>
          </template>
        </bk-dropdown>
      </span>
      <!-- <span @click="showAuthDialog(actionName)">
        <bk-button
          class="w100 ml10"
          theme="primary"
          :disabled="(detail.bk_biz_id !== -1 && isResourcePage)
            || !authVerifyData?.permissionAction[actionName]"
          :loading="cvmInfo.destroy.loading"
          @click="() => {
            handleCvmOperate('destroy')
          }"
        >
          {{ t('回收') }}
        </bk-button>
      </span> -->
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
    @closed="() => isDialogShow = false"
    @confirm="handleConfirm"
    :is-loading="isDialogBtnLoading"
  >
    <p class="mb6-text">目标业务</p>
    <business-selector
      v-model="selectedBizId"
      :authed="true"
      :auto-select="true">
    </business-selector>
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
  color: #63656E;
}
:deep(.detail-tab-main) .bk-tab-content {
  height: calc(100vh - 322px);
}
</style>
