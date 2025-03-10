<script setup lang="ts">
import { computed, nextTick, ref, useAttrs, useTemplateRef, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useVerify } from '@/hooks';
import { Senarios, useWhereAmI } from '@/hooks/useWhereAmI';
import usePage from '@/hooks/use-page';
import {
  type ISecurityGroupDetail,
  type SecurityGroupRelatedResourceName,
  type SecurityGroupRelResourceByBizItem,
  useSecurityGroupStore,
} from '@/store/security-group';
import { RELATED_RES_KEY_MAP, RELATED_RES_PROPERTIES_MAP } from '@/constants/security-group';
import { ISearchSelectValue } from '@/typings';
import { enableCount, getSimpleConditionBySearchSelect, transformSimpleCondition } from '@/utils/search';
import { getPrivateIPs } from '@/utils';
import http from '@/http';

import search from '../search/index.vue';
import dataList from '../data-list/index.vue';
import dialogFooter from '@/components/common-dialog/dialog-footer.vue';

const props = defineProps<{
  textButton?: boolean;
  tabActive: SecurityGroupRelatedResourceName;
  detail: ISecurityGroupDetail;
}>();
const emit = defineEmits(['confirm']);
const attrs: any = useAttrs();

const { t } = useI18n();
const { getBusinessApiPath, whereAmI } = useWhereAmI();
const securityGroupStore = useSecurityGroupStore();

const { handleAuth, authVerifyData } = useVerify();
const authAction = computed(() => {
  return whereAmI.value === Senarios.business ? 'biz_iaas_resource_operate' : 'iaas_resource_operate';
});
const buttonCls = computed(() => {
  const buttonClsName = props.textButton ? 'hcm-no-permision-text-btn' : 'hcm-no-permision-btn';
  return { [buttonClsName]: !authVerifyData.value?.permissionAction?.[authAction.value] };
});

const isShow = ref(false);
const handleShow = () => {
  if (!authVerifyData.value?.permissionAction?.[authAction.value]) {
    handleAuth(authAction.value);
    return;
  }
  isShow.value = true;
};

const list = ref<SecurityGroupRelResourceByBizItem[]>([]);
const { pagination, getPageParams } = usePage();
const condition = ref<Record<string, any>>({});

const loading = ref(false);
const getList = async (sort = 'created_at', order = 'DESC') => {
  loading.value = true;
  try {
    const api = `/api/v1/cloud/${getBusinessApiPath()}${RELATED_RES_KEY_MAP[props.tabActive]}s/list`;
    const data = {
      filter: transformSimpleCondition(condition.value, RELATED_RES_PROPERTIES_MAP[props.tabActive]),
      page: getPageParams(pagination, { sort, order }),
    };

    // 查询资源列表
    const [listRes, countRes] = await Promise.all([
      http.post(api, enableCount(data, false)),
      http.post(api, enableCount(data, true)),
    ]);
    const [{ details = [] }, { count = 0 }] = [listRes?.data ?? {}, countRes?.data ?? {}];

    list.value = await securityGroupStore.pullSecurityGroup(RELATED_RES_KEY_MAP[props.tabActive], details);
    // 设置页码总条数
    pagination.count = count;
  } finally {
    loading.value = false;
  }
};

watch(isShow, (val) => {
  nextTick(() => {
    if (val) searchRef.value?.clear();
  });
});

const searchRef = useTemplateRef('bind-related-resource-search');
const handleSearch = (searchValue: ISearchSelectValue) => {
  // 搜索条件变更后，重置勾选
  handleClear();

  if (!searchValue.length) {
    condition.value = { account_id: props.detail.account_id, region: props.detail.region, vendor: props.detail.vendor };
  }
  condition.value = { ...condition.value, ...getSimpleConditionBySearchSelect(searchValue) };

  if (pagination.current === 1) {
    getList();
  } else {
    pagination.current = 1;
  }
};

watch([() => pagination.current, () => pagination.limit], () => {
  getList();
});

const dataListRef = useTemplateRef('data-list');
const selected = ref<SecurityGroupRelResourceByBizItem[]>([]);
const isToBindCvmsRowSelectEnable = ({ row, isCheckAll }: any) => {
  if (isCheckAll) return true;
  return !(row as SecurityGroupRelResourceByBizItem)?.security_groups
    ?.flatMap(({ cloud_id }) => cloud_id)
    ?.includes(props.detail.cloud_id);
};
const handleClear = () => {
  dataListRef.value.handleClear();
};
const handleDelete = (cloud_id: string) => {
  dataListRef.value.handleDelete(cloud_id);
};

