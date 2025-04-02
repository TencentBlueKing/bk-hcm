<script setup lang="ts">
import { computed, ref, watchEffect } from 'vue';
import { useI18n } from 'vue-i18n';
import { type ISecurityGroupItem } from '@/store/security-group';
import { useBusinessGlobalStore } from '@/store/business-global';
import { useAccountStore } from '@/store/account';
import exclamationCircleShape from 'bkui-vue/lib/icon/exclamation-circle-shape';
import CopyToClipboard from '@/components/copy-to-clipboard/index.vue';
import unclaimedIcon from '@/assets/image/unclaimed.svg';

defineOptions({ name: 'security-group-unclaimed-comp' });

const props = defineProps<{ data: ISecurityGroupItem }>();

const { t } = useI18n();
const { getBusinessNames } = useBusinessGlobalStore();

const accountStore = useAccountStore();

const isAllBiz = computed(() => props.data.usage_biz_ids?.[0] === -1);
const accountManagers = ref<string[]>();

watchEffect(async () => {
  if (isAllBiz.value) {
    const accountDetailRes = await accountStore.getAccountDetail(props.data.account_id, { fromCache: true });
    accountManagers.value = accountDetailRes?.data?.managers ?? [];
  }
});

const tipsDesc = t(
  '该安全组当前处于未分配状态，不允许在业务下进行管理配置安全组规则等操作。如需管理规则，建议和相关负责人沟通后，确认管理业务后，再联系账号负责人分配到业务中进行管理。',
);
const mgmtBusinessName = computed(() => getBusinessNames(props.data?.mgmt_biz_id)?.[0] ?? '--');
const usageBusinessName = computed(() =>
  isAllBiz.value ? '全部业务' : getBusinessNames(props.data?.usage_biz_ids)?.join('、') ?? '--',
);
const usageBizInfosContent = computed(() =>
  props.data?.usage_biz_infos
    ?.map(({ bk_biz_name, bk_biz_maintainer }) => `${bk_biz_name}：${bk_biz_maintainer}`)
    ?.join('\n'),
);
</script>

<template>
  <div class="container">
    <span class="text">{{ mgmtBusinessName }}</span>
    <bk-popover width="480" placement="top" theme="light">
      <img class="unclaimed-icon" :src="unclaimedIcon" alt="unclaimed" />
      <template #content>
        <div class="tips-header">
          <exclamation-circle-shape fill="#EA3636" class="icon" />
          <span>{{ tipsDesc }}</span>
        </div>
        <div class="tips-info">
          <div class="info-item">
            <span class="label">{{ t('使用业务：') }}</span>
            <span>{{ usageBusinessName }}</span>
          </div>
          <div class="info-item" v-if="!isAllBiz">
            <span class="label">{{ t('业务负责人：') }}</span>
            <copy-to-clipboard class="copy" :content="usageBizInfosContent">
              <bk-button theme="primary" text>{{ t('一键复制') }}</bk-button>
            </copy-to-clipboard>
          </div>
          <div class="info-item">
            <span class="label">{{ t('账号负责人：') }}</span>
            <span v-if="!isAllBiz">{{ data?.account_managers?.join(',') }}</span>
            <span v-else>{{ accountManagers?.join(',') }}</span>
          </div>
        </div>
      </template>
    </bk-popover>
  </div>
</template>

<style scoped lang="scss">
.container {
  .text {
    margin-right: 4px;
    vertical-align: middle;
  }

  .unclaimed-icon {
    width: 24px;
    vertical-align: middle;
    cursor: pointer;
  }
}

.tips-header {
  margin-bottom: 8px;
  display: flex;
  align-items: flex-start;
  font-size: 12px;

  .icon {
    margin: 4px 6px 0 0;
  }
}

.tips-info {
  padding: 6px 16px;
  display: flex;
  flex-direction: column;
  gap: 4px;
  background: #f5f7fa;
  border-radius: 2px;

  .info-item {
    display: flex;
    align-items: flex-start;

    .label {
      width: 80px;
      text-align: right;
      flex-shrink: 0;
    }

    .copy {
      :deep(.bk-button-text) {
        line-height: inherit;
      }
    }
  }
}
</style>
