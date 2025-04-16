<script lang="ts" setup>
import { reactive, ref, watch, watchEffect } from 'vue';
import { useI18n } from 'vue-i18n';
import { VueDraggable } from 'vue-draggable-plus';
import CombineRequest from '@blueking/combine-request';
import { ModelPropertySearch } from '@/model/typings';
import { SECURITY_GROUP_RULE_TYPE, VendorEnum } from '@/common/constant';
import { useSecurityGroupStore, type ISecurityGroupItem, type ISecurityGroupRuleItem } from '@/store/security-group';
import { transformSimpleCondition } from '@/utils/search';
import useColumns from '@/views/resource/resource-manage/hooks/use-columns';

export interface ISecurityGroupSelectorDialogProps {
  title: string;
  accountId: string;
  vendor: string;
  region: string;
  loading: boolean;
  checked?: string[];
  bizId?: number;
  vpcId?: string;
  multiple?: boolean;
  sortOnly?: boolean;
  idKey?: string;
}

export type RuleGroup = Pick<ISecurityGroupItem, 'id' | 'name'>[];

const props = withDefaults(defineProps<ISecurityGroupSelectorDialogProps>(), {
  checked: () => [],
  multiple: true,
  sortOnly: false,
  title: '选择安全组',
  idKey: 'id',
});

const emit = defineEmits<{
  confirm: [sgIds: string[], sg: ISecurityGroupItem[]];
  closed: [];
}>();

const { t } = useI18n();

const securityGroupStore = useSecurityGroupStore();

const model = defineModel<boolean>();

const localChecked = ref<string[]>(structuredClone(props.checked));

const selectedRuleType = ref<SECURITY_GROUP_RULE_TYPE>(SECURITY_GROUP_RULE_TYPE.INGRESS);

const ruleCollapseActiveIds = ref([]);
const isRuleCollapseAllExpanded = ref(false);

const securityGroupList = ref<ISecurityGroupItem[]>([]);
const displaySecurityGroupList = ref<ISecurityGroupItem[]>([]);

const ruleGroup = ref<RuleGroup>([]);

const ruleLoading = reactive<Record<string, boolean>>({});

const ruleColumns = useColumns('securityCommon', false, props.vendor).columns.filter(
  ({ field }: { field: string }) => !['updated_at'].includes(field),
);

watch(
  props.checked,
  (newVal) => {
    localChecked.value = newVal;
  },
  { deep: true },
);

watch(
  [localChecked, securityGroupList],
  ([currentChecked, currentList]) => {
    const group: RuleGroup = [];
    currentChecked.forEach((sgId) => {
      const sg = currentList.find((sg) => sg.id === sgId);
      if (sg) {
        group.push({
          id: sg.id,
          name: sg.name,
        });
      }
    });

    ruleGroup.value = group;
  },
  { deep: true, immediate: true },
);

watch(ruleCollapseActiveIds, async (ids) => {
  const currentIds = ids.flat();
  const newIds = currentIds.filter((id) => ![...ruleListCacheMap.value.keys()].includes(id));

  // ids每次都会是当前展开态的全部id，没有请求过的才继续获取
  if (!newIds.length) {
    return;
  }

  const combineRequest = new CombineRequest<ISecurityGroupRuleItem[]>(async (data) => {
    const [sgId] = data;
    return securityGroupStore.getFullRuleList({
      filter: { op: 'and', rules: [] },
      vendor: props.vendor,
      id: sgId as string,
    });
  }, 1);

  newIds.forEach((id) => {
    combineRequest.add(id);
  });

  let i = 0;
  for (const result of await combineRequest.getSplitPromise()) {
    const id = newIds[i];
    ruleLoading[id] = true;
    ruleListCacheMap.value.set(id, await result);
    ruleLoading[id] = false;
    i = i + 1;
  }
});

