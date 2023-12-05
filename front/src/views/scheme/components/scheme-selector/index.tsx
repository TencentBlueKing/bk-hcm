import { defineComponent, PropType, reactive, ref, withModifiers  } from "vue";
import { useRouter } from 'vue-router';
import { ArrowsLeft, AngleUpFill, EditLine } from "bkui-vue/lib/icon";
import { Popover } from "bkui-vue";
import { useSchemeStore } from "@/store";
import { ISchemeSelectorItem } from '@/typings/scheme';
import { DEPLOYMENT_ARCHITECTURE_MAP } from '@/constants';
import SchemeEditDialog from "../scheme-edit-dialog";
import CloudServiceTag from "../cloud-service-tag";

import './index.scss';

export default defineComponent({
  name: 'scheme-selector',
  emits: ['update'],
  props: {
    schemeList: Array as PropType<ISchemeSelectorItem[]>,
    schemeListLoading: Boolean,
    showEditIcon: Boolean,
    schemeData: Object,
    selectFn: Function,
    onBack: {
      type: Function,
      required: false,
    }
  },
  setup (props, ctx) {
    const schemeStore = useSchemeStore();
    const router = useRouter();

    const isSelectorOpen = ref(false);
    const isEditDialogOpen = ref(false);
    let editedSchemeData = reactive({});

    const handleBack = () => {
      if (typeof props.onBack === 'function') {
        props.onBack();
      } else {
        router.push({ name: 'scheme-list' });
      }
    }

    const handleSelect = (scheme: ISchemeSelectorItem) => {
      if (scheme.id !== props.schemeData.id) {
        if (typeof props.selectFn === 'function') {
          props.selectFn(scheme)
        } else {
          router.push({ name: 'scheme-detail', query: { sid: scheme.id } })
        }
      }
    };

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
          <ArrowsLeft class="back-icon" onClick={props.onBack || goToSchemeList} />
          <Popover
            extCls="resource-selection-scheme-list-popover"
            theme="light"
            placement="bottom-start"
            trigger="click"
            arrow={false}
            onAfterShow={() => { isSelectorOpen.value = true }}
            onAfterHidden={() => { isSelectorOpen.value = false }}>
            {{
              default: () => (
                <div class={['selector-trigger', isSelectorOpen.value ? 'opened' : '']}>
                  <div class="scheme-name">{props.schemeData.name}</div>
                  <AngleUpFill class="arrow-icon" />
                </div>
              ),
              content: () => (
                <div class="scheme-list">
                  {
                    props.schemeListLoading
                      ?
                      <bk-loading loading={true}/>
                      :
                      props.schemeList.map(scheme => {
                        return (
                          <div
                            class={['scheme-item', scheme.id === props.schemeData.id ? 'actived' : '']}
                            onClick={() => { handleSelect(scheme) }}>
                            <div class="scheme-name-area">
                              <div class="name-text">{scheme.name}</div>
                              <div class="tag-list">
                                {
                                  Array.isArray(scheme.deployment_architecture)
                                    ? scheme.deployment_architecture?.map((item) => {
                                      return (<div class="tag-item deploy-type-tag" key={item}>{ DEPLOYMENT_ARCHITECTURE_MAP[item] }</div>);
                                    })
                                    : (
                                      <div class="tag-item deploy-type-tag">{ DEPLOYMENT_ARCHITECTURE_MAP[scheme.deployment_architecture] }</div>
                                    )
                                }
                                {
                                  scheme.vendors?.map(item => {
                                    return (<CloudServiceTag class="tag-item" key={item} type={item} showIcon={true} />)
                                  })
                                }
                              </div>
                            </div>
                            <div class="score-area">
                              <div class="score-item">
                                <span class="label">综合评分：</span>
                                <span class="value">{scheme.composite_score}</span>
                              </div>
                              <div class="score-item">
                                <span class="label">网络评分：</span>
                                <span class="value">{scheme.net_score}</span>
                              </div>
                              <div class="score-item">
                                <span class="label">方案成本：</span>
                                <span class="value">$ {scheme.cost_score}</span>
                              </div>
                            </div>
                          </div>
                        )
                      }) 
                  }
                </div>
              )
            }}
          </Popover>
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
