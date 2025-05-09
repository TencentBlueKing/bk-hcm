<script setup lang="ts">
import { ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import {
  ISecurityGroupDetail,
  useSecurityGroupStore,
  type SecurityGroupRelResourceByBizItem,
} from '@/store/security-group';
import { getPrivateIPs } from '@/utils';
import {
  RELATED_RES_KEY_MAP,
  RELATED_RES_NAME_MAP,
  SecurityGroupRelatedResourceName,
} from '@/constants/security-group';

import { Message } from 'bkui-vue';
import { ThemeEnum } from 'bkui-vue/lib/shared';
import hintIcon from '@/assets/image/hint.svg';
import dialogFooter from '@/components/common-dialog/dialog-footer.vue';

const props = defineProps<{
  row: SecurityGroupRelResourceByBizItem;
  detail: ISecurityGroupDetail;
  tabActive: SecurityGroupRelatedResourceName;
}>();
const model = defineModel<boolean>();
const emit = defineEmits(['success']);

const { t } = useI18n();
const securityGroupStore = useSecurityGroupStore();

const resName = RELATED_RES_NAME_MAP[props.tabActive];

const info = ref<SecurityGroupRelResourceByBizItem>(props.row);
watch(
  model,
  async () => {
    const res = await securityGroupStore.pullSecurityGroup(RELATED_RES_KEY_MAP[props.tabActive], [props.row]);
    [info.value] = res;
  },
  { immediate: true },
);

const handleConfirm = async () => {
  await securityGroupStore.batchDisassociateCvms({
    security_group_id: props.detail.id,
    cvm_ids: [props.row.id],
  });
  Message({ theme: 'success', message: t('解绑成功') });
  handleClosed();
  emit('success');
};
const handleClosed = () => {
  model.value = false;
};
</script>

<template>
  <bk-dialog class="unbind-dialog" v-model:is-show="model" dialog-type="show" @closed="handleClosed">
    <div class="hint-wrap">
      <img :src="hintIcon" />
      <div>{{ t('确认与该主机解绑') }}</div>
    </div>

    <bk-loading loading v-if="securityGroupStore.isBatchQuerySecurityGroupByResIdsLoading">
      <div style="width: 100%; height: 100px" />
    </bk-loading>

    <template v-else>
      <template v-if="info">
        <div class="mt16 mb16">
          <span>{{ t('内网 IP') }}：</span>
          <span>{{ getPrivateIPs(info) }}</span>
        </div>

        <div class="tips" v-if="info.security_groups">
          <template v-if="info.security_groups.length > 1">
            {{ t(`请确保${resName}上绑定的其他安全组是有效的，避免出现${resName}安全风险。`) }}
          </template>
          <template v-else>
            <span class="text-danger">{{ t('解绑被限制') }}</span>
            <span>
              {{
                t(
                  `，您的${resName}当前只绑定了${
                    info.security_groups.length ?? 0
                  }个安全组，为了确保您的${resName}安全，`,
                )
              }}
            </span>
            <span class="text-danger">{{ t('请至少保留1个以上的安全组，并确保安全组规则有效。') }}</span>
          </template>
        </div>

        <div class="operate-wrap">
          <dialog-footer
            :disabled="!info.security_groups || info.security_groups?.length <= 1"
            :loading="securityGroupStore.isBatchDisassociateCvmsLoading"
            :confirm-text="t('解绑')"
            :confirm-button-theme="ThemeEnum.DANGER"
            @confirm="handleConfirm"
            @closed="handleClosed"
          />
        </div>
      </template>
    </template>
  </bk-dialog>
</template>

<style scoped lang="scss">
.hint-wrap {
  margin-top: -35px;
  text-align: center;

  img {
    width: 42px;
    height: 42px;
  }

  div {
    font-size: 20px;
    color: #313238;
  }
}

.tips {
  padding: 12px 16px;
  background: #f5f6fa;
  border-radius: 2px;
}

.operate-wrap {
  margin-top: 16px;
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 8px;

  :deep(.bk-button) {
    min-width: 88px;
  }
}
</style>
