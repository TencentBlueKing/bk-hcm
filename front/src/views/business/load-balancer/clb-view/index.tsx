import { defineComponent, ref, provide } from 'vue';
import './index.scss';
import { SearchSelect } from 'bkui-vue';
import allVendors from '@/assets/image/all-vendors.png';
import DynamicTree from '../components/dynamic-tree';
import LoadBalancerDropdownMenu from '../components/clb-dropdown-menu';
// import Funnel from 'bkui-vue/lib/icon/funnel';

export default defineComponent({
  name: 'LoadBalancerView',
  setup() {
    const treeData = ref([]);
    const baseUrl = 'http://localhost:3000';
    const treeRef = ref();
    const currentExpandItems = ref([]);
    const isAdvancedSearchShow = ref(false);
    provide('treeRef', treeRef)
    provide('currentExpandItems', currentExpandItems)

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
                <div class='count'>{6654}</div>
                <LoadBalancerDropdownMenu uuid='all' type='all' />
              </div>
            </div>
            <DynamicTree v-model:treeData={treeData.value} baseUrl={baseUrl}></DynamicTree>
          </div>
        </div>
        {
          isAdvancedSearchShow.value && <div class='advanced-search'>高级搜索</div>
        }
        <div class='main-container'>
          <bk-button style={{margin: '0 10px 10px 0'}} theme='primary' onClick={() => {
            currentExpandItems.value.length && treeRef.value.setNodeOpened(currentExpandItems.value.pop(), false);
            }}>
            收起当前节点，支持逐级级收起
          </bk-button>
          <bk-button theme='warning' onClick={() => {
            currentExpandItems.value.length && currentExpandItems.value.forEach(node => {
              treeRef.value.setNodeOpened(node, false);
            })
          }}>
            收起全部节点
          </bk-button>
          <div>
            <p>当前展开的节点记录如下：</p>
            {
              currentExpandItems.value.map((item) => {
                return <p style={{paddingLeft: '2em'}}>{item.name}</p>
              })
            }
          </div>
        </div>
      </div>
    );
  },
});
