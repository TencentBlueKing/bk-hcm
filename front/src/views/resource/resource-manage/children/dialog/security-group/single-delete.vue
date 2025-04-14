<script setup lang="ts">
import { computed, h, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { type ISecurityGroupOperateItem } from '@/store/security-group';
import { TagThemeEnum } from 'bkui-vue/lib/shared';
import GridContainer from '@/components/layout/grid-container/grid-container.vue';
import GridItem from '@/components/layout/grid-container/grid-item.vue';
import RelResourcesDisplay from '../../components/security/rel-resources-display.vue';
import hintIcon from '@/assets/image/hint.svg';
import DeleteButton from './single-delete-button.plugin.vue';

defineOptions({ name: 'security-group-delete-dialog' });
const props = defineProps<{ detail: ISecurityGroupOperateItem; loading: boolean }>();
const model = defineModel<boolean>();
const emit = defineEmits(['success']);

const { t } = useI18n();

const boundResources = computed(() => props.detail?.resources?.filter(({ count }) => count > 0) ?? []);

const fields = [
  { id: 'name', name: t('安全组名称'), type: 'string' },
  { id: 'rule_count', name: t('规则数'), render: () => `${props.detail?.rule_count} ${t('个')}` },
  {
    id: 'rel_res',
    name: t('绑定实例'),
    render: () => {
      const { resources = [] } = props.detail ?? {};
      const displayResources = resources.filter(({ count }) => count > 0);

      if (!displayResources.length) return '--';
      return h(RelResourcesDisplay, { resources: displayResources, tagTheme: TagThemeEnum.DANGER });
    },
  },
  { id: 'manager', name: t('主负责人'), type: 'user' },
  { id: 'bak_manager', name: t('备份负责人'), type: 'user' },
  { id: 'mgmt_biz_id', name: t('管理业务'), type: 'business' },
  { id: 'usage_biz_ids', name: t('使用业务'), type: 'business' },
];

const isConfirmLoading = ref(false);
const handleDeleteSuccess = async () => {
  handleClosed();
  emit('success');
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
      <bk-alert v-if="boundResources.length > 0" theme="danger" class="mt16 mb16">
        {{ t('当前安全组绑定了') }}
        {{ boundResources.map(({ res_name: resName, count }) => `${count} 个${resName}资源`).join('，') }}
        {{ t('，无法删除。') }}
      </bk-alert>

      <grid-container class="detail-preview-wrap" :column="1" :label-width="120">
        <grid-item v-for="field in fields" :key="field.id" :label="field.name">
          <template v-if="field.render">
            <component :is="field.render" />
          </template>
          <display-value v-else :property="field" :value="detail?.[field.id]" :display="{ on: 'info' }" />
        </grid-item>
      </grid-container>
    </template>

    <template #footer>
      <div class="footer">
        <delete-button
          :id="detail?.id"
          :disabled="boundResources.length > 0"
          v-model:loading="isConfirmLoading"
          @success="handleDeleteSuccess"
        />
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

.detail-preview-wrap {
  :deep(.rel-res-wrap) {
    display: flex;
    align-items: center;
    flex-wrap: wrap;
    row-gap: 4px;

    .rel-res-item {
      display: flex;
      align-items: center;

      .number {
        padding: 0 4px;
        font-size: 12px;
      }

      &:not(:last-of-type)::after {
        content: '|';
        margin: 0 4px;
      }
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