watchEffect(async () => {
  const properties: ModelPropertySearch[] = [
    { id: 'account_id', type: 'string', name: '账号' },
    { id: 'region', type: 'string', name: '云地域' },
    { id: 'extension.vpc_id', type: 'string', name: 'VPC' },
  ];

  const conditions = {
    account_id: props.accountId,
    region: props.region,
    'extension.vpc_id': props.vpcId,
  };

  const params = { filter: transformSimpleCondition(conditions, properties) };

  const list = await securityGroupStore.getFullList(params);
  securityGroupList.value = list;
  displaySecurityGroupList.value = list;
});

const ruleListCacheMap = ref<Map<string, ISecurityGroupRuleItem[]>>(new Map());

const closeDialog = () => {
  model.value = false;
  emit('closed');
};

const handleCheck = (isChecked: boolean, item: ISecurityGroupItem) => {
  if (isChecked) {
    localChecked.value.unshift(item.id);
  } else {
    const index = localChecked.value.indexOf(item.id);
    if (index !== -1) {
      localChecked.value.splice(index, 1);
    }
  }
};

const handleExpandToggle = () => {
  isRuleCollapseAllExpanded.value = !isRuleCollapseAllExpanded.value;
  if (isRuleCollapseAllExpanded.value) {
    ruleCollapseActiveIds.value = ruleGroup.value.map((item) => item.id);
  } else {
    ruleCollapseActiveIds.value = [ruleGroup.value[0].id];
  }
};

const handleClosed = () => {
  closeDialog();
};

const handleSearchInput = (value: string) => {
  if (value?.length) {
    displaySecurityGroupList.value = securityGroupList.value.filter((item) => item.name.includes(value));
  } else {
    displaySecurityGroupList.value = securityGroupList.value;
  }
};

const handleDialogConfirm = () => {
  const checkedIds = ruleGroup.value.map((item) => item.id);
  emit(
    'confirm',
    checkedIds,
    securityGroupList.value.filter((item) => checkedIds.includes(item.id)),
  );
};
</script>
<template>
  <bk-dialog :title="title" width="60vw" :quick-close="false" :is-show="model" @closed="handleClosed">
    <div class="security-group-selector">
      <div class="selector-aside" v-bkloading="{ loading: securityGroupStore.isFullListLoading }">
        <div class="searchbar">
          <bk-input placeholder="搜索安全组" type="search" clearable @input="handleSearchInput" />
        </div>
        <div class="security-group-list g-scroller">
          <div
            :class="['security-group-item', { 'cross-business': item.usage_biz_ids?.length > 1 }]"
            v-for="item in displaySecurityGroupList"
            :key="item.id"
          >
            <bk-checkbox
              :model-value="localChecked.includes(item.id)"
              :disabled="
                sortOnly ||
                checked.includes(item.id) ||
                (localChecked.length > 0 && (vendor === VendorEnum.AZURE || !multiple))
              "
              @change="(isChecked: boolean) => handleCheck(isChecked, item)"
            >
              <span :title="item.name">{{ item.name }}</span>
            </bk-checkbox>
            <bk-tag theme="success" radius="10px" type="filled" size="small" v-if="item.usage_biz_ids?.length > 1">
              {{ t('跨业务') }}
            </bk-tag>
          </div>
          <bk-exception
            type="empty"
            scene="part"
            :description="t('没有数据')"
            v-show="!displaySecurityGroupList?.length"
          />
        </div>
      </div>
      <div class="selector-main">
        <div class="rule-toolbar">
          <bk-button-group>
            <bk-button
              :selected="selectedRuleType === SECURITY_GROUP_RULE_TYPE.EGRESS"
              @click="selectedRuleType = SECURITY_GROUP_RULE_TYPE.EGRESS"
            >
              {{ t('出站规则') }}
            </bk-button>
            <bk-button
              :selected="selectedRuleType === SECURITY_GROUP_RULE_TYPE.INGRESS"
              @click="selectedRuleType = SECURITY_GROUP_RULE_TYPE.INGRESS"
            >
              {{ t('入站规则') }}
            </bk-button>
          </bk-button-group>
          <bk-button :disabled="!ruleGroup.length" @click="handleExpandToggle">
            <template v-if="isRuleCollapseAllExpanded">
              <i class="hcm-icon bkhcm-icon-zoomout"></i>
              <span class="ml8">{{ t('全部收起') }}</span>
            </template>
            <template v-else>
              <i class="hcm-icon bkhcm-icon-fullscreen"></i>
              <span class="ml8">{{ t('全部展开') }}</span>
            </template>
          </bk-button>
        </div>
        <div class="rule-group g-scroller">
          <bk-collapse v-model="ruleCollapseActiveIds" use-block-theme>
            <vue-draggable v-model="ruleGroup" handle=".draggable-anchor">
              <bk-collapse-panel class="rule-panel" v-for="(item, index) in ruleGroup" :key="item.id" :name="item.id">
                <div class="panel-title">
                  <span class="rule-name">{{ item.name }}</span>
                  <div class="rule-draggable">
                    <span class="rule-index">{{ index + 1 }}</span>
                    <i class="hcm-icon bkhcm-icon-grag-fill draggable-anchor"></i>
                  </div>
                </div>
                <template #content>
                  <bk-table
                    :data="ruleListCacheMap.get(item.id)?.filter((rule) => rule.type === selectedRuleType)"
                    :columns="ruleColumns"
                    :max-height="300"
                    row-hover="auto"
                    :stripe="true"
                    show-overflow-tooltip
                    v-bkloading="{ loading: ruleLoading[item.id], size: 'small' }"
                  />
                </template>
              </bk-collapse-panel>
            </vue-draggable>
          </bk-collapse>
        </div>
      </div>
    </div>
    <template #footer>
      <div class="dialog-custom-footer">
        <bk-button theme="primary" :disabled="!localChecked.length" :loading="loading" @click="handleDialogConfirm">
          确定
        </bk-button>
        <bk-button @click="closeDialog">取消</bk-button>
      </div>
    </template>
  </bk-dialog>
