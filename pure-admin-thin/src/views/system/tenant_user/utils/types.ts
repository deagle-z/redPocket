interface FormItemProps {
  title: string;
  id: number;
  tenantId: number;
  username: string;
  passwordHash: string;
  passwordAlgo: string;
  email?: string;
  mobile?: string;
  roleCode: string;
  isOwner: boolean;
  status: number;
  require2fa: boolean;
  twofaSecret?: string;
  remark?: string;
}

interface FormProps {
  formInline: FormItemProps;
}

export type { FormItemProps, FormProps };
