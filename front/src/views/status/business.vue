<script setup lang="ts">
import { computed, ref } from 'vue';
import { useRoute } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { useAuthStore } from '@/store/auth';
import type { IPermission } from '@/common/auth-service';
import type { BusinessStatus } from '@/router/business-interceptor';

const route = useRoute();
const { t } = useI18n();
const authStore = useAuthStore();

const businessStatus = computed<BusinessStatus>(() => (route.meta.extra as any)?.businessStatus ?? 'noBiz');

const permissionData = computed<IPermission | null>(() => (route.meta.permissionData as IPermission) ?? null);

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
  <!-- 没有任何可用业务 -->
  <bk-exception
    v-if="businessStatus === 'noBiz'"
    type="403"
    :title="t('无任何业务权限')"
    :description="t('你没有任何业务的访问权限，请先申请业务权限后再访问')"
  />

  <!-- 业务不存在 -->
  <bk-exception
    v-else-if="businessStatus === 'bizNotFound'"
    type="404"
    :title="t('业务不存在')"
    :description="t('你访问的业务不存在，请确认后重试')"
  />

  <!-- 无当前业务权限 -->
  <bk-exception
    v-else-if="businessStatus === 'bizUnauthed'"
    type="403"
    :title="t('无当前业务权限')"
    :description="t('你没有当前业务的访问权限，可以申请权限后再访问')"
  >
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
