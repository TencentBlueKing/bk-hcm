<script lang="ts" setup>
import { computed, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { useAuthStore, type IVerifyResourceInstance } from '@/store/auth';
import usePermissionDialog, { type PermissionDialogContext } from '@/hooks/use-permission-dialog';

export interface IPermApplyDialogProps {
  permission: PermissionDialogContext['permission'];
  done?: PermissionDialogContext['done'];
}

defineOptions({ name: 'permission-apply-dialog' });

const props = defineProps<IPermApplyDialogProps>();

const model = defineModel<boolean>();

const { t } = useI18n();

// 是否点击了去申请
const applied = ref(false);

const authStore = useAuthStore();
const permissionDialog = usePermissionDialog();

const list = computed(() => {
  const { actions, system_name } = props.permission;

  const list: { system: string; action: string; resources: IVerifyResourceInstance[] }[] = [];
  actions.forEach((action) => {
    // TODO: 支持多条
    // 暂只取第一条，实际是一个多条的结构形如A/B/C的多条记录
    const [firstResourceType] = action.related_resource_types;

    // TODO: 支持多层级
    // instances会有多个，可能重复，多个相同的资源鉴同一个权限的时候就会返回重复的实例
    // instances也是多层级的形如A-B/C-D多条记录，这里暂简化打平处理
    // 打平并且去复
    const resources = firstResourceType?.instances?.flat().reduce((acc, cur) => {
      const key = `${cur.type}_${cur.id}`;
      const exists = acc.some((item) => `${item.type}_${item.id}` === key);
      return exists ? acc : [...acc, cur];
    }, []);

    list.push({
      resources: resources ?? [],
      system: system_name,
      action: action.name,
    });
  });

  return list;
});

const handleApply = async () => {
  const url = await authStore.getApplyPermUrl(props.permission);
  applied.value = true;

  window.open(url);
};

const closeDialog = () => {
  model.value = false;
};

const handleDone = () => {
  closeDialog();
  if (props.done) {
    props.done?.();
  } else {
    window.location.reload();
  }
};

const handleClosed = () => {
  closeDialog();
};

const handleHidden = () => {
  // 重置状态
  setTimeout(() => {
    // 确保弹窗不可见了再重置
    applied.value = false;
  }, 500);
};

defineExpose({ show: permissionDialog.show });
</script>

<template>
  <bk-dialog
    class="permission-dialog"
    :quick-close="false"
    :is-show="model"
    :width="740"
    :is-loading="authStore.applyPermUrlLoading"
    @confirm="handleApply"
    @closed="handleClosed"
    @hidden="handleHidden"
  >
    <div class="permission-header">
      <bk-exception type="403" scene="part" :title="t('没有权限访问或操作此资源')"></bk-exception>
    </div>
    <bk-table class="permission-table" :data="list" :max-height="190">
      <bk-table-column :label="t('系统')" :width="150" prop="system" show-overflow-tooltip></bk-table-column>
      <bk-table-column :label="t('需要申请的权限')" :width="200" prop="action" show-overflow-tooltip></bk-table-column>
      <bk-table-column :label="t('关联的资源实例')" :width="342" prop="resources">
        <template #default="{ row }">
          <template v-if="row.resources?.length">
            <div class="resource-item" v-for="instance in row.resources" :key="instance.id">
              【{{ instance.type_name }}】{{ instance.name }}
            </div>
          </template>
          <span v-else>--</span>
        </template>
      </bk-table-column>
    </bk-table>
    <template #footer>
      <div class="permission-footer">
        <bk-button v-if="!applied" theme="primary" :loading="authStore.applyPermUrlLoading" @click="handleApply">
          {{ t('去申请') }}
        </bk-button>
        <bk-button v-else theme="primary" :loading="authStore.applyPermUrlLoading" @click="handleDone">
          {{ t('我已申请') }}
        </bk-button>
        <bk-button @click="closeDialog">{{ !applied ? t('取消') : t('关闭') }}</bk-button>
      </div>
    </template>
  </bk-dialog>
</template>

<style lang="scss" scoped>
.permission-dialog {
  :deep(.bk-modal-header) {
    .bk-dialog-header {
      padding: 0;
    }
  }
  :deep(.bk-modal-content) {
    .bk-dialog-content {
      margin-top: 0;
    }
  }
}
.permission-header {
  :deep(.bk-exception) {
    .bk-exception-img {
      height: 150px;
    }
    .bk-exception-title {
      font-size: 22px;
      color: #63656e;
      line-height: normal;
      margin: 6px 0 30px;
    }
  }
}
.permission-table {
  .resource-item {
    line-height: 24px;
  }
}
.permission-footer {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 10px;
}
</style>
