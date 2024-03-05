import { computed, defineComponent, onMounted, ref, watch } from 'vue';
import DetailHeader from '../../../common/header/detail-header';
import { Senarios, useWhereAmI } from '@/hooks/useWhereAmI';
import { useRoute } from 'vue-router';
import useQueryListCommon from '../../../hooks/use-query-list-common';
import { QueryRuleOPEnum } from '@/typings';
import './index.scss';
import DetailInfo from '../../../common/info/detail-info';
import { TemplateType, TemplateTypeMap } from '../../dialog/template-dialog';
import { Table } from 'bkui-vue';

export default defineComponent({
  setup() {
    const { whereAmI } = useWhereAmI();
    const route = useRoute();
    const templateId = ref(route.query?.id);
    const singleTableData = ref([]);
    const fetchUrl = ref('argument_templates/list');
    const multipleFilter = ref({
      filter: {
        op: QueryRuleOPEnum.AND,
        rules: [
          {
            field: 'cloud_id',
            op: QueryRuleOPEnum.IN,
            value: [],
          },
        ],
      },
    });
    const ipColumns = [
      {
        label: '地址',
        field: 'address',
      },
      {
        label: '描述',
        field: 'description',
      },
    ];
    const { datas, getList } = useQueryListCommon(
      {
        filter: {
          op: 'and',
          rules: [
            {
              field: 'cloud_id',
              op: QueryRuleOPEnum.CS,
              value: route.query?.id,
            },
          ],
        },
      },
      fetchUrl,
    );

    const { datas: multipleTableData } = useQueryListCommon(multipleFilter.value, fetchUrl);

    onMounted(() => {
      getList();
    });

    const details = computed(() => {
      return datas.value?.[0] || {};
    });

    watch(
      () => details.value,
      (detail) => {
        console.log(111, detail);
        if (!detail) return;
        if ([TemplateType.IP, TemplateType.PORT].includes(detail.type)) {
          singleTableData.value = detail.templates;
        }
        if ([TemplateType.IP_GROUP, TemplateType.PORT_GROUP].includes(detail.type)) {
          multipleFilter.value.filter.rules = [
            {
              field: 'cloud_id',
              op: QueryRuleOPEnum.IN,
              value: detail.group_templates,
            },
          ];
        }
      },
      {
        deep: true,
        immediate: true,
      },
    );

    return () => (
      <div class={'template-detail-container'}>
        <DetailHeader>参数模板: ID {templateId.value}</DetailHeader>

        <div class={`detial-wrap ${whereAmI.value === Senarios.business ? 'm24' : ''}`}>
          <p class={'title'}>基本信息</p>
          <DetailInfo
            class={'mb16'}
            fields={[
              {
                name: '账号ID',
                prop: 'account_id',
              },
              {
                name: '资源ID',
                prop: 'cloud_id',
              },
              {
                name: '创建时间',
                prop: 'created_at',
              },
              {
                name: '创建者',
                prop: 'creator',
              },
              {
                name: '资源名称',
                prop: 'name',
              },
              {
                name: '备注',
                prop: 'memo',
              },
              {
                name: '更新时间',
                prop: 'updated_at',
              },
              {
                name: '云厂商',
                prop: 'vendor',
              },
            ]}
            detail={details.value}
          />

          <p class={'title'}>{TemplateTypeMap[details?.value?.type]}</p>
          {[TemplateType.IP, TemplateType.PORT].includes(details.value.type) ? (
            <Table columns={ipColumns} data={singleTableData.value} />
          ) : null}
          {
            [TemplateType.IP_GROUP, TemplateType.PORT_GROUP].includes(details.value.type) ? (
              <div>
                {
                  multipleTableData.value.map(({
                    name,
                    cloud_id,
                    templates,
                  }) => (
                    <div>
                      <p class={'subtitle'}>
                        {name} ({cloud_id})
                      </p>
                      <Table columns={ipColumns} data={templates} />
                    </div>
                  ))
                }
              </div>
            ) : null
          }
        </div>
      </div>
    );
  },
});
