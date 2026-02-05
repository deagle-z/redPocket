interface FormItemProps {
  title: string;
  id?: number;
  nickName: string;
  username: string;
  enabled: boolean;
  amount: number;
  userType: number;
  gender: number;
  mark: string;
  password: string;
  roles: Record<string, unknown>[];
  roleOptions: Role[];
}

interface Role {
  id: number;
  code: string;
  name: string;
}

interface FormProps {
  formInline: FormItemProps;
}

interface RoleFormItemProps {
  userId: number;
  username: string;
  roleOptions: any[];
  codes: Record<string, unknown>[];
}

interface RoleFormProps {
  formInline: RoleFormItemProps;
}

export type { FormItemProps, FormProps, RoleFormItemProps, RoleFormProps };
