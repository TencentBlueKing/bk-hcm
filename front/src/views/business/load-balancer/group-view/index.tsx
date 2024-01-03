import { defineComponent, ref } from 'vue';
import { Input, VirtualRender } from 'bkui-vue';
import allIcon from '@/assets/image/all-vendors.png';
import AllGroupsManager from './all-groups-manager';
import SpecificTargetGroupManager from './specific-target-group-manager';
import './index.scss';

export default defineComponent({
  name: 'TargetGroupView',
  setup() {
    const searchValue = ref('');

    // 模拟左侧列表数据
    const renderList = Array(100)
      .fill({})
      .map((_, index: number) => {
        return {
          id: index + 1,
          name: `clb-group-${index + 1}`,
          count: Math.floor(Math.random() * 10) + 1,
        };
      });

    const componentMap = {
      all: <AllGroupsManager />,
      specific: <SpecificTargetGroupManager />,
    };
    const renderComponent = (type: 'all' | 'specific') => {
      return componentMap[type];
    };
    const activeType = ref('all' as 'all' | 'specific');
    const handleTypeChange = (type: 'all' | 'specific') => {
      activeType.value = type;
    };

    return () => (
      <div class='group-view-page'>
        <div class='left-container'>
          <div class='search-wrap'>
            <Input v-model={searchValue.value} type='search' clearable placeholder='搜索目标组' />
          </div>
          <div class='group-list-wrap'>
            <div class='all-groups-wrap' onClick={() => handleTypeChange('all')}>
              <div class='left-wrap'>
                <img src={allIcon} alt='' class='prefix-icon' />
                <span>全部目标组</span>
              </div>
              <div class='right-wrap'>
                <div class='count'>{6654}</div>
              </div>
            </div>
            <VirtualRender list={renderList} height='calc(100% - 36px)' lineHeight={36}>
              {{
                default: ({ data }: any) => {
                  return data.map((item: any) => {
                    return (
                      <div key={item.id} class='group-item-wrap' onClick={() => handleTypeChange('specific')}>
                        <div class='left-wrap'>
                          <img src={allIcon} alt='' class='prefix-icon' />
                          <span>{item.name}</span>
                        </div>
                        <div class='right-wrap'>
                          <div class='count'>{item.count}</div>
                        </div>
                      </div>
                    );
                  });
                },
              }}
            </VirtualRender>
          </div>
        </div>
        <div class='main-container'>{renderComponent(activeType.value)}</div>
      </div>
    );
  },
});
