import { defineComponent, ref, provide, watch } from 'vue';
import './index.scss';
import allVendors from '@/assets/image/all-vendors.png';
import DynamicTree from '../components/dynamic-tree';
import LoadBalancerDropdownMenu from '../components/clb-dropdown-menu';
// import Funnel from 'bkui-vue/lib/icon/funnel';

export default defineComponent({
  name: 'LoadBalancerView',
  setup() {
    const treeRef = ref();
    const currentExpandItems = ref([]);
    const lastExpandItems = ref([]);
    const isAdvancedSearchShow = ref(false);
    provide('treeRef', treeRef);
    provide('currentExpandItems', currentExpandItems);

    const searchValue = ref('');
    const searchResultCount = ref(0);
    const toggleResultExpand = ref(true);
    provide('searchResultCount', searchResultCount);
    provide('toggleResultExpand', toggleResultExpand);

    const handleToggleResultExpand = (isExpand: boolean) => {
      toggleResultExpand.value = isExpand;
      if (isExpand) {
        lastExpandItems.value.forEach(node => { treeRef.value.setNodeOpened(node, isExpand); }) 
        lastExpandItems.value = [];
      } else {
        lastExpandItems.value = currentExpandItems.value;
        currentExpandItems.value.forEach(node => { treeRef.value.setNodeOpened(node, isExpand); }) 
      }
    }

    watch(searchValue, () => {
      searchResultCount.value = 0;
    })

    return () => (
      <div class='clb-view-page'>
        <div class='left-container'>
          <div class='search-wrap'>
            <bk-input v-model={searchValue.value} type='search' clearable placeholder='搜索负载均衡名称、VIP'></bk-input>
            {/* <Funnel class='advanced-search-icon' onClick={() => isAdvancedSearchShow.value = !isAdvancedSearchShow.value}></Funnel> */}
          </div>
          <div class='tree-wrap'>
            {
              searchValue.value 
                ? (searchResultCount.value ? (
                  <div class='search-result-wrap'>
                    <span class='left-text'>共 {searchResultCount.value} 条搜索结果</span>
                    {
                      currentExpandItems.value.length && toggleResultExpand.value
                        ? ( <span class='right-text' onClick={() => handleToggleResultExpand(false)}>全部收起</span> )
                        : ( <span class='right-text' onClick={() => handleToggleResultExpand(true)}>全部展开</span> )
                    }
                    
                  </div>
                ) : null) 
                : (
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
              )
            }
            <DynamicTree searchValue={searchValue.value} />
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
          <div>
            <p>上次全部收起的节点记录如下：</p>
            {
              lastExpandItems.value.map((item) => {
                return <p style={{paddingLeft: '2em'}}>{item.name}</p>
              })
            }
          </div>
        </div>
      </div>
    );
  },
});
