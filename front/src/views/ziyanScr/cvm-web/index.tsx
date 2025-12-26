import { defineComponent, ref, computed, onMounted } from 'vue';
import useSelection from '@/views/resource/resource-manage/hooks/use-selection';
import useColumns from '@/views/resource/resource-manage/hooks/use-scr-columns';
import { useTable } from '@/hooks/useTable/useTable';
import AreaSelector from '../hostApplication/components/AreaSelector';
import ZoneSelector from '../hostApplication/components/ZoneSelector';
import { Search, AngleDown } from 'bkui-vue/lib/icon';
import { updateSubnetProperties } from '@/api/host/config-management';
import './index.scss';

export default defineComponent({
  components: {
    AreaSelector,
    ZoneSelector,
  },
  setup() {
    const defaultCvmWebForm = () => ({
      region: [],
      zone: [],
      // vpc_name: '',
      cloud_vpc_id: '',
      cloud_id: '',
      name: '',
      enable: '',
    });
    const cvmWebForm = ref<Record<string, any>>(defaultCvmWebForm());
    const vpcFilterType = ref('cloud_vpc_id');
    const vpcLabel = ref('VPC ID');
    const subnetFilterType = ref('name');
    const subnetLabel = ref('Subnet 名');
    const vpcList = [
      // {
      //   label: 'VPC 名',
      //   value: 'vpc_name',
      // },
      {
        label: 'VPC ID',
        value: 'cloud_vpc_id',
      },
    ];
    const subnetList = [
      {
        label: 'Subnet 名',
        value: 'name',
      },
      {
        label: 'Subnet ID',
        value: 'cloud_id',
      },
    ];
    const useList = [
      { label: '是', value: true },
      { label: '否', value: false },
    ];
    const { selections, handleSelectionChange, resetSelections } = useSelection();
    const { columns } = useColumns('cvmWebQuery');
    const tableColumns = [
      ...columns,
      {
        label: '启用',
        field: 'enable',
        width: 80,
        render: ({ row }) => {
          const handleSave = (val) => {
            const params = {
              ids: [row.id],
              properties: {
                enable: val,
              },
            };
            return updateSubnetProperty(params).then(() => {
              row.enable = val;
            });
          };
          return <edit-item type='boolean' modelValue={row.enable} save={handleSave} />;
        },
      },
      {
        label: '备注',
        field: 'comment',
        render: ({ row }) => {
          const handleSave = (val) => {
            const params = {
              ids: [row.id],
              properties: {
                comment: val,
              },
            };
            return updateSubnetProperty(params).then(() => {
              row.comment = val;
            });
          };
          return (
            <edit-item
              type='textarea'
              control-attrs={{ maxlength: 128, showWordLimit: true }}
              modelValue={row.comment}
              save={handleSave}
            />
          );
        },
      },
    ];
    const requestParams = ref({
      page: {
        start: 0,
        limit: 10,
        count: false,
      },
      filter: {
        op: 'and',
        rules: [],
      },
    });
    const { CommonTable, getListData } = useTable({
      tableOptions: {
        columns: tableColumns,
        extra: {
          onSelect: (selections: any) => {
            handleSelectionChange(selections, () => true, false);
          },
          onSelectAll: (selections: any) => {
            handleSelectionChange(selections, () => true, true);
          },
        },
      },
      requestOption: {
        dataPath: 'data.info',
        sortOption: {
          legacy: false,
        },
      },
      scrConfig: () => {
        return {
          payload: {
            ...requestParams.value,
          },
          url: '/api/v1/woa/config/findmany/config/cvm/subnet/list',
          pageEnableCountKey: 'count',
          clearRules: false,
        };
      },
    });
    const paramRules = computed(() => {
      const { enable } = cvmWebForm.value;
      const rules = [];
      ['vpc_name', 'cloud_vpc_id', 'cloud_id', 'name'].map((item) => {
        if (cvmWebForm.value[item]) {
          rules.push({
            field: item,
            op: 'cs',
            value: cvmWebForm.value[item],
          });
        }
        return null;
      });
      ['region', 'zone'].map((item) => {
        if (Array.isArray(cvmWebForm.value[item]) && cvmWebForm.value[item].length) {
          rules.push({
            field: item,
            op: 'in',
            value: cvmWebForm.value[item],
          });
        }
        return null;
      });
      if (String(enable))
        rules.push({
          field: 'extension.enable_cvm',
          op: 'json_eq',
          value: String(enable),
        });
      return rules;
    });
    const getCvmWeblist = () => {
      requestParams.value.filter.rules = paramRules.value;
      resetSelections();
      getListData();
    };
    const filterOrders = () => {
      getCvmWeblist();
    };
    const clearFilter = () => {
      cvmWebForm.value = defaultCvmWebForm();
      filterOrders();
    };
    const isShow = ref(false);
    const updateForm = ref({
      enable: '',
      comment: '',
    });
    const handleBatchEdit = () => {
      isShow.value = true;
    };
    const handleBatchEditSubmit = () => {
      const params = {
        ids: selections.value.map((item) => item.id),
      };
      const properties = {};
      const { enable, comment } = updateForm.value;
      if (String(enable)) properties.enable = enable;
      if (comment) properties.comment = comment;
      // 如果 properties 是空对象则不发送请求
      if (Object.keys(properties).length === 0) return;
      updateSubnetProperties({ ...params, properties }, {}).then(() => {
        handleCancel();
        filterOrders();
      });
    };
    const handleCancel = () => {
      isShow.value = false;
      updateForm.value = {
        enable: '',
        comment: '',
      };
    };
    const updateSubnetProperty = (params) => {
      return updateSubnetProperties(params, {});
    };
    const getPrefix = (flag: string, item) => {
      switch (flag) {
        case 'vpc':
          cvmWebForm.value[vpcFilterType.value] = '';
          vpcFilterType.value = item.value;
          vpcLabel.value = item.label;
          break;
        case 'subnet':
          cvmWebForm.value[subnetFilterType.value] = '';
          subnetFilterType.value = item.value;
          subnetLabel.value = item.label;
          break;
        default:
          break;
      }
    };
    onMounted(() => {});
    return () => (
      <div class='apply-list-container cvm-web-wrapper'>
        <div class={'filter-container'}>
          <bk-form formType='vertical' class={'scr-form-wrapper'} model={cvmWebForm}>
            <bk-form-item label='地域'>
              <area-selector multiple v-model={cvmWebForm.value.region} params={{ resourceType: 'QCLOUDCVM' }} />
            </bk-form-item>
            <bk-form-item label='园区'>
              <zone-selector
                multiple
                v-model={cvmWebForm.value.zone}
                params={{ resourceType: 'QCLOUDCVM', region: cvmWebForm.value.region }}
              />
            </bk-form-item>
            <bk-form-item label='VPC'>
              <bk-input v-model={cvmWebForm.value[vpcFilterType.value]} placeholder='支持模糊匹配'>
                {{
                  prefix: () => (
                    <bk-dropdown>
                      {{
                        default: () => (
                          <div class='menu-item'>
                            <span>{vpcLabel.value}</span>
                            <AngleDown />
                          </div>
                        ),
                        content: () => (
                          <bk-dropdown-menu>
                            {vpcList.map((item) => {
                              return (
                                <bk-dropdown-item key={item.value} onClick={() => getPrefix('vpc', item)}>
                                  {item.label}
                                </bk-dropdown-item>
                              );
                            })}
                          </bk-dropdown-menu>
                        ),
                      }}
                    </bk-dropdown>
                  ),
                }}
              </bk-input>
            </bk-form-item>
            <bk-form-item label='Subnet'>
              <bk-input v-model={cvmWebForm.value[subnetFilterType.value]} placeholder='支持模糊匹配'>
                {{
                  prefix: () => (
                    <bk-dropdown popover-options={{ boundary: 'body' }}>
                      {{
                        default: () => (
                          <div class='menu-item'>
                            <span>{subnetLabel.value}</span>
                            <AngleDown />
                          </div>
                        ),
                        content: () => (
                          <bk-dropdown-menu>
                            {subnetList.map((item) => {
                              return (
                                <bk-dropdown-item key={item.value} onClick={() => getPrefix('subnet', item)}>
                                  {item.label}
                                </bk-dropdown-item>
                              );
                            })}
                          </bk-dropdown-menu>
                        ),
                      }}
                    </bk-dropdown>
                  ),
                }}
              </bk-input>
            </bk-form-item>
            <bk-form-item label='启用'>
              <bk-select v-model={cvmWebForm.value.enable} clearable allowEmptyValues={[false]}>
                {useList.map(({ label, value }) => {
                  return <bk-option key={value} name={label} id={value}></bk-option>;
                })}
              </bk-select>
            </bk-form-item>
          </bk-form>
          <div class='btn-container'>
            <bk-button theme='primary' onClick={filterOrders}>
              <Search />
              查询
            </bk-button>
            <bk-button onClick={clearFilter}>重置</bk-button>
          </div>
        </div>
        <div class='btn-container oper-btn-pad'>
          <bk-button disabled={!selections.value.length} onClick={handleBatchEdit}>
            批量更新
          </bk-button>
        </div>
        <CommonTable class={'filter-common-table'} />
        <bk-dialog v-model:is-show={isShow.value} width='600' title='批量更新'>
          {{
            default: () => (
              <bk-form label-width='110' model={updateForm}>
                <bk-form-item label='启用'>
                  <bk-select v-model={updateForm.value.enable} clearable allowEmptyValues={[false]}>
                    <bk-option name='保持不变' id=''></bk-option>
                    {useList.map(({ label, value }) => {
                      return <bk-option key={value} label={label} id={value}></bk-option>;
                    })}
                  </bk-select>
                </bk-form-item>
                <bk-form-item label='备注'>
                  <bk-input v-model={updateForm.value.comment} placeholder='请输入备注' type='textarea' />
                </bk-form-item>
              </bk-form>
            ),
            footer: () => (
              <div class='dialog-footer-btn'>
                <bk-button theme='primary' onClick={handleBatchEditSubmit}>
                  提交
                </bk-button>
                <bk-button onClick={handleCancel}>取消</bk-button>
              </div>
            ),
          }}
        </bk-dialog>
      </div>
    );
  },
});
