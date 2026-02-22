import type { LuckyMoney, LuckyHistory } from "@/api/luckyMoney";

interface FormItemProps {
  id: number;
  senderId: number;
  senderName: string;
  amount: number;
  received: number;
  number: number;
  lucky: number;
  thunder: number;
  chatId: number;
  redList: string;
  loseRate: number;
  status: number;
}

interface FormProps {
  formInline: FormItemProps;
}

export type { FormItemProps, FormProps };
export type { LuckyMoney, LuckyHistory };
