/*
 * @Author: yujiang yujiang@gmail.com
 * @Date: 2025-10-11 13:17:11
 * @LastEditors: yujiang yujiang@gmail.com
 * @LastEditTime: 2025-10-11 13:18:30
 * @FilePath: \BaseGoUni\pure-admin-thin\src\views\system\dictionary\utils\enums.ts
 * @Description: 这是默认设置,请设置`customMade`, 打开koroFileHeader查看配置 进行设置: https://github.com/OBKoro1/koro1FileHeader/wiki/%E9%85%8D%E7%BD%AE
 */
import type { OptionsType } from "@/components/ReSegmented";


const dictTypeOptions: Array<OptionsType> = [
  {
    label: "启用",
    tip: "字典类型启用",
    value: true
  },
  {
    label: "禁用",
    tip: "字典类型禁用",
    value: false
  }
];
const dictItemOptions: Array<OptionsType> = [
  {
    label: "启用",
    tip: "字典项启用",
    value: true
  },
  {
    label: "禁用",
    tip: "字典项禁用",
    value: false
  }
];

export {
dictTypeOptions,
dictItemOptions
};
