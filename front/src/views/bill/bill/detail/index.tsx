import { defineComponent, ref } from 'vue';
import { Tab } from 'bkui-vue';
import { BkTabPanel } from 'bkui-vue/lib/tab';
import BillDetailRenderTable from './RenderTable';
import Search from '@/views/bill/bill/components/search';
import { VendorEnum } from '@/common/constant';
import './index.scss';

export default defineComponent({
  name: 'BillDetail',
  setup() {
    const types = ref([
      { name: VendorEnum.AWS, label: '亚马逊云' },
      { name: VendorEnum.GCP, label: '谷歌云' },
      { name: VendorEnum.AZURE, label: '微软云' },
      { name: VendorEnum.HUAWEI, label: '华为云' },
      { name: VendorEnum.ZENLAYER, label: 'zenlayer' },
    ]);
    const activeType = ref(VendorEnum.AWS);

    return () => (
      <div class='bill-detail-module'>
        <Tab v-model:active={activeType.value} type='card-grid'>
          <Search />
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
