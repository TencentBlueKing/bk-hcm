import { ModelPropertyColumn } from '@/model/typings';
import { type Settings } from 'bkui-vue/lib/table/props';
import { ref } from 'vue';

type UseTableSettingsOptions = {
  defaults?: ModelPropertyColumn['id'][];
};

export default function useTableSettings(columns: ModelPropertyColumn[], options?: UseTableSettingsOptions) {
  const { defaults = [] } = options || {};
  const settings = ref<Settings>({
    fields: [],
    checked: [],
    trigger: 'manual',
  });

  columns.forEach((col, index) => {
    settings.value.fields.push({
      label: col.name,
      field: col.id,
      disabled: index < 3,
    });
  });

  if (defaults?.length) {
    settings.value.checked = defaults.slice();
  } else {
    settings.value.checked = columns.filter((col) => !col.defaultHidden).map((col) => col.id);
  }

  return {
    settings,
  };
}
