import { PropType, defineComponent, ref } from 'vue';
import { Dialog, SearchSelect, Tree } from 'bkui-vue';
import './index.scss';

export default defineComponent({
  name: 'SchemeUserProportionShowDialog',
  props: {
    isShow: {
      type: Boolean as PropType<boolean>,
      default: false,
    },
  },
  emits: ['update:isShow'],
  setup(props, ctx) {
    const toggleShow = (isShow: boolean) => {
      ctx.emit('update:isShow', isShow);
    };
    const treeData = ref([
      {
        name: '方案成熟',
        isOpen: true,
        content:
          '拥有支撑数百款腾讯业务的经验沉淀，兼容各种复杂的系统架构，生于运维 · 精于运维',
        id: '/',
        children: [
          {
            name: 'child-1-方案成熟-拥有支撑数百款腾讯业务的经验沉淀，兼容各种复杂的系统架构，生于运维 · 精于运维',
            content:
              '拥有支撑数百款腾讯业务的经验沉淀，兼容各种复杂的系统架构，生于运维 · 精于运维',
            children: [],
            __uuid: '7dd80a5e-f43e-476a-9b88-0a08f8801162',
          },
          {
            name: 'child-1-覆盖全面',
            content:
              '从配置管理，到作业执行、任务调度和监控自愈，再通过运维大数据分析辅助运营决策，全方位覆盖业务运营的全周期保障管理。',
            children: [],
            __uuid: '9195461a-24cd-46b1-9d7d-08234e7279be',
          },
          {
            name: 'child-1-开放平台',
            content:
              '开放的PaaS，具备强大的开发框架和调度引擎，以及完整的运维开发培训体系，助力运维快速转型升级。',
            children: [
              {
                name: 'child-1-方案成熟',
                content:
                  '拥有支撑数百款腾讯业务的经验沉淀，兼容各种复杂的系统架构，生于运维 · 精于运维',
                children: [],
                __uuid: 'd137d6b3-313e-49df-8e62-9b0549e5906a',
              },
              {
                name: 'child-1-覆盖全面',
                content:
                  '从配置管理，到作业执行、任务调度和监控自愈，再通过运维大数据分析辅助运营决策，全方位覆盖业务运营的全周期保障管理。',
                children: [],
                __uuid: 'ff19cff2-8ab5-4101-a823-63b5fd7c8e73',
              },
              {
                name: 'child-1-开放平台',
                isOpen: true,
                content:
                  '开放的PaaS，具备强大的开发框架和调度引擎，以及完整的运维开发培训体系，助力运维快速转型升级。',
                children: [],
                __uuid: '1b9effd2-d3b4-40be-a5ee-eb30242e3654',
              },
            ],
            __uuid: '69a6edec-7a75-432f-b1a1-2162e8c42293',
          },
        ],
        __uuid: '272197ba-2571-455d-8e89-9edc821f4b8b',
      },
      {
        name: '覆盖全面',
        content:
          '从配置管理，到作业执行、任务调度和监控自愈，再通过运维大数据分析辅助运营决策，全方位覆盖业务运营的全周期保障管理。',
        id: '//',
        children: [
          {
            name: 'child-2-方案成熟',
            content:
              '拥有支撑数百款腾讯业务的经验沉淀，兼容各种复杂的系统架构，生于运维 · 精于运维',
            children: [],
            __uuid: 'c3193bf8-10d5-4b7b-9ddf-5056e79c0024',
          },
          {
            name: 'child-2-覆盖全面',
            content:
              '从配置管理，到作业执行、任务调度和监控自愈，再通过运维大数据分析辅助运营决策，全方位覆盖业务运营的全周期保障管理。',
            children: [],
            __uuid: 'ea53c5da-a150-4164-bafd-06372db9a7d9',
          },
          {
            name: 'child-2-开放平台',
            content:
              '开放的PaaS，具备强大的开发框架和调度引擎，以及完整的运维开发培训体系，助力运维快速转型升级。',
            children: [],
            checked: true,
            __uuid: 'e281b5ca-fdfc-487d-b94b-1d481bd75cb3',
          },
        ],
        __uuid: 'ee578f5a-3928-4077-b4f8-0e714cf8fda3',
      },
      {
        name: '开放平台',
        content:
          '开放的PaaS，具备强大的开发框架和调度引擎，以及完整的运维开发培训体系，助力运维快速转型升级。',
        children: [
          {
            name: 'child-3-方案成熟',
            content:
              '拥有支撑数百款腾讯业务的经验沉淀，兼容各种复杂的系统架构，生于运维 · 精于运维',
            children: [],
            __uuid: 'a9073e51-7702-4f2c-9d5f-f64d84a40680',
          },
          {
            name: 'child-3-覆盖全面',
            content:
              '从配置管理，到作业执行、任务调度和监控自愈，再通过运维大数据分析辅助运营决策，全方位覆盖业务运营的全周期保障管理。',
            children: [],
            __uuid: '4a670b99-9a38-4624-94b7-486b7dac89a0',
          },
          {
            name: 'child-3-开放平台',
            content:
              '开放的PaaS，具备强大的开发框架和调度引擎，以及完整的运维开发培训体系，助力运维快速转型升级。',
            children: [],
            __uuid: '62267a43-bf35-4eaf-8df1-42e463059740',
          },
        ],
        __uuid: 'b334cc34-c5b0-4c25-b926-33aaadb82e7e',
      },
    ]);

    return () => (
      <Dialog
        dialogType='show'
        class='user-proportion-detail-dialog'
        isShow={props.isShow}
        title='分布权重占比'
        onClosed={() => toggleShow(false)}>
        <div class='tips-wrap mb16'>
          <i class='hcm-icon bkhcm-icon-info-line'></i>
          <div class='tips-text'>占比权重说明说明说明</div>
        </div>
        <SearchSelect
          class='mb16'
          modelValue={[]}
          data={[]}
          placeholder='请输入'
        />
        <Tree
          data={treeData.value}
          label='name'
          children='children'
          search=''
          show-node-type-icon={false}
          prefixIcon={(params: any, renderType: any) => {
            if (params.children.length === 0) return null;
            console.log(params, renderType);
            return params.isOpen ? (
              <i class='hcm-icon bkhcm-icon-minus-circle'></i>
            ) : (
              <i class='hcm-icon bkhcm-icon-plus-circle'></i>
            );
          }}>
          {{
            nodeAppend: () => <span class='proportion-num'>10</span>,
          }}
        </Tree>
      </Dialog>
    );
  },
});
