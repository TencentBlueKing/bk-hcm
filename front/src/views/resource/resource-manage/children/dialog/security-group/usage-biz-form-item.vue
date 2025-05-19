<script setup lang="ts">
import { computed, watch, toRaw, ref, watchEffect, inject } from 'vue';
import { ArrowsRight } from 'bkui-vue/lib/icon';
import isEqual from 'lodash/isEqual';
import { type IBusinessItem } from '@/store/business-global';
import { useSecurityGroupStore, type ISecurityGroupItem } from '@/store/security-group';

const props = defineProps<{
  detail: ISecurityGroupItem;
  accountBizList: IBusinessItem[];
}>();

const model = defineModel<number[]>();

const securityGroupStore = useSecurityGroupStore();

const isBusinessEditUsageBiz = inject('isBusinessEditUsageBiz');

const beforeModel = structuredClone(toRaw(model.value || []));

const relBizIdList = ref([]);
watchEffect(async () => {
  const relBusiness = await securityGroupStore.queryRelBusiness(props.detail.id);
  const bizIds = Object.values(relBusiness)
    .reduce((acc, cur) => acc.concat(cur), [])
    .map((item) => item.bk_biz_id);
  relBizIdList.value = [...new Set(bizIds)];
});

const businessOptionDisabled = computed(() => {
  return (option: IBusinessItem) => {
    const isMgmtBiz = model.value?.includes?.(props.detail.mgmt_biz_id) && option.id === props.detail.mgmt_biz_id;
    const hasRel = relBizIdList.value.length && relBizIdList.value.includes(option.id);

    // 管理业务或存在关联实例的业务
    return isMgmtBiz || hasRel;
  };
});

const afterList = ref([]);
const hasChanged = ref(false);

watch(model, (val, oldVal) => {
  hasChanged.value = !isEqual(val, oldVal);

  afterList.value = [];

  // 删除的
  beforeModel.forEach((bizId) => {
    afterList.value.push({ bizId, remove: !val.includes(bizId) });
  });

  // 新增的
  const newAdd = val.filter((bizId) => !beforeModel.includes(bizId)).map((bizId) => ({ bizId, new: true }));
  afterList.value.push(...newAdd);
});
</script>

<template>
  <bk-form-item label="使用业务">
    <!-- 业务下编辑且为当前使用业务为全部业务时禁用 -->
    <hcm-form-business
      class="usage-biz-selector"
      multiple
      :data="accountBizList"
      v-model="model"
      :option-disabled="businessOptionDisabled"
      :show-all="accountBizList ? false : true"
      :all-option-id="-1"
      :disabled="isBusinessEditUsageBiz && detail.usage_biz_ids?.[0] === -1"
    />
  </bk-form-item>
  <div class="data-diff" v-show="hasChanged">
    <dl class="biz-list before">
      <dt class="list-title">变更前</dt>
      <dd class="list-item" v-for="bizId in beforeModel" :key="bizId">
        <span v-if="bizId === -1">全部业务</span>
        <display-value v-else :property="{ type: 'business' }" :value="bizId" />
      </dd>
    </dl>
    <arrows-right class="right-icon" />
    <div class="biz-list after">
      <dt class="list-title">变更后</dt>
      <dd v-for="item in afterList" :key="item.bizId" :class="['list-item', { remove: item.remove, new: item.new }]">
        <span v-if="item.bizId === -1">全部业务</span>
        <display-value
          v-else
          class="diff-name"
          :property="{ type: 'business' }"
          :display="{ showOverflowTooltip: true }"
          :value="item.bizId"
        />
        <span class="diff-tag" v-if="item.remove">删除</span>
        <span class="diff-tag" v-if="item.new">新增</span>
      </dd>
    </div>
  </div>
  <bk-alert theme="warning" title="变更后将导致部分业务无法使用当前安全组，请谨慎变更" />
</template>

<style lang="scss" scoped>
.data-diff {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
  position: relative;
  z-index: 0;

  .biz-list {
    width: 244px;
    height: 200px;
    border: 1px solid #dcdee5;
    border-radius: 2px;
    overflow-y: auto;

    .list-title {
      color: #4d4f56;
      font-weight: 700;
      height: 32px;
      line-height: 32px;
      padding: 0 16px;
      position: sticky;
      top: 0;
      z-index: 1;
    }

    .list-item {
      display: flex;
      align-items: center;
      gap: 4px;
      height: 32px;
      line-height: 32px;
      padding: 0 16px;
      color: #4d4f56;
      background: #fff;

      &:nth-child(odd) {
        background: #fafbfd;
      }

      .diff-name {
        width: 170px;
      }

      .diff-tag {
        font-size: 12px;
        padding: 2px 4px;
        border-radius: 2px;
        display: inline-block;
        line-height: normal;
        transform: scale(0.8) translateY(-1px);
      }

      &.remove {
        color: #e71818;

        .diff-name {
          text-decoration: line-through;
        }

        .diff-tag {
          background: #ffebeb;
        }
      }

      &.new {
        color: #2caf5e;

        .diff-tag {
          background: #daf6e5;
        }
      }
    }

    &.before {
      .list-title {
        background: #f0f1f5;
      }
    }

    &.after {
      .list-title {
        background: #fdf4e8;
      }
    }
  }

  .right-icon {
    font-size: 24px;
    color: #3a84ff;
  }
}

// hack 避免禁用的option会被tag的close去掉
.usage-biz-selector {
  :deep(.bk-select-trigger) {
    .bk-tag-closable {
      .bk-tag-close {
        display: none !important;
      }
    }
  }
}
</style>
