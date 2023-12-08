import { defineComponent, ref } from 'vue';
import './index.scss';
import { SearchSelect } from 'bkui-vue';
import allVendors from '@/assets/image/all-vendors.png';
import clbIcon from "@/assets/image/clb.png";
import listenerIcon from "@/assets/image/listener.png";
import domainIcon from "@/assets/image/domain.png";
import DynamicTree from '../components/dynamic-tree';
// import Funnel from 'bkui-vue/lib/icon/funnel';

export default defineComponent({
  name: 'LoadBalancerView',
  setup() {
    const treeData = ref([]);
    const baseUrl = 'http://localhost:3000';
    const rootType = 'clb';

    const typeIconMap = {
      clb: clbIcon,
      listener: listenerIcon,
      domain: domainIcon
    };

    const isAdvancedSearchShow = ref(false);

    return () => (
      <div class='clb-view-page'>
        <div class='left-container'>
          <div class='search-wrap'>
            <SearchSelect placeholder='搜索负载均衡名称、VIP'></SearchSelect>
            {/* <Funnel class='advanced-search-icon' onClick={() => isAdvancedSearchShow.value = !isAdvancedSearchShow.value}></Funnel> */}
          </div>
          <div class='tree-wrap'>
            <div class='all-clbs'>
              <div class='left-wrap'>
                <img src={allVendors} alt='' class='prefix-icon' />
                <span class='text'>全部负载均衡</span>
              </div>
              <div class='right-wrap'>
                6654
              </div>
            </div>
            <DynamicTree v-model:treeData={treeData.value} baseUrl={baseUrl} rootType={rootType} typeIconMap={typeIconMap} class='dynamic-tree-wrap'></DynamicTree>
          </div>
        </div>
        {
          isAdvancedSearchShow.value && <div class='advanced-search'>高级搜索</div>
        }
        <div class='main-container'>右侧内容</div>
      </div>
    );
  },
});
