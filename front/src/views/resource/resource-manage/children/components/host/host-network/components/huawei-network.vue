<script lang="ts" setup>
import { useI18n } from 'vue-i18n';
import {
  ref,
  h,
  reactive,
} from 'vue';
import {
  Button,
  InfoBox,
} from 'bkui-vue';

const { t } = useI18n();
const showNetworkDialog = ref(false);
const showVirtual = ref(false);
const showPublic = ref(false);
const showVpc = ref(false);
const showGroup = ref(false);
const showPrivate = ref(false);
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
    label: '弹性公网IP',
    render() {
      return [
        h(
          'span',
          {},
          [
            '无',
          ],
        ),
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
          [
            '绑定公网IP',
          ],
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
            onClick() {
              handleToggleShow('private');
            },
          },
          [
            '修改私有IP',
          ],
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
          [
            '释放虚拟IP',
          ],
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
  console.log('type', type);
  if (type === 'network') {
    showNetworkDialog.value = !showNetworkDialog.value;
  } else if (type === 'virtual') {
    showVirtual.value = !showVirtual.value;
  } else if (type === 'public') {
    showPublic.value = !showPublic.value;
  } else if (type === 'vpc') {
    showVpc.value = !showVpc.value;
  } else if (type === 'securityGroup') {
    showGroup.value = !showGroup.value;
  } else if (type === 'private') {
    showPrivate.value = !showPrivate.value;
  }
};

const handleConfirmBind = () => {
  handleToggleShow('already');
};

