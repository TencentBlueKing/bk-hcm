import { defineComponent, ref } from 'vue';
// import components
import CommonSideslider from '@/components/common-sideslider';
import RsConfigTable from '../RsConfigTable';
// import utils
import bus from '@/common/bus';
import './index.scss';

export default defineComponent({
  name: 'BatchAddRsSideslider',
  setup() {
    const isShow = ref(false);

    const handleSubmit = () => {};

    return () => (
      <CommonSideslider title='批量添加 RS' width={960} v-model:isShow={isShow.value} onHandleSubmit={handleSubmit}>
        <div class='rs-sideslider-content'>
          <div class='selected-target-groups'>
            <span class='label'>已选择目标组</span>
            <div>
              {/* {props.selectedTargetGroups.map((item) => (
            <Tag>{item.target_group_name}</Tag>
          ))} */}
            </div>
          </div>
          <RsConfigTable onShowAddRsDialog={() => bus.$emit('showAddRsDialog')} noSearch />
        </div>
      </CommonSideslider>
    );
  },
});
