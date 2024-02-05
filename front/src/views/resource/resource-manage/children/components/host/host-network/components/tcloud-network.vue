<script lang="ts" setup>
import type {
  // PlainObject,
  FilterType,
} from '@/typings/resource';
import { useI18n } from 'vue-i18n';
import { ref, h, reactive, PropType } from 'vue';
import { Button, InfoBox } from 'bkui-vue';
import { useResourceStore } from '@/store/resource';
import useQueryList from '@/views/resource/resource-manage/hooks/use-query-list';

const props = defineProps({
  filter: {
    type: Object as PropType<FilterType>,
  },
  data: {
    type: Object,
  },
});
const resourceStore = useResourceStore();

console.log('props.data.vendor', props.data.vendor);
const { datas, pagination, isLoading, handlePageChange, handlePageSizeChange } = useQueryList(props, '', () => {
  return Promise.all([resourceStore.cvmNetwork(props.data.vendor, props.data.id)]);
});

console.log(datas, pagination, isLoading, handlePageChange, handlePageSizeChange);

const { t } = useI18n();
const showBind = ref(false);
const showSub = ref(false);
const showPublic = ref(false);
const columns = [
  {
    label: '类型',
    field: 'id',
  },
  {
    label: '内网IP',
    field: 'id',
  },
  {
    label: '普通公网IP/EIP',
    render() {
      return [
        h('span', {}, ['无']),
        h(
          Button,
          {
            text: true,
            theme: 'primary',
            class: 'ml10',
            onClick() {
              handleToggleShow('public');
            },
          },
          ['绑定公网IP'],
        ),
      ];
    },
  },
  {
    label: '备注',
    field: 'id',
  },
  {
    label: '操作',
    render() {
      return [
        h(
          Button,
          {
            text: true,
            theme: 'primary',
            class: 'mr10',
            onClick() {},
          },
          ['修改主IP'],
        ),
        h(
          Button,
          {
            text: true,
            theme: 'primary',
            class: 'mr10',
            onClick() {
              handleFreedIp('freed');
            },
          },
          ['释放辅助IP'],
        ),
      ];
    },
  },
];
const tableData = [
  {
    id: 233,
  },
];

const fromData = reactive({
  name: '',
});

const handleToggleShow = (type: string) => {
  if (type === 'already') {
    showBind.value = !showBind.value;
  } else if (type === 'sub') {
    showSub.value = !showSub.value;
  } else if (type === 'public') {
    showPublic.value = !showPublic.value;
  }
};

const handleConfirmBind = () => {
  handleToggleShow('already');
};

const handleFreedIp = (type: string) => {
  InfoBox({
    title: type === 'freed' ? '确定释放此内网IP' : '确定解绑弹性网卡',
    subTitle: type === 'freed' ? '内网IP释放后将自动解关联弹性公网IP和负载均衡' : '',
    headerAlign: 'center',
    footerAlign: 'center',
    contentAlign: 'center',
    onConfirm() {
      console.log('111');
    },
  });
};

const handleRadio = (item: any) => {
  console.log(item);
};
</script>

