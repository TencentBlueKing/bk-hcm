import { defineComponent, PropType } from 'vue';

import StepDialog from '@/components/step-dialog/step-dialog';

export default defineComponent({
  components: {
    StepDialog,
  },

  props: {
    title: {
      type: String,
    },
    isShow: {
      type: Boolean,
    },
    isDeleting: {
      type: Boolean,
    },
    data: {
      type: Array,
    },
    columns: {
      type: Array as PropType<any[]>,
    },
  },

  emits: ['update:isShow', 'confirm', 'close'],

  setup(_, { emit }) {
    // 方法
    const handleClose = () => {
      emit('update:isShow', false);
      emit('close');
    };

    const handleConfirm = () => {
      emit('confirm');
    };

    return {
      handleClose,
      handleConfirm,
    };
  },

  render() {
    const steps = [
      {
        component: () => (
          <>
            <bk-table
              class='mb20'
              row-hover='auto'
              columns={this.columns.filter((column) => !column.onlyShowOnList)}
              data={this.data}
              show-overflow-tooltip
            />
            <h3 class='g-resource-tips'>{this.$slots.tips ?? this.$slots?.default?.()}</h3>
          </>
        ),
        isConfirmLoading: this.isDeleting,
      },
    ];

    return (
      <>
        <step-dialog
          title={this.title}
          isShow={this.isShow}
          steps={steps}
          onConfirm={this.handleConfirm}
          onCancel={this.handleClose}></step-dialog>
      </>
    );
  },
});
