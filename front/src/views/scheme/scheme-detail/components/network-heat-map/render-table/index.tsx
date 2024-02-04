import { defineComponent, ref, watch, PropType } from 'vue';
import { DownShape } from 'bkui-vue/lib/icon';
import { IIdcInfo, IIdcServiceAreaRel } from '@/typings/scheme';
import { getScoreColor } from '@/common/util';

import './index.scss';

interface ITableDataItem {
  [key: string]: any;
  children?: ITableDataItem[];
}

export default defineComponent({
  name: 'RenderTableComp',
  props: {
    idcList: Array as PropType<IIdcInfo[]>,
    data: Array as PropType<{ [key: string]: string | number | boolean }[]>,
    searchStr: String,
    isHighlight: Boolean,
    highlightArea: Array as PropType<IIdcServiceAreaRel[]>,
  },
  emits: ['toggleFold'],
  setup(props, ctx) {
    const tableData = ref<ITableDataItem[]>([...props.data]);

    watch(
      () => props.data,
      (val) => {
        tableData.value = val.slice(0);
      },
    );

    watch(
      () => props.searchStr,
      (val) => {
        if (val) {
          const result: ITableDataItem[] = [];
          props.data.forEach((country) => {
            if ((country.rowName as string).toLowerCase().includes(val.toLowerCase())) {
              result.push(country);
            } else if (Array.isArray(country.children)) {
              const children = country.children.filter((item) =>
                item.rowName.toLowerCase().includes(val.toLowerCase()),
              );
              if (children.length) {
                result.push({
                  ...country,
                  children,
                });
              }
            }
          });
          tableData.value = result;
        } else {
          tableData.value = [...props.data];
        }
      },
    );

    const renderValCell = (val: ITableDataItem, idc: IIdcInfo, countryName = '') => {
      const cellValue = Number(val[idc.name]);

      let cls = '';
      if (props.isHighlight) {
        const area = props.highlightArea.find((area) => area.idc_id === idc.id);
        if (!area) {
          cls = 'no-highlight-cell';
        } else {
          if (val.isCountry) {
            const countryData = props.data.find((item) => item.rowName === val.rowName);
            if (countryData && Array.isArray(countryData.children)) {
              cls = countryData.children.some((region) => {
                return area.service_areas?.some((item) => item.province_name === region.rowName);
              })
                ? 'highlight-cell'
                : 'no-highlight-cell';
            } else {
              cls = 'no-highlight-cell';
            }
          } else {
            cls = area.service_areas?.some((item) => {
              return item.province_name === val.rowName && countryName === item.country_name;
            })
              ? 'highlight-cell'
              : 'no-highlight-cell';
          }
        }
      }
      return (
        <td class='tbody-col'>
          <div class={['cell', cls]} style={{ color: getScoreColor(cellValue) }}>
            {Number.isNaN(cellValue) ? '' : `${cellValue.toFixed(2)}ms`}
          </div>
        </td>
      );
    };

    return () => (
      <div class='render-table-comp'>
        <table>
          <thead>
            <tr>
              <th class='thead-col row-name-col'>
                <div class='cell'>国家 \ IDC</div>
              </th>
              {props.idcList.map((idc) => {
                return <th class='thead-col'>{idc.region}机房</th>;
              })}
            </tr>
          </thead>
          <tbody>
            {props.searchStr && tableData.value.length === 0 ? (
              <tr>
                <td colspan={props.idcList.length + 1}>
                  <bk-exception
                    class='search-empty-exception'
                    type='search-empty'
                    scene='part'
                    description='搜索为空'
                  />
                </td>
              </tr>
            ) : (
              tableData.value.map((item) => {
                return (
                  <>
                    <tr>
                      <td class='tbody-col row-name-col'>
                        <div class='cell country-cell'>
                          <DownShape
                            class={['arrow-icon', item.isFold ? 'fold' : '']}
                            onClick={() => {
                              ctx.emit('toggleFold', item.rowName);
                            }}
                          />
                          <div class='name-text'>{item.rowName}</div>
                        </div>
                      </td>
                      {props.idcList.map((idc) => {
                        return renderValCell(item, idc);
                      })}
                    </tr>
                    {!item.isFold && Array.isArray(item.children)
                      ? item.children.map((subItem) => {
                          return (
                            <tr>
                              <td class='tbody-col row-name-col'>
                                <div class='cell'>
                                  <div class='name-text'>{subItem.rowName}</div>
                                </div>
                              </td>
                              {props.idcList.map((idc) => {
                                return renderValCell(subItem, idc, item.rowName);
                              })}
                            </tr>
                          );
                        })
                      : null}
                  </>
                );
              })
            )}
          </tbody>
        </table>
      </div>
    );
  },
});
