import { computed, defineComponent, ref } from 'vue';
import { Button, Select, Tab } from 'bkui-vue';
import { BkTabPanel } from 'bkui-vue/lib/tab';
import FilterFormContainer, { IFormItemProps } from '@/components/FilterFormContainer';
import BillDetailRenderTable from './render-table';
import { VendorEnum } from '@/common/constant';
import './index.scss';

export default defineComponent({
  name: 'BillDetail',
  setup() {
    const types = ref([
      { name: VendorEnum.TCLOUD, label: '腾讯云' },
      { name: VendorEnum.AWS, label: '亚马逊云' },
      { name: VendorEnum.GCP, label: '谷歌云' },
      { name: VendorEnum.AZURE, label: '微软云' },
      { name: VendorEnum.HUAWEI, label: '华为云' },
      { name: VendorEnum.ZENLAYER, label: 'zenlayer' },
      { name: VendorEnum.KAOPU, label: '靠谱云' },
    ]);
    const activeType = ref(VendorEnum.TCLOUD);

    const formConfig = computed((): IFormItemProps[] => [
      {
        label: '一级账号',
        render: () => <Select />,
      },
      {
        label: '运营产品',
        render: () => <Select />,
      },
      {
        label: '二级账号',
        render: () => <Select />,
      },
    ]);

    return () => (
      <div class='bill-detail-module'>
        <Tab v-model:active={activeType.value} type='card-grid' class='bill-detail-tab'>
          <div class='filter-container'>
            <FilterFormContainer
              class='tab-panel-content'
              col={3}
              gutter={24}
              margin={0}
              formConfig={formConfig.value}
            />
            <Button theme='primary' class='w88 mr8'>
              查询
            </Button>
            <Button class='w88'>重置</Button>
          </div>
          {types.value.map(({ name, label }) => (
            <BkTabPanel key={name} name={name} label={label} renderDirective='if'>
              <BillDetailRenderTable vendor={name} />
            </BkTabPanel>
          ))}
        </Tab>
      </div>
    );
  },
});
