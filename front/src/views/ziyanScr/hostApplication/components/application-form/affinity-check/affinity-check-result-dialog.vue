<script setup lang="ts">
import { WeixinPro } from 'bkui-vue/lib/icon';
import QcloudZoneValue from '@/views/ziyanScr/components/qcloud-resource/zone-value.vue';
import QcloudRegionValue from '@/views/ziyanScr/components/qcloud-resource/region-value.vue';
import FlexTag from '@/components/flex-tag/index.vue';
import WName from '@/components/w-name';
import type { IAffinityCheckResultItem } from './use-affinity-check';

const isShow = defineModel<boolean>('isShow');

const props = defineProps<{ result: IAffinityCheckResultItem[] }>();
</script>

<template>
  <bk-dialog
    v-model:is-show="isShow"
    dialog-type="show"
    width="1080px"
    title="亲和性检查结果"
    class="affinity-check-result-dialog"
  >
    <bk-alert theme="warning" class="alert-content">
      <template #title>
        <div class="title-content">
          亲和性检查说明：
          <span style="color: #ff3b30">亲和性检查数据仅供参考，</span>
          云主机的最终分布情况，请以实际资源交付结果为准，您可关注相关交付通知以获取最终信息
          <div>
            如对亲和性有任何疑问，欢迎通过群聊咨询。联系人：
            <div class="wx-name-list">
              <w-name name="ICR" alias="ICR助手" class="wx-name">
                <template #icon>
                  <weixin-pro width="16" height="16" />
                </template>
              </w-name>
              、
              <w-name name="lotuschen" alias="lotuschen(陈曦)" class="wx-name">
                <template #icon>
                  <weixin-pro width="16" height="16" />
                </template>
              </w-name>
              、
              <w-name name="chunzhang" alias="chunzhang(张纯)" class="wx-name">
                <template #icon>
                  <weixin-pro width="16" height="16" />
                </template>
              </w-name>
            </div>
          </div>
        </div>
      </template>
    </bk-alert>
    <bk-table :data="props.result" max-height="360px" class="result-table">
      <bk-table-column label="机型" prop="device_type" width="130px" show-overflow-tooltip />
      <bk-table-column label="申请数量" prop="replicas" width="90px" align="right" />
      <bk-table-column label="地域" prop="region" width="110px" show-overflow-tooltip>
        <template #default="{ row }">
          <QcloudRegionValue :value="row.region" />
        </template>
      </bk-table-column>
      <bk-table-column label="园区" prop="zone" width="110px" show-overflow-tooltip>
        <template #default="{ row }">
          <QcloudZoneValue :value="row.zone" />
        </template>
      </bk-table-column>
      <bk-table-column label="预测状态" prop="status" width="150px">
        <template #default="{ row }">
          <bk-tag radius="10px" theme="success" v-if="row.status === 1">从CRP预检有数据</bk-tag>
          <bk-tag radius="10px" theme="danger" v-else-if="row.status === 2">从CRP预检无数据</bk-tag>
          <span v-else>--</span>
        </template>
      </bk-table-column>
      <bk-table-column label="虚拟比" prop="max_cut_num" width="70px">
        <template #default="{ row }">
          <span v-if="row.max_cut_num > 0">1:{{ row.max_cut_num }}</span>
          <span v-else>--</span>
        </template>
      </bk-table-column>
      <bk-table-column label="申请后分布的母机" prop="ips" min-width="200px">
        <template #default="{ row }">
          <flex-tag :list="row.ips?.map((ip: string) => ({ name: ip }))" v-if="row.ips?.length" />
          <span v-else>--</span>
        </template>
      </bk-table-column>
    </bk-table>
  </bk-dialog>
</template>

<style scoped lang="scss">
.affinity-check-result-dialog {
  .alert-content {
    margin-bottom: 16px;
  }

  .result-table {
    margin-bottom: 36px;
  }

  .title-content {
    line-height: 20px;

    .wx-name-list {
      display: inline-flex;
      align-items: center;
      transform: translateY(3px);
    }
  }

  .wx-name {
    :deep(.bk-button-text) {
      gap: 4px;
    }
  }
}
</style>
