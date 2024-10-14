<script setup lang="ts">
import { OverflowTitle, Tag } from 'bkui-vue';
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';

interface INodeProps {
  displayValue: string;
  count: number | (() => number);
  handleMoreActionClick: (e: MouseEvent) => void;
  noCount?: boolean;
  showDefaultDomainTag?: boolean;
  searchValue?: string;
}

const props = withDefaults(defineProps<INodeProps>(), {
  noCount: false,
  showDefaultDomainTag: false,
  searchValue: '',
});
const emit = defineEmits(['click']);

const { t } = useI18n();

const highlightedParts = computed(() => {
  const text = props.displayValue;
  const keyword = props.searchValue;
  if (!keyword) return [{ text, highlight: false }];
  const parts = [];
  const regex = new RegExp(`(${keyword})`, 'gi');
  let lastIndex = 0;
  text.replace(regex, (match, p1, offset) => {
    if (offset > lastIndex) {
      parts.push({ text: text.slice(lastIndex, offset), highlight: false });
    }
    parts.push({ text: match, highlight: true });
    lastIndex = offset + match.length;
    return p1;
  });
  if (lastIndex < text.length) {
    parts.push({ text: text.slice(lastIndex), highlight: false });
  }
  return parts;
});
</script>

<template>
  <div class="node-wrapper" @click="emit('click')">
    <div class="info">
      <slot name="prefix-icon"></slot>
      <OverflowTitle type="tips" class="content">
        <template v-for="(part, index) in highlightedParts" :key="index">
          <span v-if="part.highlight" class="highlight-keyword">{{ part.text }}</span>
          <template v-else>{{ part.text }}</template>
        </template>
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

      .highlight-keyword {
        color: #3a84ff;
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