const handleConfirm = () => {
  const ids = selected.value.map((item) => item.id);
  emit('confirm', ids);
};
const handleClosed = () => {
  isShow.value = false;
  handleClear();
};

defineExpose({ handleClosed });
</script>

<template>
  <bk-button theme="primary" :text="textButton" :class="buttonCls" @click="handleShow" v-bind="attrs">
    <slot name="icon"></slot>
    {{ t('新增绑定') }}
  </bk-button>
  <bk-dialog class="bind-dialog" v-model:is-show="isShow" :width="1500" :close-icon="false" @closed="handleClosed">
    <bk-resize-layout initial-divide="25%" placement="right" min="300" class="bind-dialog-content">
      <template #main>
        <div class="main">
          <bk-alert
            theme="warning"
            class="mb16"
            :title="
              t(
                '新绑定的安全组为最高优先级。如主机上已绑定的安全组为「安全组1」，新绑定的安全组为「安全组2」，则依次生效安全组顺序为：安全组2，安全组1。',
              )
            "
          />
          <search
            class="mb16"
            ref="bind-related-resource-search"
            :resource-name="tabActive"
            operation="bind"
            @search="handleSearch"
          />
          <data-list
            v-bkloading="{ loading }"
            ref="data-list"
            :list="list"
            :resource-name="tabActive"
            operation="bind"
            :pagination="pagination"
            :is-row-select-enable="isToBindCvmsRowSelectEnable"
            :has-settings="false"
            max-height="calc(100% - 100px)"
            @select="(selections) => (selected = selections)"
          />
        </div>
      </template>
      <template #aside>
        <div class="aside">
          <div class="title">{{ t('结果预览') }}</div>
          <div class="preview-wrap">
            <div class="tools">
              <span class="sub-title">{{ t('已选择主机') }}</span>
              <bk-tag theme="info" radius="8px" class="number">{{ selected.length }}</bk-tag>
              <span class="clear-btn" @click="handleClear">
                <i class="hcm-icon bkhcm-icon-clear mr2"></i>
                {{ t('清空') }}
              </span>
            </div>
            <div class="list-wrap">
              <div class="list-item" v-for="item in selected" :key="item.cloud_id">
                <span>{{ getPrivateIPs(item) }}</span>
                <i class="hcm-icon bkhcm-icon-close close-btn" @click="handleDelete(item.cloud_id)"></i>
              </div>
            </div>
          </div>
        </div>
      </template>
    </bk-resize-layout>

    <template #footer>
      <dialog-footer
        :disabled="!selected.length"
        :loading="securityGroupStore.isBatchAssociateCvmsLoading"
        @confirm="handleConfirm"
        @closed="handleClosed"
      />
    </template>
  </bk-dialog>
</template>

<style scoped lang="scss">
.bind-dialog {
  :deep(.bk-modal-header) {
    display: none;
  }
  :deep(.bk-dialog-content) {
    margin: 0;
    padding: 0;
  }
  .bind-dialog-content {
    max-height: 80vh;
  }
}

.main {
  padding: 24px;
  height: 100%;
}

.aside {
  height: 100%;
  background: #f5f6fa;

  .title {
    padding: 0 24px;
    height: 40px;
    line-height: 40px;
    background: #fff;
    font-size: 12px;
    font-weight: 700;
    color: #313238;
  }

  .preview-wrap {
    margin-top: 16px;
    padding: 0 24px;
    height: calc(100% - 56px);
    overflow: auto;

    .tools {
      margin-bottom: 8px;
      display: flex;
      align-items: center;
      color: #979ba5;
      font-size: 12px;

      .number {
        margin-left: 4px;
      }

      .clear-btn {
        margin-left: auto;
        cursor: pointer;
      }
    }

    .list-wrap {
      display: flex;
      flex-direction: column;
      gap: 4px;

      .list-item {
        padding: 0 12px;
        display: flex;
        align-items: center;
        justify-content: space-between;
        height: 32px;
        background: #ffffff;
        border-radius: 2px;
        font-size: 12px;
        color: #313238;

        &:hover {
          box-shadow: 0 2px 4px 0 #0000001a, 0 2px 4px 0 #1919290d;
        }
      }
      .close-btn {
        font-size: 14px;
        cursor: pointer;
      }
    }
  }
}
</style>
