import { defineComponent, PropType } from 'vue';
import { Button, Dialog } from 'bkui-vue';
import CommonLocalTable from '../CommonLocalTable';
import { useI18n } from 'vue-i18n';
import type { IProp } from '@/hooks/useLocalTable';
import './index.scss';

export default defineComponent({
  name: 'BatchOperationDialog',
  props: {
    isSubmitLoading: Boolean,
    isSubmitDisabled: Boolean,
    isShow: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
    title: {
      type: String as PropType<string>,
      required: true,
    },
    theme: {
      type: String as PropType<'primary' | 'warning' | 'success' | 'danger'>,
      default: 'primary',
    },
    confirmText: {
      type: String as PropType<string>,
      default: '确定',
    },
    tableProps: {
      type: Object as PropType<IProp>,
    },
    custom: {
      type: Boolean,
      default: false,
    },
    list: {
      type: Array,
      default: [],
    },
    submitDisabledTooltipsOption: {
      type: Object as PropType<{ content?: string; disabled: boolean }>,
      default: () => ({ disabled: true }),
    },
  },
  emits: ['update:isShow', 'handleConfirm'],
  setup(props, { emit, slots }) {
    // use hooks
    const { t } = useI18n();

    const triggerShow = (isShow: boolean) => {
      emit('update:isShow', isShow);
    };
    const handleConfirm = () => {
      emit('handleConfirm');
    };
    // 默认渲染
    const renderDefaultSlot = () => {
      return (
        <div class='batch-operation-dialog-content'>
          <div class='tips'>{slots.tips?.()}</div>
          <CommonLocalTable
            searchOptions={{ searchData: props.tableProps.searchData }}
            tableOptions={{ rowKey: 'id', columns: props.tableProps.columns }}
            tableData={props.list}
          >
            {{
              operation: () => slots.tab?.(),
            }}
          </CommonLocalTable>
        </div>
      );
    };
    // 自定义渲染
    const renderCustomDefaultSlot = () => slots.default?.();

    return () => (
      <Dialog
        class='batch-operation-dialog'
        width={960}
        isShow={props.isShow}
        title={t(props.title)}
        theme={props.theme}
        quickClose={false}
        onClosed={() => triggerShow(false)}
        confirmText={t(props.confirmText)}
      >
        {{
          default: props.custom ? renderCustomDefaultSlot : renderDefaultSlot,
          footer: () => (
            <>
              <Button
                theme={props.theme}
                onClick={handleConfirm}
                loading={props.isSubmitLoading}
                disabled={props.isSubmitDisabled}
                v-bk-tooltips={props.submitDisabledTooltipsOption}
              >
                {props.confirmText}
              </Button>
              <Button class='dialog-cancel' onClick={() => triggerShow(false)}>
                取消
              </Button>
            </>
          ),
        }}
      </Dialog>
    );
  },
});
