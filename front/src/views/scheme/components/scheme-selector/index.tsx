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
    showEditIcon: Boolean,
    schemeData: Object,
  },
  setup (props, ctx) {
    const schemeStore = useSchemeStore();
    const router = useRouter();

    const isSelectorOpen = ref(false);
    const isEditDialogOpen = ref(false);
    let editedSchemeData = reactive({});

    const goToSchemeList = () => {
      router.push({ name: 'scheme-list' });
    }

    const handleSelect = (id: string) => {
      if (id !== props.schemeData.id) {
        router.push({ name: 'scheme-detail', query: { sid: id } })
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
          <ArrowsLeft class="back-icon" onClick={goToSchemeList} />
          <Popover
            extCls="resource-selection-scheme-list-popover"
            theme="light"
            placement="bottom-start"
            trigger="click"
            arrow={false}>
            {{
              default: () => (
                <div class={['selector-trigger', isSelectorOpen.value ? 'opened' : '']}>
                  <div class="scheme-name">{props.schemeData.name}</div>
                  <AngleUpFill class="arrow-icon" />
                </div>
              ),
              content: () => (
                <div class="scheme-list">
                  { props.schemeList.map(scheme => {
                    return (
                      <div
                        class={['scheme-item', scheme.id === props.schemeData.id ? 'actived' : '']}
                        onClick={() => { handleSelect(scheme.id) }}>
                        <div class="scheme-name-area">
                          <div class="name-text">{scheme.name}</div>
                          <div class="tag-list">
                            {
                              scheme.deployment_architecture?.map(item => {
                                return (<div class="deploy-type-tag" key={item}>{ DEPLOYMENT_ARCHITECTURE_MAP[item] }</div>)
                              })
                            }
                            {
                              scheme.vendors?.map(item => {
                                return (<CloudServiceTag class="cloud-service-type" key={item} type={item} showIcon={true} />)
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
                  }) }
                </div>
              )
            }}
          </Popover>
          {/* <div
            class={['selector-trigger', isSelectorOpen.value ? 'opened' : '']}
            v-bk-tooltips={{
              placement: 'bottom-start',
              arrow: false,
              trigger: 'click',
              theme: 'light',
              extCls: 'resource-selection-scheme-list-popover',
              onShow: () => { isSelectorOpen.value = true },
              onHide: () => { isSelectorOpen.value = false },
              content: (
                <div class="scheme-list" onMousedown={withModifiers(() => { debugger }, ['stop', 'prevent'])}>
                  { props.schemeList.map(scheme => {
                    return (
                      <div class={['scheme-item', scheme.id === props.schemeData.id ? 'actived' : '']}>
                        <div class="scheme-name-area">
                          <div class="name-text">{scheme.name}</div>
                          <div class="tag-list">
                            {
                              scheme.deployment_architecture?.map(item => {
                                return (<div class="deploy-type-tag" key={item}>{ DEPLOYMENT_ARCHITECTURE_MAP[item] }</div>)
                              })
                            }
                            {
                              scheme.vendors?.map(item => {
                                return (<CloudServiceTag class="cloud-service-type" key={item} type={item} showIcon={true} />)
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
                  }) }
                </div>
              )
            }}>
            <div class="scheme-name">{props.schemeData.name}</div>
            <AngleUpFill class="arrow-icon" />
          </div> */}
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
