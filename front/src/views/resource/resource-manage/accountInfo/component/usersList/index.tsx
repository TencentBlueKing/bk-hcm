import { defineComponent } from 'vue';
// import custom hooks
import useRenderTable from './hooks/useRenderTable';
import useRenderForm from './hooks/useRenderForm';
import './index.scss';

export default defineComponent({
  setup() {
    // use custom hooks
    const { CommonTable, getListData } = useRenderTable();
    const { RenderForm } = useRenderForm(getListData);

    return () => (
      <div class='user-list-module'>
        <CommonTable></CommonTable>
        <RenderForm></RenderForm>
      </div>
    );
  },
});
