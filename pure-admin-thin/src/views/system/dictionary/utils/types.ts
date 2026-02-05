interface FormItemProps {
  id: number;
  name: string;
  code: string;
  description: string;
  dictName:string;
  dictType:string;
  status:number;
  menuIds: number[];
}

interface ItemFormItemProps{
  id:number
  dictType:string;
  dictLabel:string;
  dictValue:string;
  code:string;
  color:string;
  status
  isDefault:string;
  sort:number;
  remark:string;
}

interface FormProps {
  formInline: FormItemProps;
}
interface ItemFormProps {
  formInline: ItemFormItemProps;
}
export type { FormItemProps, FormProps,ItemFormProps,ItemFormItemProps };
