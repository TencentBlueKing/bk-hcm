import { defineComponent, ref, onMounted } from 'vue';
import { useFieldVal } from '@/views/ziyanScr/cvm-produce/component/property-display/field-map';
import { QueryFilterType, QueryRuleOPEnum } from '@/typings';
import rollRequest from '@blueking/roll-request';
import http from '@/http';

interface ICvmSubnetItem {
  id: number;
  region: string;
  zone: string;
  vpc_id: string;
  vpc_name: string;
  subnet_id: string;
  subnet_name: string;
  enable: boolean;
  comment: string;
  [key: string]: any;
}

export default defineComponent({
  props: {
    k: { type: String },
    v: { type: [Object, String, Number, Array, Boolean] },
    kVisible: { type: Boolean, default: true },
    row: Object,
  },
  setup(props) {
    const { getFieldCn, getFieldCnVal } = useFieldVal();

    const reqKeyList = ['vpc', 'subnet'];
    const keyMap: Record<string, { reqKey: string; resKey: string; nameKey: string }> = {
      vpc: { reqKey: 'cloud_vpc_id', resKey: 'vpc_id', nameKey: 'vpc_name' },
      subnet: { reqKey: 'cloud_id', resKey: 'subnet_id', nameKey: 'subnet_name' },
    };

    const displayName = ref('');
    const getDisplayName = async (k: string, v: string) => {
      if (!v) return;
      const { spec } = props.row;
      const { region, zone } = spec;

      const filter: QueryFilterType = {
        op: 'and',
        rules: [
          { field: 'region', op: QueryRuleOPEnum.EQ, value: region },
          { field: 'zone', op: QueryRuleOPEnum.EQ, value: zone },
          { field: keyMap[k].reqKey, op: QueryRuleOPEnum.EQ, value: v },
        ],
      };

      const list = (await rollRequest({
        httpClient: http,
        pageEnableCountKey: 'count',
      }).rollReqUseCount(
        '/api/v1/woa/config/findmany/config/cvm/subnet/list',
        { filter },
        { limit: 500, countGetter: (res) => res.data.count, listGetter: (res) => res.data.info },
      )) as ICvmSubnetItem[];

      displayName.value = list.find((item) => item[keyMap[k].resKey] === v)?.[keyMap[k].nameKey] || '';
    };

    onMounted(() => {
      if (reqKeyList.includes(props.k)) {
        getDisplayName(props.k, props.v as string);
      }
    });

    return () => (
      <div class='cvm-produce-property-item'>
        {props.kVisible ? <div>{getFieldCn(props.k)}：</div> : null}
        <div class='cvm-produce-property-value'>
          {getFieldCnVal(props.k, props.v, props.row)}
          {displayName.value ? `(${displayName.value})` : ''}
        </div>
      </div>
    );
  },
});
