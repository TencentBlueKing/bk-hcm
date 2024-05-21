import { defineComponent } from 'vue';
import { Tag } from 'bkui-vue';
import RsConfigTable from '../rs-config-table';
import './index.scss';

export default defineComponent({
  name: 'RsSidesliderContent',
  props: {
    selectedTargetGroups: {
      type: Array<any>,
    },
  },
  emits: ['showAddRsDialog'],
  setup(props, { emit }) {
    return () => (
      <div class='rs-sideslider-content'>
        <div class='selected-target-groups'>
          <span class='label'>已选择目标组</span>
          <div>
            {props.selectedTargetGroups.map((item) => (
              <Tag>{item.target_group_name}</Tag>
            ))}
          </div>
        </div>
        <RsConfigTable onShowAddRsDialog={() => emit('showAddRsDialog')} noSearch />
      </div>
    );
  },
});
