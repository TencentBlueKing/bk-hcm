import { computed, defineComponent, PropType, ref, watch } from 'vue';
import { Ediatable } from '@blueking/ediatable';
import './index.scss';
import { VendorEnum } from '@/common/constant';
import { useVendorHandler } from '../vendors/useVendorHanlder';
import { cloneDeep } from 'lodash';
import { cleanObject, random } from '../vendors/util';
import { useResourceStore } from '@/store';
import { Message } from 'bkui-vue';

export default defineComponent({
  props: {
    vendor: {
      type: String as PropType<VendorEnum>,
      default: VendorEnum.TCLOUD,
      required: true,
    },
    templateData: Object as PropType<{
      ipList: Array<string>;
      ipGroupList: Array<string>;
    }>,
    relatedSecurityGroups: Array as PropType<Array<Object>>,
    id: String as PropType<string>,
    activeType: String as PropType<'ingress' | 'engress'>,
    isEdit: Boolean as PropType<boolean>,
  },
  setup(props, { expose }) {
    const resourceStore = useResourceStore();
    const { handler } = useVendorHandler(props.vendor, props.activeType);
    const instances = [ref()];
    const tableData = ref(props.isEdit ? [resourceStore.securityRuleDetail] : [handler.value.Record()]);

    const handleSubmit = async () => {
      const items = await Promise.all(instances.map((ins) => ins.value.getValue()));
      items.map((item) => handler.value.handleData(item));
      const promise = props.isEdit
        ? resourceStore.update(
            `vendors/${props.vendor}/security_groups/${props.id}/rules`,
            cleanObject(items[0]),
            items[0].id,
          )
        : resourceStore.add(`vendors/${props.vendor}/security_groups/${props.id}/rules/create`, {
            [`${props.activeType}_rule_set`]: items,
          });
      await promise;
      Message({
        message: '添加成功',
        theme: 'success',
      });
    };

    const handleAdd = () => {
      tableData.value.push(handler.value.Record());
      instances.push(ref());
    };

    const handleRemove = (idx: number) => {
      tableData.value.splice(idx, 1);
      instances.splice(idx, 1);
    };

    const handleCopy = (idx: number) => {
      tableData.value.push(handler.value?.preHandle(cloneDeep({ ...tableData.value[idx], key: random() })));
      instances.push(ref());
    };

    const computedTitles = computed(() =>
      handler.value.titles.filter(({ title }) => (props.isEdit && title !== '操作') || !props.isEdit),
    );

    expose({
      handleSubmit,
    });

    watch(
      [props.isEdit, resourceStore.securityRuleDetail],
      () => {
        tableData.value = [resourceStore.securityRuleDetail];
      },
      {
        immediate: true,
      },
    );

    return () => (
      <Ediatable thead-list={computedTitles.value}>
        {{
          data: tableData.value
            .map((item) => handler.value.preHandle(item))
            .map((item, idx) => (
              <handler.value.row
                value={item}
                key={item.key}
                vendor={props.vendor}
                {...props}
                ref={instances[idx]}
                onAdd={handleAdd}
                onCopy={() => handleCopy(idx)}
                onRemove={() => handleRemove(idx)}
                onChange={(val) => (tableData.value[idx] = val)}
                removeable={tableData.value.length > 1}
              />
            )),
        }}
      </Ediatable>
    );
  },
});
