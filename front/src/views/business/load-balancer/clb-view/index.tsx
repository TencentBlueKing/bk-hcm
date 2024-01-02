import { defineComponent, ref, provide, watch } from 'vue';
import './index.scss';
import allVendors from '@/assets/image/all-vendors.png';
import DynamicTree from '../components/dynamic-tree';
import LoadBalancerDropdownMenu from '../components/clb-dropdown-menu';
import AllClbsManager from './all-clbs-manager';
import SpecificListenerManager from './specific-listener-manager';
import SpecificClbManager from './specific-clb-manager';
import SpecificDomainManager from './specific-domain-manager';
import SimpleSearchSelect from '../components/simple-search-select';

export default defineComponent({
  name: 'LoadBalancerView',
  setup() {
    const treeData = ref([]);
    provide('treeData', treeData);
    const isAdvancedSearchShow = ref(false);

    const searchValue = ref('');
    const searchDataList = [
      { id: 'clb_name', name: '负载均衡名称' },
      { id: 'clb_vip', name: '负载均衡VIP' },
      { id: 'listener_name', name: '监听器名称' },
      { id: 'protocol', name: '协议' },
      { id: 'port', name: '端口' },
      { id: 'domain', name: '域名' },
    ];
    const searchResultCount = ref(0);
    provide('searchResultCount', searchResultCount);

    const toggleExpand = ref(false);
    const handleToggleResultExpand = (isOpen: boolean, shallow?: boolean) => {
      toggleExpand.value = !isOpen;
      treeData.value = treeData.value.map((item) => {
        item.isOpen = isOpen;
        if (!shallow && item.children?.length) {
          item.children = item.children.map((subItem: any) => {
            subItem.isOpen = isOpen;
            return subItem;
          });
        }
        return item;
      });
    };

    const componentMap = {
      all: <AllClbsManager />,
      clb: <SpecificClbManager />,
      listener: <SpecificListenerManager />,
      domain: <SpecificDomainManager />,
    };
    const renderComponent = (type: 'all' | 'clb' | 'listener' | 'domain') => {
      return componentMap[type];
    };

    const activeType = ref('all' as 'all' | 'clb' | 'listener' | 'domain');
    const handleTypeChange = (type: 'all' | 'clb' | 'listener' | 'domain') => {
      activeType.value = type;
    };

    const popInstance = ref(null);
    provide('popInstance', popInstance);

    watch(searchValue, () => {
      searchResultCount.value = 0;
    });

    watch(searchResultCount, (val) => {
      if (val <= 20) {
        handleToggleResultExpand(true, true);
      } else {
        handleToggleResultExpand(false);
      }
    });

    return () => (
      <div class='clb-view-page'>
        <div class='left-container'>
          <SimpleSearchSelect v-model:searchValue={searchValue.value} dataList={searchDataList} />
          <div class='tree-wrap'>
            {
              // eslint-disable-next-line no-nested-ternary
              searchValue.value ? (
                searchResultCount.value ? (
                  <div class='search-result-wrap'>
                    <span class='left-text'>共 {searchResultCount.value} 条搜索结果</span>
                    {toggleExpand.value ? (
                      <span class='right-text' onClick={() => handleToggleResultExpand(true)}>
                        全部展开
                      </span>
                    ) : (
                      <span class='right-text' onClick={() => handleToggleResultExpand(false)}>
                        全部收起
                      </span>
                    )}
                  </div>
                ) : null
              ) : (
                <div
                  class={`all-clbs${activeType.value === 'all' ? ' selected' : ''}`}
                  onClick={() => handleTypeChange('all')}>
                  <div class='left-wrap'>
                    <img src={allVendors} alt='' class='prefix-icon' />
                    <span class='text'>全部负载均衡</span>
                  </div>
                  <div class='right-wrap'>
                    <div class='count'>{6654}</div>
                    <LoadBalancerDropdownMenu class='more-action' type='all' />
                  </div>
                </div>
              )
            }
            <DynamicTree searchValue={searchValue.value} onHandleTypeChange={handleTypeChange} />
          </div>
        </div>
        {isAdvancedSearchShow.value && <div class='advanced-search'>高级搜索</div>}
        <div class='main-container'>{renderComponent(activeType.value)}</div>
      </div>
    );
  },
});
