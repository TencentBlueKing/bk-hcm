<script lang="ts" setup>
import type { Image, VendorResource } from '@/typings/resource-detail';
import { extensionPath } from '@/views/resource/resource-manage/common/info/extension-path';
import DetailInfo from '@/views/resource/resource-manage/common/info/detail-info';
import { PropType, ref } from 'vue';
import { useResourceDetail } from '../../../hooks/use-detail';
import { FieldList } from '../../../common/info-list/types';

const props = defineProps({
  id: {
    type: String as PropType<string>,
  },
  vendor: {
    type: String as PropType<string>,
  },
});

const tcloudField = extensionPath<VendorResource<Image, 'tcloud'>>();

const fields = ref<FieldList>([
  {
    name: '实例ID',
    prop: 'cloud_id',
  },
  {
    name: '名称',
    prop: 'name',
  },
  {
    name: '云厂商',
    prop: 'vendorName',
  },
  {
    name: '架构',
    prop: 'architecture',
  },
  {
    name: '状态',
    prop: 'state',
  },
  {
    name: '类型',
    prop: 'type',
  },
  {
    name: '平台',
    prop: 'platform',
  },
]);

const { loading, detail } = useResourceDetail<'images'>(
  'images',
  props.id,
  (detail) => {
    switch (detail.vendor) {
      case 'tcloud':
        fields.value.push(
          ...[
            {
              name: '地域',
              value: detail.extension?.region ?? detail.region,
            },
            {
              name: '镜像来源',
              prop: tcloudField('extension.image_source'),
            },
            {
              name: '镜像大小',
              prop: tcloudField('extension.image_size'),
            },
          ],
        );
        break;
      // 其它类型的待补充
    }
  },
  props.vendor,
);
</script>

<template>
  <bk-loading :loading="loading">
    <detail-info :detail="detail" :fields="fields" global-copyable />
  </bk-loading>
</template>
