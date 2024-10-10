<script setup lang="ts">
import { OverflowTitle, Tag } from 'bkui-vue';
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';
import xss from 'xss';

interface INodeProps {
  displayValue: string;
  count: number | (() => number);
  handleMoreActionClick: (e: MouseEvent) => void;
  noCount?: boolean;
  showDefaultDomainTag?: boolean;
  highlightValue?: string;
}

const props = withDefaults(defineProps<INodeProps>(), {
  noCount: false,
  showDefaultDomainTag: false,
  highlightValue: '',
});
const emit = defineEmits(['click']);

const { t } = useI18n();

const highlightDisplayValue = computed(() => {
  const safeKeyword = xss(props.highlightValue);
  const regex = new RegExp(`(${safeKeyword})`, 'gi');
  const highlightedText = props.displayValue.replace(regex, '<span style="color: #3a84ff">$1</span>');

  return xss(highlightedText, {
    whiteList: {
      span: ['style'], // 允许 <span> 标签用于高亮
    },
  });
});
</script>

<template>
  <div class="node-wrapper" @click="emit('click')">
    <div class="info">
      <slot name="prefix-icon"></slot>
      <OverflowTitle type="tips" class="content">
        <!-- eslint-disable-next-line vue/no-v-html -->
        <span v-if="highlightValue" v-html="highlightDisplayValue"></span>
        <span v-else>{{ displayValue }}</span>
        <Tag v-if="showDefaultDomainTag" theme="warning" class="is-default-tag">{{ t('默认') }}</Tag>
      </OverflowTitle>
    </div>
    <div class="suffix">
      <div v-if="!noCount" class="count">{{ count }}</div>
      <div class="more-action" @click="handleMoreActionClick">
        <i class="hcm-icon bkhcm-icon-more-fill"></i>
      </div>
    </div>
  </div>
</template>

<style scoped lang="scss">
.node-wrapper {
  display: flex;
  align-items: center;
  height: 36px;
  cursor: pointer;

  .info {
    display: flex;
    align-items: center;
    flex: 1;
    max-width: calc(100% - 32px);

    .content {
      width: 100%;
      color: #313238;

      .is-default-tag {
        margin-left: 5px;
        padding: 4px;
        height: 16px;
        font-size: 12px;
      }
    }
  }

  .suffix {
    margin: 0 8px 0 auto;
    position: relative;
    display: flex;
    align-items: center;
    min-width: 24px;
    height: 36px;
    cursor: pointer;

    .count,
    .more-action {
      position: absolute;
      min-width: 24px;
      text-align: center;
    }

    .count {
      font-size: 12px;
      color: #c4c6cc;
    }

    .more-action {
      display: flex;
      align-items: center;
      justify-content: center;
      height: 24px;
      border-radius: 50%;
      opacity: 0;
    }
  }

  &:hover .suffix {
    .count {
      opacity: 0;
    }

    .more-action {
      opacity: 1;
      transition: background-color 0.2s;

      &:hover {
        background-color: #dcdee5;
      }
    }
  }

  .is-selected {
    background-color: #e1ecff !important;
  }
}
</style>