const handleFreedIp = (type: string) => {
  InfoBox({
    title: type === 'freed' ? '确定释放此虚拟IP' : '确定删除网卡',
    subTitle: type === 'freed' ? '确定要释放吗' : '确定要删除吗',
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
    @click="() => {
      handleToggleShow('network')
    }"
  >
    添加网络接口
  </bk-button>
  <div class="main-network table-warp mt20">
    <div class="table-flex">
      <div>主网卡</div>
      <div>
        <bk-button
          theme="primary"
          @click="() => {
            handleToggleShow('virtual')
          }"
        >
          绑定虚拟IP
        </bk-button>
        <bk-button
          class="ml20"
          @click="() => {
            handleToggleShow('vpc')
          }"
        >
          更换VPC
        </bk-button>
        <bk-button
          class="ml20"
          @click="() => {
            handleToggleShow('securityGroup')
          }"
        >
          更换安全组
        </bk-button>
        <bk-button
          class="ml20"
          @click="() => {
            handleFreedIp('unbind')
          }"
        >
          删除
        </bk-button>
      </div>
    </div>
    <bk-form label-position="left" class="flex mt20 pt20 form-warp">
      <bk-form-item
        :label="t('状态')"
        label-width="100"
        class="item-warp"
      >
        <span>
          新加坡
        </span>
      </bk-form-item>
      <bk-form-item
        :label="t('所属网络')"
      >
        <span>
          新加坡
        </span>
      </bk-form-item>
      <bk-form-item
        :label="t('子网')"
      >
        <span>
          新加坡
        </span>
      </bk-form-item>
      <bk-form-item
        :label="t('安全组')"
      >
        <span>
          新加坡
        </span>
      </bk-form-item>
    </bk-form>
    <bk-table
      class="mt20"
      row-hover="auto"
      :columns="columns"
      :data="tableData"
    />
  </div>

  <div class="main-network table-warp mt20">
    <div class="table-flex">
      <div>主网卡</div>
      <div>
        <bk-button
          theme="primary"
          @click="() => {
            handleToggleShow('virtual')
          }"
        >
          绑定虚拟IP
        </bk-button>
        <bk-button
          class="ml20"
          @click="() => {
            handleToggleShow('vpc')
          }"
        >
          更换VPC
        </bk-button>
        <bk-button
          class="ml20"
          @click="() => {
            handleToggleShow('securityGroup')
          }"
        >
          更换安全组
        </bk-button>
        <bk-button
          class="ml20"
          @click="() => {
            handleFreedIp('unbind')
          }"
        >
          删除
        </bk-button>
      </div>
    </div>
    <bk-form label-position="left" class="flex mt20 pt20 form-warp">
      <bk-form-item
        :label="t('状态')"
        label-width="100"
        class="item-warp"
      >
        <span>
          新加坡
        </span>
      </bk-form-item>
      <bk-form-item
        :label="t('所属网络')"
      >
        <span>
          新加坡
        </span>
      </bk-form-item>
      <bk-form-item
        :label="t('子网')"
      >
        <span>
          新加坡
        </span>
      </bk-form-item>
      <bk-form-item
        :label="t('安全组')"
      >
        <span>
          新加坡
        </span>
      </bk-form-item>
    </bk-form>
    <bk-table
      class="mt20"
      row-hover="auto"
      :columns="columns"
      :data="tableData"
    />
  </div>

  <bk-dialog
    :is-show="showNetworkDialog"
    width="620"
    title="添加网卡"
    theme="primary"
    quick-close
    @closed="() => {
      handleToggleShow('network')
    }"
    @confirm="handleConfirmBind"
  >
    <bk-form
      class="mt20"
      label-width="100">
      <bk-form-item
        :label="t('云服务器')"
      >
        <span>
          新加坡
        </span>
      </bk-form-item>
      <bk-form-item
        :label="t('虚拟私有云')"
      >
        <span>
          新加坡
        </span>
      </bk-form-item>
      <bk-form-item
        :label="t('安全组')"
      >
        <bk-select v-model="fromData.name">
        </bk-select>
      </bk-form-item>
      <bk-form-item
        :label="t('子网')"
      >
        <bk-select v-model="fromData.name">
        </bk-select>
      </bk-form-item>
      <bk-form-item
        :label="t('私有IP地址')"
      >
        <bk-input
          v-model="fromData.name"
          :placeholder="t('请输入私有IP地址')"
        />
      </bk-form-item>
    </bk-form>
  </bk-dialog>
  <bk-dialog
    :is-show="showVirtual"
    width="620"
    title="绑定虚拟IP"
    theme="primary"
    quick-close
    @closed="() => {
      handleToggleShow('virtual')
    }">
    <bk-table
      class="mt20"
      dark-header
      :data="[{ ip: 'testetstt' }]"
      :outer-border="false"
    >
      <bk-table-column
        label="内网IP"
      >
        <template #default="{ data } ">
          <div class="cell-flex">
            <bk-radio
              label="" @click="() => {
                handleRadio(data)
              }" />
            <span class="pl10">{{ data.ip }}</span>
          </div>
        </template>
      </bk-table-column>
      <bk-table-column
        label="已绑定的EIP"
        prop="ip"
      />
    </bk-table>
  </bk-dialog>
  <bk-dialog
    :is-show="showPrivate"
    width="620"
    title="修改私有IP"
    theme="primary"
    quick-close
    @closed="() => {
      handleToggleShow('private')
    }">
    <bk-form
      class="mt20"
      label-width="100">
      <bk-form-item
        :label="t('云服务器')"
      >
        <span>
          新加坡
        </span>
      </bk-form-item>
      <bk-form-item
        :label="t('虚拟私有云')"
      >
        <span>
          新加坡
        </span>
      </bk-form-item>
      <bk-form-item
        :label="t('当前私有IP地址')"
      >
        <span>
          新加坡
        </span>
      </bk-form-item>
      <bk-form-item
        :label="t('子网')"
      >
        <bk-select v-model="fromData.name">
        </bk-select>
      </bk-form-item>
      <bk-form-item
        :label="t('私有IP地址')"
      >
        <bk-input
          v-model="fromData.name"
          :placeholder="t('请输入私有IP地址')"
        />
      </bk-form-item>
    </bk-form>
  </bk-dialog>
  <bk-dialog
    :is-show="showVpc"
    width="620"
    title="更换VPC"
    theme="primary"
    quick-close
    @closed="() => {
      handleToggleShow('vpc')
    }">
    <bk-form
      class="mt20"
      label-width="100">
      <bk-form-item
        :label="t('云服务器')"
      >
        <span>
          新加坡一区
        </span>
      </bk-form-item>
      <bk-form-item
        :label="t('虚拟私有云')"
      >
        <bk-select v-model="fromData.name">
        </bk-select>
      </bk-form-item>
      <bk-form-item
        :label="t('子网')"
      >
        <bk-select v-model="fromData.name">
        </bk-select>
      </bk-form-item>
      <bk-form-item
        :label="t('私有IP地址')"
      >
        <bk-select v-model="fromData.name">
        </bk-select>
      </bk-form-item>
      <bk-form-item
        :label="t('安全组')"
      >
        <bk-select v-model="fromData.name">
        </bk-select>
      </bk-form-item>
    </bk-form>
  </bk-dialog>
  <bk-dialog
    :is-show="showGroup"
    width="620"
    title="更换安全组"
    theme="primary"
    quick-close
    @closed="() => {
      handleToggleShow('securityGroup')
    }"
    @confirm="handleConfirmBind"
  >
    <bk-form
      class="mt20"
      label-width="100"
      label-position="left">
      <bk-form-item
        :label="t('云服务器名称')"
      >
        <span>
          新加坡
        </span>
      </bk-form-item>
      <bk-form-item
        :label="t('网卡')"
      >
        <bk-select v-model="fromData.name">
        </bk-select>
      </bk-form-item>
    </bk-form>
    <bk-table
      class="mt20"
      dark-header
      :data="[{ ip: 'testetstt' }]"
      :outer-border="false"
    >
      <bk-table-column
        label="安全组名称"
      >
        <template #default="{ data } ">
          <div class="cell-flex">
            <bk-radio
              label="" @click="() => {
                handleRadio(data)
              }" />
            <span class="pl10">{{ data.ip }}</span>
          </div>
        </template>
      </bk-table-column>
      <bk-table-column
        label="描述"
        prop="ip"
      />
    </bk-table>
  </bk-dialog>
  <bk-dialog
    :is-show="showPublic"
    width="620"
    title="绑定弹性公网IP"
    theme="primary"
    quick-close
    @closed="() => {
      handleToggleShow('public')
    }"
    @confirm="handleConfirmBind"
  >
    <bk-form
      class="mt20"
      label-position="left">
      <bk-form-item
        label-width="100"
        :label="t('弹性公网IP')"
      >
        <span>
          新加坡
        </span>
      </bk-form-item>
      <bk-form-item
        :label="t('选择实例')"
        label-width="100"
      >
        <bk-radio-group
          v-model="fromData.name"
        >
          <bk-radio label="云服务器" />
          <bk-radio label="裸金属" />
          <bk-radio label="虚拟IP地址" />
        </bk-radio-group>
      </bk-form-item>
      <bk-form-item label-width="0">
        <bk-table
          class="mt20"
          dark-header
          :data="[{ ip: 'testetstt' }]"
          :outer-border="false"
        >
          <bk-table-column
            label="安全组名称"
          >
            <template #default="{ data } ">
              <div class="cell-flex">
                <bk-radio
                  label="" @click="() => {
                    handleRadio(data)
                  }" />
                <span class="pl10">{{ data.ip }}</span>
              </div>
            </template>
          </bk-table-column>
          <bk-table-column
            label="描述"
            prop="ip"
          />
        </bk-table>
      </bk-form-item>
      <bk-form-item
        :label="t('网卡')"
        label-width="100"
      >
        <bk-select v-model="fromData.name">
        </bk-select>
      </bk-form-item>
    </bk-form>
  </bk-dialog>
</template>

<style lang="scss" scoped>
  .info-title {
    font-size: 14px;
    margin: 20px 0 5px;
  }
  .sub-title{
    font-size: 12px;
  }
  .cell-flex{
    display: flex;
    align-items: center;
  }
  .table-warp{
    padding: 20px;
    border: 1px dashed rgb(225, 221, 221);
    .table-flex{
      display: flex;
      justify-content: space-between;
      align-items: center;
    }
  }
  .flex{
    display: flex;
    align-items: center;
  }
  .form-warp{
    border-top: 1px solid rgb(225, 221, 221);
    .item-warp{
      margin-right: 40px;
    }
  }

  :deep(.detail-tab-main) .bk-tab-content {
    height: calc(100vh - 300px) !important;
  }
</style>
