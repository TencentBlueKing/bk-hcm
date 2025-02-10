<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import { ISgOperateItem } from '@/store/security-group';

import DetailDisplay from '../children/detail-display.vue';
import hintIcon from '@/assets/image/hint.svg';

defineProps<{ detail: ISgOperateItem; loading: boolean }>();
const emit = defineEmits(['confirm']);
const model = defineModel<boolean>();

const { t } = useI18n();

const fieldIds = ['name', 'rule_count', 'rel_res', 'usage_biz_ids'];

const handleClosed = () => {
  model.value = false;
};
</script>

<template>
  <bk-dialog :is-show="model" footer-align="center" @confirm="emit('confirm')" @closed="handleClosed">
    <div class="hint-wrap">
      <img :src="hintIcon" />
      <div>{{ t('变更影响确认') }}</div>
    </div>

    <template v-if="loading">
      <bk-loading mode="spin" theme="primary" loading>
        <div style="width: 100%; height: 160px"></div>
      </bk-loading>
    </template>
    <template v-else>
      <detail-display class="mt16 mb16" :field-ids="fieldIds" :detail="detail" />

      <bk-alert theme="danger" :show-icon="false">
        {{
          t(
            '该安全组跨多业务使用，变更会影响多个业务的服务，建议将当前安全组克隆后绑定到相关实例，避免跨业务操作影响其他业务。',
          )
        }}
      </bk-alert>
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
</style>
