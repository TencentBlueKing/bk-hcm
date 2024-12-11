import { computed, defineComponent, PropType, ref } from 'vue';
import cssModule from './index.module.scss';

import { Button, Input } from 'bkui-vue';
import { Eye } from 'bkui-vue/lib/icon';
import CopyToClipboard from '@/components/copy-to-clipboard/index.vue';

import { useI18n } from 'vue-i18n';
import { useFormItem } from 'bkui-vue/lib/form';

export default defineComponent({
  props: {
    modelValue: { type: String },
    generateFn: { type: Function as PropType<(...args: any) => string> },
  },
  emits: ['update:modelValue'],
  setup(props, { emit }) {
    const { t } = useI18n();
    const { validate } = useFormItem();

    const inputType = ref('password');
    const isGenerate = computed(() => inputType.value === 'text');
    const toggleInputType = (type: 'text' | 'password') => {
      inputType.value = type;
    };

    const pwd = computed({
      get() {
        return props.modelValue;
      },
      set(val) {
        emit('update:modelValue', val);
      },
    });

    // 默认规则：密码长度不少于8位且不多于20位，至少包含一个小写字母、一个大写字母、一个数字和一个特殊符号（仅限@、# 、+、_、-、[、]、{、}
    const defaultGenerateFn = (length = 20) => {
      const upper = 'ABCDEFGHIJKLMNOPQRSTUVWXYZ';
      const lower = 'abcdefghijklmnopqrstuvwxyz';
      const digits = '0123456789';
      const special = '@#+_-[]{}';

      // 确保密码包含至少一个大写字母、小写字母和数字和特殊字符
      let password = [
        upper[Math.floor(Math.random() * upper.length)],
        lower[Math.floor(Math.random() * lower.length)],
        digits[Math.floor(Math.random() * digits.length)],
        special[Math.floor(Math.random() * special.length)],
      ];

      // 随机选择剩余的字符
      const allChars = upper + lower + digits + special;
      for (let i = 4; i < length; i++) {
        password.push(allChars[Math.floor(Math.random() * allChars.length)]);
      }

      // 打乱密码顺序
      password = password.sort(() => Math.random() - 0.5);

      return password.join('');
    };

    const handleClick = () => {
      toggleInputType('text');
      pwd.value = props.generateFn ? props.generateFn() : defaultGenerateFn();
      validate();
    };

    const renderSuffix = computed(() => {
      const suffix = [
        <Button class={cssModule.button} theme='primary' outline onClick={handleClick}>
          {t('自动生成')}
        </Button>,
      ];
      if (isGenerate.value)
        suffix.unshift(
          <>
            <CopyToClipboard
              class={cssModule.copy}
              content={pwd.value}
              v-bk-tooltips={{
                content: t('请将生成的密码妥善保管，如遗失密码，可能无法找回'),
                // showOnInit: true, // todo: 组件库升级至 2.0.1-beta.62 予以支持
              }}
            />
            <Eye class={cssModule.eye} onClick={() => toggleInputType('password')} />
          </>,
        );
      return suffix;
    });

    return () => (
      <Input
        // 设置key，规避组件内部状态（如pwdVisible）对于交互的影响
        key={inputType.value}
        class={[cssModule.wrapper, { [cssModule['render-custom-suffix']]: isGenerate.value }]}
        v-model={pwd.value}
        type={inputType.value}>
        {{
          suffix: () => renderSuffix.value,
        }}
      </Input>
    );
  },
});
