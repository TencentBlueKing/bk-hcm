import { Model, Column } from '@/decorator';

@Model()
export class FieldTcloud {
  @Column('string', { apiOnly: true })
  id: string;

  @Column('list', {
    name: '二级账号',
    required: true,
    meta: { display: { props: { idKey: 'id', displayKey: 'name' } } },
  })
  account_id: string;

  @Column('string', {
    name: '模板名称',
    required: true,
    rules: [
      {
        validator: (value: string) => /^[a-zA-Z0-9_-]{1,128}$/.test(value),
        message: '长度为1~128个字符，可包含英文字母、数字和-_',
        trigger: 'blur',
      },
    ],
  })
  name: string;

  @Column('string', {
    name: '权限模板类型',
    required: true,
    option: {
      '1': {
        label: '引用策略库',
        disabled: false,
      },
      '2': {
        label: '自定义',
        disabled: true,
      },
    },
    meta: {
      display: {
        appearance: 'radio',
      },
    },
  })
  type: string;

  @Column('list', {
    name: '权限策略库',
    required: true,
    meta: { display: { props: { idKey: 'id', displayKey: 'name' } } },
  })
  policy_library_id: string;

  @Column('string', {
    name: '模板预览',
    meta: {
      display: {
        props: {
          type: 'textarea',
          readonly: true,
          rows: 10,
          placeholder: '权限模板内容由选择的权限策略库自动生成，不可手动编辑',
        },
      },
    },
  })
  policy_document: string;

  @Column('string', { name: '模板描述', meta: { display: { props: { type: 'textarea', rows: 3, maxlength: 100 } } } })
  memo: string;
}
