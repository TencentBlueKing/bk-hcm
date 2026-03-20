import { Model, Column } from '@/decorator';

@Model('gpu-demand/create-form')
export class CreateForm {
  @Column('string', {
    name: '运营产品',
    index: 0,
  })
  op_product_name: string;

  @Column('string', {
    name: '运营产品关联业务',
    index: 1,
  })
  biz_names: string;
}
