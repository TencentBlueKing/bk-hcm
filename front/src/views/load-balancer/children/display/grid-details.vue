<script setup lang="ts">
import { useAttrs } from 'vue';
import { ModelPropertyDisplay } from '@/model/typings';
import { get } from 'lodash';

import GridContainer from '@/components/layout/grid-container/grid-container.vue';
import GridItem from '@/components/layout/grid-container/grid-item.vue';
import CopyToClipboard from '@/components/copy-to-clipboard/index.vue';

interface IProps {
  fields: Array<ModelPropertyDisplay & { copy?: boolean }>;
  details: any;
  isLoading: boolean;
  column?: number;
  labelWidth?: number;
}

const props = withDefaults(defineProps<IProps>(), {
  column: 1,
  labelWidth: 120,
});

const attrs = useAttrs();

const getDisplayCompProps = (field: ModelPropertyDisplay) => {
  const { id } = field;
  if (id === 'region') {
    return { vendor: props.details?.vendor };
  }
  return {};
};
</script>

<template>
  <grid-container :column="column" :label-width="labelWidth" v-bind="attrs">
    <grid-item v-for="field in fields" :key="field.id" :label="field.name">
      <bk-loading v-if="isLoading" size="mini" mode="spin" theme="primary" loading></bk-loading>
      <template v-else>
        <component v-if="field.render" :is="() => field.render(details)" />
        <display-value
          v-else
          :property="field"
          :value="get(details, field.id)"
          :display="field?.meta?.display"
          v-bind="getDisplayCompProps(field)"
        />
        <!-- !：不要混用render和copy -->
        <copy-to-clipboard v-if="field.copy" :content="get(details, field.id)" class="ml4" />
      </template>
    </grid-item>
  </grid-container>
</template>

<style scoped lang="scss"></style>
