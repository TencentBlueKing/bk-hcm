import { defineComponent, reactive, ref } from "vue";
import {  useRouter } from 'vue-router';
import { ArrowsLeft, AngleUpFill, EditLine } from "bkui-vue/lib/icon";
import { useSchemeStore } from "@/store";
import SchemeEditDialog from "../scheme-edit-dialog";

import './index.scss';

export default defineComponent({
  name: 'scheme-selector',
  emits: ['update'],
  props: {
    schemeList: Array,
    showEditIcon: Boolean,
    schemeData: Object,
  },
  setup (props, ctx) {
    const schemeStore = useSchemeStore();
    const router = useRouter();

    const isEditDialogOpen = ref(false);
    let editedSchemeData = reactive({})

    const goToSchemeList = () => {
      router.push({ name: 'scheme-list' });
    }

    const saveSchemeFn = (data:{ name: string; bk_biz_id: number; }) => {
      editedSchemeData = data;
      return schemeStore.updateCloudSelectionScheme(props.schemeData.id, data);
    };

    const handleConfirm = () => {
      isEditDialogOpen.value = false;
      ctx.emit('update', editedSchemeData);
    }

    return () => (
      <>
        <div class="scheme-selector">
          <ArrowsLeft class="back-icon" onClick={goToSchemeList} />
          <div class="scheme-name">{props.schemeData.name}</div>
          <AngleUpFill class="arrow-icon" />
          {
            props.showEditIcon ? 
              (<div class="edit-btn" onClick={() => { isEditDialogOpen.value = true }}>
                <EditLine class="edit-icon" />
                编辑
              </div>)
              : null
          }
        </div>
        <SchemeEditDialog
          v-model:show={isEditDialogOpen.value}
          title="编辑方案"
          schemeData={props.schemeData}
          confirmFn={saveSchemeFn}
          onConfirm={handleConfirm} />
      </>
    )
  },
});
