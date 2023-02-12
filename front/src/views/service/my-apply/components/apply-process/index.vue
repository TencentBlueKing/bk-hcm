<template>
  <div class="iam-approval-process-wrapper" @click.stop="handleOpenProcessLink">
    <i class="icon hcm-icon bkhcm-icon-angle-up-fill"></i>
    {{ defaultText }}
  </div>
</template>

<script lang="ts">
import { defineComponent } from 'vue';
import { useI18n } from 'vue-i18n';
export default defineComponent({
  name: 'ApplyProcess',
  props: {
    link: {
      type: String,
    },
  },
  setup(props) {
    const { t } = useI18n();
    const defaultText = t('查看审批详情');
    const handleOpenProcessLink = () => {
      window.open(props.link);
    };

    return {
      defaultText,
      handleOpenProcessLink,
    };
  },
});
</script>
<style lang="scss" scoped>
.iam-approval-process-wrapper {
  margin: 20px 0;
  padding: 0 20px;
  line-height: 64px;
  background: #fff;
  border-radius: 2px;
  box-shadow: 0px 1px 2px 0px rgba(0, 0, 0, 0.05);
  font-size: 14px;
  color: #3a84ff;
  cursor: pointer;
  &:hover {
    color: #699df4;
  }
}
</style>