<template>
  <bk-button
    class="mt20"
    theme="primary"
    @click="
      () => {
        handleToggleShow('already');
      }
    "
  >
    绑定已有网络接口
  </bk-button>
  <div class="main-network table-warp mt20">
    <div class="table-flex">
      <div>主网卡</div>
      <div>
        <bk-button
          theme="primary"
          @click="
            () => {
              handleToggleShow('sub');
            }
          "
        >
          添加辅助IP
        </bk-button>
        <bk-button
          class="ml20"
          @click="
            () => {
              handleFreedIp('unbind');
            }
          "
        >
          解绑网卡
        </bk-button>
      </div>
    </div>
    <bk-table class="mt20" row-hover="auto" :columns="columns" :data="tableData" show-overflow-tooltip />
  </div>

  <div class="sub-network table-warp mt20">
    <div class="table-flex">
      <div>辅助网卡</div>
      <div>
        <bk-button
          theme="primary"
          @click="
            () => {
              handleToggleShow('sub');
            }
          "
        >
          添加辅助IP
        </bk-button>
        <bk-button
          class="ml20"
          @click="
            () => {
              handleFreedIp('unbind');
            }
          "
        >
          解绑网卡
        </bk-button>
      </div>
    </div>
    <bk-table class="mt20" row-hover="auto" :columns="columns" :data="tableData" show-overflow-tooltip />
  </div>

  <bk-dialog
    :is-show="showBind"
    width="620"
    title="绑定弹性IP"
    theme="primary"
    quick-close
    @closed="
      () => {
        handleToggleShow('already');
      }
    "
    @confirm="handleConfirmBind"
  >
    <bk-alert theme="info" title="辅助弹性网卡需根据实际情况单独配置安全组，请确认安全策略并为其关联安全组" />
    <div class="sub-title mt10 mb10">请选择ssss要绑定的弹性网卡</div>
    <bk-radio-group value="xxx">
      <bk-radio label="绑定已有弹性网卡" />
      <bk-radio label="新建已有弹性网卡" />
    </bk-radio-group>
    <bk-table
      class="mt20"
      dark-header
      :data="[{ ip: 'testetstt' }]"
      :outer-border="false"
      v-if="false"
      show-overflow-tooltip
    >
      <bk-table-column label="名称" prop="ip">
        <!-- eslint-disable vue/no-template-shadow -->
        <template #default="{ data }">
          <div class="cell-flex">
            <bk-radio
              label=""
              @click="
                () => {
                  handleRadio(data);
                }
              "
            />
            <span class="pl10">{{ data.ip }}</span>
          </div>
        </template>
      </bk-table-column>
      <bk-table-column label="名称" prop="ip" />
      <bk-table-column label="所属子网" prop="ip" />
      <bk-table-column label="网卡内网ip数" prop="ip" />
    </bk-table>
    <bk-form class="mt20" label-width="100">
      <bk-form-item :label="t('名称')">
        <bk-input v-model="fromData.name" :placeholder="t('请输入弹性网卡名称')" />
      </bk-form-item>
      <bk-form-item :label="t('所在地域')">
        <span>新加坡</span>
      </bk-form-item>
      <bk-form-item :label="t('所属网络')">
        <span>新加坡</span>
      </bk-form-item>
      <bk-form-item :label="t('所属子网')">
        <span>新加坡</span>
      </bk-form-item>
      <bk-form-item :label="t('可用区')">
        <span>新加坡一区</span>
      </bk-form-item>
      <bk-form-item :label="t('可分配IP数')">
        <span>新加坡一区</span>
      </bk-form-item>
      <bk-form-item :label="t('分配IP')">
        <div class="cell-flex">
          <span>新加坡一区</span>
          <bk-input style="width: 200px" class="ml20" v-model="fromData.name" type="password" />
        </div>
      </bk-form-item>
    </bk-form>
  </bk-dialog>
  <bk-dialog
    :is-show="showSub"
    width="620"
    title="绑定弹性IP"
    theme="primary"
    quick-close
    @closed="
      () => {
        handleToggleShow('sub');
      }
    "
  >
    <bk-form class="mt20" label-width="100">
      <bk-form-item :label="t('所属子网')">
        <span>新加坡</span>
      </bk-form-item>
      <bk-form-item :label="t('子网CIDR')">
        <span>新加坡</span>
      </bk-form-item>
      <bk-form-item :label="t('子网可用IP')">
        <span>新加坡</span>
      </bk-form-item>
      <bk-form-item :label="t('IP 配额')">
        <span>新加坡</span>
      </bk-form-item>
      <bk-form-item :label="t('可用配额')">
        <span>新加坡一区</span>
      </bk-form-item>
      <bk-form-item :label="t('分配IP')">
        <div class="flex">
          <bk-select v-model="fromData.name"></bk-select>
          <bk-input class="ml10 mr10" v-model="fromData.name"></bk-input>
          <bk-button text theme="primary">{{ t('删除') }}</bk-button>
        </div>
      </bk-form-item>
      <bk-form-item>
        <bk-button text theme="primary">{{ t('新增') }}</bk-button>
      </bk-form-item>
    </bk-form>
  </bk-dialog>
  <bk-dialog
    :is-show="showPublic"
    width="620"
    title="绑定弹性IP"
    theme="primary"
    quick-close
    @closed="
      () => {
        handleToggleShow('public');
      }
    "
    @confirm="handleConfirmBind"
  >
    <div class="sub-title mt10 mb10">请选择ssss要绑定的弹性公网IP网卡</div>
    <bk-table class="mt20" dark-header :data="[{ ip: 'testetstt' }]" :outer-border="false" show-overflow-tooltip>
      <bk-table-column label="ID/名称">
        <!-- eslint-disable vue/no-template-shadow -->
        <template #default="{ data }">
          <div class="cell-flex">
            <bk-radio
              label=""
              @click="
                () => {
                  handleRadio(data);
                }
              "
            />
            <span class="pl10">{{ data.ip }}</span>
          </div>
        </template>
      </bk-table-column>
      <bk-table-column label="弹性公网IP" prop="ip" />
    </bk-table>
  </bk-dialog>
</template>

<style lang="scss" scoped>
.info-title {
  font-size: 14px;
  margin-bottom: 8px;
}
.sub-title {
  font-size: 12px;
}
.cell-flex {
  display: flex;
  align-items: center;
}
.table-warp {
  padding: 20px;
  border: 1px dashed rgb(225, 221, 221);
  .table-flex {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }
}
.flex {
  display: flex;
  align-items: center;
}
</style>
