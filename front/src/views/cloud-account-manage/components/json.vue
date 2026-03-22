<script setup lang="ts">
import { computed } from 'vue';
import { Message } from 'bkui-vue';
import { isJSON } from '@/utils';
import useClipboard from 'vue-clipboard3';

interface IProps {
  content?: string;
  copy?: boolean;
}

const props = withDefaults(defineProps<IProps>(), {
  content:
    '{"version":"2.0","statement":[{"effect":"allow","action":["cvm:Describe*","cvm:Query*"],"resource":"*"},{"effect":"allow","action":["cbs:Describe*"],"resource":"*"},{"effect":"allow","action":["vpc:Describe*","vpc:Query*"],"resource":"*"}]}',
  copy: true,
});
const { toClipboard } = useClipboard();

const content = computed(() => {
  const { content } = props;
  if (isJSON(content)) return JSON.stringify(JSON.parse(content), null, 2);
  return content;
});

const handleCopy = async () => {
  try {
    await toClipboard(content.value);
    Message({ theme: 'success', message: '复制成功' });
  } catch (e) {
    Message({ theme: 'error', message: '复制失败' });
  }
};
</script>

<template>
  <div class="json-code">
    {{ content }}
    <i
      class="hcm-icon bkhcm-icon-copy json-copy"
      color="#3A84FF"
      v-if="props.copy"
      title="复制"
      @click="handleCopy"
    ></i>
  </div>
</template>

<style lang="scss" scoped>
.json-code {
  white-space: pre;
  position: relative;
  font-size: 12px;
  color: #000000;
  font-weight: 400;
  line-height: 20px;
}
.json-copy {
  color: hsl(217, 100%, 61%);
  cursor: pointer;
  position: absolute;
  right: 0;
  top: 0;
  font-size: 14px;
}
</style>
