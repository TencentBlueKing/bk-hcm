<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount, nextTick } from 'vue';

defineSlots<{
  default(): any;
  footer(): any;
}>();

const containerRef = ref<HTMLElement>();
const footerRef = ref<HTMLElement>();
const isSticky = ref(false);

let observer: ResizeObserver | null = null;

const checkSticky = () => {
  if (!containerRef.value || !footerRef.value) return;
  const containerHeight = containerRef.value?.clientHeight;
  const scrollHeight = containerRef.value?.scrollHeight;
  isSticky.value = scrollHeight > containerHeight;
};

onMounted(() => {
  nextTick(checkSticky);
  observer = new ResizeObserver(checkSticky);
  if (containerRef.value) observer.observe(containerRef.value);
});

onBeforeUnmount(() => {
  observer?.disconnect();
});
</script>

<template>
  <div ref="containerRef" class="sticky-bottom-container" @scroll="checkSticky">
    <div class="sticky-bottom-body">
      <slot />
    </div>
    <div ref="footerRef" class="sticky-bottom-bar" :class="{ 'is-sticky': isSticky }">
      <slot name="footer" />
    </div>
  </div>
</template>

<style lang="scss" scoped>
.sticky-bottom-container {
  height: 100%;
  overflow-y: auto;
}

.sticky-bottom-body {
  padding: 24px;

  :deep(.common-card:last-child) {
    margin-bottom: 0;
  }
}

.sticky-bottom-bar {
  position: sticky;
  bottom: 0;
  display: flex;
  align-items: center;
  min-height: 52px;
  padding: 0 24px;

  &.is-sticky {
    background-color: #fafbfd;
    border-top: 1px solid #dcdee5;
  }

  :deep(.bk-button) {
    min-width: 88px;
  }

  :deep(.bk-button + .bk-button) {
    margin-left: 8px;
  }
}
</style>
