import { defineComponent, ref } from 'vue';
import './index.scss';

import { Tab } from 'bkui-vue';
import { BkTabPanel } from 'bkui-vue/lib/tab';
import BillDetailRenderTable from './RenderTable';
import Search from '@/views/bill/bill/components/search';

import { useI18n } from 'vue-i18n';
import { VendorEnum } from '@/common/constant';

export default defineComponent({
  name: 'BillDetail',
  setup() {
    const { t } = useI18n();
    const types = ref([
      { name: VendorEnum.AWS, label: t('亚马逊云') },
      { name: VendorEnum.GCP, label: t('谷歌云') },
      { name: VendorEnum.AZURE, label: t('微软云') },
      { name: VendorEnum.HUAWEI, label: t('华为云') },
      { name: VendorEnum.ZENLAYER, label: t('zenlayer') },
    ]);
    const activeType = ref(VendorEnum.AWS);

    const billDetailRenderTableRef = ref();

    return () => (
      <div class='bill-detail-module'>
        <Tab v-model:active={activeType.value} type='card-grid'>
          {types.value.map(({ name, label }) => (
            <BkTabPanel key={name} name={name} label={label} renderDirective='if'>
              <Search
                searchKeys={['root_account_id', 'product_id', 'main_account_id']}
                vendor={[activeType.value]}
                onSearch={(rules) => billDetailRenderTableRef.value?.reloadTable(rules)}
                disableSearchHandler={(rules) => rules.filter((rule) => rule.field !== 'vendor').length === 0}
              />
              <BillDetailRenderTable ref={billDetailRenderTableRef} vendor={name} />
            </BkTabPanel>
          ))}
        </Tab>
      </div>
    );
  },
});
