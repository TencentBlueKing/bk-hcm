<script setup lang="ts">
// import { computed, PropType } from 'vue';
// const props = defineProps({
//   width: {
//     type: [Number, String] as PropType<number | string>,
//     default: '70%',
//   },
// });
// const defalutWidth = computed(() => props.width);
</script>
<template>
  <div class="content-container">
    <slot></slot>
  </div>
</template>

<style lang="scss" scoped>
.content-container {
  background: #fff;
  padding: 20px;
  padding-bottom: 52px;
  :deep(.action-bar) {
    padding-left: 150px;

    button {
      min-width: 86px;

      & + button {
        margin-left: 12px;
      }
    }
  }

  :deep(.form-content-list) {
    display: inline-flex;
    flex-direction: column;
    gap: 16px 0;

    &.details {
      gap: 0;

      .flex-row {
        border-bottom: 1px solid #c4c6cc;
      }
    }
  }

  :deep(.bk-form .bk-form-item) {
    .flex-row {
      display: flex;
      align-items: center;
      gap: 0 16px;
      .bk-form-item {
        display: flex;
        margin-bottom: 0;
      }
      .bk-form-label {
        flex: none;
        width: auto !important;
      }
      .bk-form-content {
        margin-left: 0 !important;
      }
    }

    .btns {
      display: flex;
      align-items: center;
      .btn {
        & + .btn {
          margin-left: 12px;
        }
      }
    }
  }

  :deep(.w110) {
    width: 110px;
  }
}

@media screen and (max-width: 1920px) {
  .content-container {
    width: 100%;
  }
}
</style>
