<script setup lang="ts">
import { ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { useResourceStore } from '@/store';
import { type ISecurityGroupOperateItem } from '@/store/security-group';

import { Message } from 'bkui-vue';
import DetailSimple from '../../components/security/detail-simple.vue';
import hintIcon from '@/assets/image/hint.svg';

defineOptions({ name: 'security-group-delete-dialog' });
const props = defineProps<{ detail: ISecurityGroupOperateItem; loading: boolean }>();
const model = defineModel<boolean>();

const { t } = useI18n();
const resourceStore = useResourceStore();

const filedIds = ['name', 'rule_count', 'rel_res', 'manager', 'bak_manager', 'mgmt_biz_id', 'usage_biz_ids'];

const isConfirmLoading = ref(false);
const handleDelete = async () => {
  isConfirmLoading.value = true;
  try {
    await resourceStore.deleteBatch('security_group', [props.detail.id]);
    Message({ theme: 'success', message: t('删除成功') });
  } finally {
    isConfirmLoading.value = false;
  }
};
const handleClosed = () => {
  model.value = false;
};
</script>

<template>
  <bk-dialog :is-show="model" @closed="handleClosed">
    <div class="hint-wrap">
      <img :src="hintIcon" />
      <div>{{ t('确认删除该安全组') }}</div>
    </div>

    <template v-if="loading">
      <bk-loading mode="spin" theme="primary" loading>
        <div style="width: 100%; height: 320px"></div>
      </bk-loading>
    </template>
    <template v-else>
      <bk-alert theme="danger" class="mt16 mb16">
        {{ t('当前安全组绑定了') }}
        <!-- todo: 遍历 resources 展示各资源数量 -->
        {{ t('无法删除。') }}
      </bk-alert>

      <detail-simple :field-ids="filedIds" :detail="detail" />
    </template>

    <template #footer>
      <div class="footer">
        <bk-button theme="primary" @click="handleDelete" :loading="isConfirmLoading">{{ t('删除') }}</bk-button>
        <bk-button :disabled="isConfirmLoading" @click="handleClosed">{{ t('取消') }}</bk-button>
      </div>
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

.detail-wrap {
  :deep(.rel-res-wrap) {
    display: flex;
    align-items: center;

    .number {
      padding: 0 8px;
      display: inline-block;
      height: 16px;
      line-height: 16px;
      background: #eaebf0;
      border-radius: 2px;
      font-size: 12px;
      color: #4d4f56;

      &.danger {
        background: #ea3636;
        color: #fff;
      }
    }

    & > span:not(:last-of-type)::after {
      content: ';';
      margin: 0 4px;
    }
  }
}

.footer {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 8px;
}
</style>
