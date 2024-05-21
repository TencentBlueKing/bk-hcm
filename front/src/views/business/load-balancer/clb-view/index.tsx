import { defineComponent, ref, provide, watch, reactive } from 'vue';
import { Popover } from 'bkui-vue';
import './index.scss';
import allVendors from '@/assets/image/all-vendors.png';
import DynamicTree from '../components/dynamic-tree';
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
    const treeRef = ref(null);
    provide('treeRef', treeRef);
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
    const currentSelectedTreeNode = ref(null);
    const handleTypeChange = (type: 'all' | 'clb' | 'listener' | 'domain') => {
      activeType.value = type;
      if (type === 'all') {
        treeRef.value.setSelect(currentSelectedTreeNode.value, false);
      }
    };

    const allClbsItem = reactive({ isDropdownListShow: false });
    const handleMoreActionClick = (e: Event, node: any) => {
      e.stopPropagation();
      node.isDropdownListShow = !node.isDropdownListShow;
    };
    const handleDropdownItemClick = () => {
      // dropdown item click event
    };
    const renderDropdownActionList = (node: any) => {
      return (
        <Popover
          trigger='click'
          theme='light'
          renderType='shown'
          placement='bottom-start'
          arrow={false}
          extCls='dropdown-popover-wrap'
          onAfterHidden={({ isShow }) => (node.isDropdownListShow = isShow)}>
          {{
            default: () => (
              <div class='more-action' onClick={(e) => handleMoreActionClick(e, node)}>
                <i class='hcm-icon bkhcm-icon-more-fill'></i>
              </div>
            ),
            content: () => (
              <div class='dropdown-list'>
                <div class='dropdown-item' onClick={handleDropdownItemClick}>
                  购买负载均衡
                </div>
              </div>
            ),
          }}
        </Popover>
      );
    };

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
                  class={['all-clbs', `${activeType.value === 'all' ? ' selected' : ''}`]}
                  onClick={() => handleTypeChange('all')}>
                  <div class='left-wrap'>
                    <img src={allVendors} alt='' class='prefix-icon' />
                    <span class='text'>全部负载均衡</span>
                  </div>
                  <div class={`right-wrap${allClbsItem.isDropdownListShow ? ' show-dropdown' : ''}`}>
                    <div class='count'>{6654}</div>
                    {renderDropdownActionList(allClbsItem)}
                  </div>
                </div>
              )
            }
            <DynamicTree
              searchValue={searchValue.value}
              v-model:currentSelectedTreeNode={currentSelectedTreeNode.value}
              onHandleTypeChange={handleTypeChange}
            />
          </div>
        </div>
        {isAdvancedSearchShow.value && <div class='advanced-search'>高级搜索</div>}
        <div class='main-container'>{renderComponent(activeType.value)}</div>
      </div>
    );
  },
});
