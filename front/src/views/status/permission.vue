<script setup lang="ts">
import { computed, ref } from 'vue';
import { useRoute } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { useAuthStore } from '@/store/auth';
import { type IAuthSign, type IPermission } from '@/common/auth-service';
import {
  AUTH_FIND_MAIN_ACCOUNT,
  AUTH_CREATE_CLOUD_SELECTION_SCHEME,
  AUTH_FIND_CLOUD_SELECTION_SCHEME,
  AUTH_FIND_IAAS_RESOURCE,
  AUTH_FIND_ACCOUNT,
  AUTH_BIZ_FIND_AUDIT,
  AUTH_FIND_RECYCLE_BIN,
} from '@/constants/auth-symbols';

const route = useRoute();
const { t } = useI18n();
const authStore = useAuthStore();

const permissionData = computed<IPermission | null>(() => (route.meta.permissionData as IPermission) ?? null);

const authType = computed(() => {
  const viewAuth = route.meta?.auth?.view;
  if (!viewAuth) return null;
  const authSign: IAuthSign = typeof viewAuth === 'function' ? viewAuth(route) : viewAuth;
  return authSign.type;
});

interface IDescriptionConfig {
  permTips: string[];
  featureDesc: string[];
}

const descriptions = new Map<symbol, IDescriptionConfig>([
  [
    AUTH_FIND_MAIN_ACCOUNT,
    {
      permTips: ['当前无"账号-二级账号查看"权限'],
      featureDesc: [
        '二级账号，是由公司和云厂商签订合同协议后，以公司为主体在云上申请独立的云账号，供业务使用。不同二级账号之间的资源是隔离的。',
      ],
    },
  ],
  [
    AUTH_CREATE_CLOUD_SELECTION_SCHEME,
    {
      permTips: ['当前无"资源选型-选型推荐"的权限'],
      featureDesc: [
        '资源选型，是根据业务需求，推荐出业务的部署地点，云资源方案的功能。当前页面访问受限，可到权限中心申请权限',
      ],
    },
  ],
  [
    AUTH_FIND_CLOUD_SELECTION_SCHEME,
    {
      permTips: ['当前无"部署方案-方案查看"的权限'],
      featureDesc: ['部署方案，是系统推荐出的，由用户保存的选型推荐方案。当前页面访问受限，可到权限中心申请权限'],
    },
  ],
  [
    AUTH_FIND_IAAS_RESOURCE,
    {
      permTips: [
        '该功能由平台资源的管理员维护，属于管理员的权限。',
        '如果您是业务方用户，无需申请该权限，请在"业务"菜单下直接使用。',
      ],
      featureDesc: [
        '资源管理功能，屏蔽了各种不同云之间的底层差异，提供了统一的管理模式，方便资源管理员统一全局的管理功能',
        '具备同时管理多云多账号的云资源，支持多种不同资源的操作',
        '提供资源的生命周期管理，如资源的创建，回收，销毁等',
        '支持资源归属不同业务',
      ],
    },
  ],
  [
    AUTH_FIND_ACCOUNT,
    {
      permTips: [
        '该功能用于管理云账号，如业务运维对云账号进行管理，可以对录入海垒的账号进行查看',
        '如果您是业务下云账号的资源使用者，无需申请该权限。对账号的数据查看，可以申请对应账号的"账号查看"权限，无需申请账号"录入权限"',
        '如果您只需要录入账号，请在业务-服务申请-云账号录入',
      ],
      featureDesc: [
        '资源账号：用于从云上同步、更新、操作、购买资源的账号，需要API密钥',
        '登记账号：云上的普通登录用户，用于被安全审计的账号对象',
        '安全审计账号：用于对云上资源进行安全审计的账号，需要API密钥，权限比资源账号低',
      ],
    },
  ],
  [
    AUTH_BIZ_FIND_AUDIT,
    {
      permTips: [
        '如果您是业务运维、SRE等角色，业务下管理了多个云账号，可申请"业务审计查看"权限，查看业务下多个账号的审计信息。',
        '如果您的账号属于某个业务，您负责其中一个账号，可申请"资源审计查看"权限，单独查看该账号的审计信息',
      ],
      featureDesc: [
        '审计信息包括包括账号信息，IaaS资源想增删改查等。有2种区别：业务操作审计，业务下的审计信息；资源操作审计，以账号为粒度的审计信息。',
      ],
    },
  ],
  [
    AUTH_FIND_RECYCLE_BIN,
    {
      permTips: [
        '该功能由平台资源的管理员维护，属于管理员的权限',
        '如果您是业务方用户，无需申请该权限，请在业务菜单主机、硬盘回收记录中查看回收信息',
      ],
      featureDesc: [
        '资源回收的管理功能，对业务回收的主机、硬盘资源进行管理，如对销毁、恢复操作',
        '资源恢复后，将恢复到从原回收的业务。',
        '资源立即销毁，将从云上直接删除资源，销毁属于不可逆操作，请谨慎操作。',
      ],
    },
  ],
]);

const currentDesc = computed(() => (authType.value ? descriptions.get(authType.value) : undefined));

const applyLoading = ref(false);
const applied = ref(false);

const handleApply = async () => {
  if (!permissionData.value) return;
  applyLoading.value = true;
  try {
    const url = await authStore.getApplyPermUrl(permissionData.value);
    applied.value = true;
    window.open(url);
  } finally {
    applyLoading.value = false;
  }
};

const handleRefresh = () => {
  window.location.reload();
};
</script>

<template>
  <bk-exception type="403">
    <template v-if="currentDesc" #description>
      <div class="permission-desc">
        <div v-if="currentDesc.permTips.length" class="desc-section">
          <h4>{{ t('权限申请说明') }}</h4>
          <p v-for="(tip, i) in currentDesc.permTips" :key="i">{{ t(tip) }}</p>
        </div>
        <div v-if="currentDesc.featureDesc.length" class="desc-section">
          <h4>{{ t('功能说明') }}</h4>
          <p v-for="(desc, i) in currentDesc.featureDesc" :key="i">{{ t(desc) }}</p>
        </div>
      </div>
    </template>

    <template v-if="permissionData">
      <bk-button v-if="!applied" theme="primary" :loading="applyLoading" @click="handleApply">
        {{ t('去申请') }}
      </bk-button>
      <bk-button v-else theme="primary" @click="handleRefresh">
        {{ t('已申请，刷新页面') }}
      </bk-button>
    </template>
  </bk-exception>
</template>

<style lang="scss" scoped>
.permission-desc {
  width: 700px;
  text-align: left;
}

.desc-section {
  margin-top: 20px;

  h4 {
    font-size: 14px;
    font-weight: 700;
    color: #313238;
    margin-bottom: 8px;
  }

  p {
    font-size: 12px;
    color: #63656e;
    line-height: 22px;
    padding-left: 16px;
  }
}
</style>
