<script setup lang="ts">
import { ref, watch } from 'vue';
import { Share } from 'bkui-vue/lib/icon';
import { DisplayType } from '../typings';

export interface LinkPopoverItem {
  id: string | number;
  label: string;
}

defineOptions({ inheritAttrs: false });

const props = withDefaults(
  defineProps<{
    value: string | number | string[] | number[];
    displayValue: string | number | string[] | number[];
    displayOn: DisplayType['on'];
    /** 弹出框触发方式，默认 click */
    trigger?: 'click' | 'hover';
    /** 弹出框方向，默认 right */
    placement?: string;
    /** 异步加载列表数据的函数 */
    loadFn?: () => Promise<LinkPopoverItem[]>;
    /** 直接传入列表数据（与 loadFn 二选一） */
    list?: LinkPopoverItem[];
    /** 空状态文案，默认"未查询到数据" */
    emptyText?: string;
    /** 弹出框宽度，默认 162px */
    popoverWidth?: number;
    /** 弹出框最大高度，默认 140px */
    popoverMaxHeight?: number;
    showLoading?: boolean;
  }>(),
  {
    trigger: 'click',
    placement: 'right',
    emptyText: '未查询到数据',
    popoverWidth: 162,
    popoverMaxHeight: 140,
    showLoading: true,
  },
);

const emit = defineEmits<{
  (e: 'linkClick', item: LinkPopoverItem): void;
}>();

const loading = ref(false);
const renderList = ref<LinkPopoverItem[]>([]);
const isLoaded = ref(false);

const loadData = async () => {
  if (props.loadFn && !isLoaded.value) {
    loading.value = true;
    try {
      renderList.value = await props.loadFn();
      isLoaded.value = true;
    } finally {
      loading.value = false;
    }
  }
};

const handleAfterShow = () => {
  loadData();
};

// 当直接传入 list 时同步
watch(
  () => props.list,
  (val) => {
    if (val) {
      renderList.value = val;
      isLoaded.value = true;
    }
  },
  { immediate: true },
);

const handleLinkClick = (item: LinkPopoverItem) => {
  emit('linkClick', item);
};
</script>

<template>
  <bk-popover
    theme="light"
    :component-event-delay="300"
    :trigger="trigger"
    render-type="shown"
    :placement="placement"
    :popover-delay="trigger === 'click' ? [300, 0] : [200, 150]"
    @after-show="handleAfterShow"
  >
    <bk-button theme="primary" text>{{ displayValue || '--' }}</bk-button>
    <template #content>
      <bk-loading v-if="showLoading && loading" theme="primary" mode="spin" size="mini" :opacity="1" />
      <ul
        v-else-if="renderList.length"
        class="link-popover-list"
        :style="{ width: `${popoverWidth}px`, maxHeight: `${popoverMaxHeight}px` }"
      >
        <li v-for="item in renderList" :key="item.id" class="link-popover-item">
          <slot name="item-label" :item="item">
            <span class="link-popover-label" v-bk-tooltips="{ content: item.label }">{{ item.label }}</span>
          </slot>
          <span class="link-popover-trigger" @click="handleLinkClick(item)">
            <slot name="item-icon">
              <Share class="link-popover-icon" />
            </slot>
          </span>
        </li>
      </ul>
      <div v-else class="link-popover-empty">{{ emptyText }}</div>
    </template>
  </bk-popover>
</template>

<style lang="scss" scoped>
.link-popover-list {
  margin: 0;
  padding: 0;
  list-style: none;
  overflow-y: auto;

  .link-popover-item {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 4px 8px;
    line-height: 28px;
    cursor: pointer;

    &:nth-child(even) {
      background: #fafbfd;
    }

    &:hover {
      background: #f0f1f5;
    }

    .link-popover-label {
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
    }

    .link-popover-icon {
      flex-shrink: 0;
      font-size: 12px;
      color: #3a84ff;
      cursor: pointer;
    }
  }
}

.link-popover-empty {
  padding: 8px 0;
  color: #c4c6cc;
  text-align: center;
}
</style>
