import { Model, Column } from '@/decorator';

@Model()
export class FieldTcloud {
  @Column('string', { apiOnly: true })
  id: string;

  @Column('string', {
    name: '二级账号',
    required: true,
  })
  account_id: string;

  @Column('string', {
    name: '模板名称',
    required: true,
    rules: [
      {
        trigger: 'change',
        message: '模板名称只能包含字母、数字和下划线，并且仅能以字母开头，长度为6-20个字符',
        validator: (val: string) => {
          return /^[a-zA-Z][a-zA-Z0-9_]{5,19}$/.test(val);
        },
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