</template>

<style lang="scss" scoped>
.security-group-selector {
  display: flex;
  gap: 16px;
  min-height: 300px;
  max-height: 60vh;

  .selector-aside {
    width: 300px;
    background: #fafbfd;
    padding: 12px 0;
  }

  .selector-main {
    flex: 1;
  }
}

.searchbar {
  padding: 0 16px;
}

.security-group-list {
  height: calc(100% - 32px);
  margin-top: 8px;

  .security-group-item {
    display: flex;
    gap: 6px;
    margin: 4px 0;
    padding: 6px 16px;

    :deep(.bk-checkbox) {
      max-width: 100%;

      .bk-checkbox-label {
        flex: 1;
        text-overflow: ellipsis;
        white-space: nowrap;
        overflow: hidden;
      }
    }

    &.cross-business {
      :deep(.bk-checkbox) {
        max-width: calc(100% - 40px);
      }
    }

    &:hover {
      background: #f0f1f5;
    }
  }
}

.rule-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
}

.rule-group {
  height: calc(100% - 48px);

  :deep(.bk-collapse-wrapper) {
    .bk-collapse-content {
      padding: 0;
    }
  }

  .rule-panel {
    .panel-title {
      display: inline-flex;
      width: calc(100% - 32px);

      .rule-draggable {
        margin-left: auto;
      }

      .rule-index {
        color: #979ba5;
        padding: 0 6px;
        background: #fff;
        border-radius: 2px;
        margin-right: 8px;
      }

      .draggable-anchor {
        cursor: grabbing;
        padding: 4px 8px;
      }
    }
  }
}

.dialog-custom-footer {
  display: flex;
  justify-content: flex-end;
  gap: 8px;

  :deep(.bk-button) {
    min-width: 88px;
  }
}
</style>
